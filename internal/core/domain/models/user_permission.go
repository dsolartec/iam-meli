package models

type UserPermission struct {
	ID             uint   `json:"id,omitempty"`
	UserID         uint   `json:"user_id,omitempty"`
	PermissionID   uint   `json:"permission_id,omitempty"`
	PermissionName string `json:"permission_name,omitempty"`
}
