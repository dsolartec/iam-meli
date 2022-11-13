package tests

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsolartec/iam-meli/internal"
	"github.com/dsolartec/iam-meli/internal/database"
)

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

func postRequest(t *testing.T, serv *internal.Server, path string, body []byte) (*http.Response, []byte) {
	// Generamos la consulta
	req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("could not created request: %v", err)
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
