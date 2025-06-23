package postgres

import (
	"context"
	"fmt"

	"mpm/internal/models"
)

// CommentStorage реализует хранилище комментариев для PostgreSQL
type CommentStorage struct {
	client *Client
}

// NewCommentStorage создает новое хранилище комментариев
func NewCommentStorage(client *Client) *CommentStorage {
	return &CommentStorage{
		client: client,
	}
}

// Create создает новый комментарий
func (s *CommentStorage) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	// TODO: Реализовать создание комментария
	return nil, fmt.Errorf("comment storage not implemented yet")
}

// GetByID получает комментарий по ID
func (s *CommentStorage) GetByID(ctx context.Context, id int) (*models.Comment, error) {
	// TODO: Реализовать получение комментария по ID
	return nil, fmt.Errorf("comment storage not implemented yet")
}

// Update обновляет комментарий
func (s *CommentStorage) Update(ctx context.Context, comment *models.Comment) error {
	// TODO: Реализовать обновление комментария
	return fmt.Errorf("comment storage not implemented yet")
}

// Delete удаляет комментарий
func (s *CommentStorage) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление комментария
	return fmt.Errorf("comment storage not implemented yet")
}

// ListByPhoto получает список комментариев для фотографии
func (s *CommentStorage) ListByPhoto(ctx context.Context, photoID int, offset, limit int) ([]*models.Comment, error) {
	// TODO: Реализовать получение списка комментариев для фотографии
	return nil, fmt.Errorf("comment storage not implemented yet")
}

// Count возвращает количество комментариев для фотографии
func (s *CommentStorage) Count(ctx context.Context, photoID int) (int64, error) {
	// TODO: Реализовать подсчет комментариев
	return 0, fmt.Errorf("comment storage not implemented yet")
}
