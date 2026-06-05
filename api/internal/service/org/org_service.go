// Package org provides business logic for organizational hierarchy operations.
package org

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/tree"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// ----------------
// OrgTreeService
// ----------------

// OrgTreeService defines the interface for org tree operations.
type OrgTreeService interface {
	GetTrees(ctx context.Context, treeType string) (*org.OrgTreeListResponse, error)
	GetTree(ctx context.Context, treeID string) (*org.OrgTreeDetailResponse, error)
	GetTreeNodes(ctx context.Context, treeID, format string, depth int) (interface{}, error)
	ExportTree(ctx context.Context, treeID string, w io.Writer) error
}

type orgTreeService struct {
	orgTreeRepo *repo.OrgTreeRepo
	orgNodeRepo *repo.OrgNodeRepo
	client      *internal.Client
}

// NewOrgTreeService creates a new OrgTreeService.
func NewOrgTreeService(orgTreeRepo *repo.OrgTreeRepo, orgNodeRepo *repo.OrgNodeRepo, client *internal.Client) OrgTreeService {
	return &orgTreeService{
		orgTreeRepo: orgTreeRepo,
		orgNodeRepo: orgNodeRepo,
		client:      client,
	}
}

func (s *orgTreeService) GetTrees(ctx context.Context, treeType string) (*org.OrgTreeListResponse, error) {
	rows, err := s.orgTreeRepo.List(ctx, treeType)
	if err != nil {
		return nil, err
	}

	resp := &org.OrgTreeListResponse{
		Data: make([]org.OrgTreeResponse, len(rows)),
	}
	for i, r := range rows {
		resp.Data[i] = org.OrgTreeResponse{
			ID:   r.ID.String(),
			Name: r.Name,
			Type: r.Type,
			NodeCount: r.NodeCount,
		}
	}
	return resp, nil
}

func (s *orgTreeService) GetTree(ctx context.Context, treeID string) (*org.OrgTreeDetailResponse, error) {
	id, err := uuid.Parse(treeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid tree ID: must be a valid UUID", err)
	}

	row, err := s.orgTreeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &org.OrgTreeDetailResponse{
		Data: org.OrgTreeResponse{
			ID:   row.ID.String(),
			Name: row.Name,
			Type: row.Type,
			NodeCount: row.NodeCount,
		},
	}, nil
}

