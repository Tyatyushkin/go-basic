package models

import "time"

type Album struct {
	ID        int       `json:"id" db:"id"`                   // Уникальный идентификатор альбома
	Name      string    `json:"name" db:"name"`               // Название альбома
	User      *User     `json:"user,omitempty" db:"user"`     // Пользователь, который создал альбом
	Photos    []Photo   `json:"photos,omitempty" db:"photos"` // Фотографии в альбоме
	Tags      []string  `json:"tags,omitempty" db:"tags"`     // Теги альбома
	CreatedAt time.Time `json:"created_at" db:"created_at"`   // Дата создания альбома
}
