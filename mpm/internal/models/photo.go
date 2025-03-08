package models

import "time"

type Photo struct {
	ID        int       `json:"id" db:"id"`     // Уникальный идентификатор фотографии
	Name      string    `json:"name" db:"name"` // Название фотографии
	Path      string    `json:"path" db:"path"` // Путь к фотографии (локальный или url)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
