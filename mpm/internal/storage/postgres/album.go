package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mpm/internal/models"
)

// AlbumStorage реализует хранилище альбомов для PostgreSQL
type AlbumStorage struct {
	client *Client
}

// NewAlbumStorage создает новое хранилище альбомов
func NewAlbumStorage(client *Client) *AlbumStorage {
	return &AlbumStorage{
		client: client,
	}
}

// Create создает новый альбом
func (s *AlbumStorage) Create(ctx context.Context, album *models.Album) (*models.Album, error) {
	query := `
		INSERT INTO albums (user_id, name, description, cover_photo_id, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	now := time.Now()
	album.CreatedAt = now
	album.UpdatedAt = now

	err := s.client.QueryRowContext(
		ctx,
		query,
		album.UserID,
		album.Name,
		album.Description,
		album.CoverPhotoID,
		album.IsPublic,
		album.CreatedAt,
		album.UpdatedAt,
	).Scan(&album.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create album: %w", err)
	}

	// Добавление тегов
	if err := s.updateTags(ctx, album.ID, album.Tags); err != nil {
		return nil, fmt.Errorf("failed to update album tags: %w", err)
	}

	return album, nil
}

// GetByID получает альбом по ID
func (s *AlbumStorage) GetByID(ctx context.Context, id int) (*models.Album, error) {
	query := `
		SELECT id, user_id, name, description, cover_photo_id, is_public, created_at, updated_at
		FROM albums
		WHERE id = $1`

	album := &models.Album{}
	err := s.client.QueryRowContext(ctx, query, id).Scan(
		&album.ID,
		&album.UserID,
		&album.Name,
		&album.Description,
		&album.CoverPhotoID,
		&album.IsPublic,
		&album.CreatedAt,
		&album.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("album not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	// Загрузка тегов
	tags, err := s.getTags(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	album.Tags = tags

	// Загрузка количества фотографий
	photoCount, err := s.getPhotoCount(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photo count: %w", err)
	}
	album.PhotoCount = photoCount

	return album, nil
}

// Update обновляет альбом
func (s *AlbumStorage) Update(ctx context.Context, album *models.Album) error {
	query := `
		UPDATE albums
		SET name = $2, description = $3, cover_photo_id = $4, is_public = $5, updated_at = $6
		WHERE id = $1`

	album.UpdatedAt = time.Now()

	result, err := s.client.ExecContext(
		ctx,
		query,
		album.ID,
		album.Name,
		album.Description,
		album.CoverPhotoID,
		album.IsPublic,
		album.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("album not found")
	}

	// Обновление тегов
	if err := s.updateTags(ctx, album.ID, album.Tags); err != nil {
		return fmt.Errorf("failed to update album tags: %w", err)
	}

	return nil
}

// Delete удаляет альбом
func (s *AlbumStorage) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM albums WHERE id = $1`

	result, err := s.client.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("album not found")
	}

	return nil
}

// List получает список альбомов с фильтрацией и пагинацией
func (s *AlbumStorage) List(ctx context.Context, userID int, offset, limit int) ([]*models.Album, error) {
	query := `
		SELECT id, user_id, name, description, cover_photo_id, is_public, created_at, updated_at
		FROM albums
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.client.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list albums: %w", err)
	}
	defer rows.Close()

	var albums []*models.Album
	for rows.Next() {
		album := &models.Album{}
		err := rows.Scan(
			&album.ID,
			&album.UserID,
			&album.Name,
			&album.Description,
			&album.CoverPhotoID,
			&album.IsPublic,
			&album.CreatedAt,
			&album.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}

		// Загрузка тегов для каждого альбома
		tags, err := s.getTags(ctx, album.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get album tags: %w", err)
		}
		album.Tags = tags

		// Загрузка количества фотографий
		photoCount, err := s.getPhotoCount(ctx, album.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get photo count: %w", err)
		}
		album.PhotoCount = photoCount

		albums = append(albums, album)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate albums: %w", err)
	}

	return albums, nil
}

// Count возвращает количество альбомов пользователя
func (s *AlbumStorage) Count(ctx context.Context, userID int) (int64, error) {
	query := `SELECT COUNT(*) FROM albums WHERE user_id = $1`

	var count int64
	err := s.client.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count albums: %w", err)
	}

	return count, nil
}

// updateTags обновляет теги альбома
func (s *AlbumStorage) updateTags(ctx context.Context, albumID int, tags []string) error {
	// Начинаем транзакцию
	tx, err := s.client.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем существующие связи
	_, err = tx.ExecContext(ctx, `DELETE FROM album_tags WHERE album_id = $1`, albumID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Добавляем новые теги
	for _, tagName := range tags {
		// Получаем или создаем тег
		var tagID int
		err := tx.QueryRowContext(ctx,
			`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = $1 RETURNING id`,
			tagName,
		).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("failed to upsert tag: %w", err)
		}

		// Создаем связь
		_, err = tx.ExecContext(ctx,
			`INSERT INTO album_tags (album_id, tag_id) VALUES ($1, $2)`,
			albumID, tagID,
		)
		if err != nil {
			return fmt.Errorf("failed to create album-tag relation: %w", err)
		}
	}

	return tx.Commit()
}

// getTags получает теги альбома
func (s *AlbumStorage) getTags(ctx context.Context, albumID int) ([]string, error) {
	query := `
		SELECT t.name
		FROM tags t
		JOIN album_tags at ON t.id = at.tag_id
		WHERE at.album_id = $1
		ORDER BY t.name`

	rows, err := s.client.QueryContext(ctx, query, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// getPhotoCount получает количество фотографий в альбоме
func (s *AlbumStorage) getPhotoCount(ctx context.Context, albumID int) (int, error) {
	query := `SELECT COUNT(*) FROM photos WHERE album_id = $1`

	var count int
	err := s.client.QueryRowContext(ctx, query, albumID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count photos: %w", err)
	}

	return count, nil
}
