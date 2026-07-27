package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/sed-evaluacion-desempeno/api/internal/seed"
)

// ssoBaseRe strips the /realms/<realm> suffix from an OIDC issuer
// to derive the base URL (e.g. for the SSO seed admin API).
var ssoBaseRe = regexp.MustCompile(`^(.*?)/realms/[^/]+/?$`)

// ssoUserNotFoundRe extracts the employee number from a 400 error like:
// {"error":"Usuario \"20569\" no encontrado"}
// The quotes may be JSON-escaped (\"…\") or plain ("…").
// Employee numbers can be negative (e.g. \"-13\").
var ssoUserNotFoundRe = regexp.MustCompile(`Usuario\s+\\?"(-?\d+)\\"?\s+no\s+encontrado`)

type reasons map[string]int

func (r reasons) add(key string)      { r[key]++ }
func (r reasons) merge(other reasons) { for k, v := range other { r[k] += v } }
func (r reasons) String() string {
	if len(r) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(r))
	for k, v := range r {
		parts = append(parts, fmt.Sprintf("%s:%d", k, v))
	}
	// sort for stable output
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

type passResult struct {
	label       string
	sourceTotal int // total rows in source table (no JOIN filter)
	fetched     int // rows returned by query (after JOINs)
	written     int // rows that would be written (dry-run) or were written
	reasons     reasons
}

func (p passResult) String() string {
	return fmt.Sprintf("  %-12s source:%-5d fetched:%-5d write:%-5d  %s",
		p.label, p.sourceTotal, p.fetched, p.written, p.reasons.String())
}

func main() {
	dryRun := flag.Bool("dry-run", false, "print queries without writing")
	flag.Parse()

	_ = godotenv.Load()

	seedDBURL := os.Getenv("SEED_DB_URL")
	if seedDBURL == "" {
		log.Fatal("[import] SEED_DB_URL is required (e.g. user:pass@tcp(host:3306)/mobonet?parseTime=true)")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("[import] DATABASE_URL is required (internal PostgreSQL)")
	}

	extDB, err := sql.Open("mysql", seedDBURL)
	if err != nil {
		log.Fatalf("[import] failed to open source DB: %v", err)
	}
	defer extDB.Close()
	extDB.SetMaxOpenConns(1)

	tgtDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("[import] failed to open target DB: %v", err)
	}
	defer tgtDB.Close()

	ctx := context.Background()

	if *dryRun {
		log.Println("[import] DRY RUN — will not write")
	}

	// ── Source row counts (active only) ──
	srcOrgs := sourceCount(ctx, extDB, `SELECT COUNT(*) FROM mobonet.Departamentos d WHERE d.Status = 1`)
	srcEmps := sourceCount(ctx, extDB, `SELECT COUNT(*) FROM mobonet.Empleados e WHERE e.Status = 1`)

	// ── Setup: Organization + EvaluationProfiles ──
	orgID := setupOrg(ctx, tgtDB, *dryRun)
	setupProfiles(ctx, tgtDB, *dryRun)

	// ── Passes ──
	r1, orgIDMap, headEmailMap, rootOrgUUID := passOrgs(ctx, tgtDB, extDB, srcOrgs, orgID, *dryRun)
	r2 := passOrgParents(ctx, tgtDB, extDB, orgIDMap, rootOrgUUID, *dryRun)
	r3, empIDMap, managerMap := passEmployees(ctx, tgtDB, extDB, srcEmps, orgIDMap, *dryRun)
	r4 := passManagers(ctx, tgtDB, empIDMap, managerMap, *dryRun)
	r5 := passHeadEmployees(ctx, tgtDB, headEmailMap, empIDMap, *dryRun)
	r6 := passSSOSeed(ctx, tgtDB, *dryRun)

	// ── Summary ──
	log.Println("[import] ═════════════════════════════════════════════════════════════")
	log.Println("[import] SUMMARY")
	log.Println("[import]")
	log.Println("[import]   source = total rows in MariaDB")
	log.Println("[import]   fetched = rows returned by query (all source rows, LEFT JOINs)")
	log.Println("[import]   write = rows that would be inserted/updated")
	log.Println("[import]")
	log.Println(r1)
	log.Println(r2)
	log.Println(r3)
	log.Println(r4)
	log.Println(r5)
	log.Println(r6)
	log.Println("[import]")
	if *dryRun {
		log.Println("[import]   (dry run — nothing was written)")
	}
	log.Println("[import] ═════════════════════════════════════════════════════════════")
}

