package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsolartec/iam-meli/internal/core/middlewares"
	"github.com/dsolartec/iam-meli/pkg"
	"github.com/dsolartec/iam-meli/pkg/dto"
	"github.com/dsolartec/iam-meli/pkg/interfaces"
	"github.com/dsolartec/iam-meli/pkg/models"
	"github.com/dsolartec/iam-meli/pkg/utils"
	"github.com/go-chi/chi"
)

type PermissionsService struct {
	Auth        interfaces.AuthorizationRepository
	Permissions interfaces.PermissionsRepository
}

func (service *PermissionsService) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "create_permission"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var data dto.CreatePermissionBody

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err := utils.ValidatePermissionName(data.Name); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	_, err := service.Permissions.GetByName(ctx, data.Name)
	if err == nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, "El nombre del permiso ya está en uso")
		return
	}

	if err = utils.ValidatePermissionDescription(data.Description); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission := models.Permission{
		Name:        data.Name,
		Description: data.Description,
		Deletable:   true,
		Editable:    true,
	}

	if err = service.Permissions.Create(ctx, &permission); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), permission.ID))

	pkg.JSON(w, r, http.StatusCreated, pkg.Map{"permission": permission})
}

func (service *PermissionsService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "delete_permission"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	if !permission.Deletable {
		pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no puede ser borrado")
		return
	}

	err = service.Permissions.Delete(ctx, uint(id))
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{})
}

func (service *PermissionsService) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	permissions, err := service.Permissions.GetAll(ctx)
	if err != nil {
		pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		return
	}

	if permissions == nil || len(permissions) == 0 {
		pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{"permissions": permissions})
}

func (service *PermissionsService) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.JSON(w, r, http.StatusNoContent, pkg.Map{})
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{"permission": permission})
}

func (service *PermissionsService) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := service.Auth.VerifyPermission(ctx, "update_permission"); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	permission, err := service.Permissions.GetByID(ctx, uint(id))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no existe")
		} else {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		}

		return
	}

	if !permission.Editable {
		pkg.HTTPError(w, r, http.StatusBadRequest, "El permiso no puede ser editado")
		return
	}

	var data dto.UpdatePermissionBody

	if err = json.NewDecoder(r.Body).Decode(&data); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if data.Name == "" {
		data.Name = permission.Name
	}

	if data.Name != permission.Name {
		if err = utils.ValidatePermissionName(data.Name); err != nil {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		_, err = service.Permissions.GetByName(ctx, data.Name)
		if err == nil {
			pkg.HTTPError(w, r, http.StatusBadRequest, "El nombre del permiso ya está en uso")
			return
		}
	}

	if data.Description == "" {
		data.Description = permission.Description
	}

	if err = utils.ValidatePermissionDescription(data.Description); err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = service.Permissions.Update(ctx, uint(id), &data)
	if err != nil {
		pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	pkg.JSON(w, r, http.StatusOK, pkg.Map{})
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
