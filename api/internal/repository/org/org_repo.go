// Package org provides the repository layer for organizational hierarchy entities.
// It uses Ent-generated queries for standard CRUD and raw SQL for ltree path
// operations, version fields, and tree traversal.
package org

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// contextKey for db role routing.
type ctxKeyDBRole struct{}

const (
	// DBRolePrimary routes queries to the primary database.
	DBRolePrimary = "primary"
	// DBRoleReplica routes queries to a read replica.
	DBRoleReplica = "replica"
)

// WithDBRole embeds a db role hint into the context.
func WithDBRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxKeyDBRole{}, role)
}

// DBRoleFromContext extracts the db role from context; returns "primary" if not set.
func DBRoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyDBRole{}).(string)
	if v == "" {
		return DBRolePrimary
	}
	return v
}

// Domain error sentinels for org hierarchy.
// These delegate to the shared errors package, using codes registered there.
var (
	ErrTreeNotFound     = errors.ErrTreeNotFound
	ErrNodeNotFound     = errors.ErrNodeNotFound
	ErrEmployeeNotFound = errors.ErrEmployeeNotFound
	ErrNodeHasChildren  = errors.ErrNodeHasChildren
	ErrInvalidParent    = errors.ErrInvalidParent
	ErrStaleVersion     = errors.ErrStaleVersion
	ErrInvalidTreeType  = errors.ErrInvalidTreeType
	ErrScopeNotFound    = errors.ErrScopeNotFound
)

// ----------------
// OrgTreeRow — read model for organizational trees
// ----------------

// OrgTreeRow represents an organization tree with node count.
type OrgTreeRow struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	NodeCount int       `json:"nodeCount"`
}

// OrgTreeRepo provides database queries for organizational trees.
type OrgTreeRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewOrgTreeRepo creates a new OrgTreeRepo.
func NewOrgTreeRepo(client *internal.Client, db *sql.DB) *OrgTreeRepo {
	return &OrgTreeRepo{client: client, db: db}
}

// readClient returns the appropriate client based on context db role hint.
func (r *OrgTreeRepo) readClient() *internal.Client {
	return r.client
}

