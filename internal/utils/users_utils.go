package utils

import (
	"errors"
	"regexp"
)

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("Debes ingresar la contraseña")
	}

	passwordMatches, err := regexp.MatchString("^[a-zA-Z0-9_]*$", password)
	if err != nil {
		return err
	}

	if !passwordMatches {
		return errors.New("La contraseña no puede contener espacios o caracteres especiales")
	}

	if len(password) < 4 || len(password) > 15 {
		return errors.New("La contraseña debe tener entre 4 y 15 caracteres")
	}

	return nil
}

func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("Debes ingresar el nombre de usuario")
	}

	usernameMatches, err := regexp.MatchString("^[a-zA-Z0-9_]*$", username)
	if err != nil {
		return err
	}

	if !usernameMatches {
		return errors.New("El nombre de usuario no puede contener espacios o caracteres especiales")
	}

	if len(username) < 4 || len(username) > 10 {
		return errors.New("El nombre de usuario debe tener entre 4 y 10 caracteres")
	}

	return nil
}
