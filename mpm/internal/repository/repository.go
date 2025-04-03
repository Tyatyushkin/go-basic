package repository

import (
	"fmt"
	"mpm/internal/models"
	"sync"
)

// Repository объединяет все хранилища сущностей
type Repository struct {
	photoStore *PhotoStore
	albumStore *AlbumStore
	tagStore   *TagStore
}

// Хранилище для всех типов сущностей
var (
	photos []models.Photo
	albums []models.Album
	tags   []models.Tag
	mutex  sync.RWMutex // Используется для синхронизации доступа к слайсам
)

// Initialize инициализирует слайсы, если это необходимо
func Initialize() {
	mutex.Lock()
	defer mutex.Unlock()

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
	mutex.Lock()
	defer mutex.Unlock()

	for _, entity := range entities {
		// Проверяем тип сущности с помощью type assertion
		switch e := entity.(type) {
		//case models.Photo:
		//	photos = append(photos, e)
		//	fmt.Printf("Сохранена фотография: ID=%d, Title=%s\n", e.ID, e.Name)

		case models.Album:
			albums = append(albums, e)
			fmt.Printf("Сохранен альбом: ID=%d, Title=%s\n", e.ID, e.Name)

		case models.Tag:
			tags = append(tags, e)
			fmt.Printf("Сохранен тег: ID=%d, Name=%s\n", e.ID, e.Name)

		default:
			return fmt.Errorf("неизвестный тип сущности: %T", entity)
		}
	}

	return nil
}

// GetPhotos возвращает все фотографии
func GetPhotos() []models.Photo {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]models.Photo, len(photos))
	copy(result, photos)
	return result
}

// GetAlbums возвращает все альбомы
func GetAlbums() []models.Album {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]models.Album, len(albums))
	return result
}

// GetTags возвращает все теги
func GetTags() []models.Tag {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]models.Tag, len(tags))
	copy(result, tags)
	return result
}

// GetEntitiesCounts возвращает количество сущностей каждого типа
func GetEntitiesCounts() (photoCount, albumCount, tagCount int) {
	mutex.RLock()
	defer mutex.RUnlock()

	return len(photos), len(albums), len(tags)
}

// GetNewEntities возвращает новые сущности, начиная с определенных индексов
func GetNewEntities(photoStartIndex, albumStartIndex, tagStartIndex int) (newPhotos []models.Photo, newAlbums []models.Album, newTags []models.Tag) {
	mutex.RLock()
	defer mutex.RUnlock()

	if photoStartIndex < len(photos) {
		newPhotos = make([]models.Photo, len(photos)-photoStartIndex)
		copy(newPhotos, photos[photoStartIndex:])
	}

	if albumStartIndex < len(albums) {
		newAlbums = make([]models.Album, len(albums)-albumStartIndex)
		copy(newAlbums, albums[albumStartIndex:])
	}

	if tagStartIndex < len(tags) {
		newTags = make([]models.Tag, len(tags)-tagStartIndex)
		copy(newTags, tags[tagStartIndex:])
	}

	return
}
