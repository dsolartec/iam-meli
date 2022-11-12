package repositories

import (
	"context"
	"time"

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
	query := "DELETE FROM permissions WHERE id = $1;"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

func (repository *PermissionsRepository) GetAll(ctx context.Context) ([]models.Permission, error) {
	query := "SELECT id, name, description, deletable, created_at, updated_at FROM permissions;"

	rows, err := repository.Database.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission

		err = rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Deletable, &permission.CreatedAt, &permission.UpdatedAt)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (repository *PermissionsRepository) GetByID(ctx context.Context, id uint) (models.Permission, error) {
	query := "SELECT id, name, description, deletable, created_at, updated_at FROM permissions WHERE id = $1;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, id)

	var permission models.Permission

	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Deletable, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		return models.Permission{}, err
	}

	return permission, nil
}

func (repository *PermissionsRepository) Update(ctx context.Context, id uint, data *models.Permission) error {
	query := "UPDATE permissions SET name = $1, description = $2, updated_at = $3 WHERE id = $4"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	data.UpdatedAt = time.Now()

	_, err = stmt.ExecContext(ctx, data.Name, data.Description, data.UpdatedAt, id)
	return err
}
