package org

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
)

// EmployeeRow is a read model for employee queries.
type EmployeeRow struct {
	ID             uuid.UUID  `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email"`
	EmployeeNumber string     `json:"employee_number"`
	IsActive       bool       `json:"is_active"`
	OrgNodeID      uuid.UUID  `json:"org_node_id"`
	ManagerID      *uuid.UUID `json:"manager_id,omitempty"`
	ProfileID      uuid.UUID  `json:"profile_id"`
}

// EmployeeDetailRow extends EmployeeRow with nested org node and manager info.
type EmployeeDetailRow struct {
	EmployeeRow
	OrgNodeName string `json:"org_node_name"`
	OrgNodePath string `json:"org_node_path"`
	ManagerName string `json:"manager_name,omitempty"`
}

// EmployeeFilter specifies the filtering criteria for listing employees.
type EmployeeFilter struct {
	TreeID    *uuid.UUID
	NodeID    *uuid.UUID
	ProfileID *uuid.UUID
	IsActive  *bool
	Query     string
	Cursor    string
	Limit     int
}

// EmployeeRepo provides database queries for employees.
type EmployeeRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewEmployeeRepo creates a new EmployeeRepo.
func NewEmployeeRepo(client *internal.Client, db *sql.DB) *EmployeeRepo {
	return &EmployeeRepo{client: client, db: db}
}

// List returns employees matching the given filter with cursor-based pagination.
// Returns limit+1 items so caller can determine hasMore.
func (r *EmployeeRepo) List(ctx context.Context, filter EmployeeFilter) ([]*EmployeeRow, error) {
	query := `SELECT e.id, e.created_at, e.updated_at, e.first_name, e.last_name, e.email,
	                 e.employee_number, e.is_active, e.org_node_id, e.manager_id, e.profile_id
	           FROM employees e`
	var conditions []string
	var joins []string
	args := []interface{}{}
	idx := 1

	if filter.TreeID != nil {
		joins = append(joins, `JOIN org_nodes on2 ON e.org_node_id = on2.id`)
		conditions = append(conditions, `on2.organization_id = $`+itoa(idx))
		args = append(args, *filter.TreeID)
		idx++
	}
	if filter.NodeID != nil {
		conditions = append(conditions, `e.org_node_id = $`+itoa(idx))
		args = append(args, *filter.NodeID)
		idx++
	}
	if filter.ProfileID != nil {
		conditions = append(conditions, `e.profile_id = $`+itoa(idx))
		args = append(args, *filter.ProfileID)
		idx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, `e.is_active = $`+itoa(idx))
		args = append(args, *filter.IsActive)
		idx++
	}
	if filter.Query != "" {
		conditions = append(conditions, `(e.first_name ILIKE $`+itoa(idx)+
			` OR e.last_name ILIKE $`+itoa(idx)+
			` OR e.email ILIKE $`+itoa(idx)+
			` OR e.employee_number ILIKE $`+itoa(idx)+`)`)
		args = append(args, "%"+filter.Query+"%")
		idx++
	}
	if filter.Cursor != "" {
		cursorID, err := uuid.Parse(filter.Cursor)
		if err == nil {
			conditions = append(conditions, `e.id > $`+itoa(idx))
			args = append(args, cursorID)
			idx++
		}
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	fullQuery := query + " " + strings.Join(joins, " ")
	if len(conditions) > 0 {
		fullQuery += ` WHERE ` + strings.Join(conditions, " AND ")
	}
	fullQuery += ` ORDER BY e.last_name, e.first_name, e.id LIMIT $` + itoa(idx)
	args = append(args, limit+1)

	return scanEmployeeRows(r.db, ctx, fullQuery, args...)
}

