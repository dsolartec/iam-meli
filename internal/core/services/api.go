package services

import (
	"net/http"

	"github.com/dsolartec/iam-meli/pkg/interfaces"
	"github.com/go-chi/chi"
)

func New(
	auth_repository interfaces.AuthorizationRepository,
	permissions_repository interfaces.PermissionsRepository,
	users_repository interfaces.UsersRepository,
) http.Handler {
	r := chi.NewRouter()

	authorization := AuthorizationService{
		Users: users_repository,
	}

	permissions := PermissionsService{
		Auth:        auth_repository,
		Permissions: permissions_repository,
	}

	users := UsersService{
		Auth:        auth_repository,
		Permissions: permissions_repository,
		Users:       users_repository,
	}

	r.Mount("/auth", authorization.Routes())
	r.Mount("/permissions", permissions.Routes())
	r.Mount("/users", users.Routes())

	return r
}
