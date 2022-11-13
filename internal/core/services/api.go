package services

import (
	"net/http"

	"github.com/dsolartec/iam-meli/internal/database"
	"github.com/dsolartec/iam-meli/internal/repositories"
	"github.com/go-chi/chi"
)

func New() http.Handler {
	r := chi.NewRouter()

	auth_repository := repositories.AuthorizationRepository{
		Database: database.New(),
	}

	permissions_repository := repositories.PermissionsRepository{
		Database: database.New(),
	}

	users_repository := repositories.UsersRepository{
		Database: database.New(),
	}

	authorization := AuthorizationService{
		Users: &users_repository,
	}

	permissions := PermissionsService{
		Auth:        &auth_repository,
		Permissions: &permissions_repository,
	}

	users := UsersService{
		Auth:        &auth_repository,
		Permissions: &permissions_repository,
		Users:       &users_repository,
	}

	r.Mount("/auth", authorization.Routes())
	r.Mount("/permissions", permissions.Routes())
	r.Mount("/users", users.Routes())

	return r
}
