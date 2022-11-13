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
	Repository interfaces.UsersRepository
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

func (service *UsersService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", service.GetAllHandler)
	r.Post("/", service.CreateHandler)

	r.Get("/{find}", service.GetOneHandler)
	r.Delete("/{id}", service.DeleteHandler)

	return r
}
