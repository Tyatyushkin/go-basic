package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mpm/internal/models"
)

// AlbumStorage реализация хранилища альбомов для MongoDB
type AlbumStorage struct {
	client     *Client
	collection *mongo.Collection
}

// NewAlbumStorage создает новое хранилище альбомов
func NewAlbumStorage(client *Client) *AlbumStorage {
	return &AlbumStorage{
		client:     client,
		collection: client.GetAlbumsCollection(),
	}
}

// Create создает новый альбом
func (s *AlbumStorage) Create(ctx context.Context, album *models.Album) (*models.Album, error) {
	// Преобразуем модель в документ MongoDB
	doc := AlbumDocumentFromModel(album)
	doc.ID = primitive.NewObjectID()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	// Вставляем документ
	result, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert album: %w", err)
	}

	// Устанавливаем ID и возвращаем созданный альбом
	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = objectID
	}

	return doc.ToModel(), nil
}

// GetByID получает альбом по ID
func (s *AlbumStorage) GetByID(ctx context.Context, id string) (*models.Album, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid album ID format: %w", err)
	}

	var doc AlbumDocument
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("album not found")
		}
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	return doc.ToModel(), nil
}

// Update обновляет существующий альбом
func (s *AlbumStorage) Update(ctx context.Context, id string, updates *AlbumUpdateRequest) (*models.Album, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid album ID format: %w", err)
	}

	// Подготавливаем обновления
	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if updates.Name != nil {
		updateDoc["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateDoc["description"] = *updates.Description
	}
	if updates.Tags != nil {
		updateDoc["tags"] = updates.Tags
	}

	// Выполняем обновление
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updateDoc}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update album: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("album not found")
	}

	// Возвращаем обновленный альбом
	return s.GetByID(ctx, id)
}

// Delete удаляет альбом
func (s *AlbumStorage) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid album ID format: %w", err)
	}

	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("album not found")
	}

	return nil
}

// List получает список альбомов с фильтрацией и пагинацией
func (s *AlbumStorage) List(ctx context.Context, opts *AlbumListOptions) ([]*models.Album, error) {
	// Подготавливаем фильтр
	filter := bson.M{}
	if opts.Filter != nil {
		if opts.Filter.UserID != nil {
			filter["user_id"] = *opts.Filter.UserID
		}
		if len(opts.Filter.Tags) > 0 {
			filter["tags"] = bson.M{"$in": opts.Filter.Tags}
		}
		if opts.Filter.Name != nil {
			filter["name"] = primitive.Regex{
				Pattern: *opts.Filter.Name,
				Options: "i", // case insensitive
			}
		}
		if opts.Filter.CreatedAt != nil {
			timeFilter := bson.M{}
			if opts.Filter.CreatedAt.From != nil {
				timeFilter["$gte"] = *opts.Filter.CreatedAt.From
			}
			if opts.Filter.CreatedAt.To != nil {
				timeFilter["$lte"] = *opts.Filter.CreatedAt.To
			}
			if len(timeFilter) > 0 {
				filter["created_at"] = timeFilter
			}
		}
	}

	// Подготавливаем опции запроса
	findOptions := options.Find()

	if opts.Skip > 0 {
		findOptions.SetSkip(opts.Skip)
	}
	if opts.Limit > 0 {
		findOptions.SetLimit(opts.Limit)
	}

	// Сортировка
	if len(opts.Sort) > 0 {
		sortDoc := bson.D{}
		for field, direction := range opts.Sort {
			sortDoc = append(sortDoc, bson.E{Key: field, Value: direction})
		}
		findOptions.SetSort(sortDoc)
	} else {
		// Сортировка по умолчанию: сначала новые
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// Выполняем запрос
	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find albums: %w", err)
	}
	defer cursor.Close(ctx)

	// Декодируем результаты
	var albums []*models.Album
	for cursor.Next(ctx) {
		var doc AlbumDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode album: %w", err)
		}
		albums = append(albums, doc.ToModel())
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return albums, nil
}

// Count подсчитывает количество альбомов с учетом фильтров
func (s *AlbumStorage) Count(ctx context.Context, filter *AlbumFilter) (int64, error) {
	mongoFilter := bson.M{}

	if filter != nil {
		if filter.UserID != nil {
			mongoFilter["user_id"] = *filter.UserID
		}
		if len(filter.Tags) > 0 {
			mongoFilter["tags"] = bson.M{"$in": filter.Tags}
		}
		if filter.Name != nil {
			mongoFilter["name"] = primitive.Regex{
				Pattern: *filter.Name,
				Options: "i",
			}
		}
		if filter.CreatedAt != nil {
			timeFilter := bson.M{}
			if filter.CreatedAt.From != nil {
				timeFilter["$gte"] = *filter.CreatedAt.From
			}
			if filter.CreatedAt.To != nil {
				timeFilter["$lte"] = *filter.CreatedAt.To
			}
			if len(timeFilter) > 0 {
				mongoFilter["created_at"] = timeFilter
			}
		}
	}

	count, err := s.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to count albums: %w", err)
	}

	return count, nil
}

// GetByTags получает альбомы по тегам
func (s *AlbumStorage) GetByTags(ctx context.Context, tags []string, limit int64) ([]*models.Album, error) {
	opts := &AlbumListOptions{
		Filter: &AlbumFilter{Tags: tags},
		Limit:  limit,
	}
	return s.List(ctx, opts)
}

// GetByUserID получает альбомы пользователя
func (s *AlbumStorage) GetByUserID(ctx context.Context, userID string, limit int64) ([]*models.Album, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	opts := &AlbumListOptions{
		Filter: &AlbumFilter{UserID: &objectID},
		Limit:  limit,
	}
	return s.List(ctx, opts)
}
