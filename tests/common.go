package tests

import (
	"database/sql/driver"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsolartec/iam-meli/internal"
	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/database"
)

type anyPassword struct{}

func (a anyPassword) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok && strings.HasPrefix(v.(string), "$2a$10$")
}

var noResultsError = errors.New("sql: no rows in result set")

func newDatabaseMock() (*database.Database, sqlmock.Sqlmock) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	database := database.Database{
		Conn: conn,
	}

	return &database, mock
}

func newTestServer() (*internal.Server, sqlmock.Sqlmock) {
	db, mock := newDatabaseMock()

	return internal.New(db, "80"), mock
}

func request(t *testing.T, serv *internal.Server, path string, method string, body io.Reader, accessToken string) (*http.Response, []byte) {
	// Generamos la consulta
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("could not created request: %v", err)
	}

	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	rec := httptest.NewRecorder()

	serv.Router().ServeHTTP(rec, req)

	// Obtenemos el resultado de la consulta.
	res := rec.Result()

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Could not read response: %v", err)
	}

	return res, b
}

func generateAccessToken(t *testing.T, mock sqlmock.Sqlmock, permission_names []string) string {
	for permission_id, permission_name := range permission_names {
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT p.id FROM permissions p
				INNER JOIN user_permissions up ON up.user_id = $1 AND up.permission_id = p.id
				WHERE name = $2
		`)).
			WithArgs(1, permission_name).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(permission_id + 1))
	}

	// Generamos el token para el usuario con ID 1
	claim := domain.Claim{ID: 1}

	token, err := claim.GenerateToken("MeLiTest")
	if err != nil {
		t.Fatalf("Could not generate access token %v", err)
	}

	return token
}
