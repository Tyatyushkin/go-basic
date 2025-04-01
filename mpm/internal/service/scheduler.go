package service

import (
	"fmt"
	"mpm/internal/models"
	"mpm/internal/repository"
	"sync"
	"time"
)

// EntityJob структура для передачи сущности и ее типа в горутину
type EntityJob struct {
	Entity models.Entity
	Type   string // Тип сущности для определения куда сохранять
}

// GenerateAndSaveEntities функция, которая создает разные структуры
// из internal/model и передает их в функцию слоя internal/repository
func GenerateAndSaveEntities() error {
	// Инициализируем хранилище, если оно еще не инициализировано
	repository.Initialize()

	entityChannel := make(chan EntityJob, 100)

	var wg sync.WaitGroup

	wg.Add(1)
	go generateEntities(entityChannel, &wg)

	// Запускаем несколько горутин для сохранения сущностей
	const numWorkers = 3
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go saveEntities(entityChannel, &wg, i)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()

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

// generateEntities генерирует различные сущности и отправляет их в канал
func generateEntities(entityChannel chan<- EntityJob, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(entityChannel) // Закрываем канал после генерации всех сущностей

	// Текущее время для установки в поля CreatedAt
	now := time.Now()

	// Создаем альбом по умолчанию
	defaultAlbum := models.Album{
		ID:          0,
		Name:        "Default",
		Description: "Альбом по умолчанию для всех фотографий",
		CreatedAt:   now,
	}

	// Отправляем альбом в канал
	entityChannel <- EntityJob{Entity: defaultAlbum, Type: "album"}

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

	// Отправляем теги в канал
	for _, tag := range basicTags {
		entityChannel <- EntityJob{Entity: tag, Type: "tag"}
	}

	fmt.Println("Генерация сущностей завершена")

}

// saveEntities получает сущности из канала и сохраняет их
func saveEntities(entityChannel <-chan EntityJob, wg *sync.WaitGroup, workerID int) {

}
