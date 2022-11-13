package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsolartec/iam-meli/internal/core/domain"
)

func TestUsers_InvalidToken(t *testing.T) {
	expected := "El token de acceso no es válido"

	cases := []string{"", "token.cualquiera.meli", "token cualquiera"}
	for _, accessToken := range cases {
		serv, _ := newTestServer()

		urls := []struct {
			method string
			path   string
			body   io.Reader
		}{
			{method: "GET", path: "/api/users", body: nil},
			{method: "GET", path: "/api/users/1", body: nil},
			{method: "GET", path: "/api/users/superadmin", body: nil},
			{method: "DELETE", path: "/api/users/1", body: nil},
			{method: "DELETE", path: "/api/users/superadmin", body: nil},
			{method: "GET", path: "/api/users/1/permissions", body: nil},
			{method: "GET", path: "/api/users/superadmin/permissions", body: nil},
			{method: "PATCH", path: "/api/users/1/permissions/permission_test", body: nil},
			{method: "PATCH", path: "/api/users/superadmin/permissions/permission_test", body: nil},
			{method: "DELETE", path: "/api/users/1/permissions/permission_test", body: nil},
			{method: "DELETE", path: "/api/users/superadmin/permissions/permission_test", body: nil},
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

func TestUsers_NoAuthorized(t *testing.T) {
	expected := "¿Qué intentas hacer? No tienes permisos suficientes para hacer esta acción"

	urls := []struct {
		method string
		path   string
		body   io.Reader
	}{
		{method: "DELETE", path: "/api/users/1", body: nil},
		{method: "DELETE", path: "/api/users/superadmin", body: nil},
		{method: "PATCH", path: "/api/users/1/permissions/permission_test", body: nil},
		{method: "PATCH", path: "/api/users/superadmin/permissions/permission_test", body: nil},
		{method: "DELETE", path: "/api/users/1/permissions/permission_test", body: nil},
		{method: "DELETE", path: "/api/users/superadmin/permissions/permission_test", body: nil},
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

func TestDeleteUser_NotFound(t *testing.T) {
	expected := "El usuario no existe"

	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"delete_user"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find, "DELETE", nil, accessToken)
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

func TestDeleteUser_NoDeletable(t *testing.T) {
	expected := "No puedes eliminar el usuario con el que estás autenticado"

	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"delete_user"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		res, b := request(t, serv, "/api/users/"+find, "DELETE", nil, accessToken)
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

func TestDeleteUser_Deletable(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"delete_user"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM users WHERE id = $1;")).
			ExpectExec().WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))

		res, _ := request(t, serv, "/api/users/"+find, "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
		}
	}
}

func TestGetAllUsers_NoData(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users;")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}))

	res, _ := request(t, serv, "/api/users", "GET", nil, accessToken)
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("Expected %d, got: %d", http.StatusNoContent, res.StatusCode)
	}
}

func TestGetAllUsers_Success(t *testing.T) {
	serv, mock := newTestServer()

	accessToken := generateAccessToken(t, mock, []string{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users;")).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "created_at"}).
				AddRow(1, "superadmin", time.Now()).
				AddRow(2, "meli", time.Now()),
		)

	res, b := request(t, serv, "/api/users", "GET", nil, accessToken)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}

	var data domain.Map
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	users := data["users"].([]interface{})
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got: %d", len(users))
	}

	superadmin := users[0].(map[string]interface{})
	if superadmin["username"].(string) != "superadmin" {
		t.Errorf("Expected superadmin, got: %s", superadmin["username"].(string))
	}

	meli := users[1].(map[string]interface{})
	if meli["username"].(string) != "meli" {
		t.Errorf("Expected meli, got: %s", meli["username"].(string))
	}
}

func TestGetOneUser_NoData(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnError(noResultsError)

		res, _ := request(t, serv, "/api/users/"+find, "GET", nil, accessToken)
		if res.StatusCode != http.StatusNoContent {
			t.Errorf("Expected %d, got: %d", http.StatusNoContent, res.StatusCode)
		}
	}
}

func TestGetOneUser_Success(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		res, b := request(t, serv, "/api/users/"+find, "GET", nil, accessToken)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
		}

		var data domain.Map
		if err := json.Unmarshal(b, &data); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		superadmin := data["user"].(map[string]interface{})
		if superadmin["id"].(float64) != 1 {
			t.Errorf("Expected 1, got: %f", superadmin["id"].(float64))
		}

		if superadmin["username"].(string) != "superadmin" {
			t.Errorf("Expected superadmin, got: %s", superadmin["username"].(string))
		}
	}
}

