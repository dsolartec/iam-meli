package models

import "time"

type Permission struct {
	ID          uint      `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Deletable   uint8     `json:"deletable"`
	Editable    uint8     `json:"editable"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
