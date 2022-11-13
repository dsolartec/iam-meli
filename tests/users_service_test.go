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