// GetByID retrieves a single employee by ID.
func (r *EmployeeRepo) GetByID(ctx context.Context, empID uuid.UUID) (*EmployeeRow, error) {
	return scanEmployeeRow(r.db.QueryRowContext(ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, email,
		        employee_number, is_active, org_node_id, manager_id, profile_id
		 FROM employees WHERE id = $1`, empID))
}

// GetDetailByID retrieves an employee with nested org node and manager info.
func (r *EmployeeRepo) GetDetailByID(ctx context.Context, empID uuid.UUID) (*EmployeeDetailRow, error) {
	detail := &EmployeeDetailRow{}
	var managerID sql.NullString
	var managerName sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT e.id, e.created_at, e.updated_at, e.first_name, e.last_name, e.email,
		        e.employee_number, e.is_active, e.org_node_id, e.manager_id, e.profile_id,
		        COALESCE(on2.name, '') as org_node_name,
		        COALESCE(on2.path::text, '') as org_node_path,
		        COALESCE(m.first_name || ' ' || m.last_name, '') as manager_name
		 FROM employees e
		 LEFT JOIN org_nodes on2 ON e.org_node_id = on2.id
		 LEFT JOIN employees m ON e.manager_id = m.id
		 WHERE e.id = $1`, empID,
	).Scan(
		&detail.ID, &detail.CreatedAt, &detail.UpdatedAt,
		&detail.FirstName, &detail.LastName, &detail.Email,
		&detail.EmployeeNumber, &detail.IsActive,
		&detail.OrgNodeID, &managerID, &detail.ProfileID,
		&detail.OrgNodeName, &detail.OrgNodePath, &managerName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	if managerID.Valid {
		mid, _ := uuid.Parse(managerID.String)
		detail.ManagerID = &mid
	}
	if managerName.Valid {
		detail.ManagerName = managerName.String
	}

	return detail, nil
}

// GetByIDs performs a batch lookup for up to 100 employee IDs.
// Returns employees in the order of the input IDs.
func (r *EmployeeRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*EmployeeRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 100 {
		ids = ids[:100]
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + itoa(i+1)
		args[i] = id
	}

	return scanEmployeeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, email,
		        employee_number, is_active, org_node_id, manager_id, profile_id
		 FROM employees WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
}

// ListByManager returns employees managed by the given manager.
// If activeOnly is true, only returns active employees.
func (r *EmployeeRepo) ListByManager(ctx context.Context, managerID uuid.UUID, activeOnly bool) ([]*EmployeeRow, error) {
	query := `SELECT id, created_at, updated_at, first_name, last_name, email,
	                 employee_number, is_active, org_node_id, manager_id, profile_id
	           FROM employees WHERE manager_id = $1`
	args := []interface{}{managerID}

	if activeOnly {
		query += ` AND is_active = true`
	}

	query += ` ORDER BY last_name, first_name`
	return scanEmployeeRows(r.db, ctx, query, args...)
}

// Search performs a full-text search on employees using ILIKE.
// Results are limited to the specified limit (max 50).
func (r *EmployeeRepo) Search(ctx context.Context, query string, limit int) ([]*EmployeeRow, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	searchTerm := "%" + query + "%"
	return scanEmployeeRows(r.db, ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, email,
		        employee_number, is_active, org_node_id, manager_id, profile_id
		 FROM employees
		 WHERE first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1 OR employee_number ILIKE $1
		 ORDER BY last_name, first_name
		 LIMIT $2`, searchTerm, limit)
}

// GetManager returns the manager of an employee.
func (r *EmployeeRepo) GetManager(ctx context.Context, empID uuid.UUID) (*EmployeeRow, error) {
	var managerID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(manager_id, '00000000-0000-0000-0000-000000000000') FROM employees WHERE id = $1`,
		empID,
	).Scan(&managerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	if managerID == uuid.Nil {
		return nil, nil // no manager
	}
	return r.GetByID(ctx, managerID)
}

// ---------- helpers ----------

func scanEmployeeRow(row *sql.Row) (*EmployeeRow, error) {
	e := &EmployeeRow{}
	var managerID sql.NullString
	err := row.Scan(
		&e.ID, &e.CreatedAt, &e.UpdatedAt,
		&e.FirstName, &e.LastName, &e.Email,
		&e.EmployeeNumber, &e.IsActive,
		&e.OrgNodeID, &managerID, &e.ProfileID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	if managerID.Valid {
		mid, _ := uuid.Parse(managerID.String)
		e.ManagerID = &mid
	}
	return e, nil
}

func scanEmployeeRows(db *sql.DB, ctx context.Context, query string, args ...interface{}) ([]*EmployeeRow, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*EmployeeRow
	for rows.Next() {
		e := &EmployeeRow{}
		var managerID sql.NullString
		err := rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt,
			&e.FirstName, &e.LastName, &e.Email,
			&e.EmployeeNumber, &e.IsActive,
			&e.OrgNodeID, &managerID, &e.ProfileID,
		)
		if err != nil {
			return nil, err
		}
		if managerID.Valid {
			mid, _ := uuid.Parse(managerID.String)
			e.ManagerID = &mid
		}
		results = append(results, e)
	}
	return results, rows.Err()
}
