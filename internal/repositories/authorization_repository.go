package repositories

import (
	"context"
	"errors"

	"github.com/dsolartec/iam-meli/internal/core/domain/models"
	"github.com/dsolartec/iam-meli/internal/database"
)

type AuthorizationRepository struct {
	Database *database.Database
}

func (repository *AuthorizationRepository) VerifyPermission(ctx context.Context, permissionName string) error {
	userID, ok := ctx.Value("current_user_id").(int)
	if !ok {
		return errors.New("¿Qué intentas hacer? No tienes permisos suficientes para hacer esta acción")
	}

	query := `
		SELECT p.id FROM permissions p
			INNER JOIN user_permissions up ON up.user_id = $1 AND up.permission_id = p.id
			WHERE name = $2
	`

	row := repository.Database.Conn.QueryRowContext(ctx, query, userID, permissionName)

	permission := models.Permission{}

	if err := row.Scan(&permission.ID); err != nil {
		return errors.New("¿Qué intentas hacer? No tienes permisos suficientes para hacer esta acción")
	}

	return nil
}
