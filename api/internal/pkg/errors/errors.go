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
	OTPRequired          DomainCode = "OTP_REQUIRED"
	NotAuthenticated     DomainCode = "NOT_AUTHENTICATED"

	// Códigos de error de dominio específicos de objetivos
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
	InvalidBaselineValue    DomainCode = "INVALID_BASELINE_VALUE"
	InvalidDirection        DomainCode = "INVALID_DIRECTION"
	InvalidQuadrant         DomainCode = "INVALID_QUADRANT"

	// Códigos de error de dominio de jerarquía organizacional
	TreeNotFound         DomainCode = "TREE_NOT_FOUND"
	NodeNotFound         DomainCode = "NODE_NOT_FOUND"
	EmployeeNotFound     DomainCode = "EMPLOYEE_NOT_FOUND"
	OrganizationNotFound DomainCode = "ORGANIZATION_NOT_FOUND"
	NodeHasChildren      DomainCode = "NODE_HAS_CHILDREN"
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
	ErrCycleNotFound        = &DomainError{Code: CycleNotFound, Message: "El ciclo solicitado no fue encontrado."}
	ErrInvalidTransition    = &DomainError{Code: InvalidTransition, Message: "La transición de fase solicitada no es válida desde la fase actual."}
	ErrCycleAlreadyActive   = &DomainError{Code: CycleAlreadyActive, Message: "Ya existe un ciclo para esta organización y año."}
	ErrPhaseNotAdvanceable  = &DomainError{Code: PhaseNotAdvanceable, Message: "La fase actual no se puede avanzar; condiciones no cumplidas."}
	ErrConcurrentUpdate     = &DomainError{Code: ConcurrentUpdate, Message: "El recurso fue modificado por otra solicitud; reintente con la versión más reciente."}
	ErrIdempotencyConflict  = &DomainError{Code: IdempotencyKeyConflict, Message: "La clave de idempotencia ya fue usada con un payload diferente."}
	ErrRateLimitExceeded    = &DomainError{Code: RateLimitExceeded, Message: "Límite de velocidad excedido para esta organización."}
	ErrInvalidRequest       = &DomainError{Code: InvalidRequest, Message: "La solicitud contiene parámetros inválidos."}
	ErrMissingIfMatch       = &DomainError{Code: MissingIfMatch, Message: "El encabezado If-Match es requerido para esta operación."}
	ErrInvalidIfMatch       = &DomainError{Code: InvalidIfMatch, Message: "El encabezado If-Match está malformado; se esperaba una versión entera."}
	ErrRequestTimeout       = &DomainError{Code: RequestTimeout, Message: "La solicitud expiró antes de completarse."}

	// Goal-specific sentinel errors
	ErrCategoryNotFound        = &DomainError{Code: CategoryNotFound, Message: "La categoría solicitada no fue encontrada."}
	ErrGoalNotFound            = &DomainError{Code: GoalNotFound, Message: "El objetivo solicitado no fue encontrado."}
	ErrKpiNotFound             = &DomainError{Code: KpiNotFound, Message: "El KPI solicitado no fue encontrado."}
	ErrWeightSumInvalid        = &DomainError{Code: WeightSumInvalid, Message: "La validación de la suma de pesos falló; las categorías y/o objetivos deben sumar cada uno 100%."}
	ErrPhaseRestricted         = &DomainError{Code: PhaseRestricted, Message: "Esta operación no está permitida en la fase actual del ciclo."}
	ErrDuplicateCategoryName   = &DomainError{Code: DuplicateCategoryName, Message: "Ya existe una categoría con este nombre para este empleado."}
	ErrInvalidWeightRange      = &DomainError{Code: InvalidWeightRange, Message: "El peso debe estar entre 0 y 100."}
	ErrInvalidTargetValue      = &DomainError{Code: InvalidTargetValue, Message: "El valor objetivo debe ser mayor que 0."}
	ErrInvalidUnit             = &DomainError{Code: InvalidUnit, Message: "La unidad debe ser una de: porcentaje, moneda, número, binario."}
	ErrGoalWeightOverflow      = &DomainError{Code: GoalWeightOverflow, Message: "Agregar este objetivo excedería el límite de peso del 100% para esta categoría."}
	ErrGoalNotDeletableInPhase = &DomainError{Code: GoalNotDeletableInPhase, Message: "Los objetivos no pueden ser eliminados en la fase actual del ciclo."}
	ErrKpiLinkedCannotDelete   = &DomainError{Code: KpiLinkedCannotDelete, Message: "No se puede eliminar un KPI que está vinculado a uno o más objetivos."}
	ErrConcurrentModification  = &DomainError{Code: ConcurrentModification, Message: "El recurso fue modificado por otra solicitud; reintente con la versión más reciente."}
	ErrBatchSizeExceeded       = &DomainError{Code: BatchSizeExceeded, Message: "El tamaño del lote excede el máximo permitido (50)."}
	ErrKpiLinkLimitExceeded    = &DomainError{Code: KpiLinkLimitExceeded, Message: "Un objetivo no puede tener más de 5 KPIs vinculados."}
	ErrInvalidBaselineValue = &DomainError{Code: InvalidBaselineValue, Message: "El valor base debe ser mayor que el valor objetivo para objetivos descendentes."}
	ErrInvalidDirection     = &DomainError{Code: InvalidDirection, Message: "La dirección debe ser 'ascendente' o 'descendente'."}
	ErrInvalidQuadrant      = &DomainError{Code: InvalidQuadrant, Message: "El cuadrante debe ser un entero entre 1 y 9."}

	// Forbidden
	ErrForbidden = &DomainError{Code: "FORBIDDEN", Message: "No tiene permiso para acceder a este recurso."}

	// Org-hierarchy sentinel errors
	ErrTreeNotFound         = &DomainError{Code: TreeNotFound, Message: "El árbol organizacional no fue encontrado."}
	ErrNodeNotFound         = &DomainError{Code: NodeNotFound, Message: "El nodo organizacional no fue encontrado."}
	ErrEmployeeNotFound     = &DomainError{Code: EmployeeNotFound, Message: "El empleado no fue encontrado."}
	ErrOrganizationNotFound = &DomainError{Code: OrganizationNotFound, Message: "La organización no fue encontrada."}
	ErrNodeHasChildren      = &DomainError{Code: NodeHasChildren, Message: "No se puede eliminar un nodo con hijos."}
	ErrInvalidParent    = &DomainError{Code: InvalidParent, Message: "Padre inválido: crearía un ciclo."}
	ErrStaleVersion     = &DomainError{Code: StaleVersion, Message: "El bloqueo optimista falló; discordancia de versión."}
	ErrInvalidTreeType  = &DomainError{Code: InvalidTreeType, Message: "El tipo de árbol debe ser 'corporate' o 'retail'."}
	ErrScopeNotFound    = &DomainError{Code: ScopeNotFound, Message: "El ámbito del evaluador no fue encontrado."}
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
	case CycleNotFound, CategoryNotFound, GoalNotFound, KpiNotFound, TreeNotFound, NodeNotFound, EmployeeNotFound, OrganizationNotFound, ScopeNotFound,
		"EVALUATION_NOT_FOUND", "MATRIX_NOT_FOUND", "ENTRY_NOT_FOUND":
		return 404
	case InvalidTransition, CycleAlreadyActive, PhaseNotAdvanceable, ConcurrentUpdate, IdempotencyKeyConflict, DuplicateCategoryName, KpiLinkedCannotDelete, ConcurrentModification, NodeHasChildren, StaleVersion,
		"EVALUATION_ALREADY_FINALIZED":
		return 409
	case PhaseRestricted, GoalNotDeletableInPhase, "FORBIDDEN", OTPRequired:
		return 403
	case WeightSumInvalid, GoalWeightOverflow:
		return 422
	case RateLimitExceeded:
		return 429
	case MissingIfMatch:
		return 428
	case InvalidRequest, InvalidIfMatch, InvalidWeightRange, InvalidTargetValue, InvalidUnit, BatchSizeExceeded, KpiLinkLimitExceeded, InvalidBaselineValue, InvalidDirection, InvalidParent, InvalidTreeType, InvalidQuadrant,
		"QUADRANT_OUT_OF_RANGE":
		return 400
	case RequestTimeout:
		return 408
	case NotAuthenticated:
		return 401
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
