package repository

import (
	"context"
	"fmt"
	"log"
	"mpm/internal/models"
)

// Repository объединяет все хранилища сущностей
// Repository объединяет доступ к хранилищу сущностей
type Repository struct {
	storage EntityStorage
}

// NewRepository создает новый экземпляр репозитория
func NewRepository() *Repository {
	// Используем фабрику для создания хранилища нужного типа
	storage, err := CreateStorage()
	if err != nil {
		log.Printf("Ошибка при создании хранилища: %v", err)
		log.Println("Будет использовано JSON-хранилище по умолчанию")
		storage = NewJSONStorage("", 30) // Используем значения по умолчанию
	}

	// Загружаем данные при создании репозитория
	if err := storage.Load(); err != nil {
		log.Printf("Предупреждение: ошибка при загрузке данных: %v", err)
	}

	return &Repository{
		storage: storage,
	}
}

// InitStorage инициализирует хранилище и запускает автоматическое сохранение
// для JSON-хранилища. Для других типов хранилищ может выполнять другие действия.
func (r *Repository) InitStorage(ctx context.Context) {
	// Проверяем, поддерживает ли хранилище автоматическое сохранение
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		jsonStorage.StartAutoSave(ctx)
	}
}

// SaveEntities сохраняет сущности в хранилище
func (r *Repository) SaveEntities(entities []models.Entity) error {
	return r.storage.SaveBatch(entities)
}

// SaveEntity сохраняет одну сущность в хранилище
func (r *Repository) SaveEntity(entity models.Entity) error {
	return r.storage.Save(entity)
}

// PersistData принудительно сохраняет все данные
func (r *Repository) PersistData() error {
	return r.storage.Persist()
}

// LoadData загружает данные из хранилища
func (r *Repository) LoadData() error {
	return r.storage.Load()
}

// GetAllPhotos возвращает все фотографии
func (r *Repository) GetAllPhotos() []models.Photo {
	// Проверяем тип хранилища
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		return jsonStorage.GetPhotos()
	}

	// В будущем здесь будут проверки других типов хранилищ
	return []models.Photo{}
}

// GetAllAlbums возвращает все альбомы
func (r *Repository) GetAllAlbums() []models.Album {
	// Проверяем тип хранилища
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		return jsonStorage.GetAlbums()
	}

	// В будущем здесь будут проверки других типов хранилищ
	return []models.Album{}
}

// GetAllTags возвращает все теги
func (r *Repository) GetAllTags() []models.Tag {
	// Проверяем тип хранилища
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		return jsonStorage.GetTags()
	}

	// В будущем здесь будут проверки других типов хранилищ
	return []models.Tag{}
}

// GetEntitiesCounts возвращает количество сущностей каждого типа
func (r *Repository) GetEntitiesCounts() (photoCount, albumCount, tagCount int) {
	// Проверяем, может ли хранилище предоставить количество сущностей
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		return jsonStorage.GetCounts()
	}

	// Если хранилище другого типа, возвращаем нули
	// В будущем здесь будет проверка других типов хранилищ
	return 0, 0, 0
}

// GetNewEntities возвращает новые сущности с момента последнего вызова
func (r *Repository) GetNewEntities() (newPhotos []models.Photo, newAlbums []models.Album, newTags []models.Tag) {
	// Проверяем, может ли хранилище предоставить новые сущности
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		return jsonStorage.GetNewPhotos(), jsonStorage.GetNewAlbums(), jsonStorage.GetNewTags()
	}

	// Если хранилище другого типа, возвращаем пустые слайсы
	return []models.Photo{}, []models.Album{}, []models.Tag{}
}

// FindPhotoByID находит фотографию по ID
func (r *Repository) FindPhotoByID(id int) (models.Photo, error) {
	photos := r.GetAllPhotos()
	for _, photo := range photos {
		if photo.ID == id {
			return photo, nil
		}
	}
	return models.Photo{}, fmt.Errorf("фотография с ID=%d не найдена", id)
}

// FindAlbumByID находит альбом по ID
func (r *Repository) FindAlbumByID(id int) (models.Album, error) {
	albums := r.GetAllAlbums()
	for _, album := range albums {
		if album.ID == id {
			return album, nil
		}
	}
	return models.Album{}, fmt.Errorf("альбом с ID=%d не найден", id)
}

// FindTagByID находит тег по ID
func (r *Repository) FindTagByID(id int) (models.Tag, error) {
	tags := r.GetAllTags()
	for _, tag := range tags {
		if tag.ID == id {
			return tag, nil
		}
	}
	return models.Tag{}, fmt.Errorf("тег с ID=%d не найден", id)
}
