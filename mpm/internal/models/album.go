package models

import (
	"fmt"
	"strings"
	"time"
)

type Album struct {
	ID          int       `json:"id" db:"id"` // Уникальный идентификатор альбома
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"` // Название альбома
	User        *User     `json:"user,omitempty" db:"user"`     // Пользователь, который создал альбом
	Photos      []Photo   `json:"photos,omitempty" db:"photos"` // Фотографии в альбоме
	Tags        []string  `json:"tags,omitempty" db:"tags"`     // Теги альбома
	CreatedAt   time.Time `json:"created_at" db:"created_at"`   // Дата создания альбома
}

func (a Album) GetID() int {
	return a.ID
}

func (a Album) GetType() string {
	return "album"
}

func (a Album) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("название альбома не может быть пустым")
	}
	if len(a.Name) > 100 {
		return fmt.Errorf("название альбома слишком длинное")
	}
	if len(a.Description) > 500 {
		return fmt.Errorf("описание альбома слишком длинное")
	}

	for _, tag := range a.Tags {
		if tag == "" || strings.Contains(tag, " ") {
			return fmt.Errorf("теги содержат недопустимые символы")
		}
	}

	return nil
}
