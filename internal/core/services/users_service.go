package services

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsolartec/iam-meli/internal/core/middlewares"
	"github.com/dsolartec/iam-meli/pkg"
	"github.com/dsolartec/iam-meli/pkg/interfaces"
	"github.com/dsolartec/iam-meli/pkg/models"
	"github.com/go-chi/chi"
)

type UsersService struct {
	Auth        interfaces.AuthorizationRepository
	Users       interfaces.UsersRepository
	Permissions interfaces.PermissionsRepository
}

func (service *UsersService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "delete_user"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	find := chi.URLParam(r, "find")

	user := models.User{}

	id, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Users.GetByUsername(ctx, find, false)
	} else {
		user, err = service.Users.GetByID(ctx, uint(id))
	}

	if currentUserID := ctx.Value("current_user_id").(int); err == nil && currentUserID == int(user.ID) {
		err = errors.New("No puedes eliminar el usuario con el que estás autenticado")
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	err = service.Users.Delete(ctx, user.ID)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{})
}

func (service *UsersService) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := service.Users.GetAll(ctx)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if users == nil || len(users) == 0 {
		pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{"users": users})
}

func (service *UsersService) GetOneHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	find := chi.URLParam(r, "find")

	user := models.User{}

	id, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Users.GetByUsername(ctx, find, false)
	} else {
		user, err = service.Users.GetByID(ctx, uint(id))
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{"user": user})
}

func (service *UsersService) GetAllUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	find := chi.URLParam(r, "find")

	user := models.User{}

	userID, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Users.GetByUsername(ctx, find, false)
	} else {
		user, err = service.Users.GetByID(ctx, uint(userID))
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	user_permissions, err := service.Users.GetAllUserPermissions(ctx, user.ID)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if user_permissions == nil || len(user_permissions) == 0 {
		pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{"user_permissions": user_permissions})
}

func (service *UsersService) GrantPermissionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "grant_permission"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	find := chi.URLParam(r, "find")
	permissionName := chi.URLParam(r, "permission_name")

	user := models.User{}

	userID, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Users.GetByUsername(ctx, find, false)
	} else {
		user, err = service.Users.GetByID(ctx, uint(userID))
	}

	if currentUserID := ctx.Value("current_user_id").(int); err == nil && currentUserID == int(user.ID) {
		err = errors.New("No puedes otorgarte un permiso a ti mismo")
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	permission, err := service.Permissions.GetByName(ctx, permissionName)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	data := models.UserPermission{
		UserID:       user.ID,
		PermissionID: permission.ID,
	}

	_, err = service.Users.GetUserPermission(ctx, data.UserID, data.PermissionID)
	if err == nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario ya tiene el permiso asignado")
		return
	}

	if err = service.Users.GrantPermission(ctx, &data); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), data.ID))
	pkg.JSON(w, r, http.StatusOK, pkg.Map{"user_permission": data})
}

func (service *UsersService) RevokePermissionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "revoke_permission"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	find := chi.URLParam(r, "find")
	permissionName := chi.URLParam(r, "permission_name")

	user := models.User{}

	userID, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Users.GetByUsername(ctx, find, false)
	} else {
		user, err = service.Users.GetByID(ctx, uint(userID))
	}

	if currentUserID := ctx.Value("current_user_id").(int); err == nil && currentUserID == int(user.ID) {
		err = errors.New("No puedes quitarte un permiso a ti mismo")
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	permission, err := service.Permissions.GetByName(ctx, permissionName)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	user_permission, err := service.Users.GetUserPermission(ctx, user.ID, permission.ID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El usuario no tiene este permiso asignado")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	err = service.Users.RevokePermission(ctx, user_permission.ID)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{})
}

func (service *UsersService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.Authorizator)

	r.Get("/", service.GetAllHandler)

	r.Get("/{find}", service.GetOneHandler)
	r.Delete("/{find}", service.DeleteHandler)

	r.Get("/{find}/permissions", service.GetAllUserPermissionsHandler)
	r.Patch("/{find}/permissions/{permission_name}", service.GrantPermissionHandler)
	r.Delete("/{find}/permissions/{permission_name}", service.RevokePermissionHandler)

	return r
}
