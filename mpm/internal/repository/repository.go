package repository

import (
	"fmt"
	"mpm/internal/models"
)

// хранилище для всех типов сущностей
var (
	photos []models.Photo
	albums []models.Album
	tags   []models.Tag
)

// Initialize инициализирует слайсы, если это необходимо
func Initialize() {
	if photos == nil {
		photos = make([]models.Photo, 0)
	}

	if albums == nil {
		albums = make([]models.Album, 0)
	}

	if tags == nil {
		tags = make([]models.Tag, 0)
	}
}

// SaveEntities функция, которая принимает слайс интерфейсов Entity,
// проверяет реальный тип каждой сущности и добавляет ее в соответствующий слайс
func SaveEntities(entities []models.Entity) error {
	for _, entity := range entities {
		// Проверяем тип сущности с помощью type assertion
		switch e := entity.(type) {
		case models.Photo:
			photos = append(photos, e)
			fmt.Printf("Сохранена фотография: ID=%d, Title=%s\n", e.ID, e.Name)

		case *models.Photo:
			photos = append(photos, *e)
			fmt.Printf("Сохранена фотография: ID=%d, Title=%s\n", e.ID, e.Name)

		case models.Album:
			albums = append(albums, e)
			fmt.Printf("Сохранен альбом: ID=%d, Title=%s\n", e.ID, e.Name)

		case *models.Album:
			albums = append(albums, *e)
			fmt.Printf("Сохранен альбом: ID=%d, Title=%s\n", e.ID, e.Name)

		case models.Tag:
			tags = append(tags, e)
			fmt.Printf("Сохранен тег: ID=%d, Name=%s\n", e.ID, e.Name)

		case *models.Tag:
			tags = append(tags, *e)
			fmt.Printf("Сохранен тег: ID=%d, Name=%s\n", e.ID, e.Name)

		default:
			return fmt.Errorf("неизвестный тип сущности: %T", entity)
		}
	}

	return nil
}

// GetPhotos возвращает все фотографии
func GetPhotos() []models.Photo {
	return photos
}

// GetAlbums возвращает все альбомы
func GetAlbums() []models.Album {
	return albums
}

// GetTags возвращает все теги
func GetTags() []models.Tag {
	return tags
}
