package internal

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dsolartec/iam-meli/internal/core/services"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	server *http.Server
}

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile("./docs/index.html")
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(b)
}

func New(port string) (*Server, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", documentationHandler)
	r.Mount("/api", services.New())

	serv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server := Server{server: serv}

	return &server, nil
}

func (serv *Server) Close() error {
	return nil
}

func (serv *Server) Start() {
	log.Printf("Server running on http://localhost%s", serv.server.Addr)
	log.Fatal(serv.server.ListenAndServe())
}
