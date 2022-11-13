package dto

type CreatePermissionBody struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdatePermissionBody struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
