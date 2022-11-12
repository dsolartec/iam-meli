package database

import (
	"database/sql"
	"log"
	"sync"
)

type Database struct {
	Conn *sql.DB
}

func (db *Database) Close() error {
	return db.Conn.Close()
}

var (
	db   *Database
	once sync.Once
)

func initDatabase() {
	conn, err := getConnection()
	if err != nil {
		log.Panic(err)
	}

	if err := RunMigration(conn, "INITIAL_DATA"); err != nil {
		log.Panic(err)
	}

	db = &Database{
		Conn: conn,
	}
}

func New() *Database {
	once.Do(initDatabase)

	return db
}
