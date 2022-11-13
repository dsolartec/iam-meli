package interfaces

import "context"

type AuthorizationRepository interface {
	VerifyPermission(ctx context.Context, permissionName string) error
}
