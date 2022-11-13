package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/core/domain/dto"
	"github.com/dsolartec/iam-meli/internal/core/domain/interfaces"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/core/middlewares"
	"github.com/dsolartec/iam-meli/internal/utils"
	"github.com/go-chi/chi"
)

type PermissionsService struct {
	Auth        interfaces.AuthorizationRepository
	Permissions interfaces.PermissionsRepository
}

func (service *PermissionsService) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "create_permission"); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var data dto.CreatePermissionBody

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err := utils.ValidatePermissionName(data.Name); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	_, err := service.Permissions.GetByName(ctx, data.Name)
	if err == nil {
		domain.HTTPError(w, r, http.StatusBadRequest, "El nombre del permiso ya está en uso")
		return
	}

	if err = utils.ValidatePermissionDescription(data.Description); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission := models.Permission{
		Name:        data.Name,
		Description: data.Description,
		Deletable:   true,
		Editable:    true,
	}

	if err = service.Permissions.Create(ctx, &permission); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), permission.ID))

	domain.JSON(w, r, http.StatusCreated, domain.Map{"permission": permission})
}

func (service *PermissionsService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "delete_permission"); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if !permission.Deletable {
		domain.HTTPError(w, r, http.StatusBadRequest, "El permiso no puede ser borrado")
		return
	}

	err = service.Permissions.Delete(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *PermissionsService) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	permissions, err := service.Permissions.GetAll(ctx)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"permissions": permissions})
}

func (service *PermissionsService) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"permission": permission})
}

func (service *PermissionsService) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "update_permission"); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var data dto.UpdatePermissionBody

	if err = json.NewDecoder(r.Body).Decode(&data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if !permission.Editable {
		domain.HTTPError(w, r, http.StatusBadRequest, "El permiso no puede ser editado")
		return
	}

	if data.Name == "" {
		data.Name = permission.Name
	}

	if data.Name != permission.Name {
		if err = utils.ValidatePermissionName(data.Name); err != nil {
			domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		_, err = service.Permissions.GetByName(ctx, data.Name)
		if err == nil {
			domain.HTTPError(w, r, http.StatusBadRequest, "El nombre del permiso ya está en uso")
			return
		}
	}

	if data.Description == "" {
		data.Description = permission.Description
	}

	if err = utils.ValidatePermissionDescription(data.Description); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = service.Permissions.Update(ctx, uint(id), &data)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *PermissionsService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares.Authorizator)

	r.Get("/", service.GetAllHandler)
	r.Post("/", service.CreateHandler)

	r.Get("/{id}", service.GetByIDHandler)
	r.Put("/{id}", service.UpdateHandler)
	r.Delete("/{id}", service.DeleteHandler)

	return r
}
