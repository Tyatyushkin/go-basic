package repository

import (
	"context"
	"fmt"
	"log"
	"mpm/config"
	"mpm/internal/models"
	"mpm/internal/storage/mongodb"
	"time"
)

// MongoDBStorage реализует EntityStorage для MongoDB
type MongoDBStorage struct {
	client       *mongodb.Client
	albumStorage *mongodb.AlbumStorage

	// Кэш для совместимости с существующей архитектурой
	albums []models.Album
	photos []models.Photo
	tags   []models.Tag
}

// NewMongoDBStorage создает новое MongoDB хранилище
func NewMongoDBStorage() (EntityStorage, error) {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Создаем MongoDB клиент
	client, err := mongodb.NewClient(&cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	storage := &MongoDBStorage{
		client:       client,
		albumStorage: mongodb.NewAlbumStorage(client),
		albums:       make([]models.Album, 0),
		photos:       make([]models.Photo, 0),
		tags:         make([]models.Tag, 0),
	}

	log.Println("MongoDB хранилище инициализировано")
	return storage, nil
}

// Реализация EntityStorage интерфейса

// Save сохраняет сущность в MongoDB
func (s *MongoDBStorage) Save(entity models.Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch e := entity.(type) {
	case *models.Album:
		_, err := s.albumStorage.Create(ctx, e)
		return err
	case models.Album:
		_, err := s.albumStorage.Create(ctx, &e)
		return err
	default:
		return fmt.Errorf("неподдерживаемый тип сущности: %T", entity)
	}
}

// SaveBatch сохраняет множество сущностей
func (s *MongoDBStorage) SaveBatch(entities []models.Entity) error {
	for _, entity := range entities {
		if err := s.Save(entity); err != nil {
			return fmt.Errorf("ошибка сохранения сущности %v: %w", entity.GetID(), err)
		}
	}
	return nil
}

// Load загружает данные из MongoDB в кэш
func (s *MongoDBStorage) Load() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Загружаем альбомы
	opts := &mongodb.AlbumListOptions{
		Limit: 10000, // разумный лимит
	}

	albums, err := s.albumStorage.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("ошибка загрузки альбомов: %w", err)
	}

	// Преобразуем в слайс models.Album
	s.albums = make([]models.Album, len(albums))
	for i, album := range albums {
		s.albums[i] = *album
	}

	log.Printf("Загружено из MongoDB: %d альбомов", len(s.albums))
	return nil
}

// Persist сохраняет текущее состояние кэша в MongoDB
func (s *MongoDBStorage) Persist() error {
	// MongoDB сохраняет данные автоматически при каждой операции
	// Этот метод нужен для совместимости с интерфейсом
	log.Println("MongoDB Persist() вызван - данные уже сохранены")
	return nil
}

// Close закрывает соединение с MongoDB
func (s *MongoDBStorage) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// MongoAlbumStore адаптер для альбомов MongoDB
type MongoAlbumStore struct {
	storage *mongodb.AlbumStorage
}

// NewMongoAlbumStore создает хранилище альбомов для MongoDB
func NewMongoAlbumStore(client *mongodb.Client) *MongoAlbumStore {
	return &MongoAlbumStore{
		storage: mongodb.NewAlbumStorage(client),
	}
}

// Add добавляет альбом
func (s *MongoAlbumStore) Add(entity models.Entity) error {
	album, ok := entity.(*models.Album)
	if !ok {
		return fmt.Errorf("entity is not an Album")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.storage.Create(ctx, album)
	return err
}

// GetAll получает все альбомы
func (s *MongoAlbumStore) GetAll() []models.Entity {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := &mongodb.AlbumListOptions{
		Limit: 1000, // ограничение для безопасности
	}

	albums, err := s.storage.List(ctx, opts)
	if err != nil {
		log.Printf("Ошибка получения альбомов: %v", err)
		return []models.Entity{}
	}

	entities := make([]models.Entity, len(albums))
	for i, album := range albums {
		entities[i] = album
	}

	return entities
}

// GetNew возвращает новые альбомы (созданные за последний час)
func (s *MongoAlbumStore) GetNew() []models.Entity {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	since := time.Now().Add(-time.Hour)
	opts := &mongodb.AlbumListOptions{
		Filter: &mongodb.AlbumFilter{
			CreatedAt: &mongodb.TimeRange{
				From: &since,
			},
		},
		Limit: 100,
	}

	albums, err := s.storage.List(ctx, opts)
	if err != nil {
		log.Printf("Ошибка получения новых альбомов: %v", err)
		return []models.Entity{}
	}

	entities := make([]models.Entity, len(albums))
	for i, album := range albums {
		entities[i] = album
	}

	return entities
}

// Count возвращает количество альбомов
func (s *MongoAlbumStore) Count() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := s.storage.Count(ctx, nil)
	if err != nil {
		log.Printf("Ошибка подсчета альбомов: %v", err)
		return 0
	}

	return int(count)
}

// MongoPhotoStore заглушка для фотографий (пока не реализовано)
type MongoPhotoStore struct{}

func NewMongoPhotoStore(client *mongodb.Client) *MongoPhotoStore {
	return &MongoPhotoStore{}
}

func (s *MongoPhotoStore) Add(entity models.Entity) error {
	return fmt.Errorf("фото хранилище MongoDB пока не реализовано")
}

func (s *MongoPhotoStore) GetAll() []models.Entity {
	return []models.Entity{}
}

func (s *MongoPhotoStore) GetNew() []models.Entity {
	return []models.Entity{}
}

func (s *MongoPhotoStore) Count() int {
	return 0
}

// MongoTagStore заглушка для тегов (пока не реализовано)
type MongoTagStore struct{}

func NewMongoTagStore(client *mongodb.Client) *MongoTagStore {
	return &MongoTagStore{}
}

func (s *MongoTagStore) Add(entity models.Entity) error {
	return fmt.Errorf("тег хранилище MongoDB пока не реализовано")
}

func (s *MongoTagStore) GetAll() []models.Entity {
	return []models.Entity{}
}

func (s *MongoTagStore) GetNew() []models.Entity {
	return []models.Entity{}
}

func (s *MongoTagStore) Count() int {
	return 0
}
