package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"mpm/internal/models"
	"os"
	"path/filepath"
	"time"
)

// Сделаем путь настраиваемым через переменную окружения
var usersDirectory = getDataPath()

func getDataPath() string {
	path := os.Getenv("MPM_DATA_PATH")
	if path == "" {
		path = "/opt/mpm/data/users.json"
	}
	return path
}

func getDefaultUserPassword() string {
	password := os.Getenv("DEFAULT_USER_PASSWORD")
	if password == "" {
		password = "P@ssw0rd84" // Пароль по умолчанию
	}
	return password
}

type JSONUserStorage struct{}

func NewUserStorage() *JSONUserStorage {
	if err := ensureDataDir(); err != nil {
		log.Printf("Ошибка при создании директории данных: %v", err)
	}

	if err := ensureDefaultUser(); err != nil {
		log.Printf("Ошибка при создании пользователя по умолчанию: %v", err)
	}

	return &JSONUserStorage{}
}

// LoadUsers загружает список пользователей из JSON
func (s *JSONUserStorage) LoadUsers() ([]models.User, error) {
	// Проверяем наличие файла перед чтением
	if _, err := os.Stat(usersDirectory); os.IsNotExist(err) {
		log.Printf("Файл пользователей не существует, возвращаем пустой список")
		return []models.User{}, nil
	}

	data, err := os.ReadFile(usersDirectory)
	if err != nil {
		log.Printf("Ошибка при чтении файла пользователей: %v", err)
		return []models.User{}, nil // Возвращаем пустой список вместо ошибки
	}

	// Проверка на пустой файл
	if len(data) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		log.Printf("Ошибка при анализе JSON: %v", err)
		return []models.User{}, nil
	}

	return users, nil
}

// SaveUsers сохраняет список пользователей в JSON файл
func (s *JSONUserStorage) SaveUsers(users []models.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Printf("Ошибка при сериализации пользователей: %v", err)
		return err
	}

	if err := os.WriteFile(usersDirectory, data, 0644); err != nil {
		log.Printf("Ошибка при записи файла пользователей: %v", err)
		return err
	}

	return nil
}

func ensureDataDir() error {
	dir := filepath.Dir(usersDirectory)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Создание директории: %s", dir)
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func ensureDefaultUser() error {
	storage := &JSONUserStorage{}
	users, _ := storage.LoadUsers()

	// Проверка на наличие пользователя masterplan
	for _, user := range users {
		if user.Username == "masterplan" {
			return nil
		}
	}

	// Создание пользователя по умолчанию
	defaultUser := models.User{
		ID:        1,
		Username:  "masterplan",
		Email:     "maxim.tyatyushkin@gmail.com",
		Password:  getDefaultUserPassword(), // TODO: Захешировать пароль
		CreatedAt: time.Now(),
	}

	users = append(users, defaultUser)
	return storage.SaveUsers(users)
}

// GetUserByCredentials находит пользователя по логину и паролю
func (s *JSONUserStorage) GetUserByCredentials(username, password string) (*models.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for i, user := range users {
		if user.Username == username && user.Password == password {
			return &users[i], nil
		}
	}

	return nil, nil // Пользователь не найден
}

// GetUserByID находит пользователя по ID
func (s *JSONUserStorage) GetUserByID(id int) (*models.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for i, user := range users {
		if user.ID == id {
			return &users[i], nil
		}
	}

	return nil, fmt.Errorf("пользователь с ID %d не найден", id)
}
