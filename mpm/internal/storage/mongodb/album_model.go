package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"mpm/internal/models"
)

// AlbumDocument представляет структуру альбома в MongoDB
type AlbumDocument struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"`
	Name        string              `bson:"name"`
	Description string              `bson:"description"`
	UserID      *primitive.ObjectID `bson:"user_id,omitempty"` // Ссылка на пользователя
	Tags        []string            `bson:"tags,omitempty"`
	CreatedAt   time.Time           `bson:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at"`
}

// ToModel преобразует AlbumDocument в models.Album
func (ad *AlbumDocument) ToModel() *models.Album {
	album := &models.Album{
		ID:          int(ad.ID.Timestamp().Unix()), // Временное решение для ID
		Name:        ad.Name,
		Description: ad.Description,
		Tags:        ad.Tags,
		CreatedAt:   ad.CreatedAt,
	}

	// TODO: Загрузка связанного пользователя и фотографий по мере необходимости

	return album
}

// FromModel создает AlbumDocument из models.Album
func AlbumDocumentFromModel(album *models.Album) *AlbumDocument {
	doc := &AlbumDocument{
		Name:        album.Name,
		Description: album.Description,
		Tags:        album.Tags,
		CreatedAt:   album.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	// Если у альбома есть пользователь, устанавливаем UserID
	if album.User != nil {
		// TODO: Здесь нужно будет преобразовать User ID в ObjectID
		// когда будет реализована MongoDB поддержка для пользователей
	}

	return doc
}

// AlbumCreateRequest структура для создания нового альбома
type AlbumCreateRequest struct {
	Name        string              `bson:"name" validate:"required,max=100"`
	Description string              `bson:"description" validate:"max=500"`
	Tags        []string            `bson:"tags,omitempty"`
	UserID      *primitive.ObjectID `bson:"user_id,omitempty"`
}

// ToDocument преобразует запрос создания в документ
func (req *AlbumCreateRequest) ToDocument() *AlbumDocument {
	now := time.Now()
	return &AlbumDocument{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		UserID:      req.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AlbumUpdateRequest структура для обновления альбома
type AlbumUpdateRequest struct {
	Name        *string   `bson:"name,omitempty" validate:"omitempty,max=100"`
	Description *string   `bson:"description,omitempty" validate:"omitempty,max=500"`
	Tags        []string  `bson:"tags,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// AlbumFilter структура для фильтрации альбомов
type AlbumFilter struct {
	UserID    *primitive.ObjectID `bson:"user_id,omitempty"`
	Tags      []string            `bson:"tags,omitempty"`
	Name      *string             `bson:"name,omitempty"`
	CreatedAt *TimeRange          `bson:"created_at,omitempty"`
}

// TimeRange структура для фильтрации по диапазону времени
type TimeRange struct {
	From *time.Time `bson:"$gte,omitempty"`
	To   *time.Time `bson:"$lte,omitempty"`
}

// AlbumListOptions опции для получения списка альбомов
type AlbumListOptions struct {
	Filter *AlbumFilter
	Sort   map[string]int // поле -> направление (1 для asc, -1 для desc)
	Skip   int64
	Limit  int64
}
