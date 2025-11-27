package models

type Status struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updated_at"`
}