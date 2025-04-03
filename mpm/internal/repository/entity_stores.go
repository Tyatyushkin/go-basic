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

// NewAlbumStore создает новое хранилище альбомов
func NewAlbumStore() *AlbumStore {
	return &AlbumStore{
		items: make([]models.Album, 0),
	}
}

// Add добавляет альбом в хранилище
func (s *AlbumStore) Add(album models.Album) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = append(s.items, album)
	fmt.Printf("Сохранен альбом: ID=%d, Title=%s\n", album.ID, album.Name)
	return nil
}

// AddEntity реализует интерфейс EntityStore
func (s *AlbumStore) AddEntity(entity models.Entity) error {
	if album, ok := entity.(models.Album); ok {
		return s.Add(album)
	}
	return fmt.Errorf("неверный тип сущности для AlbumStore: %T", entity)
}

// GetAll возвращает копию всех альбомов
func (s *AlbumStore) GetAll() []models.Album {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Album, len(s.items))
	copy(result, s.items)
	return result
}

// GetAllEntities реализует интерфейс EntityStore
func (s *AlbumStore) GetAllEntities() []models.Entity {
	albums := s.GetAll()
	entities := make([]models.Entity, len(albums))
	for i, album := range albums {
		entities[i] = album
	}
	return entities
}

// GetNew возвращает новые альбомы с момента последнего вызова
func (s *AlbumStore) GetNew() []models.Album {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastIndex >= len(s.items) {
		return []models.Album{}
	}

	result := make([]models.Album, len(s.items)-s.lastIndex)
	copy(result, s.items[s.lastIndex:])
	s.lastIndex = len(s.items)
	return result
}

// GetNewEntities реализует интерфейс EntityStore
func (s *AlbumStore) GetNewEntities() []models.Entity {
	albums := s.GetNew()
	entities := make([]models.Entity, len(albums))
	for i, album := range albums {
		entities[i] = album
	}
	return entities
}

// Count возвращает количество альбомов
func (s *AlbumStore) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.items)
}

// NewTagStore создает новое хранилище тегов
func NewTagStore() *TagStore {
	return &TagStore{
		items: make([]models.Tag, 0),
	}
}

// Add добавляет тег в хранилище
func (s *TagStore) Add(tag models.Tag) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = append(s.items, tag)
	fmt.Printf("Сохранен тег: ID=%d, Name=%s\n", tag.ID, tag.Name)
	return nil
}

// AddEntity реализует интерфейс EntityStore
func (s *TagStore) AddEntity(entity models.Entity) error {
	if tag, ok := entity.(models.Tag); ok {
		return s.Add(tag)
	}
	return fmt.Errorf("неверный тип сущности для TagStore: %T", entity)
}

// GetAll возвращает копию всех тегов
func (s *TagStore) GetAll() []models.Tag {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Tag, len(s.items))
	copy(result, s.items)
	return result
}

// GetAllEntities реализует интерфейс EntityStore
func (s *TagStore) GetAllEntities() []models.Entity {
	tags := s.GetAll()
	entities := make([]models.Entity, len(tags))
	for i, tag := range tags {
		entities[i] = tag
	}
	return entities
}

// GetNew возвращает новые теги с момента последнего вызова
func (s *TagStore) GetNew() []models.Tag {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastIndex >= len(s.items) {
		return []models.Tag{}
	}

	result := make([]models.Tag, len(s.items)-s.lastIndex)
	copy(result, s.items[s.lastIndex:])
	s.lastIndex = len(s.items)
	return result
}

// GetNewEntities реализует интерфейс EntityStore
func (s *TagStore) GetNewEntities() []models.Entity {
	tags := s.GetNew()
	entities := make([]models.Entity, len(tags))
	for i, tag := range tags {
		entities[i] = tag
	}
	return entities
}

// Count возвращает количество тегов
func (s *TagStore) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.items)
}
