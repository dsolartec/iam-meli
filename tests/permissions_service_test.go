package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsolartec/iam-meli/internal/core/domain"
)

func TestPermissions_InvalidToken(t *testing.T) {
	expected := "El token de acceso no es válido"

	cases := []string{"", "token.cualquiera.meli", "token cualquiera"}
	for _, accessToken := range cases {
		serv, _ := newTestServer()

		urls := []struct {
			method string
			path   string
			body   io.Reader
		}{
			{method: "GET", path: "/api/permissions", body: nil},
			{method: "POST", path: "/api/permissions", body: nil},
			{method: "GET", path: "/api/permissions/1", body: nil},
			{method: "PUT", path: "/api/permissions/1", body: nil},
			{method: "DELETE", path: "/api/permissions/1", body: nil},
		}

		for _, url := range urls {
			res, b := request(t, serv, url.path, url.method, url.body, accessToken)
			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
			}

			var errorMessage domain.ErrorMessage
			if err := json.Unmarshal(b, &errorMessage); err != nil {
				t.Fatalf("Could not unmarshall response %v", err)
			}

			if errorMessage.Message != expected {
				t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
			}
		}
	}
}

func TestPermissions_NoAuthorized(t *testing.T) {
	expected := "¿Qué intentas hacer? No tienes permisos suficientes para hacer esta acción"

	urls := []struct {
		method string
		path   string
		body   io.Reader
	}{
		{method: "POST", path: "/api/permissions", body: nil},
		{method: "PUT", path: "/api/permissions/1", body: nil},
		{method: "DELETE", path: "/api/permissions/1", body: nil},
	}

	for _, url := range urls {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		res, b := request(t, serv, url.path, url.method, url.body, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestCreatePermission_ValidationErrors(t *testing.T) {
	cases := []struct {
		permission_name        string
		permission_description string
		expected               string
	}{
		{
			expected: "Debes ingresar el nombre del permiso",
		},
		{
			permission_name: "nombre permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "nombre$permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "%nombre_permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "nom",
			expected:        "El nombre del permiso debe tener entre 4 y 25 caracteres",
		},
		{
			permission_name: "nombre_permiso_nombre_permison_nombre_permiso",
			expected:        "El nombre del permiso debe tener entre 4 y 25 caracteres",
		},
		{
			permission_name: "permission_test",
			expected:        "Debes ingresar la descripción del permiso",
		},
		{
			permission_name:        "permission_test",
			permission_description: "desc",
			expected:               "La descripción del permiso debe tener entre 10 y 150 caracteres",
		},
		{
			permission_name:        "permission_test",
			permission_description: "Esta es una descripción demasiadoooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo largaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:               "La descripción del permiso debe tener entre 10 y 150 caracteres",
		},
	}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"create_permission"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs(td.permission_name).
			WillReturnError(noResultsError)

		body := []byte(`{
			"name": "` + td.permission_name + `",
			"description": "` + td.permission_description + `"
		}`)

		res, b := request(t, serv, "/api/permissions", "POST", bytes.NewBuffer(body), accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		if errorMessage.Message != td.expected {
			t.Errorf("Expected %s, got: %s", td.expected, errorMessage.Message)
		}
	}
}

func TestCreatePermission_DuplicatedPermissionName(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"create_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
		WithArgs("permission_test").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
		)

	body := []byte(`{
		"name": "permission_test",
		"description": "Este es un permiso de prueba"
	}`)

	res, b := request(t, serv, "/api/permissions", "POST", bytes.NewBuffer(body), accessToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre del permiso ya está en uso"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestCreatePermission_Success(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"create_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
		WithArgs("permission_test").
		WillReturnError(noResultsError)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id;")).
		WithArgs("permission_test", "Este es un permiso de prueba").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body := []byte(`{
		"name": "permission_test",
		"description": "Este es un permiso de prueba"
	}`)

	res, _ := request(t, serv, "/api/permissions", "POST", bytes.NewBuffer(body), accessToken)
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected %d, got: %d", http.StatusCreated, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "/api/permissions1" {
		t.Errorf("Expected /api/permissions1, got %s", location)
	}
}

func TestDeletePermission_NoDeletable(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"delete_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
		)

	res, b := request(t, serv, "/api/permissions/1", "DELETE", nil, accessToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El permiso no puede ser borrado"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestDeletePermission_Deletable(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"delete_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", true, true, time.Now(), time.Now()),
		)

	mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM permissions WHERE id = $1 AND deletable = TRUE;")).
		ExpectExec().WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	res, _ := request(t, serv, "/api/permissions/1", "DELETE", nil, accessToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}
}

func TestGetAllPermissions_Success(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions;")).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", true, false, time.Now(), time.Now()).
				AddRow(2, "permission_test_1", "Este es un permiso de prueba", false, true, time.Now(), time.Now()),
		)

	res, b := request(t, serv, "/api/permissions", "GET", nil, accessToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}

	var data domain.Map
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	permissions := data["permissions"].([]interface{})
	if len(permissions) != 2 {
		t.Errorf("Expected 2 permissions, got: %d", len(permissions))
	}

	permission_test := permissions[0].(map[string]interface{})
	if permission_test["name"].(string) != "permission_test" {
		t.Errorf("Expected permission_test, got: %s", permission_test["name"].(string))
	}

	if !permission_test["deletable"].(bool) {
		t.Errorf("Expected deletable true, got: %v", permission_test["deletable"].(bool))
	}

	if permission_test["editable"].(bool) {
		t.Errorf("Expected editable false, got: %v", permission_test["editable"].(bool))
	}

	permission_test_1 := permissions[1].(map[string]interface{})
	if permission_test_1["name"].(string) != "permission_test_1" {
		t.Errorf("Expected permission_test_1, got: %s", permission_test_1["name"].(string))
	}

	if permission_test_1["deletable"].(bool) {
		t.Errorf("Expected deletable false, got: %v", permission_test_1["deletable"].(bool))
	}

	if !permission_test_1["editable"].(bool) {
		t.Errorf("Expected editable true, got: %v", permission_test_1["editable"].(bool))
	}
}