func (s *orgTreeService) GetTreeNodes(ctx context.Context, treeID, format string, depth int) (interface{}, error) {
	id, err := uuid.Parse(treeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid tree ID: must be a valid UUID", err)
	}

	// Verify tree exists
	_, err = s.orgTreeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get all nodes for this tree
	nodes, err := s.orgNodeRepo.ListByOrg(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert to FlatNodes
	flatNodes := make([]tree.FlatNode, len(nodes))
	for i, n := range nodes {
		parentID := ""
		if n.ParentID != nil {
			parentID = n.ParentID.String()
		}
		flatNodes[i] = tree.FlatNode{
			ID:       n.ID.String(),
			ParentID: parentID,
			Depth:    tree.Nlevel(n.Path) - 1,
			Path:     n.Path,
		}
	}

	// Apply depth filter
	if depth > 0 {
		flatNodes = tree.FilterDepth(flatNodes, depth)
	}

	if format == "nested" {
		return s.buildNestedResponse(nodes, flatNodes)
	}

	return s.buildFlatResponse(nodes, flatNodes)
}

func (s *orgTreeService) buildFlatResponse(dbNodes []*repo.OrgNodeRow, flatNodes []tree.FlatNode) (*org.OrgNodeFlatList, error) {
	nodeMap := make(map[string]*repo.OrgNodeRow, len(dbNodes))
	for _, n := range dbNodes {
		nodeMap[n.ID.String()] = n
	}

	data := make([]org.OrgNodeResponse, 0, len(flatNodes))
	for _, fn := range flatNodes {
		if dbNode, ok := nodeMap[fn.ID]; ok {
			parentID := ""
			if dbNode.ParentID != nil {
				parentID = dbNode.ParentID.String()
			}
			data = append(data, org.OrgNodeResponse{
				ID:            dbNode.ID.String(),
				ParentID:      parentID,
				Name:          dbNode.Name,
				Type:          string(dbNode.Type),
				Code:          dbNode.Code,
				Depth:         fn.Depth,
				Path:          dbNode.Path,
				EmployeeCount: dbNode.EmployeeCount,
			})
		}
	}

	resp := &org.OrgNodeFlatList{Data: data}
	resp.Meta.Format = "flat"
	resp.Meta.Total = len(data)
	return resp, nil
}

func (s *orgTreeService) buildNestedResponse(dbNodes []*repo.OrgNodeRow, flatNodes []tree.FlatNode) (*org.OrgNodeNested, error) {
	root := tree.ToNested(flatNodes)

	nodeMap := make(map[string]*repo.OrgNodeRow, len(dbNodes))
	for _, n := range dbNodes {
		nodeMap[n.ID.String()] = n
	}

	nested := s.mapNestedNode(root, nodeMap)
	return &org.OrgNodeNested{
		Data: nested,
		Meta: struct {
			Format string `json:"format"`
		}{Format: "nested"},
	}, nil
}

func (s *orgTreeService) mapNestedNode(n *tree.NestedNode, nodeMap map[string]*repo.OrgNodeRow) *org.OrgNodeNestedResponse {
	if n == nil {
		return nil
	}

	resp := &org.OrgNodeNestedResponse{
		ID:       n.ID,
		ParentID: n.ParentID,
		Children: make([]*org.OrgNodeNestedResponse, 0),
	}

	if dbNode, ok := nodeMap[n.ID]; ok {
		resp.Name = dbNode.Name
		resp.Type = string(dbNode.Type)
		resp.Code = dbNode.Code
		resp.Path = dbNode.Path
	}

	for _, child := range n.Children {
		mapped := s.mapNestedNode(child, nodeMap)
		if mapped != nil {
			resp.Children = append(resp.Children, mapped)
		}
	}

	if len(resp.Children) == 0 {
		resp.Children = nil
	}

	return resp
}

func (s *orgTreeService) ExportTree(ctx context.Context, treeID string, w io.Writer) error {
	id, err := uuid.Parse(treeID)
	if err != nil {
		return errors.NewDomainError(errors.InvalidRequest, "Invalid tree ID: must be a valid UUID", err)
	}

	_, err = s.orgTreeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	nodes, err := s.orgNodeRepo.ListByOrg(ctx, id)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	if _, err := w.Write([]byte("{\"nodes\":[")); err != nil {
		return err
	}

	for i, n := range nodes {
		if i > 0 {
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}
		parentID := ""
		if n.ParentID != nil {
			parentID = n.ParentID.String()
		}
		nodeResp := org.OrgNodeResponse{
			ID:            n.ID.String(),
			ParentID:      parentID,
			Name:          n.Name,
			Type:          string(n.Type),
			Code:          n.Code,
			Depth:         tree.Nlevel(n.Path) - 1,
			Path:          n.Path,
			EmployeeCount: n.EmployeeCount,
		}
		if err := enc.Encode(nodeResp); err != nil {
			return fmt.Errorf("export: encode error at node %d: %w", i, err)
		}
	}

	if _, err := w.Write([]byte("]}")); err != nil {
		return err
	}

	return nil
}

// ----------------
// OrgNodeService
// ----------------

// OrgNodeService defines the interface for org node operations.
type OrgNodeService interface {
	GetNode(ctx context.Context, nodeID string) (*org.OrgNodeDetailResponse, error)
	CreateNode(ctx context.Context, req org.CreateOrgNodeRequest) (*org.OrgNodeDetailResponse, error)
	UpdateNode(ctx context.Context, nodeID string, req org.UpdateOrgNodeRequest) (*org.OrgNodeDetailResponse, error)
	DeleteNode(ctx context.Context, nodeID string) error
	MoveNode(ctx context.Context, nodeID string, newParentID string) (*org.OrgNodeDetailResponse, error)
}

type orgNodeService struct {
	orgNodeRepo *repo.OrgNodeRepo
	client      *internal.Client
}

// NewOrgNodeService creates a new OrgNodeService.
func NewOrgNodeService(orgNodeRepo *repo.OrgNodeRepo, client *internal.Client) OrgNodeService {
	return &orgNodeService{
		orgNodeRepo: orgNodeRepo,
		client:      client,
	}
}

func (s *orgNodeService) GetNode(ctx context.Context, nodeID string) (*org.OrgNodeDetailResponse, error) {
	id, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid node ID: must be a valid UUID", err)
	}

	row, err := s.orgNodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return nodeRowToDetail(row), nil
}

