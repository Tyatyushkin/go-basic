package service

import (
	"context"
	"fmt"
	"log"
	"mpm/internal/models"
	"mpm/internal/repository"
	"sync"
	"time"
)

// EntityService служит для работы с сущностями через репозиторий
type EntityService struct {
	repo              *repository.Repository
	monitoringStarted sync.Once
	mutex             sync.Mutex
}

// NewEntityService создает новый сервис для работы с сущностями
func NewEntityService(repo *repository.Repository) *EntityService {
	return &EntityService{
		repo: repo,
	}
}

// EntityJob структура для передачи сущности и ее типа в горутину
type EntityJob struct {
	Entity models.Entity
	Type   string // Тип сущности для определения куда сохранять
}

// GenerateAndSaveEntities генерирует и сохраняет сущности в репозитории
func (s *EntityService) GenerateAndSaveEntities() error {
	entityChannel := make(chan EntityJob, 100)

	var wg sync.WaitGroup

	wg.Add(1)
	go s.generateEntities(entityChannel, &wg)

	// Запускаем несколько горутин для сохранения сущностей
	const numWorkers = 3
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go s.saveEntities(entityChannel, &wg, i)
	}
	// Ожидаем завершения всех горутин
	wg.Wait()

	// Получаем количество сущностей
	photoCount, albumCount, tagCount := s.repo.GetEntitiesCounts()
	fmt.Printf("Всего фотографий: %d\n", photoCount)
	fmt.Printf("Всего альбомов: %d\n", albumCount)
	fmt.Printf("Всего тегов: %d\n", tagCount)

	return nil
}

// StartMonitoring запускает мониторинг сущностей с поддержкой отмены через контекст
func (s *EntityService) StartMonitoring(ctx context.Context) {
	go s.monitorEntities(ctx)
}

// monitorEntities следит за изменениями в репозитории и логирует новые сущности
func (s *EntityService) monitorEntities(ctx context.Context) {
	log.Println("Запуск мониторинга сущностей")

	// Начальное количество элементов
	lastPhotoCount, lastAlbumCount, lastTagCount := s.repo.GetEntitiesCounts()

	// Создаем тикер для периодической проверки
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Получаем текущее количество элементов
		currentPhotoCount, currentAlbumCount, currentTagCount := s.repo.GetEntitiesCounts()

		// Проверяем, были ли добавлены новые элементы
		if currentPhotoCount > lastPhotoCount ||
			currentAlbumCount > lastAlbumCount ||
			currentTagCount > lastTagCount {

			// Получаем новые элементы
			newPhotos, newAlbums, newTags := s.repo.GetNewEntities()

			// Логируем новые фотографии
			for _, photo := range newPhotos {
				log.Printf("МОНИТОР: Обнаружена новая фотография - ID: %d, Название: %s",
					photo.ID, photo.Name)
			}

			// Логируем новые альбомы
			for _, album := range newAlbums {
				log.Printf("МОНИТОР: Обнаружен новый альбом - ID: %d, Название: %s",
					album.ID, album.Name)
			}

			// Логируем новые теги
			for _, tag := range newTags {
				log.Printf("МОНИТОР: Обнаружен новый тег - ID: %d, Название: %s",
					tag.ID, tag.Name)
			}

			// Обновляем счетчики
			lastPhotoCount = currentPhotoCount
			lastAlbumCount = currentAlbumCount
			lastTagCount = currentTagCount
		}
	}
}

// generateEntities генерирует различные сущности и отправляет их в канал
func (s *EntityService) generateEntities(entityChannel chan<- EntityJob, wg *sync.WaitGroup) {
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
func (s *EntityService) saveEntities(entityChannel <-chan EntityJob, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	// Создаем отдельные слайсы для каждого типа сущностей
	var photos []models.Entity
	var albums []models.Entity
	var tags []models.Entity

	// Обрабатываем сущности из канала
	for job := range entityChannel {
		switch job.Type {
		case "photo":
			photos = append(photos, job.Entity)
			fmt.Printf("Worker %d: получена фотография ID=%d\n", workerID, job.Entity.GetID())
		case "album":
			albums = append(albums, job.Entity)
			fmt.Printf("Worker %d: получен альбом ID=%d\n", workerID, job.Entity.GetID())
		case "tag":
			tags = append(tags, job.Entity)
			fmt.Printf("Worker %d: получен тег ID=%d\n", workerID, job.Entity.GetID())
		default:
			fmt.Printf("Worker %d: неизвестный тип сущности: %s\n", workerID, job.Type)
		}
	}

	// Сохраняем каждый тип сущностей пакетно
	if len(photos) > 0 {
		s.mutex.Lock()
		err := s.repo.SaveEntities(photos)
		s.mutex.Unlock()
		if err != nil {
			return
		}
		fmt.Printf("Worker %d: сохранено %d фотографий\n", workerID, len(photos))
	}

	if len(albums) > 0 {
		s.mutex.Lock()
		err := s.repo.SaveEntities(albums)
		s.mutex.Unlock()
		if err != nil {
			return
		}
		fmt.Printf("Worker %d: сохранено %d альбомов\n", workerID, len(albums))
	}

	if len(tags) > 0 {
		s.mutex.Lock()
		err := s.repo.SaveEntities(tags)
		s.mutex.Unlock()
		if err != nil {
			return
		}
		fmt.Printf("Worker %d: сохранено %d тегов\n", workerID, len(tags))
	}
}
