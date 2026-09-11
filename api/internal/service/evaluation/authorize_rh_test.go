package evaluation_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	orgrepo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var employeeColumns = []string{"id", "created_at", "updated_at", "first_name", "last_name", "email",
	"employee_number", "is_active", "org_node_id", "manager_id", "profile_id", "job_title"}

func TestAuthorizeRHEvaluationWrite(t *testing.T) {
	evalID := uuid.New()
	subordinateID := uuid.New()
	jefeID := uuid.New()
	otherID := uuid.New()

	ctxWith := func(role auth.Role, empID uuid.UUID) context.Context {
		sess := &auth.Session{ID: uuid.New(), EmployeeID: empID}
		return auth.WithSession(context.Background(), sess, role, uuid.New())
	}

	managerValue := func(id *uuid.UUID) interface{} {
		if id == nil {
			return nil
		}
		return id.String()
	}

	setup := func(t *testing.T, managerID *uuid.UUID, expectQuery bool) (*svc.EvaluationService, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		if expectQuery {
			mock.ExpectQuery("FROM employees WHERE id").
				WithArgs(subordinateID).
				WillReturnRows(sqlmock.NewRows(employeeColumns).AddRow(
					subordinateID, time.Now(), time.Now(), "Ana", "Perez", "ana@example.com",
					"E-001", true, uuid.New(), managerValue(managerID), uuid.New(), "vendedor",
				))
		}
		mockEval := &mockEvalRepo{
			row: &repo.EvaluationRow{ID: evalID, EmployeeID: subordinateID, State: "en_progreso"},
		}
		service := svc.NewEvaluationService(mockEval, nil, nil, &mockCycleChecker{phase: "cierre"}, nil, orgrepo.NewEmployeeRepo(nil, db), nil)
		return service, mock
	}

	t.Run("rh permitido", func(t *testing.T) {
		service, mock := setup(t, nil, false)
		require.NoError(t, service.AuthorizeRHEvaluationWrite(ctxWith(auth.RoleRH, uuid.New()), evalID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("jefe asignado permitido", func(t *testing.T) {
		service, mock := setup(t, &jefeID, true)
		require.NoError(t, service.AuthorizeRHEvaluationWrite(ctxWith(auth.RoleJefe, jefeID), evalID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("jefe no asignado 403", func(t *testing.T) {
		service, mock := setup(t, &otherID, true)
		err := service.AuthorizeRHEvaluationWrite(ctxWith(auth.RoleJefe, jefeID), evalID)
		require.Error(t, err)
		assert.Equal(t, 403, pkgerrors.HTTPStatus(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sin sesion 401", func(t *testing.T) {
		service, mock := setup(t, nil, false)
		err := service.AuthorizeRHEvaluationWrite(context.Background(), evalID)
		require.Error(t, err)
		assert.Equal(t, 401, pkgerrors.HTTPStatus(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