func TestGetPermissionByID_Success(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
		)

	res, b := request(t, serv, "/api/permissions/1", "GET", nil, accessToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var data domain.Map
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	permission := data["permission"].(map[string]interface{})
	if permission["name"].(string) != "permission_test" {
		t.Errorf("Expected permission_test, got: %s", permission["name"].(string))
	}

	if permission["deletable"].(bool) {
		t.Errorf("Expected deletable false, got: %v", permission["deletable"].(bool))
	}

	if permission["editable"].(bool) {
		t.Errorf("Expected editable false, got: %v", permission["editable"].(bool))
	}
}

func TestUpdatePermission_ValidationErrors(t *testing.T) {
	cases := []struct {
		permission_name        string
		permission_description string
		expected               string
	}{
		{
			permission_name: "nombre permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "nombre$permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "%nombre_permiso",
			expected:        "El nombre del permiso no puede contener espacios o caracteres especiales",
		},
		{
			permission_name: "nom",
			expected:        "El nombre del permiso debe tener entre 4 y 25 caracteres",
		},
		{
			permission_name: "nombre_permiso_nombre_permison_nombre_permiso",
			expected:        "El nombre del permiso debe tener entre 4 y 25 caracteres",
		},
		{
			permission_description: "desc",
			expected:               "La descripción del permiso debe tener entre 10 y 150 caracteres",
		},
		{
			permission_description: "Esta es una descripción demasiadoooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo largaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:               "La descripción del permiso debe tener entre 10 y 150 caracteres",
		},
	}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"update_permission"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
			WithArgs(1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
					AddRow(1, "permission_test", "Este es un permiso de prueba", false, true, time.Now(), time.Now()),
			)

		body := []byte(`{
			"name": "` + td.permission_name + `",
			"description": "` + td.permission_description + `"
		}`)

		res, b := request(t, serv, "/api/permissions/1", "PUT", bytes.NewBuffer(body), accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		if errorMessage.Message != td.expected {
			t.Errorf("Expected %s, got: %s", td.expected, errorMessage.Message)
		}
	}
}

func TestUpdatePermission_DuplicatePermissionName(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"update_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, true, time.Now(), time.Now()),
		)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
		WithArgs("permission_test_1").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(2, "permission_test_1", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
		)

	body := []byte(`{"name": "permission_test_1"}`)

	res, b := request(t, serv, "/api/permissions/1", "PUT", bytes.NewBuffer(body), accessToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre del permiso ya está en uso"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestUpdatePermission_NoEditable(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"update_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
		)

	res, b := request(t, serv, "/api/permissions/1", "PUT", nil, accessToken)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El permiso no puede ser editado"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestUpdatePermission_Editable(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{"update_permission"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
				AddRow(1, "permission_test", "Este es un permiso de prueba", false, true, time.Now(), time.Now()),
		)

	mock.ExpectPrepare(regexp.QuoteMeta("UPDATE permissions SET name = $1, description = $2, updated_at = $3 WHERE id = $4 AND editable = TRUE;")).
		ExpectExec().
		WithArgs("permission_test_1", "Este es un permiso de prueba editado", anyTime{}, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := []byte(`{
		"name": "permission_test_1",
		"description": "Este es un permiso de prueba editado"
	}`)

	res, _ := request(t, serv, "/api/permissions/1", "PUT", bytes.NewBuffer(body), accessToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}
}
