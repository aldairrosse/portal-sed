// Package evaluation provides shared DTOs for the evaluation and 9×9 API.
package evaluation

import (
	"github.com/google/uuid"
)

// EvaluationExportRow is one row of GET /evaluations/export: the 7 export
// columns (backend-computed, active-phase scoped).
type EvaluationExportRow struct {
	EmployeeID     uuid.UUID `json:"employeeId"`
	EmployeeNumber string    `json:"employeeNumber"`
	EmployeeName   string    `json:"employeeName"`
	GoalProgress   float64   `json:"goalProgress"`
	SelfAvg        float64   `json:"selfAvg"`
	RhAvg          float64   `json:"rhAvg"`
	Rating         float64   `json:"rating"`
	Status         string    `json:"status"`
}

// EvaluationExportMeta carries the export scope.
type EvaluationExportMeta struct {
	CycleID uuid.UUID `json:"cycleId"`
	Phase   string    `json:"phase"`
	Total   int       `json:"total"`
}

// EvaluationExportResponse is the response for GET /evaluations/export.
type EvaluationExportResponse struct {
	Data []EvaluationExportRow `json:"data"`
	Meta EvaluationExportMeta  `json:"meta"`
}
