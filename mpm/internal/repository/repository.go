package repository

import (
	"fmt"
	"mpm/internal/models"
)

// Repository объединяет все хранилища сущностей
type Repository struct {
	photoStore *PhotoStore
	albumStore *AlbumStore
	tagStore   *TagStore
}

// NewRepository создаёт новый экземпляр репозитория
func NewRepository() *Repository {
	return &Repository{
		photoStore: NewPhotoStore(),
		albumStore: NewAlbumStore(),
		tagStore:   NewTagStore(),
	}
}

// Photos возвращает хранилище фотографий
func (r *Repository) Photos() *PhotoStore {
	return r.photoStore
}

// Albums возвращает хранилище альбомов
func (r *Repository) Albums() *AlbumStore {
	return r.albumStore
}

// Tags возвращает хранилище тегов
func (r *Repository) Tags() *TagStore {
	return r.tagStore
}

// SaveEntities функция, которая распределяет сущности по соответствующим хранилищам
func (r *Repository) SaveEntities(entities []models.Entity) error {
	for _, entity := range entities {
		// Проверяем тип сущности с помощью type assertion
		switch e := entity.(type) {
		case models.Photo:
			if err := r.photoStore.Add(e); err != nil {
				return err
			}

		case models.Album:
			if err := r.albumStore.Add(e); err != nil {
				return err
			}

		case models.Tag:
			if err := r.tagStore.Add(e); err != nil {
				return err
			}

		default:
			return fmt.Errorf("неизвестный тип сущности: %T", entity)
		}
	}

	return nil
}

// GetEntitiesCounts возвращает количество сущностей каждого типа
func (r *Repository) GetEntitiesCounts() (photoCount, albumCount, tagCount int) {
	return r.photoStore.Count(), r.albumStore.Count(), r.tagStore.Count()
}

// GetNewEntities возвращает новые сущности с момента последнего вызова
func (r *Repository) GetNewEntities() (newPhotos []models.Photo, newAlbums []models.Album, newTags []models.Tag) {
	return r.photoStore.GetNew(), r.albumStore.GetNew(), r.tagStore.GetNew()
}