func TestGetAllUserPermissions_UserNotFound(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions", "GET", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El usuario no existe"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestGetAllUserPermissions_NoData(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT
				up.id,
				up.user_id,
				up.permission_id,
				p.name as permission_name
			FROM user_permissions up
				INNER JOIN permissions p ON p.id = up.permission_id
			WHERE user_id = $1;
		`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission_id", "permission_name"}))

		res, _ := request(t, serv, "/api/users/"+find+"/permissions", "GET", nil, accessToken)
		if res.StatusCode != http.StatusNoContent {
			t.Errorf("Expected %d, got: %d", http.StatusNoContent, res.StatusCode)
		}
	}
}

func TestGetAllUserPermissions_Success(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT
				up.id,
				up.user_id,
				up.permission_id,
				p.name as permission_name
			FROM user_permissions up
				INNER JOIN permissions p ON p.id = up.permission_id
			WHERE user_id = $1;
		`)).
			WithArgs(1).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "user_id", "permission_id", "permission_name"}).
					AddRow(1, 1, 1, "permission_test").
					AddRow(2, 1, 2, "permission_test_1"),
			)

		res, b := request(t, serv, "/api/users/"+find+"/permissions", "GET", nil, accessToken)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
		}

		var data domain.Map
		if err := json.Unmarshal(b, &data); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		user_permissions := data["user_permissions"].([]interface{})
		if len(user_permissions) != 2 {
			t.Errorf("Expected 2 permissions, got: %d", len(user_permissions))
		}

		permission_test := user_permissions[0].(map[string]interface{})
		if permission_test["permission_name"].(string) != "permission_test" {
			t.Errorf("Expected permission_test, got: %s", permission_test["permission_name"].(string))
		}

		permission_test_1 := user_permissions[1].(map[string]interface{})
		if permission_test_1["permission_name"].(string) != "permission_test_1" {
			t.Errorf("Expected permission_test_1, got: %s", permission_test_1["permission_name"].(string))
		}
	}
}

func TestGrantUserPermission_UserNotFound(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"grant_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "PATCH", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El usuario no existe"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestGrantUserPermission_CurrentUser(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"grant_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "PATCH", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "No puedes otorgarte un permiso a ti mismo"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestGrantUserPermission_PermissionNotFound(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"grant_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "PATCH", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El permiso no existe"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestGrantUserPermission_AlreadyHasPermission(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"grant_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
					AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
			)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;")).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission_id"}).AddRow(1, 2, 1))

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "PATCH", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El usuario ya tiene el permiso asignado"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestGrantUserPermission_Success(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"grant_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
					AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
			)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;")).
			WithArgs(2, 1).
			WillReturnError(noResultsError)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) RETURNING id;")).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "PATCH", nil, accessToken)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
		}

		location := res.Header.Get("Location")
		if location != "/api/users/"+find+"/permissions/permission_test1" {
			t.Errorf("Expected /api/users/%s/permissions/permission_test1, got: %s", find, location)
		}

		var data domain.Map
		if err := json.Unmarshal(b, &data); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		user_permission := data["user_permission"].(map[string]interface{})
		if user_permission["id"].(float64) != 1 {
			t.Errorf("Expected 1, got: %f", user_permission["id"].(float64))
		}
	}
}

func TestRevokeUserPermission_UserNotFound(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"revoke_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El usuario no existe"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestRevokeUserPermission_CurrentUser(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 1}, {username: "superadmin"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"revoke_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "No puedes quitarte un permiso a ti mismo"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestRevokeUserPermission_PermissionNotFound(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"revoke_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El permiso no existe"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestRevokeUserPermission_NoHasPermission(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"revoke_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
					AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
			)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;")).
			WithArgs(2, 1).
			WillReturnError(noResultsError)

		res, b := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
		}

		var errorMessage domain.ErrorMessage
		if err := json.Unmarshal(b, &errorMessage); err != nil {
			t.Fatalf("Could not unmarshall response %v", err)
		}

		expected := "El usuario no tiene este permiso asignado"
		if errorMessage.Message != expected {
			t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
		}
	}
}

func TestRevokeUserPermission_Success(t *testing.T) {
	cases := []struct {
		ID       int
		username string
	}{{ID: 2}, {username: "meli"}}

	for _, td := range cases {
		serv, mock := newTestServer()

		accessToken := generateAccessToken(t, mock, []string{"revoke_permission"})

		var (
			query *sqlmock.ExpectedQuery
			find  string
		)

		if td.username != "" {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).WithArgs(td.username)
			find = td.username
		} else {
			query = mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE id = $1;")).WithArgs(td.ID)
			find = fmt.Sprint(td.ID)
		}

		query.WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(2, "meli", time.Now()))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;")).
			WithArgs("permission_test").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "description", "deletable", "editable", "created_at", "updated_at"}).
					AddRow(1, "permission_test", "Este es un permiso de prueba", false, false, time.Now(), time.Now()),
			)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;")).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "permission_id"}).AddRow(1, 2, 1))

		mock.ExpectPrepare(regexp.QuoteMeta("DELETE FROM user_permissions WHERE id = $1;")).
			ExpectExec().
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		res, _ := request(t, serv, "/api/users/"+find+"/permissions/permission_test", "DELETE", nil, accessToken)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
		}
	}
}
