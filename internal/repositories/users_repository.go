package repositories

import (
	"context"
	"time"

	"github.com/dsolartec/iam-meli/internal/database"
	"github.com/dsolartec/iam-meli/pkg/models"
)

type UsersRepository struct {
	Database *database.Database
}

func (repository *UsersRepository) Create(ctx context.Context, data *models.User) error {
	query := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id;"

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
	query := "SELECT id, username, created_at FROM users;"

	rows, err := repository.Database.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err = rows.Scan(&user.ID, &user.Username, &user.CreatedAt)
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
		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) GetByUsername(ctx context.Context, username string, with_password bool) (models.User, error) {
	query := "SELECT id, username, created_at FROM users WHERE username = $1;"
	if with_password {
		query = "SELECT id, username, password, created_at FROM users WHERE username = $1;"
	}

	row := repository.Database.Conn.QueryRowContext(ctx, query, username)

	var err error
	var user models.User

	if with_password {
		err = row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	} else {
		err = row.Scan(&user.ID, &user.Username, &user.CreatedAt)
	}

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (repository *UsersRepository) GetAllUserPermissions(ctx context.Context, userID uint) ([]models.UserPermission, error) {
	query := `
		SELECT
			up.id,
			up.user_id,
			up.permission_id,
			p.name as permission_name
		FROM user_permissions up
			INNER JOIN permissions p ON p.id = up.permission_id
		WHERE user_id = $1;
	`

	rows, err := repository.Database.Conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var user_permissions []models.UserPermission
	for rows.Next() {
		var user_permission models.UserPermission

		err = rows.Scan(&user_permission.ID, &user_permission.UserID, &user_permission.PermissionID, &user_permission.PermissionName)
		if err != nil {
			return nil, err
		}

		user_permissions = append(user_permissions, user_permission)
	}

	return user_permissions, nil
}

func (repository *UsersRepository) GetUserPermission(ctx context.Context, userID uint, permissionID uint) (models.UserPermission, error) {
	query := "SELECT id, user_id, permission_id FROM user_permissions WHERE user_id = $1 AND permission_id = $2;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, userID, permissionID)

	var user_permission models.UserPermission

	if err := row.Scan(&user_permission.ID, &user_permission.UserID, &user_permission.PermissionID); err != nil {
		return models.UserPermission{}, err
	}

	return user_permission, nil
}

func (repository *UsersRepository) GrantPermission(ctx context.Context, data *models.UserPermission) error {
	query := "INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) RETURNING id;"

	row := repository.Database.Conn.QueryRowContext(ctx, query, data.UserID, data.PermissionID)

	return row.Scan(&data.ID)
}

func (repository *UsersRepository) RevokePermission(ctx context.Context, id uint) error {
	query := "DELETE FROM user_permissions WHERE id = $1;"

	stmt, err := repository.Database.Conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}
