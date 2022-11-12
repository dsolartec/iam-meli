package database

import (
	"database/sql"
	"io/ioutil"
	"os"

	_ "github.com/lib/pq"
)

func getConnection() (*sql.DB, error) {
	uri := os.Getenv("DATABASE_URI")
	return sql.Open("postgres", uri)
}

func RunMigration(db *sql.DB, name string) error {
	b, err := ioutil.ReadFile("./internal/database/migrations/" + name + ".sql")
	if err != nil {
		return err
	}

	rows, err := db.Query(string(b))
	if err != nil {
		return err
	}

	return rows.Close()
}