// List returns all organizations that have at least one org_node (i.e., trees).
// If treeType is non-empty, filters by node type.
func (r *OrgTreeRepo) List(ctx context.Context, treeType string) ([]*OrgTreeRow, error) {
	query := `SELECT o.id, o.name,
	           COALESCE((SELECT on2.type FROM org_nodes on2 WHERE on2.organization_id = o.id LIMIT 1), '') as type,
	           (SELECT COUNT(1) FROM org_nodes on2 WHERE on2.organization_id = o.id) as node_count
	           FROM organizations o
	           WHERE EXISTS (SELECT 1 FROM org_nodes on2 WHERE on2.organization_id = o.id)`
	args := []interface{}{}

	if treeType != "" {
		query += ` AND EXISTS (SELECT 1 FROM org_nodes on2 WHERE on2.organization_id = o.id AND on2.type = $1)`
		args = append(args, treeType)
	}

	query += ` ORDER BY o.name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*OrgTreeRow
	for rows.Next() {
		row := &OrgTreeRow{}
		if err := rows.Scan(&row.ID, &row.Name, &row.Type, &row.NodeCount); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetByID returns a single tree row by org ID with node count.
func (r *OrgTreeRepo) GetByID(ctx context.Context, orgID uuid.UUID) (*OrgTreeRow, error) {
	row := &OrgTreeRow{}
	err := r.db.QueryRowContext(ctx,
		`SELECT o.id, o.name,
		        COALESCE((SELECT on2.type FROM org_nodes on2 WHERE on2.organization_id = o.id LIMIT 1), '') as type,
		        (SELECT COUNT(1) FROM org_nodes on2 WHERE on2.organization_id = o.id) as node_count
		 FROM organizations o WHERE o.id = $1`, orgID,
	).Scan(&row.ID, &row.Name, &row.Type, &row.NodeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTreeNotFound
		}
		return nil, err
	}
	return row, nil
}

// ----------------
// OrgNodeRow — read model for org nodes
// ----------------

// OrgNodeRow represents a single org node with path, version, and employee count.
type OrgNodeRow struct {
	ID             uuid.UUID  `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	Code           string     `json:"code"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Path           string     `json:"path"`
	Version        int        `json:"version"`
	EmployeeCount  int        `json:"employeeCount"`
}

// OrgNodeRepo provides CRUD and tree traversal for org nodes.
type OrgNodeRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewOrgNodeRepo creates a new OrgNodeRepo.
func NewOrgNodeRepo(client *internal.Client, db *sql.DB) *OrgNodeRepo {
	return &OrgNodeRepo{client: client, db: db}
}

// readClient returns the appropriate client based on context db role hint.
func (r *OrgNodeRepo) readClient() *internal.Client {
	return r.client
}

// Create inserts a new org node, computing the ltree path from its parent.
func (r *OrgNodeRepo) Create(ctx context.Context, orgID uuid.UUID, parentID *uuid.UUID, name, nodeType, code string, metadata map[string]interface{}) (*OrgNodeRow, error) {
	now := time.Now()
	id := uuid.New()

	var parentPath string
	if parentID != nil {
		err := r.db.QueryRowContext(ctx,
			`SELECT COALESCE(path::text, '') FROM org_nodes WHERE id = $1`, *parentID,
		).Scan(&parentPath)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrNodeNotFound
			}
			return nil, err
		}
	}

	path := makeLtreePath(parentPath, id)

	metaJSON := []byte("{}")
	if metadata != nil {
		metaJSON, _ = json.Marshal(metadata)
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO org_nodes (id, created_at, updated_at, name, type, code, metadata, organization_id, parent_id, path, version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10::ltree, 0)`,
		id, now, now, name, nodeType, code, string(metaJSON), orgID, parentID, path,
	)
	if err != nil {
		return nil, err
	}

	return &OrgNodeRow{
		ID: id, CreatedAt: now, UpdatedAt: now, Name: name,
		Type: nodeType, Code: code, OrganizationID: orgID,
		ParentID: parentID, Path: path, Version: 0,
	}, nil
}

// GetByID retrieves a single org node by ID.
func (r *OrgNodeRepo) GetByID(ctx context.Context, nodeID uuid.UUID) (*OrgNodeRow, error) {
	return scanNodeRow(r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0) FROM org_nodes WHERE id = $1`,
		nodeID,
	))
}

// UpdateWithVersion updates an org node with optimistic locking.
// If the version doesn't match the current row, returns ErrStaleVersion.
func (r *OrgNodeRepo) UpdateWithVersion(ctx context.Context, nodeID uuid.UUID, version int, name, code string, metadata map[string]interface{}) (*OrgNodeRow, error) {
	var sets []string
	args := []interface{}{nodeID, version}
	idx := 3

	if name != "" {
		sets = append(sets, "name = $"+itoa(idx))
		args = append(args, name)
		idx++
	}
	if code != "" {
		sets = append(sets, "code = $"+itoa(idx))
		args = append(args, code)
		idx++
	}
	if metadata != nil {
		metaJSON, _ := json.Marshal(metadata)
		sets = append(sets, "metadata = $"+itoa(idx)+"::jsonb")
		args = append(args, string(metaJSON))
		idx++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, nodeID)
	}

	allSets := "version = version + 1, updated_at = NOW(), " + strings.Join(sets, ", ")
	query := `UPDATE org_nodes SET ` + allSets +
		` WHERE id = $1 AND version = $2
		  RETURNING id, created_at, updated_at, name, type, code, organization_id, parent_id,
		            COALESCE(path::text, '') as path, COALESCE(version, 0)`

	return scanNodeRow(r.db.QueryRowContext(ctx, query, args...))
}

// Delete performs a hard delete of an org node by ID.
func (r *OrgNodeRepo) Delete(ctx context.Context, nodeID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM org_nodes WHERE id = $1`, nodeID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNodeNotFound
	}
	return nil
}

