package pkg

import (
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
)

type Claim struct {
	jwt.StandardClaims
	ID int `json:"id"`
}

func (claim *Claim) GenerateToken(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (*Claim, error) {
	if tokenString == "" {
		return nil, errors.New("El token de acceso no es válido")
	}

	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("El token de acceso no es válido")
	}

	if !token.Valid {
		return nil, errors.New("El token de acceso no es válido")
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("El token de acceso no es válido")
	}

	iID, ok := claim["id"]
	if !ok {
		return nil, errors.New("El token de acceso no es válido")
	}

	id, ok := iID.(float64)
	if !ok {
		return nil, errors.New("El token de acceso no es válido")
	}

	return &Claim{ID: int(id)}, nil
}
