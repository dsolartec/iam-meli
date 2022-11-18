package utils

import (
	"errors"
	"regexp"
)

func ValidatePermissionDescription(description string) error {
	if description == "" {
		return errors.New("Debes ingresar la descripción del permiso")
	}

	if len(description) < 10 || len(description) > 150 {
		return errors.New("La descripción del permiso debe tener entre 10 y 150 caracteres")
	}

	return nil
}

func ValidatePermissionName(name string) error {
	if name == "" {
		return errors.New("Debes ingresar el nombre del permiso")
	}

	nameMatches, err := regexp.MatchString("^[a-zA-Z0-9_]*$", name)
	if err != nil {
		return err
	}

	if !nameMatches {
		return errors.New("El nombre del permiso no puede contener espacios o caracteres especiales")
	}

	if len(name) < 4 || len(name) > 25 {
		return errors.New("El nombre del permiso debe tener entre 4 y 25 caracteres")
	}

	return nil
}
