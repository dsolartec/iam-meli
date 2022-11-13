package services

import (
	"net/http"

	"github.com/dsolartec/iam-meli/internal/database"
	"github.com/dsolartec/iam-meli/internal/repositories"
	"github.com/go-chi/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	permissions_repository := repositories.PermissionsRepository{
		Database: database.New(),
	}

	permissions := &PermissionsService{
		Repository: &permissions_repository,
	}

	users := &UsersService{
		Permissions: &permissions_repository,
		Repository: &repositories.UsersRepository{
			Database: database.New(),
		},
	}

	r.Mount("/permissions", permissions.Routes())
	r.Mount("/users", users.Routes())

	return r
}
