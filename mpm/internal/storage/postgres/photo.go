package postgres

import (
	"context"
	"fmt"

	"mpm/internal/models"
)

// PhotoStorage реализует хранилище фотографий для PostgreSQL
type PhotoStorage struct {
	client *Client
}

// NewPhotoStorage создает новое хранилище фотографий
func NewPhotoStorage(client *Client) *PhotoStorage {
	return &PhotoStorage{
		client: client,
	}
}

// Create создает новую фотографию
func (s *PhotoStorage) Create(ctx context.Context, photo *models.Photo) (*models.Photo, error) {
	// TODO: Реализовать создание фотографии
	return nil, fmt.Errorf("photo storage not implemented yet")
}

// GetByID получает фотографию по ID
func (s *PhotoStorage) GetByID(ctx context.Context, id int) (*models.Photo, error) {
	// TODO: Реализовать получение фотографии по ID
	return nil, fmt.Errorf("photo storage not implemented yet")
}

// Update обновляет фотографию
func (s *PhotoStorage) Update(ctx context.Context, photo *models.Photo) error {
	// TODO: Реализовать обновление фотографии
	return fmt.Errorf("photo storage not implemented yet")
}

// Delete удаляет фотографию
func (s *PhotoStorage) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление фотографии
	return fmt.Errorf("photo storage not implemented yet")
}

// List получает список фотографий с фильтрацией
func (s *PhotoStorage) List(ctx context.Context, albumID int, offset, limit int) ([]*models.Photo, error) {
	// TODO: Реализовать получение списка фотографий
	return nil, fmt.Errorf("photo storage not implemented yet")
}

// Count возвращает количество фотографий в альбоме
func (s *PhotoStorage) Count(ctx context.Context, albumID int) (int64, error) {
	// TODO: Реализовать подсчет фотографий
	return 0, fmt.Errorf("photo storage not implemented yet")
}
