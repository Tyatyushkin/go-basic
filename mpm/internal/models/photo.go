package models

import "time"

type Photo struct {
	ID          int        `json:"id" db:"id"`                 // Уникальный идентификатор фотографии
	Name        string     `json:"name" db:"name"`             // Название фотографии
	Path        string     `json:"path" db:"path"`             // Путь к фотографии (локальный или url)
	Album       *Album     `json:"album,omitempty" db:"album"` // Альбом, к которому принадлежит фотография
	User        *User      `json:"user,omitempty" db:"user"`   // Пользователь, который загрузил фотографию
	Tags        []string   `json:"tags" db:"tags"`             // Теги фотографии
	Metadata    []Metadata `json:"metadata" db:"metadata"`
	StorageType string     `json:"storage_type" db:"storage_type"` // Тип хранения фотографии (local, google, dropbox)
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
