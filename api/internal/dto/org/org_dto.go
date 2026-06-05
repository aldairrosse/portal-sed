// Package org provides request and response DTOs for the org hierarchy API.
package org

// ----------------
// Organization Trees
// ----------------

// OrgTreeResponse is the response for a single org tree.
type OrgTreeResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // corporate | retail
	NodeCount int    `json:"nodeCount"`
}

// OrgTreeListResponse is the response for listing org trees.
type OrgTreeListResponse struct {
	Data []OrgTreeResponse `json:"data"`
}

// OrgTreeDetailResponse is the response for a single org tree detail.
type OrgTreeDetailResponse struct {
	Data OrgTreeResponse `json:"data"`
}

// ----------------
// Org Nodes
// ----------------

// OrgNodeResponse is a single org node in a list.
type OrgNodeResponse struct {
	ID            string `json:"id"`
	ParentID      string `json:"parentId,omitempty"` // nullable, omit if empty
	Name          string `json:"name"`
	Type          string `json:"type"`
	Code          string `json:"code"`
	Depth         int    `json:"depth"`
	Path          string `json:"path"`
	EmployeeCount int    `json:"employeeCount"`
}

// OrgNodeNestedResponse is a node with children nested.
type OrgNodeNestedResponse struct {
	ID       string                   `json:"id"`
	ParentID string                   `json:"parentId,omitempty"`
	Name     string                   `json:"name"`
	Type     string                   `json:"type"`
	Code     string                   `json:"code"`
	Depth    int                      `json:"depth"`
	Path     string                   `json:"path"`
	Children []*OrgNodeNestedResponse `json:"children,omitempty"`
}

// OrgNodeFlatList is the flat format response.
type OrgNodeFlatList struct {
	Data []OrgNodeResponse `json:"data"`
	Meta struct {
		Format string `json:"format"`
		Total  int    `json:"total"`
	} `json:"meta"`
}

// OrgNodeNested is the nested format response.
type OrgNodeNested struct {
	Data *OrgNodeNestedResponse `json:"data"`
	Meta struct {
		Format string `json:"format"`
	} `json:"meta"`
}

// OrgNodeDetailResponse is the response for a single node.
type OrgNodeDetailResponse struct {
	Data OrgNodeResponse `json:"data"`
}

// CreateOrgNodeRequest is the request body for creating an org node.
type CreateOrgNodeRequest struct {
	OrganizationID string                 `json:"organizationId"`
	ParentID       string                 `json:"parentId,omitempty"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // corporate | retail
	Code           string                 `json:"code"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateOrgNodeRequest is the request body for updating an org node.
type UpdateOrgNodeRequest struct {
	Name     string                 `json:"name,omitempty"`
	Code     string                 `json:"code,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Version  int                    `json:"version"`
}

// MoveOrgNodeRequest is the request body for moving an org node.
type MoveOrgNodeRequest struct {
	NewParentID string `json:"newParentId"`
}

// ----------------
// Employees
// ----------------

// EmployeeListItem is a light employee projection for lists.
type EmployeeListItem struct {
	ID             string `json:"id"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	EmployeeNumber string `json:"employeeNumber"`
	OrgNodeID      string `json:"orgNodeId"`
	ManagerID      string `json:"managerId,omitempty"`
	ProfileID      string `json:"profileId"`
	IsActive       bool   `json:"isActive"`
}

// EmployeeDetail is the detailed employee response with nested orgNode and manager.
type EmployeeDetail struct {
	ID             string `json:"id"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	EmployeeNumber string `json:"employeeNumber"`
	OrgNodeID      string `json:"orgNodeId"`
	ManagerID      string `json:"managerId,omitempty"`
	ProfileID      string `json:"profileId"`
	IsActive       bool   `json:"isActive"`
	OrgNode        *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"orgNode"`
	Manager *struct {
		ID        string `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"manager,omitempty"`
}

// EmployeeListResponse is the cursor-paginated employee list response.
type EmployeeListResponse struct {
	Data []EmployeeListItem `json:"data"`
	Meta struct {
		NextCursor string `json:"nextCursor,omitempty"`
		HasMore    bool   `json:"hasMore"`
		Limit      int    `json:"limit"`
	} `json:"meta"`
}

// EmployeeDetailResponse wraps a single employee detail.
type EmployeeDetailResponse struct {
	Data EmployeeDetail `json:"data"`
}

// BatchEmployeeRequest is the request body for batch lookup.
type BatchEmployeeRequest struct {
	IDs []string `json:"ids"` // max 100 UUIDs
}

// SearchEmployeesRequest carries search query parameters.
type SearchEmployeesRequest struct {
	Q     string `json:"q"`
	Limit int    `json:"limit"` // default 20, max 50
}

// ----------------
// Ancestors
// ----------------

// AncestorChainResponse is the response for chain of command.
type AncestorChainResponse struct {
	Data []AncestorItem `json:"data"`
}

// AncestorItem is a single entry in the chain of command.
type AncestorItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Depth    int    `json:"depth"`
	Relation string `json:"relation"` // self | direct_manager | director | vp | ceo
}

// ----------------
// Evaluator Scopes
// ----------------

// EvaluatorScopeResponse is the response for evaluator scope.
type EvaluatorScopeResponse struct {
	EvaluatorID    string `json:"evaluatorId"`
	CycleID        string `json:"cycleId,omitempty"`
	ScopeType      string `json:"scopeType"` // department | team | individual
	ScopeData      struct {
		OrgNodeIDs  []string `json:"orgNodeIds,omitempty"`
		EmployeeIDs []string `json:"employeeIds,omitempty"`
	} `json:"scopeData"`
	EvaluateeCount int `json:"evaluateeCount"`
}
