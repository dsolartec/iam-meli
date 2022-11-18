package middlewares

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/dsolartec/iam-meli/pkg"
)

func parseTokenFromAuthorization(authorization string) (string, error) {
	parts := strings.Split(authorization, " ")
	if authorization == "" || !strings.HasPrefix(authorization, "Bearer") || len(parts) != 2 {
		return "", errors.New("El token de acceso no es válido")
	}

	return parts[1], nil
}

func Authorizator(next http.Handler) http.Handler {
	jwtKey := os.Getenv("JWT_KEY")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		token, err := parseTokenFromAuthorization(authorization)
		if err != nil {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		claim, err := pkg.ParseToken(token, jwtKey)
		if err != nil {
			pkg.HTTPError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "current_user_id", claim.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
