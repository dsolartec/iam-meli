package internal

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dsolartec/iam-meli/internal/core/services"
	"github.com/dsolartec/iam-meli/internal/database"
	"github.com/dsolartec/iam-meli/internal/repositories"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	server *http.Server

	router http.Handler
}

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile("./docs/index.html")
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(b)
}

func New(db *database.Database, port string) *Server {
	// Iniciamos los repositorios.
	auth_repository := repositories.AuthorizationRepository{
		Database: db,
	}

	permissions_repository := repositories.PermissionsRepository{
		Database: db,
	}

	users_repository := repositories.UsersRepository{
		Database: db,
	}

	// Enrutador
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", documentationHandler)
	r.Mount("/api", services.New(&auth_repository, &permissions_repository, &users_repository))

	// Servidor
	serv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server := Server{server: serv, router: r}

	return &server
}

func (serv *Server) Start() {
	log.Printf("Server running on http://localhost%s", serv.server.Addr)
	log.Fatal(serv.server.ListenAndServe())
}

func (serv *Server) Router() http.Handler {
	return serv.router
}

func (serv *Server) Close() error {
	return nil
}
