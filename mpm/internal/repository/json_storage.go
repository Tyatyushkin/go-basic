package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mpm/internal/models"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONStorage реализует интерфейс EntityStorage для временного хранения в JSON-файлах
type JSONStorage struct {
	dataDir string // Директория для хранения JSON-файлов

	// Отдельные мьютексы для каждого типа данных
	photosMutex sync.RWMutex // Мьютекс для доступа к фотографиям
	albumsMutex sync.RWMutex // Мьютекс для доступа к альбомам
	tagsMutex   sync.RWMutex // Мьютекс для доступа к тегам

	// Общий мьютекс для метаданных (dirtyFlag, lastSaveTime)
	metaMutex sync.RWMutex

	dirtyFlag    bool          // Флаг наличия несохраненных изменений
	lastSaveTime time.Time     // Время последнего сохранения
	saveInterval time.Duration // Интервал между автоматическими сохранениями

	// Хранилища данных
	photos []models.Photo
	albums []models.Album
	tags   []models.Tag

	// Счетчики для определения новых сущностей
	lastPhotoIndex int
	lastAlbumIndex int
	lastTagIndex   int
}

// NewJSONStorage создает новое хранилище с сохранением в JSON
func NewJSONStorage(dataDir string, saveInterval time.Duration) *JSONStorage {

	return &JSONStorage{
		dataDir:      dataDir,
		saveInterval: saveInterval,
		photos:       make([]models.Photo, 0),
		albums:       make([]models.Album, 0),
		tags:         make([]models.Tag, 0),
		lastSaveTime: time.Now(),
	}
}

// Save Сохраняет одну сущность в JSON хранилище
func (s *JSONStorage) Save(entity models.Entity) error {
	// Сначала блокируем метаданные для изменения флага dirtyFlag
	s.metaMutex.Lock()
	s.dirtyFlag = true
	s.metaMutex.Unlock()

	// Определяем тип сущности и добавляем в соответствующий слайс
	switch e := entity.(type) {
	case models.Photo:
		s.photosMutex.Lock()
		s.photos = append(s.photos, e)
		s.photosMutex.Unlock()
		log.Printf("Добавлена фотография: ID=%d, Название=%s", e.ID, e.Name)

	case models.Album:
		s.albumsMutex.Lock()
		s.albums = append(s.albums, e)
		s.albumsMutex.Unlock()
		log.Printf("Добавлен альбом: ID=%d, Название=%s", e.ID, e.Name)

	case models.Tag:
		s.tagsMutex.Lock()
		s.tags = append(s.tags, e)
		s.tagsMutex.Unlock()
		log.Printf("Добавлен тег: ID=%d, Название=%s", e.ID, e.Name)

	default:
		return fmt.Errorf("неизвестный тип сущности: %T", entity)
	}

	// Проверяем, нужно ли сохранить данные
	s.metaMutex.RLock()
	needSafe := time.Since(s.lastSaveTime) > s.saveInterval
	s.metaMutex.RUnlock()
	if needSafe {
		return s.persistData()
	}

	return nil
}

// SaveBatch сохраняет несколько сущностей
func (s *JSONStorage) SaveBatch(entities []models.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.dirtyFlag = true

	// Группируем сущности по типу
	var photos []models.Photo
	var albums []models.Album
	var tags []models.Tag

	for _, entity := range entities {
		switch e := entity.(type) {
		case models.Photo:
			photos = append(photos, e)
		case models.Album:
			albums = append(albums, e)
		case models.Tag:
			tags = append(tags, e)
		default:
			log.Printf("Предупреждение: неизвестный тип сущности: %T", entity)
		}
	}

	// Добавляем сущности в соответствующие слайсы
	if len(photos) > 0 {
		s.photos = append(s.photos, photos...)
		log.Printf("Добавлено %d фотографий", len(photos))
	}

	if len(albums) > 0 {
		s.albums = append(s.albums, albums...)
		log.Printf("Добавлено %d альбомов", len(albums))
	}

	if len(tags) > 0 {
		s.tags = append(s.tags, tags...)
		log.Printf("Добавлено %d тегов", len(tags))
	}

	// Проверяем, нужно ли сохранить данные
	if time.Since(s.lastSaveTime) > s.saveInterval {
		return s.persistData()
	}

	return nil
}

// Load загружает данные из JSON-файлов
func (s *JSONStorage) Load() error {
	// TODO добавить мьютексы

	// Убеждаемся, что директория существует
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("ошибка при создании директории данных: %v", err)
	}

	// Загружаем фотографии
	photosPath := filepath.Join(s.dataDir, "photos.json")
	s.photosMutex.Lock()
	photosErr := s.loadFile(photosPath, &s.photos)
	s.photosMutex.Unlock()
	if photosErr != nil {
		return fmt.Errorf("ошибка при загрузке фотографий: %v", err)
	}

	// Загружаем альбомы
	albumsPath := filepath.Join(s.dataDir, "albums.json")
	if err := s.loadFile(albumsPath, &s.albums); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка при загрузке альбомов: %v", err)
	}

	// Загружаем теги
	tagsPath := filepath.Join(s.dataDir, "tags.json")
	if err := s.loadFile(tagsPath, &s.tags); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка при загрузке тегов: %v", err)
	}

	// Устанавливаем индексы для отслеживания новых сущностей
	s.lastPhotoIndex = len(s.photos)
	s.lastAlbumIndex = len(s.albums)
	s.lastTagIndex = len(s.tags)

	log.Printf("Загружено: %d фотографий, %d альбомов, %d тегов",
		len(s.photos), len(s.albums), len(s.tags))

	return nil
}

