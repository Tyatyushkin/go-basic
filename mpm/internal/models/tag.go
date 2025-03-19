package models

import "time"

type Tag struct {
	ID        int       `json:"id" db:"id"`     // Уникальный идентификатор тега
	Name      string    `json:"name" db:"name"` // Название тега
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (t Tag) GetID() int {
	return t.ID
}

func (t Tag) GetType() string {
	return "tag"
}
