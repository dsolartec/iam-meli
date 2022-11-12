package main

import (
	"log"
	"os"
	"os/signal"

	server "github.com/dsolartec/iam-meli/internal"
	Database "github.com/dsolartec/iam-meli/internal/database"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	port := os.Getenv("PORT")

	serv, err := server.New(port)
	if err != nil {
		log.Fatal(err)
	}

	// Database connection.
	db := Database.New()
	if err := db.Conn.Ping(); err != nil {
		log.Fatal(err)
	}

	// Start the server.
	go serv.Start()

	// Wait for an interrupt.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Shutdown.
	serv.Close()
	db.Close()
}