// Persist сохраняет текущее состояние в JSON-файлы
func (s *JSONStorage) Persist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.persistData()
}

// loadFile вспомогательная функция для загрузки данных из файла
func (s *JSONStorage) loadFile(filePath string, target interface{}) error {
	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Файл не найден, будет создан новый: %s", filePath)
		return nil
	}

	// Читаем содержимое файла
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Проверка на пустой файл
	if len(data) == 0 {
		return nil
	}

	// Десериализуем JSON
	return json.Unmarshal(data, target)
}

// persistData внутренняя функция для сохранения данных
func (s *JSONStorage) persistData() error {
	if !s.dirtyFlag {
		return nil // Нет изменений для сохранения
	}

	// Убеждаемся, что директория существует
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("ошибка при создании директории данных: %v", err)
	}

	// Сохраняем фотографии (всегда, даже если пустые)
	photosPath := filepath.Join(s.dataDir, "photos.json")
	if err := s.saveFile(photosPath, s.photos); err != nil {
		return fmt.Errorf("ошибка при сохранении фотографий: %v", err)
	}

	// Сохраняем альбомы (всегда, даже если пустые)
	albumsPath := filepath.Join(s.dataDir, "albums.json")
	if err := s.saveFile(albumsPath, s.albums); err != nil {
		return fmt.Errorf("ошибка при сохранении альбомов: %v", err)
	}

	// Сохраняем теги (всегда, даже если пустые)
	tagsPath := filepath.Join(s.dataDir, "tags.json")
	if err := s.saveFile(tagsPath, s.tags); err != nil {
		return fmt.Errorf("ошибка при сохранении тегов: %v", err)
	}

	s.dirtyFlag = false
	s.lastSaveTime = time.Now()
	log.Printf("Данные успешно сохранены в %s", s.dataDir)

	return nil
}

// saveFile вспомогательная функция для сохранения данных в файл
func (s *JSONStorage) saveFile(filePath string, data interface{}) error {
	// Сериализуем данные в JSON с отступами для удобства чтения
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Записываем данные в файл
	return os.WriteFile(filePath, jsonData, 0644)
}

// GetPhotos возвращает копию всех фотографий
func (s *JSONStorage) GetPhotos() []models.Photo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Photo, len(s.photos))
	copy(result, s.photos)
	return result
}

// GetAlbums возвращает копию всех альбомов
func (s *JSONStorage) GetAlbums() []models.Album {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Album, len(s.albums))
	copy(result, s.albums)
	return result
}

// GetTags возвращает копию всех тегов
func (s *JSONStorage) GetTags() []models.Tag {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]models.Tag, len(s.tags))
	copy(result, s.tags)
	return result
}

// GetNewPhotos возвращает новые фотографии с момента последнего вызова
func (s *JSONStorage) GetNewPhotos() []models.Photo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastPhotoIndex >= len(s.photos) {
		return []models.Photo{}
	}

	result := make([]models.Photo, len(s.photos)-s.lastPhotoIndex)
	copy(result, s.photos[s.lastPhotoIndex:])
	s.lastPhotoIndex = len(s.photos)
	return result
}

// GetNewAlbums возвращает новые альбомы с момента последнего вызова
func (s *JSONStorage) GetNewAlbums() []models.Album {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastAlbumIndex >= len(s.albums) {
		return []models.Album{}
	}

	result := make([]models.Album, len(s.albums)-s.lastAlbumIndex)
	copy(result, s.albums[s.lastAlbumIndex:])
	s.lastAlbumIndex = len(s.albums)
	return result
}

// GetNewTags возвращает новые теги с момента последнего вызова
func (s *JSONStorage) GetNewTags() []models.Tag {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastTagIndex >= len(s.tags) {
		return []models.Tag{}
	}

	result := make([]models.Tag, len(s.tags)-s.lastTagIndex)
	copy(result, s.tags[s.lastTagIndex:])
	s.lastTagIndex = len(s.tags)
	return result
}

// GetCounts возвращает количество сущностей каждого типа
func (s *JSONStorage) GetCounts() (photosCount, albumsCount, tagsCount int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.photos), len(s.albums), len(s.tags)
}

// StartAutoSave запускает автоматическое сохранение данных с заданным интервалом
func (s *JSONStorage) StartAutoSave(ctx context.Context) {
	ticker := time.NewTicker(s.saveInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				// Финальное сохранение при завершении
				if s.dirtyFlag {
					_ = s.Persist()
				}
				return
			case <-ticker.C:
				if s.dirtyFlag {
					if err := s.Persist(); err != nil {
						log.Printf("Ошибка при автоматическом сохранении: %v", err)
					}
				}
			}
		}
	}()

	log.Printf("Запущено автоматическое сохранение с интервалом %v", s.saveInterval)
}
