package services

import (
	"net/http"

	"github.com/dsolartec/iam-meli/internal/database"
	"github.com/dsolartec/iam-meli/internal/repositories"
	"github.com/go-chi/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	users := &UsersService{
		Repository: &repositories.UsersRepository{
			Database: database.New(),
		},
	}

	permissions := &PermissionsService{
		Repository: &repositories.PermissionsRepository{
			Database: database.New(),
		},
	}

	r.Mount("/users", users.Routes())
	r.Mount("/permissions", permissions.Routes())

	return r
}
