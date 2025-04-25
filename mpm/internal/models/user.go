package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`                 // Уникальный идентификатор пользователя
	Username  string    `json:"username" db:"username"`     // Имя пользователя
	Password  string    `json:"password" db:"password"`     // Хэш пароля (не возвращается в API)
	Email     string    `json:"email" db:"email"`           // Email пользователя
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Дата регистрации пользователя
}