// GetDescendants returns all nodes whose path is a descendant of the given path (inclusive).
func (r *OrgNodeRepo) GetDescendants(ctx context.Context, path string) ([]*OrgNodeRow, error) {
	return queryNodeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0)
		 FROM org_nodes WHERE path::text LIKE $1 || '.%' OR path::text = $1
		 ORDER BY path::text`, path)
}

// GetAncestors returns all nodes whose path is an ancestor of the given path (inclusive).
func (r *OrgNodeRepo) GetAncestors(ctx context.Context, path string) ([]*OrgNodeRow, error) {
	return queryNodeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0)
		 FROM org_nodes WHERE $1 LIKE path::text || '.%' OR path::text = $1
		 ORDER BY path::text`, path)
}

// GetPathToRoot returns ancestor nodes from the given node up to the root using recursive CTE.
// Ordered from self (depth=1) to root (max depth).
func (r *OrgNodeRepo) GetPathToRoot(ctx context.Context, nodeID uuid.UUID) ([]*OrgNodeRow, error) {
	return queryNodeRows(r.db, ctx,
		`WITH RECURSIVE ancestors AS (
		    SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		           COALESCE(path::text, '') as path, COALESCE(version, 0), 1 as depth
		    FROM org_nodes WHERE id = $1
		    UNION ALL
		    SELECT n.id, n.created_at, n.updated_at, n.name, n.type, n.code, n.organization_id,
		           n.parent_id, COALESCE(n.path::text, '') as path, COALESCE(n.version, 0), a.depth + 1
		    FROM org_nodes n
		    JOIN ancestors a ON n.id = a.parent_id
		)
		SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		       path, version
		FROM ancestors ORDER BY depth`, nodeID)
}

// GetSubtree returns all descendants of a node with optional depth limit.
// maxDepth<0 returns all. maxDepth=0 returns only the node itself.
func (r *OrgNodeRepo) GetSubtree(ctx context.Context, nodeID uuid.UUID, maxDepth int) ([]*OrgNodeRow, error) {
	node, err := r.GetByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	if maxDepth < 0 {
		return queryNodeRows(r.db, ctx,
			`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
			        COALESCE(path::text, '') as path, COALESCE(version, 0)
			 FROM org_nodes WHERE (path::text = $1 OR path::text LIKE $1 || '.%')
			 ORDER BY path::text`, node.Path)
	}

	return queryNodeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0)
		 FROM org_nodes WHERE (path::text = $1 OR path::text LIKE $1 || '.%')
		   AND nlevel(path::text::ltree) - nlevel($1::ltree) <= $2
		 ORDER BY path::text`, node.Path, maxDepth)
}

// GetRootNode returns the node with parent_id IS NULL for the given tree.
func (r *OrgNodeRepo) GetRootNode(ctx context.Context, treeID uuid.UUID) (*OrgNodeRow, error) {
	return scanNodeRow(r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0)
		 FROM org_nodes WHERE organization_id = $1 AND parent_id IS NULL LIMIT 1`, treeID))
}

// CountChildren returns the number of direct children for a node.
func (r *OrgNodeRepo) CountChildren(ctx context.Context, nodeID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM org_nodes WHERE parent_id = $1`, nodeID,
	).Scan(&count)
	return count, err
}

// ListByOrg returns all nodes for an organization, ordered by path.
func (r *OrgNodeRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrgNodeRow, error) {
	rows, err := queryNodeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, name, type, code, organization_id, parent_id,
		        COALESCE(path::text, '') as path, COALESCE(version, 0)
		 FROM org_nodes WHERE organization_id = $1
		 ORDER BY path::text`, orgID)
	if err != nil {
		return nil, err
	}
	// Load employee counts in batch
	loadEmployeeCounts(r.db, ctx, rows)
	return rows, nil
}

