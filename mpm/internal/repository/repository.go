package repository

import (
	"context"
	"fmt"
	"log"
	"mpm/internal/models"
	"time"
)

// Repository объединяет все хранилища сущностей
// Repository объединяет доступ к хранилищу сущностей
type Repository struct {
	storage EntityStorage
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(storageType, dataDir string, saveInterval time.Duration) *Repository {
	// Используем фабрику для создания хранилища нужного типа
	storage, err := CreateStorage(storageType, dataDir, saveInterval)
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

// GetAllAlbums возвращает все альбомы, исключая дубликаты по ID
func (r *Repository) GetAllAlbums() []models.Album {
	var allAlbums []models.Album

	// Проверяем тип хранилища и получаем альбомы
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		allAlbums = jsonStorage.GetAlbums()
	} else {
		return []models.Album{} // Возвращаем пустой слайс для других типов хранилищ
	}

	// Используем map для исключения дубликатов по ID
	uniqueAlbums := make(map[int]models.Album)

	// Находим максимальный существующий ID
	maxID := 0
	for _, album := range allAlbums {
		if album.ID > maxID {
			maxID = album.ID
		}
	}

	// Генерируем новые ID для альбомов с ID=0 или дубликатами
	nextID := maxID + 1
	for _, album := range allAlbums {
		a := album // создаем копию альбома

		// Если ID=0 или такой ID уже есть в мапе и это не пустой альбом
		existingAlbum, exists := uniqueAlbums[a.ID]
		if a.ID == 0 || (exists && existingAlbum.ID != 0) {
			a.ID = nextID
			nextID++
		}

		uniqueAlbums[a.ID] = a
	}

	// Преобразуем map обратно в slice
	result := make([]models.Album, 0, len(uniqueAlbums))
	for _, album := range uniqueAlbums {
		result = append(result, album)
	}

	return result
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

// AddAlbum добавляет новый альбом с уникальным ID
func (r *Repository) AddAlbum(album models.Album) (int, error) {
	albums := r.GetAllAlbums()

	// Находим максимальный ID
	maxID := 0
	for _, a := range albums {
		if a.ID > maxID {
			maxID = a.ID
		}
	}

	// Всегда генерируем новый ID, даже если в запросе был указан ID
	album.ID = maxID + 1

	// Устанавливаем дату создания
	if album.CreatedAt.IsZero() {
		album.CreatedAt = time.Now()
	}

	// Добавляем альбом к существующим
	albums = append(albums, album)

	// Сохраняем обновленные данные
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		jsonStorage.albums = albums
		jsonStorage.albumsModified = true
		jsonStorage.dirtyFlag = true
		return album.ID, jsonStorage.Persist()
	}

	return album.ID, nil
}

// UpdateAlbum обновляет данные альбома по ID
func (r *Repository) UpdateAlbum(id int, updatedAlbum models.Album) error {
	// Получаем текущий список альбомов
	albums := r.GetAllAlbums()

	// Флаг для проверки, найден ли альбом
	found := false

	// Обновляем данные альбома
	for i, album := range albums {
		if album.ID == id {
			updatedAlbum.ID = id                     // Сохраняем ID
			updatedAlbum.CreatedAt = album.CreatedAt // Сохраняем дату создания
			albums[i] = updatedAlbum
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("альбом с ID=%d не найден", id)
	}

	// Сохраняем обновленный список альбомов
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		jsonStorage.albums = albums
		jsonStorage.albumsModified = true
		jsonStorage.dirtyFlag = true
		return jsonStorage.Persist()
	}

	return fmt.Errorf("обновление альбомов не поддерживается текущим хранилищем")
}

// Удалить альбом по ID
func (r *Repository) DeleteAlbum(id int) error {
	// Проверяем существование альбома
	_, err := r.FindAlbumByID(id)
	if err != nil {
		return err // Возвращаем ошибку, если альбом не найден
	}

	// Получаем текущий список альбомов
	albums := r.GetAllAlbums()

	// Создаем новый слайс без удаляемого альбома
	newAlbums := make([]models.Album, 0, len(albums))
	for _, album := range albums {
		if album.ID != id {
			newAlbums = append(newAlbums, album)
		}
	}

	// Обновляем состояние в JSON-хранилище
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		// Обновляем кэш альбомов и флаги
		jsonStorage.albums = newAlbums
		jsonStorage.albumsModified = true
		jsonStorage.dirtyFlag = true

		// Сохраняем изменения на диск
		return jsonStorage.Persist()
	}

	return fmt.Errorf("удаление альбомов не поддерживается текущим хранилищем")
}
