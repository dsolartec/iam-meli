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
	if os.Getenv("JWT_KEY") == "" {
		log.Fatal("Se debe iniciar la variable `JWT_KEY`.")
	}

	// Database connection.
	db := Database.New()
	if err := db.Conn.Ping(); err != nil {
		log.Fatal(err)
	}

	// Start the server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	serv := server.New(db, port)
	go serv.Start()

	// Wait for an interrupt.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Shutdown.
	serv.Close()
	db.Close()
}
