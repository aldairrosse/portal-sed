package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Service sincroniza is_active de employees contra Mobonet.
type Service struct {
	db     *sql.DB
	client *http.Client
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// fetchMobonet lista employee_number activos en Mobonet.
// Si MOBONET_URL y MOBONET_KEY vacíos, retorna mock (activos actuales en BD) — idempotente.
func (s *Service) fetchMobonet(ctx context.Context) ([]string, error) {
	url := strings.TrimSpace(os.Getenv("MOBONET_URL"))
	key := strings.TrimSpace(os.Getenv("MOBONET_KEY"))
	if url == "" && key == "" {
		// Mock: lista de activos actuales en BD (usa Status=1 como en import/main.go)
		rows, err := s.db.QueryContext(ctx, `SELECT employee_number FROM employees WHERE is_active = true`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				continue
			}
			if n != "" {
				out = append(out, n)
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

	// Intenta parsear múltiples formatos: []string, []{employee_number}, []{NumEmpleado}
	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	// 1) []string
	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err == nil && len(asStrings) > 0 {
		return asStrings, nil
	}
	// 2) []map
	var asMaps []map[string]interface{}
	if err := json.Unmarshal(raw, &asMaps); err == nil {
		var out []string
		for _, m := range asMaps {
			for _, k := range []string{"employee_number", "NumEmpleado", "num_empleado", "employeeNumber", "numero_empleado"} {
				if v, ok := m[k]; ok {
					if s, ok := v.(string); ok && s != "" {
						out = append(out, s)
						break
					}
				}
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	// 3) wrapper {data: [...]} o {employees: [...]}
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err == nil {
		for _, wk := range []string{"data", "employees", "result"} {
			if v, ok := wrapper[wk]; ok {
				var ws []string
				if err := json.Unmarshal(v, &ws); err == nil {
					return ws, nil
				}
				var wm []map[string]interface{}
				if err := json.Unmarshal(v, &wm); err == nil {
					var out []string
					for _, m := range wm {
						for _, k := range []string{"employee_number", "NumEmpleado"} {
							if val, ok := m[k]; ok {
								if sv, ok := val.(string); ok && sv != "" {
									out = append(out, sv)
									break
								}
							}
						}
					}
					return out, nil
				}
			}
		}
	}
	slog.Warn("mobonet_sync: respuesta no reconocida, retorna vacío")
	return []string{}, nil
}

// Run sincroniza is_active. Retorna cantidad deshabilitados. Idempotente.
func (s *Service) Run(ctx context.Context) (disabledCount int, err error) {
	active, err := s.fetchMobonet(ctx)
	if err != nil {
		return 0, err
	}

	// Deduplicar y limpiar
	seen := make(map[string]struct{}, len(active))
	var uniq []string
	for _, n := range active {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		uniq = append(uniq, n)
	}
	active = uniq

	if len(active) == 0 {
		slog.Warn("mobonet_sync: lista de activos vacía, no se deshabilita nadie (idempotente)")
		return 0, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(active))
	args := make([]interface{}, len(active))
	for i, n := range active {
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
		// deshabilitados ya contados, retornar parcial
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

	return len(disabled), nil
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
