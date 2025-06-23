package postgres

import (
	"context"
	"fmt"

	"mpm/internal/models"
)

// UserStorage реализует хранилище пользователей для PostgreSQL
type UserStorage struct {
	client *Client
}

// NewUserStorage создает новое хранилище пользователей
func NewUserStorage(client *Client) *UserStorage {
	return &UserStorage{
		client: client,
	}
}

// Create создает нового пользователя
func (s *UserStorage) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// TODO: Реализовать создание пользователя
	return nil, fmt.Errorf("user storage not implemented yet")
}

// GetByID получает пользователя по ID
func (s *UserStorage) GetByID(ctx context.Context, id int) (*models.User, error) {
	// TODO: Реализовать получение пользователя по ID
	return nil, fmt.Errorf("user storage not implemented yet")
}

// GetByUsername получает пользователя по username
func (s *UserStorage) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	// TODO: Реализовать получение пользователя по username
	return nil, fmt.Errorf("user storage not implemented yet")
}

// Update обновляет пользователя
func (s *UserStorage) Update(ctx context.Context, user *models.User) error {
	// TODO: Реализовать обновление пользователя
	return fmt.Errorf("user storage not implemented yet")
}

// Delete удаляет пользователя
func (s *UserStorage) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление пользователя
	return fmt.Errorf("user storage not implemented yet")
}

// List получает список пользователей
func (s *UserStorage) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	// TODO: Реализовать получение списка пользователей
	return nil, fmt.Errorf("user storage not implemented yet")
}

// Count возвращает количество пользователей
func (s *UserStorage) Count(ctx context.Context) (int64, error) {
	// TODO: Реализовать подсчет пользователей
	return 0, fmt.Errorf("user storage not implemented yet")
}
