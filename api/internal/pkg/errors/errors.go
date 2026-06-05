// Package errors provides standard domain error types and HTTP status mapping
// for the SED evaluation lifecycle API.
package errors

import (
	"encoding/json"
	"fmt"
)

// DomainCode is a machine-readable error code.
type DomainCode string

const (
	CycleNotFound        DomainCode = "CYCLE_NOT_FOUND"
	InvalidTransition    DomainCode = "INVALID_TRANSITION"
	CycleAlreadyActive   DomainCode = "CYCLE_ALREADY_ACTIVE"
	PhaseNotAdvanceable  DomainCode = "PHASE_NOT_ADVANCEABLE"
	ConcurrentUpdate     DomainCode = "CONCURRENT_UPDATE"
	IdempotencyKeyConflict DomainCode = "IDEMPOTENCY_KEY_CONFLICT"
	RateLimitExceeded    DomainCode = "RATE_LIMIT_EXCEEDED"
	InvalidRequest       DomainCode = "INVALID_REQUEST"
	MissingIfMatch       DomainCode = "MISSING_IF_MATCH"
	InvalidIfMatch       DomainCode = "INVALID_IF_MATCH"
	RequestTimeout       DomainCode = "REQUEST_TIMEOUT"

	// Goal-specific domain error codes
	CategoryNotFound        DomainCode = "CATEGORY_NOT_FOUND"
	GoalNotFound            DomainCode = "GOAL_NOT_FOUND"
	KpiNotFound             DomainCode = "KPI_NOT_FOUND"
	WeightSumInvalid        DomainCode = "WEIGHT_SUM_INVALID"
	PhaseRestricted         DomainCode = "PHASE_RESTRICTED"
	DuplicateCategoryName   DomainCode = "DUPLICATE_CATEGORY_NAME"
	InvalidWeightRange      DomainCode = "INVALID_WEIGHT_RANGE"
	InvalidTargetValue      DomainCode = "INVALID_TARGET_VALUE"
	InvalidUnit             DomainCode = "INVALID_UNIT"
	GoalWeightOverflow      DomainCode = "GOAL_WEIGHT_OVERFLOW"
	GoalNotDeletableInPhase DomainCode = "GOAL_NOT_DELETABLE_IN_PHASE"
	KpiLinkedCannotDelete   DomainCode = "KPI_LINKED_CANNOT_DELETE"
	ConcurrentModification  DomainCode = "CONCURRENT_MODIFICATION"
	BatchSizeExceeded       DomainCode = "BATCH_SIZE_EXCEEDED"
	KpiLinkLimitExceeded    DomainCode = "KPI_LINK_LIMIT_EXCEEDED"

	// Org-hierarchy domain error codes
	TreeNotFound        DomainCode = "TREE_NOT_FOUND"
	NodeNotFound        DomainCode = "NODE_NOT_FOUND"
	EmployeeNotFound    DomainCode = "EMPLOYEE_NOT_FOUND"
	NodeHasChildren     DomainCode = "NODE_HAS_CHILDREN"
	InvalidParent       DomainCode = "INVALID_PARENT"
	StaleVersion        DomainCode = "STALE_VERSION"
	InvalidTreeType     DomainCode = "INVALID_TREE_TYPE"
	ScopeNotFound       DomainCode = "SCOPE_NOT_FOUND"
)

