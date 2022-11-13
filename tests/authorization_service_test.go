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

func TestLogin_EmptyUsername(t *testing.T) {
	serv, _ := newTestServer()

	body := []byte(`{
		"username":"",
		"password":""
	}`)

	res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "Debes ingresar el nombre de usuario"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	serv, _ := newTestServer()

	body := []byte(`{
		"username":"superadmin",
		"password":""
	}`)

	res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "Debes ingresar la contraseña"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestLogin_UserNotExists(t *testing.T) {
	serv, mock := newTestServer()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, password, created_at FROM users WHERE username = $1;")).
		WithArgs("superadmin").WillReturnError(noResultsError)

	body := []byte(`{
		"username":"superadmin",
		"password":"superadmin"
	}`)

	res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre de usuario o la contraseña es incorrecta"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestLogin_IncorrectPassword(t *testing.T) {
	serv, mock := newTestServer()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT id, username, password, created_at FROM users WHERE username = $1;"),
	).WithArgs("superadmin").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
				AddRow(1, "superadmin", "superadmin", time.Now()),
		)

	body := []byte(`{
		"username":"superadmin",
		"password":"superadmin"
	}`)

	res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre de usuario o la contraseña es incorrecta"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestLogin_AccessToken(t *testing.T) {
	if err := os.Setenv("JWT_KEY", "MeLiTest"); err != nil {
		t.Fatalf("Coult not set `JWT_KEY` environment variable %v", err)
	}

	serv, mock := newTestServer()

	user := models.User{Password: "12345"}
	if err := user.EncryptPassword(); err != nil {
		t.Fatalf("Could not encrypt password %v", err)
	}

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT id, username, password, created_at FROM users WHERE username = $1;"),
	).WithArgs("superadmin").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "password", "created_at"}).
				AddRow(1, "superadmin", user.Password, time.Now()),
		)

	body := []byte(`{
		"username":"superadmin",
		"password":"12345"
	}`)

	res, b := request(t, serv, "/api/auth/login", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}

	var data domain.Map
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	accessToken := data["accessToken"].(string)
	id := data["id"].(float64)

	claim, err := domain.ParseToken(accessToken, "MeLiTest")
	if err != nil {
		t.Fatalf("Could not parse access token %v", err)
	}

	if claim.ID != int(id) {
		t.Errorf("Expected %f, got %d", id, claim.ID)
	}

	if id != 1 {
		t.Errorf("Expected 1, got %f", id)
	}
}

func TestSignUp_ValidationErrors(t *testing.T) {
	cases := []struct {
		username string
		password string
		expected string
	}{
		{
			username: "",
			password: "",
			expected: "Debes ingresar el nombre de usuario",
		},
		{
			username: "super admin",
			password: "",
			expected: "El nombre de usuario no puede contener espacios o caracteres especiales",
		},
		{
			username: "super_admin$%",
			password: "",
			expected: "El nombre de usuario no puede contener espacios o caracteres especiales",
		},
		{
			username: "&super",
			password: "",
			expected: "El nombre de usuario no puede contener espacios o caracteres especiales",
		},
		{
			username: "sup",
			password: "",
			expected: "El nombre de usuario debe tener entre 4 y 10 caracteres",
		},
		{
			username: "super_admin_super_admin_super_admin",
			password: "",
			expected: "El nombre de usuario debe tener entre 4 y 10 caracteres",
		},
		{
			username: "superadmin",
			password: "",
			expected: "Debes ingresar la contraseña",
		},
		{
			username: "superadmin",
			password: "super admin",
			expected: "La contraseña no puede contener espacios o caracteres especiales",
		},
		{
			username: "superadmin",
			password: "super_admin$%",
			expected: "La contraseña no puede contener espacios o caracteres especiales",
		},
		{
			username: "superadmin",
			password: "&super",
			expected: "La contraseña no puede contener espacios o caracteres especiales",
		},
		{
			username: "superadmin",
			password: "123",
			expected: "La contraseña debe tener entre 4 y 15 caracteres",
		},
		{
			username: "superadmin",
			password: "super_admin_super_admin_super_admin",
			expected: "La contraseña debe tener entre 4 y 15 caracteres",
		},
	}

	for _, td := range cases {
		serv, _ := newTestServer()

		body := []byte(`{
			"username":"` + td.username + `",
			"password":"` + td.password + `"
		}`)

		res, b := request(t, serv, "/api/auth/signup", "POST", bytes.NewBuffer(body), "")
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

func TestSignUp_DuplicatedUsername(t *testing.T) {
	serv, mock := newTestServer()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).
		WithArgs("superadmin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at"}).AddRow(1, "superadmin", time.Now()))

	body := []byte(`{
		"username":"superadmin",
		"password":"superadmin"
	}`)

	res, b := request(t, serv, "/api/auth/signup", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre de usuario ya está en uso"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, got: %s", expected, errorMessage.Message)
	}
}

func TestSignUp_AccessToken(t *testing.T) {
	if err := os.Setenv("JWT_KEY", "MeLiTest"); err != nil {
		t.Fatalf("Coult not set `JWT_KEY` environment variable %v", err)
	}

	serv, mock := newTestServer()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, created_at FROM users WHERE username = $1;")).
		WithArgs("superadmin").WillReturnError(noResultsError)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id;")).
		WithArgs("superadmin", anyPassword{}).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body := []byte(`{
		"username":"superadmin",
		"password":"superadmin"
	}`)

	res, b := request(t, serv, "/api/auth/signup", "POST", bytes.NewBuffer(body), "")
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got: %d", http.StatusOK, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "/api/auth/signup1" {
		t.Errorf("Expected /api/auth/signup1, got %s", location)
	}

	var data domain.Map
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	accessToken := data["accessToken"].(string)
	id := data["id"].(float64)

	claim, err := domain.ParseToken(accessToken, "MeLiTest")
	if err != nil {
		t.Fatalf("Could not parse access token %v", err)
	}

	if claim.ID != int(id) {
		t.Errorf("Expected %f, got %d", id, claim.ID)
	}

	if id != 1 {
		t.Errorf("Expected 1, got %f", id)
	}
}