// UpdatePathAndDescendants updates the path for a node and all its descendants
// within a transaction. Used after a MoveNode operation.
func (r *OrgNodeRepo) UpdatePathAndDescendants(ctx context.Context, tx *sql.Tx, oldPath, newPath string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE org_nodes
		 SET path = ($1::ltree || subpath(path::ltree, nlevel($2::text)))
		 WHERE path::text LIKE $2 || '.%' OR path::text = $2`,
		newPath, oldPath,
	)
	return err
}

// AcquireTreeLock acquires a PostgreSQL advisory lock for a tree.
// Returns a release function that must be called (deferred) by the caller.
func (r *OrgNodeRepo) AcquireTreeLock(ctx context.Context, treeID uuid.UUID) (func() error, error) {
	lockKey := "org:tree:" + treeID.String()
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext($1))`, lockKey)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return func() error {
		defer conn.Close()
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, lockKey)
		return err
	}, nil
}

// BeginTx starts a *sql.Tx for transactional operations.
func (r *OrgNodeRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// ---------- package-level helpers ----------

// scanNodeRow scans a single *sql.Row into an OrgNodeRow.
func scanNodeRow(row *sql.Row) (*OrgNodeRow, error) {
	n := &OrgNodeRow{}
	var parentID sql.NullString
	var path sql.NullString
	var version sql.NullInt64

	err := row.Scan(
		&n.ID, &n.CreatedAt, &n.UpdatedAt, &n.Name, &n.Type, &n.Code,
		&n.OrganizationID, &parentID, &path, &version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}
	if parentID.Valid {
		pid, _ := uuid.Parse(parentID.String)
		n.ParentID = &pid
	}
	if path.Valid {
		n.Path = path.String
	}
	if version.Valid {
		n.Version = int(version.Int64)
	}
	return n, nil
}

// queryNodeRows scans multiple *sql.Rows into an OrgNodeRow slice.
func queryNodeRows(db *sql.DB, ctx context.Context, query string, args ...interface{}) ([]*OrgNodeRow, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*OrgNodeRow
	for rows.Next() {
		n := &OrgNodeRow{}
		var parentID sql.NullString
		var path sql.NullString
		var version sql.NullInt64

		err := rows.Scan(
			&n.ID, &n.CreatedAt, &n.UpdatedAt, &n.Name, &n.Type, &n.Code,
			&n.OrganizationID, &parentID, &path, &version,
		)
		if err != nil {
			return nil, err
		}
		if parentID.Valid {
			pid, _ := uuid.Parse(parentID.String)
			n.ParentID = &pid
		}
		if path.Valid {
			n.Path = path.String
		}
		if version.Valid {
			n.Version = int(version.Int64)
		}
		results = append(results, n)
	}
	return results, rows.Err()
}

// loadEmployeeCounts batch-loads the employee count for each node.
func loadEmployeeCounts(db *sql.DB, ctx context.Context, nodes []*OrgNodeRow) {
	if len(nodes) == 0 {
		return
	}

	byID := make(map[string]*OrgNodeRow, len(nodes))
	for _, n := range nodes {
		byID[n.ID.String()] = n
	}

	// Build IN clause
	placeholders := make([]string, len(nodes))
	args := make([]interface{}, len(nodes))
	for i, n := range nodes {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = n.ID.String()
	}

	query := `SELECT org_node_id, COUNT(1) FROM employees WHERE org_node_id IN (` +
		strings.Join(placeholders, ",") + `) AND is_active = true GROUP BY org_node_id`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID string
		var count int
		if err := rows.Scan(&nodeID, &count); err == nil {
			if n, ok := byID[nodeID]; ok {
				n.EmployeeCount = count
			}
		}
	}
}

// makeLtreePath generates a sanitized ltree path from a parent path and a child UUID.
func makeLtreePath(parentPath string, childID uuid.UUID) string {
	cleanID := strings.ReplaceAll(childID.String(), "-", "_")
	if parentPath == "" {
		return cleanID
	}
	return parentPath + "." + cleanID
}

// itoa converts int to string without strconv import per-call.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
