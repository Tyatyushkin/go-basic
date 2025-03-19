package service

import (
	"fmt"
	"mpm/internal/models"
	"mpm/internal/repository"
	"time"
)

// GenerateAndSaveEntities функция, которая создает разные структуры
// из internal/model и передает их в функцию слоя internal/repository
func GenerateAndSaveEntities() error {
	// Инициализируем хранилище, если оно еще не инициализировано
	repository.Initialize()

	// Текущее время для установки в поля CreatedAt
	now := time.Now()

	// Создаем альбом по умолчанию
	defaultAlbum := models.Album{
		ID:          0,
		Name:        "Default",
		Description: "Альбом по умолчанию для всех фотографий",
		CreatedAt:   now,
	}

	// Создаем базовые теги для фотографий
	basicTags := []models.Entity{
		models.Tag{ID: 1, Name: "природа", CreatedAt: now},
		models.Tag{ID: 2, Name: "закат", CreatedAt: now},
		models.Tag{ID: 3, Name: "портрет", CreatedAt: now},
		models.Tag{ID: 4, Name: "архитектура", CreatedAt: now},
		models.Tag{ID: 5, Name: "пейзаж", CreatedAt: now},
		models.Tag{ID: 6, Name: "город", CreatedAt: now},
		models.Tag{ID: 7, Name: "животные", CreatedAt: now},
		models.Tag{ID: 8, Name: "макро", CreatedAt: now},
		models.Tag{ID: 9, Name: "чб", CreatedAt: now},
		models.Tag{ID: 10, Name: "street", CreatedAt: now},
		models.Tag{ID: 11, Name: "люди", CreatedAt: now},
		models.Tag{ID: 12, Name: "путешествия", CreatedAt: now},
		models.Tag{ID: 13, Name: "еда", CreatedAt: now},
		models.Tag{ID: 14, Name: "море", CreatedAt: now},
		models.Tag{ID: 15, Name: "горы", CreatedAt: now},
	}

	// Создаем слайс разных сущностей
	entities := []models.Entity{
		//Добавляем альбом по умолчанию
		defaultAlbum,
	}

	// Добавляем базовые теги в общий слайс сущностей
	entities = append(entities, basicTags...)

	// Передаем сущности в функцию репозитория
	err := repository.SaveEntities(entities)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении сущностей: %v", err)
	}

	// Для демонстрации выводим количество сущностей в каждом слайсе
	fmt.Printf("Всего фотографий: %d\n", len(repository.GetPhotos()))
	fmt.Printf("Всего альбомов: %d\n", len(repository.GetAlbums()))
	fmt.Printf("Всего тегов: %d\n", len(repository.GetTags()))

	return nil
}
