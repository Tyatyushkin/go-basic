package postgres

import (
	"context"
	"fmt"

	"mpm/internal/models"
)

// TagStorage реализует хранилище тегов для PostgreSQL
type TagStorage struct {
	client *Client
}

// NewTagStorage создает новое хранилище тегов
func NewTagStorage(client *Client) *TagStorage {
	return &TagStorage{
		client: client,
	}
}

// Create создает новый тег
func (s *TagStorage) Create(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	// TODO: Реализовать создание тега
	return nil, fmt.Errorf("tag storage not implemented yet")
}

// GetByID получает тег по ID
func (s *TagStorage) GetByID(ctx context.Context, id int) (*models.Tag, error) {
	// TODO: Реализовать получение тега по ID
	return nil, fmt.Errorf("tag storage not implemented yet")
}

// GetByName получает тег по имени
func (s *TagStorage) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	// TODO: Реализовать получение тега по имени
	return nil, fmt.Errorf("tag storage not implemented yet")
}

// Update обновляет тег
func (s *TagStorage) Update(ctx context.Context, tag *models.Tag) error {
	// TODO: Реализовать обновление тега
	return fmt.Errorf("tag storage not implemented yet")
}

// Delete удаляет тег
func (s *TagStorage) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление тега
	return fmt.Errorf("tag storage not implemented yet")
}

// List получает список тегов
func (s *TagStorage) List(ctx context.Context, offset, limit int) ([]*models.Tag, error) {
	// TODO: Реализовать получение списка тегов
	return nil, fmt.Errorf("tag storage not implemented yet")
}

// Count возвращает количество тегов
func (s *TagStorage) Count(ctx context.Context) (int64, error) {
	// TODO: Реализовать подсчет тегов
	return 0, fmt.Errorf("tag storage not implemented yet")
}
