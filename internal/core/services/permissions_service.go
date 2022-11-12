package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/core/domain/interfaces"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/go-chi/chi"
)

type PermissionsService struct {
	Repository interfaces.PermissionsRepository
}

func (service *PermissionsService) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var permission models.Permission

	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	ctx := r.Context()

	if err := service.Repository.Create(ctx, &permission); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), permission.ID))

	domain.JSON(w, r, http.StatusCreated, domain.Map{"permission": permission})
}

func (service *PermissionsService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	err = service.Repository.Delete(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *PermissionsService) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	permissions, err := service.Repository.GetAll(ctx)
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

	permission, err := service.Repository.GetByID(ctx, uint(id))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"permission": permission})
}

func (service *PermissionsService) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	var data models.Permission

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	ctx := r.Context()

	err = service.Repository.Update(ctx, uint(id), &data)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{})
}

func (service *PermissionsService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", service.GetAllHandler)
	r.Post("/", service.CreateHandler)

	r.Get("/{id}", service.GetByIDHandler)
	r.Put("/{id}", service.UpdateHandler)
	r.Delete("/{id}", service.DeleteHandler)

	return r
}
