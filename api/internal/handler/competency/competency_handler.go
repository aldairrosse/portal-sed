// Package competency provides HTTP handlers for the competency framework API.
//
// All 16 endpoints under /api/v1/ are implemented here. Handlers are thin:
// they parse input, delegate to the service layer, and write JSON responses.
// Business logic lives in the service package.
//
// Auth markers: TODO(auth:C7) — replace with real RBAC middleware when C7 is implemented.
package competency

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/competency"
)

// contextKey for handler-specific values.
type contextKey struct{ name string }

var idempCtxKey = &contextKey{"idempotency-key"}

// Handler holds all handler dependencies for the competency API.
type Handler struct {
	pillarSvc      svc.PillarService
	competencySvc  svc.CompetencyService
	scaleSvc       svc.ScaleService
	catalogSvc     svc.CatalogService
	acceptanceSvc  svc.AcceptanceService
}

// NewHandler creates a new Handler.
func NewHandler(
	pillarSvc svc.PillarService,
	competencySvc svc.CompetencyService,
	scaleSvc svc.ScaleService,
	catalogSvc svc.CatalogService,
	acceptanceSvc svc.AcceptanceService,
) *Handler {
	return &Handler{
		pillarSvc:     pillarSvc,
		competencySvc: competencySvc,
		scaleSvc:      scaleSvc,
		catalogSvc:    catalogSvc,
		acceptanceSvc: acceptanceSvc,
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func generateTraceID() string {
	id := uuid.New().String()
	return id[:8] + "-" + id[9:13]
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: failed to encode JSON response: %v", err)
	}
}

// httpStatus maps competency-specific error codes to HTTP status codes,
// falling back to pkgerrors.HTTPStatus for standard codes.
func httpStatus(err error) int {
	var de *pkgerrors.DomainError
	if !pkgerrors.AsDomainError(err, &de) {
		return http.StatusInternalServerError
	}
	switch de.Code {
	case "PILLAR_NOT_FOUND", "COMPETENCY_NOT_FOUND", "RESOURCE_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_CURSOR", "INVALID_PARAMETER", "VALIDATION_ERROR",
		"INVALID_LEVEL", "DUPLICATE_LEVEL", "INVALID_REQUEST":
		return http.StatusBadRequest
	case "DUPLICATE_NAME", "PILLAR_HAS_COMPETENCIES", "COMPETENCY_HAS_CRITERIA",
		"CONCURRENT_UPDATE", "DUPLICATE_ENTRY":
		return http.StatusConflict
	case "PRECONDITION_REQUIRED":
		return http.StatusPreconditionRequired
	default:
		return pkgerrors.HTTPStatus(err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	traceID := generateTraceID()
	status := httpStatus(err)
	var de *pkgerrors.DomainError
	if pkgerrors.AsDomainError(err, &de) {
		writeJSON(w, status, pkgerrors.NewAPIErrorResponse(de, traceID))
		return
	}
	writeJSON(w, status, pkgerrors.NewAPIErrorResponse(
		pkgerrors.NewDomainError(pkgerrors.InvalidRequest, err.Error(), err),
		traceID,
	))
}

// parsePagination parses cursor and limit from query params with defaults.
func parsePagination(r *http.Request) (cursor string, limit int, err error) {
	cursor = r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = 20
	} else {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 {
			return "", 0, pkgerrors.NewDomainError("INVALID_PARAMETER",
				"limit must be a positive integer", nil)
		}
		limit = l
		if limit > 100 {
			limit = 100
		}
	}
	return cursor, limit, nil
}

// parseIDParam parses a UUID path parameter.
func parseIDParam(r *http.Request, param string) (string, error) {
	val := chi.URLParam(r, param)
	if val == "" {
		return "", pkgerrors.NewDomainError("INVALID_PARAMETER",
			"missing path parameter: "+param, nil)
	}
	if _, err := uuid.Parse(val); err != nil {
		return "", pkgerrors.NewDomainError("INVALID_PARAMETER",
			"invalid UUID: "+param, nil)
	}
	return val, nil
}

// parseIfMatch parses the If-Match header as an RFC3339Nano timestamp.
func parseIfMatch(r *http.Request) (time.Time, error) {
	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		return time.Time{}, pkgerrors.NewDomainError(pkgerrors.MissingIfMatch,
			"If-Match header is required for this operation", nil)
	}
	t, err := time.Parse(time.RFC3339Nano, ifMatch)
	if err != nil {
		return time.Time{}, pkgerrors.NewDomainError(pkgerrors.InvalidIfMatch,
			"If-Match must be an RFC3339Nano timestamp", err)
	}
	return t, nil
}

// getIdempotencyKey extracts the Idempotency-Key from context (set by middleware).
func getIdempotencyKey(ctx context.Context) string {
	v := ctx.Value(idempCtxKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

// ---------------------------------------------------------------------------
// 1. GET /api/v1/pillars — ListPillars
// ---------------------------------------------------------------------------

// ListPillars handles GET /api/v1/pillars.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) ListPillars(w http.ResponseWriter, r *http.Request) {
	cursor, limit, err := parsePagination(r)
	if err != nil {
		writeError(w, err)
		return
	}

	include := strings.Split(r.URL.Query().Get("include"), ",")
	// Filter out empty strings from splitting
	var cleanInclude []string
	for _, s := range include {
		if s != "" {
			cleanInclude = append(cleanInclude, s)
		}
	}

	result, err := h.pillarSvc.List(r.Context(), svc.ListOptions{
		Cursor:  cursor,
		Limit:   limit,
		Include: cleanInclude,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PaginatedResponse[dto.PillarListItem]{
		Data: result.Data,
		Pagination: dto.Pagination{
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		},
	})
}

// ---------------------------------------------------------------------------
// 2. POST /api/v1/pillars — CreatePillar
// ---------------------------------------------------------------------------

// CreatePillar handles POST /api/v1/pillars.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) CreatePillar(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePillarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.pillarSvc.Create(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// ---------------------------------------------------------------------------
// 3. GET /api/v1/pillars/{id} — GetPillar
// ---------------------------------------------------------------------------

// GetPillar handles GET /api/v1/pillars/{id}.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) GetPillar(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	includeCompetencies := r.URL.Query().Get("include") == "competencies"
	result, err := h.pillarSvc.Get(r.Context(), id, includeCompetencies)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 4. PUT /api/v1/pillars/{id} — UpdatePillar
// ---------------------------------------------------------------------------

// UpdatePillar handles PUT /api/v1/pillars/{id}.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) UpdatePillar(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	ifMatch, err := parseIfMatch(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req dto.UpdatePillarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.pillarSvc.Update(r.Context(), id, req, ifMatch)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 5. DELETE /api/v1/pillars/{id} — DeletePillar
// ---------------------------------------------------------------------------

// DeletePillar handles DELETE /api/v1/pillars/{id}.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) DeletePillar(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	if err := h.pillarSvc.Delete(r.Context(), id, force); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// 6. GET /api/v1/pillars/{pillarId}/competencies — ListCompetenciesByPillar
// ---------------------------------------------------------------------------

// ListCompetenciesByPillar handles GET /api/v1/pillars/{pillarId}/competencies.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) ListCompetenciesByPillar(w http.ResponseWriter, r *http.Request) {
	pillarID, err := parseIDParam(r, "pillarId")
	if err != nil {
		writeError(w, err)
		return
	}

	cursor, limit, err := parsePagination(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.competencySvc.ListByPillar(r.Context(), pillarID, svc.ListOptions{
		Cursor: cursor,
		Limit:  limit,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PaginatedResponse[dto.CompetencyLite]{
		Data: result.Data,
		Pagination: dto.Pagination{
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		},
	})
}

// ---------------------------------------------------------------------------
// 7. POST /api/v1/pillars/{pillarId}/competencies — CreateCompetency
// ---------------------------------------------------------------------------

// CreateCompetency handles POST /api/v1/pillars/{pillarId}/competencies.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) CreateCompetency(w http.ResponseWriter, r *http.Request) {
	pillarID, err := parseIDParam(r, "pillarId")
	if err != nil {
		writeError(w, err)
		return
	}

	var req dto.CreateCompetencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.competencySvc.Create(r.Context(), pillarID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// ---------------------------------------------------------------------------
// 8. GET /api/v1/competencies/{id} — GetCompetency
// ---------------------------------------------------------------------------

// GetCompetency handles GET /api/v1/competencies/{id}.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) GetCompetency(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.competencySvc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 9. PUT /api/v1/competencies/{id} — UpdateCompetency
// ---------------------------------------------------------------------------

// UpdateCompetency handles PUT /api/v1/competencies/{id}.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) UpdateCompetency(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	ifMatch, err := parseIfMatch(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req dto.UpdateCompetencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.competencySvc.Update(r.Context(), id, req, ifMatch)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 10. DELETE /api/v1/competencies/{id} — DeleteCompetency
// ---------------------------------------------------------------------------

// DeleteCompetency handles DELETE /api/v1/competencies/{id}.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) DeleteCompetency(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	if err := h.competencySvc.Delete(r.Context(), id, force); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// 11. GET /api/v1/competencies/{id}/scale-criteria — GetScaleCriteria
// ---------------------------------------------------------------------------

// GetScaleCriteria handles GET /api/v1/competencies/{id}/scale-criteria.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) GetScaleCriteria(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.scaleSvc.GetByCompetency(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 12. POST /api/v1/competencies/{id}/scale-criteria — UpsertScaleCriteria
// ---------------------------------------------------------------------------

// UpsertScaleCriteria handles POST /api/v1/competencies/{id}/scale-criteria.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) UpsertScaleCriteria(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}

	var req dto.ScaleCriteriaBulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.scaleSvc.Upsert(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 13. GET /api/v1/levels — ListLevels
// ---------------------------------------------------------------------------

// ListLevels handles GET /api/v1/levels.
// Returns a 304 Not Modified when the client sends If-None-Match with a
// matching ETag.
// TODO(auth:C7): any authenticated user.
func (h *Handler) ListLevels(w http.ResponseWriter, r *http.Request) {
	levels, err := h.catalogSvc.ListLevels(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	etagVal, err := svc.ComputeETag("levels:v1", levels)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("ETag", etagVal)

	// Check If-None-Match
	if noneMatch := r.Header.Get("If-None-Match"); noneMatch == etagVal {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	writeJSON(w, http.StatusOK, levels)
}

// ---------------------------------------------------------------------------
// 14. GET /api/v1/acceptance-levels — ListAcceptanceLevels
// ---------------------------------------------------------------------------

// ListAcceptanceLevels handles GET /api/v1/acceptance-levels.
// TODO(auth:C7): any authenticated org user.
func (h *Handler) ListAcceptanceLevels(w http.ResponseWriter, r *http.Request) {
	var filter svc.AcceptanceFilter

	if pid := r.URL.Query().Get("profile_id"); pid != "" {
		if _, err := uuid.Parse(pid); err != nil {
			writeError(w, pkgerrors.NewDomainError("INVALID_PARAMETER",
				"invalid profile_id UUID", err))
			return
		}
		filter.ProfileID = &pid
	}
	if cid := r.URL.Query().Get("competency_id"); cid != "" {
		if _, err := uuid.Parse(cid); err != nil {
			writeError(w, pkgerrors.NewDomainError("INVALID_PARAMETER",
				"invalid competency_id UUID", err))
			return
		}
		filter.CompetencyID = &cid
	}

	items, err := h.acceptanceSvc.List(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// ---------------------------------------------------------------------------
// 15. POST /api/v1/acceptance-levels — UpsertAcceptanceLevel
// ---------------------------------------------------------------------------

// UpsertAcceptanceLevel handles POST /api/v1/acceptance-levels.
// TODO(auth:C7): rh or admin role required.
func (h *Handler) UpsertAcceptanceLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.UpsertAcceptanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"invalid JSON body", err))
		return
	}

	result, err := h.acceptanceSvc.Upsert(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// 16. GET /api/v1/profiles — ListProfiles
// ---------------------------------------------------------------------------

// ListProfiles handles GET /api/v1/profiles.
// Returns a 304 Not Modified when the client sends If-None-Match with a
// matching ETag.
// TODO(auth:C7): any authenticated user.
func (h *Handler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.catalogSvc.ListProfiles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	etagVal, err := svc.ComputeETag("profiles:v1", profiles)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("ETag", etagVal)

	if noneMatch := r.Header.Get("If-None-Match"); noneMatch == etagVal {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	writeJSON(w, http.StatusOK, profiles)
}
