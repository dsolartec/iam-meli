package tests

import (
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

	res, b := postRequest(t, serv, "/api/auth/login", body)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "Debes ingresar el nombre de usuario"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, go: %s", expected, errorMessage.Message)
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	serv, _ := newTestServer()

	body := []byte(`{
		"username":"superadmin",
		"password":""
	}`)

	res, b := postRequest(t, serv, "/api/auth/login", body)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "Debes ingresar la contraseña"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, go: %s", expected, errorMessage.Message)
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

	res, b := postRequest(t, serv, "/api/auth/login", body)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got: %d", http.StatusBadRequest, res.StatusCode)
	}

	var errorMessage domain.ErrorMessage
	if err := json.Unmarshal(b, &errorMessage); err != nil {
		t.Fatalf("Could not unmarshall response %v", err)
	}

	expected := "El nombre de usuario o la contraseña es incorrecta"
	if errorMessage.Message != expected {
		t.Errorf("Expected %s, go: %s", expected, errorMessage.Message)
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

	res, b := postRequest(t, serv, "/api/auth/login", body)
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
