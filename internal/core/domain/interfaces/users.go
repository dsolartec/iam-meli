package interfaces

import (
	"context"

	"github.com/dsolartec/iam-meli/internal/core/domain/models"
)

type UsersRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]models.User, error)
	GetByID(ctx context.Context, id uint) (models.User, error)
	GetByUsername(ctx context.Context, username string) (models.User, error)

	GetAllUserPermissions(ctx context.Context, userID uint) ([]models.UserPermission, error)
	GetUserPermission(ctx context.Context, userID uint, permissionID uint) (models.UserPermission, error)
	GrantPermission(ctx context.Context, data *models.UserPermission) error
	RevokePermission(ctx context.Context, id uint) error
}
