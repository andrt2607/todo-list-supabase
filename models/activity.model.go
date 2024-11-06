package models

type ActivityModel struct {
	Id          int    `json:"id"`
	Title       string `json:"title" validate:"required, min=5, max=40"`
	Category    string `json:"category" validate:"required, oneof=TASK EVENT"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required, oneof=NEW ON_PROGRESS"`
	CreatedAt   string `json:"created_at"`
}
