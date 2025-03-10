package storage

import (
	"encoding/json"
	"errors"
	"mpm/internal/models"
	"os"
	"path/filepath"
	"time"
)

const usersDirectory = "/opt/mpm/data/users.json"

type JSONUserStorage struct{}

func NewUserStorage(directory string) *JSONUserStorage {
	ensureDataDir()
	ensureDefaultUser()
	return &JSONUserStorage{}
}

// LoadUsers загружает список пользователей из JSON
func (s *JSONUserStorage) LoadUsers() ([]models.User, error) {
	data, err := os.ReadFile(usersDirectory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.User{}, nil
		}
		return nil, err
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *JSONUserStorage) SaveUsers(users []models.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(usersDirectory, data, 0644)
}

// AddUser добавляет нового пользователя в JSON-хранилище
func (s *JSONUserStorage) AddUser(user models.User) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	users = append(users, user)
	return s.SaveUsers(users)
}

func ensureDataDir() {
	dir := filepath.Dir(usersDirectory)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// GetUserByUsername ищет пользователя по имени
func (s *JSONUserStorage) GetUserByUsername(username string) (*models.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, errors.New("пользователь не найден")
}

func ensureDefaultUser() {
	users, err := (&JSONUserStorage{}).LoadUsers()
	if err != nil {
		return
	}

	// Проверяем, есть ли пользователь masterplan
	for _, user := range users {
		if user.Username == "masterplan" {
			return
		}
	}

	// Создаём пользователя masterplan
	defaultUser := models.User{
		ID:        1,
		Username:  "masterplan",
		Email:     "masterplan@example.com",
		Password:  "changeme", // TODO: Захешировать пароль
		CreatedAt: time.Now(),
	}

	users = append(users, defaultUser)
	(&JSONUserStorage{}).SaveUsers(users)
}
