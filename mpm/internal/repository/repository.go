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

// GetAllAlbums возвращает все альбомы, исключая дубликаты и оставляя только один дефолтный альбом
func (r *Repository) GetAllAlbums(ctx context.Context) ([]models.Album, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

	}

	var allAlbums []models.Album

	// Получаем альбомы из хранилища
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		allAlbums = jsonStorage.GetAlbums()
	} else {
		return []models.Album{}, nil
	}

	// Отдельно обрабатываем дефолтные альбомы
	var defaultAlbum *models.Album
	var otherAlbums []models.Album

	// Разделяем дефолтные и обычные альбомы
	for _, album := range allAlbums {
		if album.Name == "Default" && album.Description == "Альбом по умолчанию для всех фотографий" {
			// Сохраняем только первый найденный дефолтный альбом
			if defaultAlbum == nil {
				copyAlbum := album
				defaultAlbum = &copyAlbum
			}
		} else {
			otherAlbums = append(otherAlbums, album)
		}
	}

	// Добавляем дефолтный альбом, если он был найден
	processedAlbums := []models.Album{}
	if defaultAlbum != nil {
		// Устанавливаем ID дефолтного альбома равным 0
		defaultAlbum.ID = 0
		processedAlbums = append(processedAlbums, *defaultAlbum)
	}

	// Добавляем остальные альбомы
	processedAlbums = append(processedAlbums, otherAlbums...)

	// Обрабатываем все альбомы для исключения дубликатов по ID
	uniqueAlbums := make(map[int]models.Album)

	// Находим максимальный существующий ID
	maxID := 0
	for _, album := range processedAlbums {
		// Пропускаем дефолтный альбом при поиске максимального ID
		if album.ID == 0 && album.Name == "Default" {
			continue
		}

		if album.ID > maxID {
			maxID = album.ID
		}
	}

	// Добавляем дефолтный альбом в map
	if defaultAlbum != nil {
		uniqueAlbums[0] = *defaultAlbum
	}

	// Генерируем новые ID для альбомов с дублирующимися ID
	nextID := maxID + 1
	for _, album := range processedAlbums {
		// Пропускаем дефолтный альбом, который уже добавлен
		if album.ID == 0 && album.Name == "Default" {
			continue
		}

		a := album
		if a.ID == 0 || uniqueAlbums[a.ID].ID != 0 {
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

	// Обновляем хранилище, чтобы исправления сохранились
	if jsonStorage, ok := r.storage.(*JSONStorage); ok {
		jsonStorage.albums = result
		jsonStorage.albumsModified = true
		jsonStorage.dirtyFlag = true
		err := jsonStorage.Persist()
		if err != nil {
			return nil, err
		}
	}

	return result, nil
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
func (r *Repository) FindAlbumByID(ctx context.Context, id int) (models.Album, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return models.Album{}, ctx.Err()
	default:
		// Продолжаем выполнение
	}

	albums, err := r.GetAllAlbums(ctx)
	if err != nil {
		return models.Album{}, err
	}

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
func (r *Repository) AddAlbum(ctx context.Context, album models.Album) (int, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// Продолжаем выполнение
	}

	albums, err := r.GetAllAlbums(ctx)
	if err != nil {
		return 0, err
	}

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
func (r *Repository) UpdateAlbum(ctx context.Context, id int, updatedAlbum models.Album) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Продолжаем выполнение
	}

	// Получаем текущий список альбомов
	albums, err := r.GetAllAlbums(ctx)
	if err != nil {
		return err // Исправлено: возвращаем только ошибку
	}

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

// DeleteAlbum Удалить альбом по ID
func (r *Repository) DeleteAlbum(ctx context.Context, id int) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Продолжаем выполнение
	}

	// Проверяем существование альбома - исправлено передачей контекста
	_, err := r.FindAlbumByID(ctx, id)
	if err != nil {
		return err
	}

	// Получаем текущий список альбомов
	albums, err := r.GetAllAlbums(ctx)
	if err != nil {
		return err
	}

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
