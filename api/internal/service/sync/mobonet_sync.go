package sync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/seed"
)

// MobonetEmployee representa un registro activo de Mobonet (Status=1).
// Campos alineados a import/main.go: email, Nombre, Apellidos, NumEmpleado, Puestos.Nombre, Departamentos.Nombre.
type MobonetEmployee struct {
	Email          string  `json:"email"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	EmployeeNumber string  `json:"employee_number"`
	JobTitle       string  `json:"job_title"`
	DeptName       string  `json:"dept_name"`
	ManagerEmail   *string `json:"manager_email,omitempty"`
}

// Service sincroniza is_active + upsert de employees contra Mobonet y SSO.
type Service struct {
	db     *sql.DB
	client *http.Client
	// sso opcional inyectado; si nil se usa syncSSO interno best-effort
	sso SSOSyncer
}

// SSOSyncer abstrae el sync de usuarios al SSO seed (reutiliza lógica import/passSSOSeed).
type SSOSyncer interface {
	SyncUsers(ctx context.Context, employeeNumbers []string) error
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// WithSSO inyecta dependencia SSO (opcional, usado desde main.go).
func (s *Service) WithSSO(ss SSOSyncer) *Service {
	s.sso = ss
	return s
}

// fetchMobonetEmployees lista empleados activos completos en Mobonet.
// Si MOBONET_URL y MOBONET_KEY vacíos, retorna mock (activos actuales en BD) — idempotente.
func (s *Service) fetchMobonetEmployees(ctx context.Context) ([]MobonetEmployee, error) {
	url := strings.TrimSpace(os.Getenv("MOBONET_URL"))
	key := strings.TrimSpace(os.Getenv("MOBONET_KEY"))
	if url == "" && key == "" {
		rows, err := s.db.QueryContext(ctx, `SELECT email, first_name, last_name, employee_number, COALESCE(job_title,'') FROM employees WHERE is_active = true`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var out []MobonetEmployee
		for rows.Next() {
			var e MobonetEmployee
			if err := rows.Scan(&e.Email, &e.FirstName, &e.LastName, &e.EmployeeNumber, &e.JobTitle); err != nil {
				continue
			}
			if e.EmployeeNumber != "" {
				out = append(out, e)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		slog.Info("mobonet_sync: mock mode (MOBONET_URL/KEY vacíos)", "active_count", len(out))
		return out, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("X-API-Key", key)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	// 1) []MobonetEmployee directo
	var direct []MobonetEmployee
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 && direct[0].EmployeeNumber != "" {
		return direct, nil
	}
	// 2) []map genérico
	var asMaps []map[string]interface{}
	if err := json.Unmarshal(raw, &asMaps); err == nil && len(asMaps) > 0 {
		var out []MobonetEmployee
		for _, m := range asMaps {
			e := mapToMobonetEmployee(m)
			if e.EmployeeNumber != "" {
				out = append(out, e)
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	// 3) wrapper {data: [...]}
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err == nil {
		for _, wk := range []string{"data", "employees", "result"} {
			if v, ok := wrapper[wk]; ok {
				var ws []MobonetEmployee
				if err := json.Unmarshal(v, &ws); err == nil && len(ws) > 0 {
					return ws, nil
				}
				var wm []map[string]interface{}
				if err := json.Unmarshal(v, &wm); err == nil {
					var out []MobonetEmployee
					for _, m := range wm {
						e := mapToMobonetEmployee(m)
						if e.EmployeeNumber != "" {
							out = append(out, e)
						}
					}
					if len(out) > 0 {
						return out, nil
					}
				}
			}
		}
	}
	// 4) fallback legacy []string -> convertir a MobonetEmployee mínimo
	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err == nil && len(asStrings) > 0 {
		var out []MobonetEmployee
		for _, n := range asStrings {
			n = strings.TrimSpace(n)
			if n != "" {
				out = append(out, MobonetEmployee{EmployeeNumber: n, Email: n + "@mock.local", FirstName: n, LastName: ""})
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	slog.Warn("mobonet_sync: respuesta no reconocida, retorna vacío")
	return []MobonetEmployee{}, nil
}

func mapToMobonetEmployee(m map[string]interface{}) MobonetEmployee {
	strVal := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
		return ""
	}
	email := strVal("email", "Email", "correo")
	first := strVal("first_name", "Nombre", "firstName", "nombre")
	last := strVal("last_name", "Apellidos", "lastName", "apellidos")
	empNum := strVal("employee_number", "NumEmpleado", "num_empleado", "employeeNumber", "numero_empleado", "Num_Empleado")
	job := strVal("job_title", "Puesto", "Nombre_puesto", "node", "puesto")
	dept := strVal("dept_name", "Departamento", "dept", "Nombre_departamento")
	var mgr *string
	if v := strVal("manager_email", "mgrEmail", "email_jefe"); v != "" {
		mgr = &v
	} else if raw, ok := m["manager_email"]; ok && raw != nil {
		if s, ok := raw.(string); ok && s != "" {
			mgr = &s
		}
	}
	if email == "" && empNum != "" {
		email = empNum + "@mock.local"
	}
	return MobonetEmployee{
		Email: email, FirstName: first, LastName: last,
		EmployeeNumber: empNum, JobTitle: job, DeptName: dept, ManagerEmail: mgr,
	}
}

// fetchMobonet lista employee_number activos (compat legacy).
func (s *Service) fetchMobonet(ctx context.Context) ([]string, error) {
	emps, err := s.fetchMobonetEmployees(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(emps))
	for _, e := range emps {
		out = append(out, e.EmployeeNumber)
	}
	return out, nil
}

// Run sincroniza is_active + upsert. Retorna cantidad deshabilitados. Idempotente.
func (s *Service) Run(ctx context.Context) (disabledCount int, err error) {
	emps, err := s.fetchMobonetEmployees(ctx)
	if err != nil {
		return 0, err
	}

	// Deduplicar por employee_number y limpiar
	seen := make(map[string]MobonetEmployee, len(emps))
	var uniq []MobonetEmployee
	for _, e := range emps {
		e.EmployeeNumber = strings.TrimSpace(e.EmployeeNumber)
		if e.EmployeeNumber == "" {
			continue
		}
		if _, ok := seen[e.EmployeeNumber]; ok {
			continue
		}
		seen[e.EmployeeNumber] = e
		uniq = append(uniq, e)
	}
	emps = uniq

	if len(emps) == 0 {
		slog.Warn("mobonet_sync: lista de activos vacía, no se deshabilita nadie (idempotente)")
		return 0, nil
	}

	// Extraer solo numbers para IN clause
	activeNums := make([]string, len(emps))
	for i, e := range emps {
		activeNums[i] = e.EmployeeNumber
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(activeNums))
	args := make([]interface{}, len(activeNums))
	for i, n := range activeNums {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = n
	}
	inClause := strings.Join(placeholders, ",")

	// Desactivar los que ya no están en Mobonet
	qDisable := `UPDATE employees SET is_active = false, updated_at = NOW()
		WHERE employee_number NOT IN (` + inClause + `) AND is_active = true
		RETURNING employee_number`

	rows, err := s.db.QueryContext(ctx, qDisable, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var disabled []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err == nil {
			disabled = append(disabled, n)
			slog.Info("mobonet_sync: empleado deshabilitado", "employee_number", n)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Reactivar los que volvieron a aparecer
	qEnable := `UPDATE employees SET is_active = true, updated_at = NOW()
		WHERE employee_number IN (` + inClause + `) AND is_active = false
		RETURNING employee_number`
	rows2, err := s.db.QueryContext(ctx, qEnable, args...)
	if err != nil {
		slog.Error("mobonet_sync: error reactivando", "error", err)
		return len(disabled), err
	}
	defer rows2.Close()
	for rows2.Next() {
		var n string
		if err := rows2.Scan(&n); err == nil {
			slog.Info("mobonet_sync: empleado reactivado", "employee_number", n)
		}
	}
	_ = rows2.Err()

	// ── Upsert phase: map existentes por employee_number ──
	existing := make(map[string]struct {
		id        uuid.UUID
		email     string
		firstName string
		lastName  string
		jobTitle  string
	})
	rows3, err := s.db.QueryContext(ctx, `SELECT id, employee_number, email, first_name, last_name, COALESCE(job_title,'') FROM employees`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var id uuid.UUID
			var empNum, email, fn, ln, jt string
			if err := rows3.Scan(&id, &empNum, &email, &fn, &ln, &jt); err == nil {
				existing[empNum] = struct {
					id        uuid.UUID
					email     string
					firstName string
					lastName  string
					jobTitle  string
				}{id, email, fn, ln, jt}
			}
		}
		_ = rows3.Err()
	}

	// Resolver org_node fallback y profile fallback (sin duplicar SQL de import)
	var fallbackOrgID uuid.UUID
	_ = s.db.QueryRowContext(ctx, `SELECT id FROM org_nodes LIMIT 1`).Scan(&fallbackOrgID)
	var fallbackProfileID uuid.UUID
	_ = s.db.QueryRowContext(ctx, `SELECT id FROM evaluation_profiles WHERE name='colaborador' LIMIT 1`).Scan(&fallbackProfileID)
	if fallbackProfileID == uuid.Nil {
		_ = s.db.QueryRowContext(ctx, `SELECT id FROM evaluation_profiles LIMIT 1`).Scan(&fallbackProfileID)
	}
	systemID := seed.SeedID("system")

	var upserted, updated int
	for _, m := range emps {
		if m.Email == "" {
			m.Email = m.EmployeeNumber + "@mock.local"
		}
		if rec, ok := existing[m.EmployeeNumber]; ok {
			// Existe → refrescar si cambió (UpdateFromMobonet)
			if rec.email != m.Email || rec.firstName != m.FirstName || rec.lastName != m.LastName || rec.jobTitle != m.JobTitle {
				_, err := s.db.ExecContext(ctx,
					`UPDATE employees SET email=$1, first_name=$2, last_name=$3, job_title=$4, updated_at=NOW() WHERE id=$5`,
					m.Email, m.FirstName, m.LastName, m.JobTitle, rec.id)
				if err != nil {
					slog.Error("mobonet_sync: update failed", "employee_number", m.EmployeeNumber, "error", err)
					continue
				}
				slog.Info("mobonet_sync: empleado actualizado", "employee_number", m.EmployeeNumber)
				updated++
			}
		} else {
			// No existe → upsert (INSERT ON CONFLICT id)
			orgID := fallbackOrgID
			if m.DeptName != "" {
				var deptID uuid.UUID
				if err := s.db.QueryRowContext(ctx, `SELECT id FROM org_nodes WHERE name=$1 LIMIT 1`, m.DeptName).Scan(&deptID); err == nil {
					orgID = deptID
				}
			}
			if orgID == uuid.Nil {
				orgID = fallbackOrgID
			}
			profileID := fallbackProfileID
			id := seed.SeedID(m.Email)
			_, err := s.db.ExecContext(ctx,
				`INSERT INTO employees (id, email, first_name, last_name, employee_number, job_title, is_active, org_node_id, profile_id, created_by, updated_by, created_at, updated_at)
				 VALUES ($1,$2,$3,$4,$5,$6,true,$7,$8,$9,$10,NOW(),NOW())
				 ON CONFLICT (id) DO UPDATE SET
				   email=EXCLUDED.email, first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name,
				   employee_number=EXCLUDED.employee_number, job_title=EXCLUDED.job_title,
				   org_node_id=EXCLUDED.org_node_id, profile_id=EXCLUDED.profile_id,
				   is_active=true, updated_at=NOW()`,
				id, m.Email, m.FirstName, m.LastName, m.EmployeeNumber, m.JobTitle, orgID, profileID, systemID, systemID)
			if err != nil {
				slog.Error("mobonet_sync: upsert failed", "employee_number", m.EmployeeNumber, "error", err)
				continue
			}
			slog.Info("mobonet_sync: empleado upsert", "employee_number", m.EmployeeNumber, "email", m.Email)
			upserted++
		}
	}
	if upserted > 0 || updated > 0 {
		slog.Info("mobonet_sync: upsert phase", "upserted", upserted, "updated", updated)
	}

	// ── SSO sync al finalizar si hubo cambios ──
	if upserted+updated > 0 {
		if s.sso != nil {
			var nums []string
			for _, e := range emps {
				nums = append(nums, e.EmployeeNumber)
			}
			if err := s.sso.SyncUsers(ctx, nums); err != nil {
				slog.Error("mobonet_sync: sso sync failed", "error", err)
			}
		} else {
			if err := s.syncSSO(ctx); err != nil {
				slog.Error("mobonet_sync: sso sync failed", "error", err)
			}
		}
	}

	return len(disabled), nil
}

// ── SSO sync interno (reutiliza lógica import/passSSOSeed, best-effort) ──
var (
	ssoBaseRe        = regexp.MustCompile(`^(.*?)/realms/[^/]+/?$`)
	ssoUserNotFoundRe = regexp.MustCompile(`Usuario\s+\\?"(-?\d+)\\"?\s+no\s+encontrado`)
)

func ssoBaseURL(issuer string) string {
	if m := ssoBaseRe.FindStringSubmatch(issuer); m != nil {
		return m[1]
	}
	return issuer
}

func (s *Service) syncSSO(ctx context.Context) error {
	adminUser := strings.TrimSpace(os.Getenv("SSO_SEED_ADMIN_USER"))
	adminPass := strings.TrimSpace(os.Getenv("SSO_SEED_ADMIN_PASSWORD"))
	if adminUser == "" || adminPass == "" {
		slog.Info("mobonet_sync: sso sync skipped (no creds)")
		return nil
	}
	baseURL := strings.TrimSpace(os.Getenv("SSO_SEED_BASE_URL"))
	if baseURL == "" {
		baseURL = ssoBaseURL(strings.TrimSpace(os.Getenv("SSO_KC_ISSUER")))
	}
	clientID := strings.TrimSpace(os.Getenv("SSO_CLIENT_ID"))
	if clientID == "" {
		slog.Info("mobonet_sync: sso sync skipped (no client_id)")
		return nil
	}
	if baseURL == "" {
		slog.Info("mobonet_sync: sso sync skipped (no base url)")
		return nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT employee_number FROM employees WHERE is_active = true`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type ssoUser struct {
		User        string   `json:"user"`
		RoleCodigos []string `json:"role_codigos,omitempty"`
	}
	var usuarios []ssoUser
	skipEmp := strings.TrimSpace(os.Getenv("SSO_SEED_DEV_EMPLOYEE_NUMBER"))
	for rows.Next() {
		var empNum string
		if err := rows.Scan(&empNum); err != nil {
			continue
		}
		if skipEmp != "" && strings.EqualFold(empNum, skipEmp) {
			continue
		}
		usuarios = append(usuarios, ssoUser{User: empNum, RoleCodigos: []string{"usuario"}})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(usuarios) == 0 {
		return nil
	}
	// login
	loginURL := baseURL + "/api/seed/login"
	loginBody := map[string]string{"user": adminUser, "password": adminPass}
	var loginResp struct {
		Token string `json:"token"`
	}
	status, body, err := doJSON(ctx, s.client, http.MethodPost, loginURL, "", loginBody, &loginResp)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		slog.Error("mobonet_sync: sso login failed", "status", status, "body", string(body))
		return nil
	}
	seedURL := baseURL + "/api/seed/sistemas/usuarios"
	payload := map[string]interface{}{
		"sync_keycloak": true,
		"items": []map[string]interface{}{
			{"client_id": clientID, "usuarios": usuarios},
		},
	}
	// retry on Usuario X no encontrado
	discarded := 0
	for {
		status, body, err := doJSON(ctx, s.client, http.MethodPost, seedURL, loginResp.Token, payload, nil)
		if err != nil {
			return err
		}
		if status >= 200 && status < 300 {
			slog.Info("mobonet_sync: sso sync OK", "users", len(usuarios), "discarded", discarded)
			return nil
		}
		m := ssoUserNotFoundRe.FindStringSubmatch(string(body))
		if m == nil || status != 400 {
			slog.Error("mobonet_sync: sso seed failed", "status", status, "body", string(body))
			return nil
		}
		notFound := m[1]
		found := false
		for i, u := range usuarios {
			if u.User == notFound {
				usuarios = append(usuarios[:i], usuarios[i+1:]...)
				found = true
				break
			}
		}
		if !found || len(usuarios) == 0 {
			return nil
		}
		discarded++
		payload["items"] = []map[string]interface{}{{"client_id": clientID, "usuarios": usuarios}}
	}
}

func doJSON(ctx context.Context, client *http.Client, method, url, token string, reqBody, respOut interface{}) (int, []byte, error) {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if respOut != nil && len(respData) > 0 {
		if err := json.Unmarshal(respData, respOut); err != nil {
			return resp.StatusCode, respData, err
		}
	}
	return resp.StatusCode, respData, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