// DomainError is the standard error type for domain-level errors.
type DomainError struct {
	Code    DomainCode `json:"code"`
	Message string     `json:"message"`
	Details []string   `json:"details,omitempty"`
	Err     error      `json:"-"` // wrapped error, not serialised
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a DomainError with an optional underlying cause.
func NewDomainError(code DomainCode, message string, err error) *DomainError {
	return &DomainError{Code: code, Message: message, Err: err}
}

// WithDetails adds detail lines to a DomainError.
func (e *DomainError) WithDetails(details ...string) *DomainError {
	e.Details = append(e.Details, details...)
	return e
}

// Sentinel error values for switch/type-assertion checks.
var (
	ErrCycleNotFound        = &DomainError{Code: CycleNotFound, Message: "The requested cycle was not found."}
	ErrInvalidTransition    = &DomainError{Code: InvalidTransition, Message: "The requested phase transition is not valid from the current phase."}
	ErrCycleAlreadyActive   = &DomainError{Code: CycleAlreadyActive, Message: "A cycle already exists for this organization and year."}
	ErrPhaseNotAdvanceable  = &DomainError{Code: PhaseNotAdvanceable, Message: "The current phase cannot be advanced; conditions not met."}
	ErrConcurrentUpdate     = &DomainError{Code: ConcurrentUpdate, Message: "The resource was modified by another request; retry with the latest version."}
	ErrIdempotencyConflict  = &DomainError{Code: IdempotencyKeyConflict, Message: "Idempotency-Key already used with a different payload."}
	ErrRateLimitExceeded    = &DomainError{Code: RateLimitExceeded, Message: "Rate limit exceeded for this organization."}
	ErrInvalidRequest       = &DomainError{Code: InvalidRequest, Message: "The request contains invalid parameters."}
	ErrMissingIfMatch       = &DomainError{Code: MissingIfMatch, Message: "If-Match header is required for this operation."}
	ErrInvalidIfMatch       = &DomainError{Code: InvalidIfMatch, Message: "If-Match header is malformed; expected an integer version."}
	ErrRequestTimeout       = &DomainError{Code: RequestTimeout, Message: "The request timed out before completion."}

	// Goal-specific sentinel errors
	ErrCategoryNotFound        = &DomainError{Code: CategoryNotFound, Message: "The requested category was not found."}
	ErrGoalNotFound            = &DomainError{Code: GoalNotFound, Message: "The requested goal was not found."}
	ErrKpiNotFound             = &DomainError{Code: KpiNotFound, Message: "The requested KPI was not found."}
	ErrWeightSumInvalid        = &DomainError{Code: WeightSumInvalid, Message: "Weight sum validation failed; categories and/or goals must each sum to 100%."}
	ErrPhaseRestricted         = &DomainError{Code: PhaseRestricted, Message: "This operation is not allowed in the current cycle phase."}
	ErrDuplicateCategoryName   = &DomainError{Code: DuplicateCategoryName, Message: "A category with this name already exists for this employee."}
	ErrInvalidWeightRange      = &DomainError{Code: InvalidWeightRange, Message: "Weight must be between 0 and 100."}
	ErrInvalidTargetValue      = &DomainError{Code: InvalidTargetValue, Message: "Target value must be greater than 0."}
	ErrInvalidUnit             = &DomainError{Code: InvalidUnit, Message: "Unit must be one of: porcentaje, moneda, numero."}
	ErrGoalWeightOverflow      = &DomainError{Code: GoalWeightOverflow, Message: "Adding this goal would exceed the 100% weight limit for this category."}
	ErrGoalNotDeletableInPhase = &DomainError{Code: GoalNotDeletableInPhase, Message: "Goals cannot be deleted in the current cycle phase."}
	ErrKpiLinkedCannotDelete   = &DomainError{Code: KpiLinkedCannotDelete, Message: "Cannot delete a KPI that is linked to one or more goals."}
	ErrConcurrentModification  = &DomainError{Code: ConcurrentModification, Message: "The resource was modified by another request; retry with the latest version."}
	ErrBatchSizeExceeded       = &DomainError{Code: BatchSizeExceeded, Message: "Batch size exceeds the maximum allowed (50)."}
	ErrKpiLinkLimitExceeded    = &DomainError{Code: KpiLinkLimitExceeded, Message: "A goal cannot have more than 5 linked KPIs."}

	// Org-hierarchy sentinel errors
	ErrTreeNotFound     = &DomainError{Code: TreeNotFound, Message: "Organizational tree not found."}
	ErrNodeNotFound     = &DomainError{Code: NodeNotFound, Message: "Org node not found."}
	ErrEmployeeNotFound = &DomainError{Code: EmployeeNotFound, Message: "Employee not found."}
	ErrNodeHasChildren  = &DomainError{Code: NodeHasChildren, Message: "Cannot delete node with children."}
	ErrInvalidParent    = &DomainError{Code: InvalidParent, Message: "Invalid parent: would create a cycle."}
	ErrStaleVersion     = &DomainError{Code: StaleVersion, Message: "Optimistic lock failed; version mismatch."}
	ErrInvalidTreeType  = &DomainError{Code: InvalidTreeType, Message: "Tree type must be 'corporate' or 'retail'."}
	ErrScopeNotFound    = &DomainError{Code: ScopeNotFound, Message: "Evaluator scope not found."}
)

// HTTPStatus returns the HTTP status code for a domain error.
// If err is not a recognised DomainError, returns 500.
func HTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	var de *DomainError
	if !AsDomainError(err, &de) {
		return 500
	}
	switch de.Code {
	case CycleNotFound, CategoryNotFound, GoalNotFound, KpiNotFound, TreeNotFound, NodeNotFound, EmployeeNotFound, ScopeNotFound,
		"EVALUATION_NOT_FOUND", "MATRIX_NOT_FOUND", "ENTRY_NOT_FOUND":
		return 404
	case InvalidTransition, CycleAlreadyActive, PhaseNotAdvanceable, ConcurrentUpdate, IdempotencyKeyConflict, DuplicateCategoryName, KpiLinkedCannotDelete, ConcurrentModification, NodeHasChildren, StaleVersion,
		"EVALUATION_ALREADY_FINALIZED":
		return 409
	case PhaseRestricted, GoalNotDeletableInPhase:
		return 403
	case WeightSumInvalid, GoalWeightOverflow:
		return 422
	case RateLimitExceeded:
		return 429
	case MissingIfMatch:
		return 428
	case InvalidRequest, InvalidIfMatch, InvalidWeightRange, InvalidTargetValue, InvalidUnit, BatchSizeExceeded, KpiLinkLimitExceeded, InvalidParent, InvalidTreeType,
		"QUADRANT_OUT_OF_RANGE":
		return 400
	case RequestTimeout:
		return 408
	default:
		return 500
	}
}

// AsDomainError is a wrapper around errors.As for *DomainError.
func AsDomainError(err error, target **DomainError) bool {
	if err == nil {
		return false
	}
	de, ok := err.(*DomainError)
	if ok {
		*target = de
		return true
	}
	// try unwrapping
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
		if err == nil {
			return false
		}
		de, ok := err.(*DomainError)
		if ok {
			*target = de
			return true
		}
	}
}

// APIError is the JSON-serialisable error body returned to the client.
type APIError struct {
	Error   APIErrorBody `json:"error"`
}

// APIErrorBody holds the fields of an API error response.
type APIErrorBody struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
	TraceID string   `json:"trace_id"`
}

// NewAPIErrorResponse builds an *APIError from a DomainError and a trace ID.
func NewAPIErrorResponse(de *DomainError, traceID string) *APIError {
	return &APIError{
		Error: APIErrorBody{
			Code:    string(de.Code),
			Message: de.Message,
			Details: de.Details,
			TraceID: traceID,
		},
	}
}

// MustMarshalJSON serialises the APIError to JSON bytes. Panics on failure
// (should never happen for these simple structs).
func (ae *APIError) MustMarshalJSON() []byte {
	b, err := json.Marshal(ae)
	if err != nil {
		panic("errors: failed to marshal APIError: " + err.Error())
	}
	return b
}
