package interfaces

import (
	"context"

	"github.com/dsolartec/iam-meli/pkg/dto"
	"github.com/dsolartec/iam-meli/pkg/models"
)

type PermissionsRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]models.Permission, error)
	GetByID(ctx context.Context, id uint) (models.Permission, error)
	GetByName(ctx context.Context, name string) (models.Permission, error)
	Update(ctx context.Context, id uint, permission *dto.UpdatePermissionBody) error
}
