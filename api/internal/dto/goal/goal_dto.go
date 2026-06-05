// Package goal provides request and response DTOs for the goals API.
package goal

// Category DTOs

// CreateCategoryRequest is the request body for creating a goal category.
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Weight      float64 `json:"weight"`
}

// UpdateCategoryRequest is the request body for updating a goal category.
type UpdateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Weight      float64 `json:"weight"`
}

// CategoryResponse is the response body for a single category with nested goals.
type CategoryResponse struct {
	ID          string           `json:"id"`
	EmployeeID  string           `json:"employee_id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Weight      float64          `json:"weight"`
	Goals       []GoalResponse   `json:"goals,omitempty"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// CategoryListResponse is the paginated response for listing categories.
type CategoryListResponse struct {
	Items      []CategoryResponse `json:"items"`
	NextCursor *string            `json:"next_cursor,omitempty"`
}

// Goal DTOs

// CreateGoalRequest is the request body for creating a goal.
type CreateGoalRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Unit        string   `json:"unit"`
	Weight      float64  `json:"weight"`
	TargetValue float64  `json:"target_value"`
	KpiIDs      []string `json:"kpi_ids,omitempty"`
}

// UpdateGoalRequest is the request body for updating a goal.
type UpdateGoalRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Unit        string   `json:"unit"`
	Weight      float64  `json:"weight"`
	TargetValue float64  `json:"target_value"`
	Version     int      `json:"version"`
	KpiIDs      []string `json:"kpi_ids,omitempty"`
}

// UpdateProgressRequest is the request body for updating goal progress.
type UpdateProgressRequest struct {
	CurrentValue float64 `json:"current_value"`
}

// GoalResponse is the response body for a single goal.
type GoalResponse struct {
	ID           string         `json:"id"`
	CategoryID   string         `json:"category_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Unit         string         `json:"unit"`
	Weight       float64        `json:"weight"`
	TargetValue  float64        `json:"target_value"`
	CurrentValue float64        `json:"current_value"`
	State        string         `json:"state"`
	Version      int            `json:"version"`
	KPIs         []KpiResponse  `json:"kpis,omitempty"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

// KPI DTOs

// CreateKpiRequest is the request body for creating a KPI.
type CreateKpiRequest struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
}

// UpdateKpiRequest is the request body for updating a KPI.
type UpdateKpiRequest struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
}

// KpiResponse is the response body for a single KPI.
type KpiResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// KpiListResponse is the paginated response for listing KPIs.
type KpiListResponse struct {
	Items      []KpiResponse `json:"items"`
	NextCursor *string       `json:"next_cursor,omitempty"`
}

// LinkKpiRequest is the request body for linking a KPI to a goal.
type LinkKpiRequest struct {
	KpiID string `json:"kpi_id"`
}

// Batch DTOs

// BatchGoalItem represents a single operation in a batch request.
type BatchGoalItem struct {
	Operation  string             `json:"operation"` // "create" or "update"
	CategoryID string             `json:"category_id,omitempty"`
	GoalID     string             `json:"goal_id,omitempty"`
	Goal       CreateGoalRequest  `json:"goal"`
}

// BatchGoalRequest is the request body for batch operations.
type BatchGoalRequest struct {
	Items []BatchGoalItem `json:"items"`
}

// BatchGoalResponse is the response body for batch operations.
type BatchGoalResponse struct {
	Items []GoalResponse `json:"items"`
}

// Weight Validation DTOs

// WeightValidationResponse is the response body for weight validation.
type WeightValidationResponse struct {
	Valid       bool               `json:"valid"`
	CategorySum float64            `json:"category_sum"`
	ExpectedSum float64            `json:"expected_sum"`
	Deficit     float64            `json:"deficit"`
	GoalSums    []CategoryGoalSum  `json:"goal_sums,omitempty"`
}

// CategoryGoalSum provides per-category weight sum details.
type CategoryGoalSum struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Sum          float64 `json:"sum"`
	ExpectedSum  float64 `json:"expected_sum"`
	Deficit      float64 `json:"deficit"`
}

// Assignment DTOs

// CreateAssignmentRequest is the request body for creating an assignment.
type CreateAssignmentRequest struct {
	CycleID string `json:"cycle_id"`
}

// AssignmentResponse is the response body for a goal assignment.
type AssignmentResponse struct {
	ID         string             `json:"id"`
	EmployeeID string             `json:"employee_id"`
	CycleID    string             `json:"cycle_id"`
	Categories []CategoryResponse `json:"categories,omitempty"`
	CreatedAt  string             `json:"created_at"`
}