// ── Helpers ────────────────────────────────────────────────────────────

func sourceCount(ctx context.Context, db *sql.DB, query string) int {
	var n int
	if err := db.QueryRowContext(ctx, query).Scan(&n); err != nil {
		log.Printf("[import] source count query failed: %v", err)
		return -1
	}
	return n
}

func setupOrg(ctx context.Context, tgtDB *sql.DB, dryRun bool) uuid.UUID {
	orgID := seed.SeedID("org-import")
	if dryRun {
		log.Printf("[import] [dry] setup org: %s", orgID)
		return orgID
	}
	_, err := tgtDB.ExecContext(ctx,
		`INSERT INTO organizations (id, name, slug, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (id) DO NOTHING`,
		orgID, "Organización Import", "org-import",
	)
	if err != nil {
		log.Printf("[import] setup org: %v", err)
	}
	return orgID
}

func setupProfiles(ctx context.Context, tgtDB *sql.DB, dryRun bool) {
	profiles := []struct{ name, desc string }{
		{"colaborador", "Colaborador de la organización"},
		{"jefe", "Jefe de equipo o departamento"},
		{"director", "Director de área"},
		{"director-general", "Director general de la organización"},
		{"rh", "Recursos humanos"},
	}
	for _, p := range profiles {
		pid := seed.SeedID("profile-" + p.name)
		if dryRun {
			continue
		}
		_, err := tgtDB.ExecContext(ctx,
			`INSERT INTO evaluation_profiles (id, name, description)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO NOTHING`,
			pid, p.name, p.desc,
		)
		if err != nil {
			log.Printf("[import] setup profile %s: %v", p.name, err)
		}
	}
}

// ── Pass 1: OrgNodes ───────────────────────────────────────────────────

