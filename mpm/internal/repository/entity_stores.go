package repository

import (
	"fmt"
	"mpm/internal/models"
	"sync"
)

// EntityStore определяет общий интерфейс для всех хранилищ сущностей
type EntityStore interface {
	Add(entity models.Entity) error
	GetAll() []models.Entity
	GetNew() []models.Entity
	Count() int
}

// PhotoStore - типизированное хранилище фотографий
type PhotoStore struct {
	items     []models.Photo
	mutex     sync.RWMutex
	lastIndex int
}

// AlbumStore - типизированное хранилище альбомов
type AlbumStore struct {
	items     []models.Album
	mutex     sync.RWMutex
	lastIndex int
}

// TagStore - типизированное хранилище тегов
type TagStore struct {
	items     []models.Tag
	mutex     sync.RWMutex
	lastIndex int
}

// NewPhotoStore создает новое хранилище фотографий
func NewPhotoStore() *PhotoStore {
	return &PhotoStore{
		items: make([]models.Photo, 0),
	}
}

// Add добавляет фотографию в хранилище
func (s *PhotoStore) Add(photo models.Photo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = append(s.items, photo)
	fmt.Printf("Сохранена фотография: ID=%d, Title=%s\n", photo.ID, photo.Name)
	return nil
}

// AddEntity реализует интерфейс EntityStore
func (s *PhotoStore) AddEntity(entity models.Entity) error {
	if photo, ok := entity.(models.Photo); ok {
		return s.Add(photo)
	}
	return fmt.Errorf("неверный тип сущности для PhotoStore: %T", entity)
}

// GetAll возвращает копию всех фотографий
func (s *PhotoStore) GetAll() []models.Photo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Photo, len(s.items))
	copy(result, s.items)
	return result
}

// GetAllEntities реализует интерфейс EntityStore
func (s *PhotoStore) GetAllEntities() []models.Entity {
	photos := s.GetAll()
	entities := make([]models.Entity, len(photos))
	for i, photo := range photos {
		entities[i] = photo
	}
	return entities
}

// GetNew возвращает новые фотографии с момента последнего вызова
func (s *PhotoStore) GetNew() []models.Photo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastIndex >= len(s.items) {
		return []models.Photo{}
	}

	result := make([]models.Photo, len(s.items)-s.lastIndex)
	copy(result, s.items[s.lastIndex:])
	s.lastIndex = len(s.items)
	return result
}

// GetNewEntities реализует интерфейс EntityStore
func (s *PhotoStore) GetNewEntities() []models.Entity {
	photos := s.GetNew()
	entities := make([]models.Entity, len(photos))
	for i, photo := range photos {
		entities[i] = photo
	}
	return entities
}

// Count возвращает количество фотографий
func (s *PhotoStore) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.items)
}
