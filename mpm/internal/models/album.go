package models

import "time"

type Album struct {
	ID        int       `json:"id" db:"id"`                 // Уникальный идентификатор альбома
	Name      string    `json:"name" db:"name"`             // Название альбома
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Дата создания альбома
}
