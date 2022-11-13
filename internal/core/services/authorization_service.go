package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dsolartec/iam-meli/internal/core/domain"
	"github.com/dsolartec/iam-meli/internal/core/domain/dto"
	"github.com/dsolartec/iam-meli/internal/core/domain/interfaces"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/utils"
	"github.com/go-chi/chi"
)

type AuthorizationService struct {
	Users interfaces.UsersRepository
}

func (service *AuthorizationService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var data dto.LoginAndSignUpBody

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if data.Username == "" {
		domain.HTTPError(w, r, http.StatusBadRequest, "Debes ingresar el nombre de usuario")
		return
	}

	if data.Password == "" {
		domain.HTTPError(w, r, http.StatusBadRequest, "Debes ingresar la contraseña")
		return
	}

	ctx := r.Context()

	user, err := service.Users.GetByUsername(ctx, data.Username, true)
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if !user.IsPassword(data.Password) {
		domain.HTTPError(w, r, http.StatusBadRequest, "El nombre de usuario o la contraseña es incorrecta")
		return
	}

	claim := domain.Claim{ID: int(user.ID)}

	token, err := claim.GenerateToken(os.Getenv("JWT_KEY"))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	domain.JSON(w, r, http.StatusOK, domain.Map{"accessToken": token, "id": user.ID})
}

func (service *AuthorizationService) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var data dto.LoginAndSignUpBody

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

	_, err := service.Users.GetByUsername(ctx, data.Username, false)
	if err == nil {
		domain.HTTPError(w, r, http.StatusBadRequest, "El nombre de usuario ya está en uso")
		return
	}

	if err = utils.ValidatePassword(data.Password); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := models.User{
		Username: data.Username,
		Password: data.Password,
	}

	if err := service.Users.Create(ctx, &user); err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	data.Password = ""

	claim := domain.Claim{ID: int(user.ID)}

	token, err := claim.GenerateToken(os.Getenv("JWT_KEY"))
	if err != nil {
		domain.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Add("Location", fmt.Sprintf("%s%d", r.URL.String(), user.ID))

	domain.JSON(w, r, http.StatusOK, domain.Map{"accessToken": token, "id": user.ID})
}

func (service *AuthorizationService) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/login", service.LoginHandler)
	r.Post("/signup", service.SignUpHandler)

	return r
}
