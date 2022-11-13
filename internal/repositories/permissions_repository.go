package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/dsolartec/iam-meli/internal/core/domain/dto"
	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/database"
)

type PermissionsRepository struct {
	Database *database.Database
}

func (repository *PermissionsRepository) Create(ctx context.Context, data *models.Permission) error {
	query := "INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id;"

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	row := repository.Database.Conn.QueryRowContext(ctx, query, data.Name, data.Description)

	return row.Scan(&data.ID)
}

func (repository *PermissionsRepository) Delete(ctx context.Context, id uint) error {
	query := "DELETE FROM permissions WHERE id = $1 AND deletable = TRUE;"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

func (repository *PermissionsRepository) GetAll(ctx context.Context) ([]models.Permission, error) {
	query := "SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions;"

	rows, err := repository.Database.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission

		err = rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Deletable, &permission.Editable, &permission.CreatedAt, &permission.UpdatedAt)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (repository *PermissionsRepository) GetByID(ctx context.Context, id uint) (models.Permission, error) {
	query := "SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE id = $1;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, id)

	var permission models.Permission

	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Deletable, &permission.Editable, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		return models.Permission{}, err
	}

	return permission, nil
}

func (repository *PermissionsRepository) GetByName(ctx context.Context, name string) (models.Permission, error) {
	query := "SELECT id, name, description, deletable, editable, created_at, updated_at FROM permissions WHERE name = $1;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, name)

	var permission models.Permission

	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Deletable, &permission.Editable, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = errors.New("El permiso no existe")
		}

		return models.Permission{}, err
	}

	return permission, nil
}

func (repository *PermissionsRepository) Update(ctx context.Context, id uint, data *dto.UpdatePermissionBody) error {
	query := "UPDATE permissions SET name = $1, description = $2, updated_at = $3 WHERE id = $4 AND editable = TRUE;"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, data.Name, data.Description, time.Now(), id)
	return err
}
