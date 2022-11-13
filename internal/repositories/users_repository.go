package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/database"
)

type UsersRepository struct {
	Database *database.Database
}

func (repository *UsersRepository) Create(ctx context.Context, data *models.User) error {
	query := "INSERT INTO users (id, username, password) VALUES (DEFAULT, $1, $2) RETURNING id;"

	if err := data.EncryptPassword(); err != nil {
		return err
	}

	data.CreatedAt = time.Now()

	row := repository.Database.Conn.QueryRowContext(ctx, query, data.Username, data.Password)

	return row.Scan(&data.ID)
}

func (repository *UsersRepository) Delete(ctx context.Context, id uint) error {
	query := "DELETE FROM users WHERE id = $1;"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)

	return err
}

func (repository *UsersRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := "SELECT id, username, password, created_at FROM users;"

	rows, err := repository.Database.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (repository *UsersRepository) GetByID(ctx context.Context, id uint) (models.User, error) {
	query := "SELECT id, username, created_at FROM users WHERE id = $1;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, id)

	var user models.User

	if err := row.Scan(&user.ID, &user.Username, &user.CreatedAt); err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = errors.New("El usuario no existe")
		}

		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) GetByUsername(ctx context.Context, username string) (models.User, error) {
	query := "SELECT id, username, created_at FROM users WHERE username = $1;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, username)

	var user models.User

	if err := row.Scan(&user.ID, &user.Username, &user.CreatedAt); err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = errors.New("El usuario no existe")
		}

		return models.User{}, err
	}

	return user, nil
}
