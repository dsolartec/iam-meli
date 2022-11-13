package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
)

func TestEndToEnd(t *testing.T) {
	if err := os.Setenv("JWT_KEY", "MeLiTest"); err != nil {
		t.Fatalf("Coult not set `JWT_KEY` environment variable %v", err)
	}

	// Registro de nuevo usuario
	t.Run("Registro del usuario meli", func(t *testing.T) {
		serv, mock := newTestServer()

		body := []byte(`{"username":"meli", "password":"meli"}`)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).
			WithArgs("meli").
			WillReturnError(noResultsError)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id;")).
			WithArgs("meli", anyPassword{}).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

		res, b := request(t, serv, "/api/auth/signup", "POST", bytes.NewBuffer(body), "")
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d - %s", http.StatusOK, res.StatusCode, b)
		}

		var data domain.Map
		if err := json.Unmarshal(b, &data); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		accessToken := data["accessToken"].(string)

		t.Run("Error al intentar crear un permiso", func(t *testing.T) {
			body := []byte(`{"name": "permission_test", "description": "Este es un permiso de prueba"}`)

			mock.ExpectQuery("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;").
				WithArgs("permission_test").
				WillReturnError(noResultsError)

			mock.ExpectQuery("INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id;").
				WithArgs("permission_test", "Este es un permiso de prueba").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			res, b := request(t, serv, "/api/permissions", "POST", bytes.NewBuffer(body), accessToken)
			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
			}

			var errorMessage domain.ErrorMessage
			if err := json.Unmarshal(b, &errorMessage); err != nil {
				t.Fatalf("Could not unmarshall response %v", err)
			}

			expected := "¿Qué intentas hacer? No tienes permisos suficientes para hacer esta acción"
			if errorMessage.Message != expected {
				t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
			}
		})
	})

	// Inicio de sesión como superadmin
	t.Run("Inicio de sesión como superadmin", func(t *testing.T) {
		serv, mock := newTestServer()

		user := models.User{Password: "superadmin"}
		if err := user.EncryptPassword(); err != nil {
			t.Fatalf("Could not encrypt password %v", err)
		}

		body := []byte(`{"username":"superadmin","password":"superadmin"}`)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, password, created_at FROM users WHERE username = $1;")).
			WithArgs("superadmin").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
					AddRow(1, "superadmin", user.Password, time.Now()),
			)

		res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d - %s", http.StatusOK, res.StatusCode, b)
		}

		var data domain.Map
		if err := json.Unmarshal(b, &data); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		accessToken := data["accessToken"].(string)

		t.Run("Creamos el permiso permission_test con el usuario superadmin", func(t *testing.T) {
			body := []byte(`{
				"name": "permission_test",
				"description": "Este es un permiso de prueba"
			}`)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT p.id FROM permissions p
					INNER JOIN user_permissions up ON up.user_id = $1 AND up.permission_id = p.id
					WHERE name = $2
			`)).
				WithArgs(1, "create_permission").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

			mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
				WithArgs("permission_test").
				WillReturnError(noResultsError)

			mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id;")).
				WithArgs("permission_test", "Este es un permiso de prueba").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

			res, _ := request(t, serv, "/api/permissions", "POST", bytes.NewBuffer(body), accessToken)
			if res.StatusCode != http.StatusCreated {
				t.Errorf("Expected %d, got: %d", http.StatusCreated, res.StatusCode)
			}

			location := res.Header.Get("Location")
			if location != "/api/permissions7" {
				t.Errorf("Expected /api/permissions7, got %s", location)
			}

			t.Run("Asignamos el permiso al usuario meli", func(t *testing.T) {
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT p.id FROM permissions p
						INNER JOIN user_permissions up ON up.user_id = $1 AND up.permission_id = p.id
						WHERE name = $2
				`)).
					WithArgs(1, "grant_permission").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
					WithArgs("permission_test").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
							AddRow(7, "permission_test", "Este es un permiso de prueba", true, true, time.Now(), time.Now()),
					)

				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;")).
					WithArgs(2, 7).
					WillReturnError(noResultsError)

				mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) RETURNING id;")).
					WithArgs(2, 7).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
			})

			res, b := request(t, serv, "/api/users/2/permissions/permission_test", "PATCH", nil, accessToken)
			if res.StatusCode != http.StatusOK {
				t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
			}

			location = res.Header.Get("Location")
			if location != "/api/users/2/permissions/permission_test7" {
				t.Errorf("Expected /api/users/2/permissions/permission_test7, got: %s", location)
			}

			var data domain.Map
			if err := json.Unmarshal(b, &data); err != nil {
				t.Fatalf("Could not unmarshall response %v", err)
			}

			user_permission := data["user_permission"].(map[string]interface{})
			if user_permission["id"].(float64) != 7 {
				t.Errorf("Expected 7, got: %f", user_permission["id"].(float64))
			}
		})
	})
}