func passOrgs(ctx context.Context, tgtDB, extDB *sql.DB, srcTotal int, orgID uuid.UUID, dryRun bool) (passResult, map[string]uuid.UUID, map[uuid.UUID]*string, uuid.UUID) {
	pr := passResult{label: "orgs", sourceTotal: srcTotal, reasons: make(reasons)}
	idMap := make(map[string]uuid.UUID)
	headMap := make(map[uuid.UUID]*string) // orgNodeUUID → responsible employee email
	var rootOrgUUID uuid.UUID
	logged := 0

	rows, err := extDB.QueryContext(ctx, `
		SELECT d.IdDepartamento, d.Nombre, e.email AS head_employee, d2.Nombre AS parent
		FROM mobonet.Departamentos d
		LEFT JOIN mobonet.Empleados e ON e.IdEmpleado = d.FkIdEmpleadoResponsable AND e.Status = 1
		LEFT JOIN mobonet.Departamentos d2 ON d2.IdDepartamento = d.IdDepartamentoPadre
		WHERE d.Status = 1
	`)
	if err != nil {
		pr.reasons.add("query_error")
		log.Fatalf("[import] org query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		pr.fetched++
		var deptID int
		var name string
		var headEmail, parentName *string
		if err := rows.Scan(&deptID, &name, &headEmail, &parentName); err != nil {
			pr.reasons.add("scan_error")
			continue
		}

		orgUUID := seed.SeedID(name)

		if headEmail == nil {
			pr.reasons.add("no_head_employee")
		} else {
			headMap[orgUUID] = headEmail
		}

		idMap[name] = orgUUID

		if deptID == 1 {
			rootOrgUUID = orgUUID
		}

		if dryRun {
			pr.written++
			continue
		}

		if _, err := tgtDB.ExecContext(ctx,
			`INSERT INTO org_nodes (id, name, type, code, organization_id, created_at, updated_at)
			 VALUES ($1, $2, 'corporate', $3, $4, NOW(), NOW())
			 ON CONFLICT (id) DO UPDATE SET name = $2`,
			orgUUID, name, shortCode(name), orgID,
		); err != nil {
			pr.reasons.add("insert_error")
			if logged < 3 {
				log.Printf("[import] org insert error [%s]: %v", name, err)
				logged++
			}
			continue
		}
		pr.written++
	}

	if pr.fetched == 0 {
		pr.reasons.add("no_rows")
	}

	// Set root_node_id on the organization
	if rootOrgUUID != uuid.Nil && !dryRun {
		if _, err := tgtDB.ExecContext(ctx,
			`UPDATE organizations SET root_node_id = $1, updated_at = NOW() WHERE id = $2`,
			rootOrgUUID, orgID,
		); err != nil {
			log.Printf("[import] update root_node_id: %v", err)
		}
	}

	return pr, idMap, headMap, rootOrgUUID
}

// ── Pass 2: OrgNode parent edges ───────────────────────────────────────

func passOrgParents(ctx context.Context, tgtDB, extDB *sql.DB, idMap map[string]uuid.UUID, rootOrgUUID uuid.UUID, dryRun bool) passResult {
	pr := passResult{label: "parents", reasons: make(reasons)}

	// Source count: active departments only
	pr.sourceTotal = sourceCount(ctx, extDB, `SELECT COUNT(*) FROM mobonet.Departamentos d WHERE d.Status = 1`)

	rows, err := extDB.QueryContext(ctx, `
		SELECT d.Nombre, d2.Nombre AS parent, d.IdDepartamentoPadre
		FROM mobonet.Departamentos d
		LEFT JOIN mobonet.Departamentos d2 ON d2.IdDepartamento = d.IdDepartamentoPadre
		WHERE d.Status = 1
	`)
	if err != nil {
		pr.reasons.add("query_error")
		log.Fatalf("[import] org parent query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		pr.fetched++
		var deptName string
		var parentName *string
		var idDepartamentoPadre sql.NullInt64
		if err := rows.Scan(&deptName, &parentName, &idDepartamentoPadre); err != nil {
			pr.reasons.add("scan_error")
			continue
		}

		childID, ok := idMap[deptName]
		if !ok {
			pr.reasons.add("dept_not_found")
			continue
		}

		// No parent: link to root (except root itself)
		if parentName == nil {
			if rootOrgUUID != uuid.Nil && childID != rootOrgUUID {
				if dryRun {
					pr.written++
					continue
				}
				if _, err := tgtDB.ExecContext(ctx,
					`UPDATE org_nodes SET parent_id = $1 WHERE id = $2`,
					rootOrgUUID, childID,
				); err != nil {
					pr.reasons.add("update_error")
					continue
				}
				pr.written++
			} else {
				pr.reasons.add("root_or_no_parent")
			}
			continue
		}

		parentID, ok2 := idMap[*parentName]
		if !ok2 {
			pr.reasons.add("parent_not_found")
			continue
		}

		if dryRun {
			pr.written++
			continue
		}

		if _, err := tgtDB.ExecContext(ctx,
			`UPDATE org_nodes SET parent_id = $1 WHERE id = $2`,
			parentID, childID,
		); err != nil {
			pr.reasons.add("insert_error")
			continue
		}
		pr.written++
	}

	return pr
}

// ── Pass 3: Employees ──────────────────────────────────────────────────

func passEmployees(
	ctx context.Context,
	tgtDB, extDB *sql.DB,
	srcTotal int,
	orgIDMap map[string]uuid.UUID,
	dryRun bool,
) (passResult, map[string]uuid.UUID, map[string]*string) {
	pr := passResult{label: "employees", sourceTotal: srcTotal, reasons: make(reasons)}
	empIDMap := make(map[string]uuid.UUID)
	managerMap := make(map[string]*string)
	seenEmails := make(map[string]bool)    // tracks original emails already used
	emailAlias := make(map[string]string)  // original → aliased (for duplicates)
	logged := 0

	rows, err := extDB.QueryContext(ctx, `
		-- TODO: agregar condiciones para vendedor, regional, divisional cuando
		-- se defina la lógica de asignación por puesto o jerarquía retail.
		SELECT e.email, e.Nombre, e.Apellidos, e.NumEmpleado,
		       CASE
		         WHEN e.FkIdDepartamento = 1 THEN 'director-general'
		         WHEN e.FkIdDepartamento = 5 THEN 'rh'
		         WHEN EXISTS (
		           SELECT 1 FROM mobonet.Departamentos d3
		           WHERE d3.FkIdEmpleadoResponsable = e.IdEmpleado
		             AND d3.IdDireccion = d3.IdDepartamento
		             AND d3.IdDepartamento <> 1
		         ) THEN 'director'
		         WHEN EXISTS (
		           SELECT 1 FROM mobonet.Empleados e3
		           WHERE e3.FkIdEmpleadoJefeInmediato = e.IdEmpleado
		             AND e3.Status = 1
		         ) THEN 'jefe'
		         ELSE 'colaborador'
		       END,
		       d.Nombre, p.Nombre, e2.email
		FROM mobonet.Empleados e
		LEFT JOIN mobonet.Departamentos d ON d.IdDepartamento = e.FkIdDepartamento
		LEFT JOIN mobonet.Puestos p ON p.IdPuesto = e.FkIdPuesto
		LEFT JOIN mobonet.Empleados e2 ON e2.IdEmpleado = e.FkIdEmpleadoJefeInmediato AND e2.Status = 1
		WHERE e.Status = 1
	`)
	if err != nil {
		pr.reasons.add("query_error")
		log.Fatalf("[import] employee query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		pr.fetched++
		var email, firstName, lastName, empNum, profile string
		var dept, node *string
		var mgrEmail *string
		if err := rows.Scan(&email, &firstName, &lastName, &empNum,
			&profile, &dept, &node, &mgrEmail); err != nil {
			pr.reasons.add("scan_error")
			continue
		}

		if dept == nil {
			pr.reasons.add("dept_missing")
		}
		if node == nil {
			pr.reasons.add("position_missing")
		}

		// Handle duplicate emails: second occurrence gets +employee_number@domain
		importEmail := email
		if seenEmails[email] {
			at := strings.LastIndex(email, "@")
			if at > 0 {
				importEmail = email[:at] + "+" + empNum + email[at:]
			}
			emailAlias[email] = importEmail
			pr.reasons.add("duplicate_email")
			log.Printf("[import] duplicate email: %s → %s (%s %s)", email, importEmail, firstName, lastName)
		}
		seenEmails[email] = true

		empUUID := seed.SeedID(importEmail)
		empIDMap[importEmail] = empUUID

		// Resolve manager email: if original was aliased, use aliased version
		resolvedMgr := mgrEmail
		if mgrEmail != nil {
			if aliased, ok := emailAlias[*mgrEmail]; ok {
				resolvedMgr = &aliased
			}
		}
		managerMap[importEmail] = resolvedMgr

		if mgrEmail == nil {
			pr.reasons.add("no_manager")
		}

		if dryRun {
			pr.written++
			continue
		}

		var orgID *uuid.UUID
		if dept != nil {
			if id, ok := orgIDMap[*dept]; ok {
				orgID = &id
			} else {
				pr.reasons.add("dept_not_found")
			}
		}
		profileID := seed.SeedID("profile-" + profile)

		var jobTitle string
		if node != nil {
			jobTitle = *node
		}

		if _, err := tgtDB.ExecContext(ctx,
			`INSERT INTO employees (id, email, first_name, last_name, employee_number,
			 job_title, is_active, org_node_id, profile_id, created_by, updated_by,
			 created_at, updated_at)
			 VALUES ($1,$2,$3,$4,$5,$6,true,$7,$8,$9,$10,NOW(),NOW())
			 ON CONFLICT (id) DO UPDATE SET
			   email = $2, first_name = $3, last_name = $4, employee_number = $5,
			   job_title = $6, org_node_id = $7, profile_id = $8`,
			empUUID, importEmail, firstName, lastName, empNum, jobTitle,
			orgID, profileID,
			seed.SeedID("system"), seed.SeedID("system"),
		); err != nil {
			pr.reasons.add("insert_error")
			if logged < 3 {
				log.Printf("[import] employee insert error [%s]: %v", importEmail, err)
				logged++
			}
			continue
		}
		pr.written++
	}

	if pr.fetched == 0 {
		pr.reasons.add("no_rows")
	}
	return pr, empIDMap, managerMap
}

// ── Pass 4: Employee manager edges ─────────────────────────────────────

func passManagers(ctx context.Context, tgtDB *sql.DB, empIDMap map[string]uuid.UUID, managerMap map[string]*string, dryRun bool) passResult {
	pr := passResult{label: "managers", sourceTotal: len(managerMap), reasons: make(reasons)}

	for empEmail, mgrEmail := range managerMap {
		pr.fetched++
		if mgrEmail == nil {
			pr.reasons.add("no_manager_ref")
			continue
		}

		empID, ok1 := empIDMap[empEmail]
		mgrID, ok2 := empIDMap[*mgrEmail]
		if !ok1 || !ok2 {
			pr.reasons.add("manager_not_found")
			continue
		}

		if dryRun {
			pr.written++
			continue
		}

		if _, err := tgtDB.ExecContext(ctx,
			`UPDATE employees SET manager_id = $1 WHERE id = $2`,
			mgrID, empID,
		); err != nil {
			pr.reasons.add("insert_error")
			continue
		}
		pr.written++
	}

	return pr
}

// ── Pass 5: OrgNode head_employee_id ───────────────────────────────────

func passHeadEmployees(ctx context.Context, tgtDB *sql.DB, headMap map[uuid.UUID]*string, empIDMap map[string]uuid.UUID, dryRun bool) passResult {
	pr := passResult{label: "head_emps", sourceTotal: len(headMap), reasons: make(reasons)}

	for orgID, headEmail := range headMap {
		pr.fetched++
		if headEmail == nil {
			pr.reasons.add("no_head_email")
			continue
		}

		empID, ok := empIDMap[*headEmail]
		if !ok {
			pr.reasons.add("head_emp_not_found")
			continue
		}

		if dryRun {
			pr.written++
			continue
		}

		if _, err := tgtDB.ExecContext(ctx,
			`UPDATE org_nodes SET head_employee_id = $1 WHERE id = $2`,
			empID, orgID,
		); err != nil {
			pr.reasons.add("update_error")
			continue
		}
		pr.written++
	}

	return pr
}

// ── Pass 6: SSO seed ────────────────────────────────────────────────────

type ssoSeedUser struct {
	User        string   `json:"user"`
	RoleCodigos []string `json:"role_codigos,omitempty"`
}

type ssoSeedItem struct {
	ClientID string        `json:"client_id"`
	Usuarios []ssoSeedUser `json:"usuarios"`
}

type ssoSeedPayload struct {
	SyncKeycloak bool          `json:"sync_keycloak"`
	Items        []ssoSeedItem `json:"items"`
}

type ssoLoginResponse struct {
	Token string `json:"token"`
}

func ssoBaseURL(issuer string) string {
	if m := ssoBaseRe.FindStringSubmatch(issuer); m != nil {
		return m[1]
	}
	log.Printf("[import] SSO warn: cannot derive base URL from issuer %q, using as-is", issuer)
	return issuer
}

func doJSON(ctx context.Context, method, url, token string, reqBody, respOut interface{}) (int, []byte, error) {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read body: %w", err)
	}

	if respOut != nil {
		if err := json.Unmarshal(respData, respOut); err != nil {
			return resp.StatusCode, respData, fmt.Errorf("unmarshal: %w", err)
		}
	}

	return resp.StatusCode, respData, nil
}

// passSSOSeed registers all active employees in the SSO portal via the seed admin API.
// Best-effort: errors are logged but never abort the import.
func passSSOSeed(ctx context.Context, tgtDB *sql.DB, dryRun bool) passResult {
	pr := passResult{label: "sso_seed", reasons: make(reasons)}

	if dryRun {
		log.Println("[import] SSO seed skipped: dry run")
		pr.reasons.add("dry_run")
		return pr
	}

	adminUser := os.Getenv("SSO_SEED_ADMIN_USER")
	adminPass := os.Getenv("SSO_SEED_ADMIN_PASSWORD")
	if adminUser == "" || adminPass == "" {
		log.Println("[import] SSO seed skipped: credentials not configured (SSO_SEED_ADMIN_USER/SSO_SEED_ADMIN_PASSWORD)")
		pr.reasons.add("no_creds")
		return pr
	}

	skipEmp := os.Getenv("SEED_SSO_DEV_EMPLOYEE_NUMBER")

	baseURL := os.Getenv("SSO_SEED_BASE_URL")
	if baseURL == "" {
		issuer := os.Getenv("SSO_KC_ISSUER")
		baseURL = ssoBaseURL(issuer)
	}

	clientID := os.Getenv("SSO_CLIENT_ID")
	if clientID == "" {
		log.Println("[import] SSO seed skipped: SSO_CLIENT_ID not configured")
		pr.reasons.add("no_client_id")
		return pr
	}

	rows, err := tgtDB.QueryContext(ctx, `SELECT employee_number FROM employees WHERE is_active = true`)
	if err != nil {
		log.Printf("[import] SSO seed ERROR: query employees: %v", err)
		pr.reasons.add("query_error")
		return pr
	}
	defer rows.Close()

	var usuarios []ssoSeedUser
	for rows.Next() {
		var empNum string
		if err := rows.Scan(&empNum); err != nil {
			pr.reasons.add("scan_error")
			continue
		}
		if skipEmp != "" && strings.EqualFold(empNum, skipEmp) {
			pr.reasons.add("skipped")
			continue
		}
		usuarios = append(usuarios, ssoSeedUser{User: empNum, RoleCodigos: []string{"usuario"}})
	}
	if err := rows.Err(); err != nil {
		log.Printf("[import] SSO seed ERROR: rows iteration: %v", err)
		pr.reasons.add("rows_error")
		return pr
	}

	pr.sourceTotal = len(usuarios)
	if len(usuarios) == 0 {
		log.Println("[import] SSO seed skipped: no active employees to seed")
		pr.reasons.add("no_users")
		return pr
	}

	// Login to get admin token
	loginURL := baseURL + "/api/seed/login"
	loginBody := map[string]string{"user": adminUser, "password": adminPass}
	var loginResp ssoLoginResponse

	log.Printf("[import] SSO seed login → %s", loginURL)
	status, body, err := doJSON(ctx, http.MethodPost, loginURL, "", loginBody, &loginResp)
	if err != nil {
		log.Printf("[import] SSO seed login ERROR: %v (url=%s)", err, loginURL)
		pr.reasons.add("login_error")
		return pr
	}
	if status < 200 || status >= 300 {
		log.Printf("[import] SSO seed login ERROR: url=%s status=%d body=%s", loginURL, status, string(body))
		pr.reasons.add("login_http_error")
		return pr
	}

	// Register all employees — retry on "Usuario X no encontrado" by excluding that user
	seedURL := baseURL + "/api/seed/sistemas/usuarios"
	var discarded int

	for {
		seedBody := ssoSeedPayload{
			SyncKeycloak: true,
			Items: []ssoSeedItem{
				{ClientID: clientID, Usuarios: usuarios},
			},
		}

		log.Printf("[import] SSO seed → %s (%d users)", seedURL, len(usuarios))
		status, body, err := doJSON(ctx, http.MethodPost, seedURL, loginResp.Token, seedBody, nil)
		if err != nil {
			log.Printf("[import] SSO seed ERROR: %v (url=%s)", err, seedURL)
			pr.reasons.add("seed_error")
			return pr
		}
		if status >= 200 && status < 300 {
			log.Printf("[import] SSO seed OK: status=%d users=%d discarded=%d", status, len(usuarios), discarded)
			pr.written = len(usuarios)
			return pr
		}

		// Non-2xx — check if it's a "user not found" error we can recover from
		m := ssoUserNotFoundRe.FindStringSubmatch(string(body))
		if m == nil || status != 400 {
			log.Printf("[import] SSO seed ERROR: url=%s status=%d body=%s", seedURL, status, string(body))
			pr.reasons.add("seed_http_error")
			return pr
		}

		notFoundUser := m[1]
		found := false
		for i, u := range usuarios {
			if u.User == notFoundUser {
				usuarios = append(usuarios[:i], usuarios[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			log.Printf("[import] SSO seed ERROR: user %q not found in local list, url=%s status=%d body=%s",
				notFoundUser, seedURL, status, string(body))
			pr.reasons.add("seed_http_error")
			return pr
		}

		discarded++
		log.Printf("[import] SSO seed retry: removed user %s (not in SSO), retrying with %d users",
			notFoundUser, len(usuarios))

		if len(usuarios) == 0 {
			log.Println("[import] SSO seed: no users left after exclusions")
			pr.reasons.add("all_excluded")
			return pr
		}
	}
}

func shortCode(name string) string {
	// Use rune count to avoid splitting multi-byte UTF-8 characters
	runes := []rune(name)
	if len(runes) > 8 {
		return string(runes[:8])
	}
	return name
}
