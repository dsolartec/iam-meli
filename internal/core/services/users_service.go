package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/core/domain/interfaces"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/utils"
	"github.com/go-chi/chi"
)

type UsersService struct {
	Repository  interfaces.UsersRepository
	Permissions interfaces.PermissionsRepository
}

func (service *UsersService) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var data models.User

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	ctx := r.Context()

	if err := utils.ValidateUsername(data.Username); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	_, err := service.Repository.GetByUsername(ctx, data.Username)
	if err == nil {
		domain.HTTPError(w, r, http.StatusBadRequest, "El nombre de usuario ya está en uso")
		return
	}

	if err = utils.ValidatePassword(data.Password); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.Repository.Create(ctx, &data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	data.Password = ""

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), data.ID))

	domain.JSON(w, r, http.StatusCreated, domain.Map{"user": data})
}

func (service *UsersService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	_, err = service.Repository.GetByID(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = service.Repository.Delete(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *UsersService) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := service.Repository.GetAll(ctx)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"users": users})
}

func (service *UsersService) GetOneHandler(w http.ResponseWriter, r *http.Request) {
	find := chi.URLParam(r, "find")

	ctx := r.Context()

	user := models.User{}

	id, err := strconv.Atoi(find)
	if err != nil {
		user, err = service.Repository.GetByUsername(ctx, find)
	} else {
		user, err = service.Repository.GetByID(ctx, uint(id))
	}

	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"user": user})
}

func (service *UsersService) GetAllUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	user_permissions, err := service.Repository.GetAllUserPermissions(ctx, uint(userID))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"user_permissions": user_permissions})
}

func (service *UsersService) GrantPermissionHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	permissionName := chi.URLParam(r, "permission_name")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	_, err = service.Repository.GetByID(ctx, uint(userID))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission, err := service.Permissions.GetByName(ctx, permissionName)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	data := models.UserPermission{
		UserID:       uint(userID),
		PermissionID: permission.ID,
	}

	_, err = service.Repository.GetUserPermission(ctx, data.UserID, data.PermissionID)
	if err == nil {
		domain.HTTPError(w, r, http.StatusBadRequest, "El usuario ya tiene el permiso asignado")
		return
	}

	if err = service.Repository.GrantPermission(ctx, &data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), data.ID))
	domain.JSON(w, r, http.StatusCreated, domain.Map{"user_permission": data})
}

func (service *UsersService) RevokePermissionHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	permissionName := chi.URLParam(r, "permission_name")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	_, err = service.Repository.GetByID(ctx, uint(userID))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission, err := service.Permissions.GetByName(ctx, permissionName)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user_permission, err := service.Repository.GetUserPermission(ctx, uint(userID), permission.ID)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = service.Repository.RevokePermission(ctx, user_permission.ID)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *UsersService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", service.GetAllHandler)
	r.Post("/", service.CreateHandler)

	r.Get("/{find}", service.GetOneHandler)
	r.Delete("/{id}", service.DeleteHandler)

	r.Get("/{user_id}/permissions", service.GetAllUserPermissionsHandler)
	r.Patch("/{user_id}/permissions/{permission_name}", service.GrantPermissionHandler)
	r.Delete("/{user_id}/permissions/{permission_name}", service.RevokePermissionHandler)

	return r
}