func (s *orgNodeService) CreateNode(ctx context.Context, req org.CreateOrgNodeRequest) (*org.OrgNodeDetailResponse, error) {
	if req.Name == "" {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Node name is required", nil)
	}
	if req.Type != "corporate" && req.Type != "retail" {
		return nil, repo.ErrInvalidTreeType
	}
	if req.Code == "" {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Node code is required", nil)
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid organization ID", err)
	}

	var parentID *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid parent ID", err)
		}
		parentID = &pid
	}

	row, err := s.orgNodeRepo.Create(ctx, orgID, parentID, req.Name, req.Type, req.Code, req.Metadata)
	if err != nil {
		return nil, err
	}

	return nodeRowToDetail(row), nil
}

func (s *orgNodeService) UpdateNode(ctx context.Context, nodeID string, req org.UpdateOrgNodeRequest) (*org.OrgNodeDetailResponse, error) {
	id, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid node ID: must be a valid UUID", err)
	}

	row, err := s.orgNodeRepo.UpdateWithVersion(ctx, id, req.Version, req.Name, req.Code, req.Metadata)
	if err != nil {
		return nil, err
	}

	return nodeRowToDetail(row), nil
}

func (s *orgNodeService) DeleteNode(ctx context.Context, nodeID string) error {
	id, err := uuid.Parse(nodeID)
	if err != nil {
		return errors.NewDomainError(errors.InvalidRequest, "Invalid node ID: must be a valid UUID", err)
	}

	count, err := s.orgNodeRepo.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return repo.ErrNodeHasChildren
	}

	return s.orgNodeRepo.Delete(ctx, id)
}

func (s *orgNodeService) MoveNode(ctx context.Context, nodeID, newParentID string) (*org.OrgNodeDetailResponse, error) {
	nid, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid node ID: must be a valid UUID", err)
	}
	npID, err := uuid.Parse(newParentID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "Invalid new parent ID: must be a valid UUID", err)
	}

	if nid == npID {
		return nil, repo.ErrInvalidParent.WithDetails("Cannot move a node to itself")
	}

	node, err := s.orgNodeRepo.GetByID(ctx, nid)
	if err != nil {
		return nil, err
	}

	_, err = s.orgNodeRepo.GetByID(ctx, npID)
	if err != nil {
		return nil, err
	}

	// Cycle detection: check if new parent is a descendant of current node
	descendants, err := s.orgNodeRepo.GetDescendants(ctx, node.Path)
	if err != nil {
		return nil, err
	}
	for _, d := range descendants {
		if d.ID == npID {
			return nil, repo.ErrInvalidParent.WithDetails("Cannot move a node to one of its descendants")
		}
	}

	// Acquire advisory lock
	unlock, err := s.orgNodeRepo.AcquireTreeLock(ctx, node.OrganizationID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = unlock()
	}()

	// Load new parent to get its path
	npRow, err := s.orgNodeRepo.GetByID(ctx, npID)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	tx, err := s.orgNodeRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Update parent_id
	_, err = tx.ExecContext(ctx,
		`UPDATE org_nodes SET parent_id = $1, updated_at = NOW() WHERE id = $2`,
		npID, nid,
	)
	if err != nil {
		return nil, err
	}

	// Update path for subtree
	oldPath := node.Path
	newPath := makeMovePath(npRow.Path, node.ID)
	if err := s.orgNodeRepo.UpdatePathAndDescendants(ctx, tx, oldPath, newPath); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	updated, err := s.orgNodeRepo.GetByID(ctx, nid)
	if err != nil {
		return nil, err
	}

	return nodeRowToDetail(updated), nil
}

// ---------- helpers ----------

func nodeRowToDetail(row *repo.OrgNodeRow) *org.OrgNodeDetailResponse {
	parentID := ""
	if row.ParentID != nil {
		parentID = row.ParentID.String()
	}
	return &org.OrgNodeDetailResponse{
		Data: org.OrgNodeResponse{
			ID:            row.ID.String(),
			ParentID:      parentID,
			Name:          row.Name,
			Type:          string(row.Type),
			Code:          row.Code,
			Depth:         tree.Nlevel(row.Path) - 1,
			Path:          row.Path,
			EmployeeCount: row.EmployeeCount,
		},
	}
}

func makeMovePath(parentPath string, nodeID uuid.UUID) string {
	cleanID := strings.ReplaceAll(nodeID.String(), "-", "_")
	if parentPath == "" {
		return cleanID
	}
	return parentPath + "." + cleanID
}
