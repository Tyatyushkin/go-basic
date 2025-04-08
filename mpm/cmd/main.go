package main

import (
	"context"
	"errors"
	"log"
	"mpm/internal/handlers"
	"mpm/internal/repository"
	"mpm/internal/service"
	"mpm/internal/storage"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Выводим информацию о доступных переменных окружения
	log.Println("Доступные переменные окружения для настройки хранилища:")
	log.Println("MPM_STORAGE_TYPE - тип хранилища (json или postgres, по умолчанию json)")
	log.Println("MPM_DATA_PATH - путь к директории с данными для JSON-хранилища")

	// Создаем репозиторий
	repo := repository.NewRepository()
	log.Println("Репозиторий инициализирован")

	// Создаем контекст, который будет отменен при получении указанных сигналов
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// Освобождаем ресурсы после завершения
	defer stop()

	// Инициализируем хранилище и запускаем автоматическое сохранение
	repo.InitStorage(ctx)

	go func() {
		<-ctx.Done()
		log.Println("Получен сигнал, начинаем graceful shutdown")
	}()
	// Инициализация хранилища пользователей
	userStorage := storage.NewUserStorage()

	// Создание обработчика для пользователей
	userHandler := handlers.NewUserHandler(userStorage)

	entityService := service.NewEntityService(repo)

	// Запускаем мониторинг с контекстом
	entityService.StartMonitoring(ctx)

	// Вызываем функцию генерации и сохранения сущностей сразу
	err := entityService.GenerateAndSaveEntities(ctx)
	if err != nil {
		log.Printf("Ошибка при генерации и сохранении сущностей: %v", err)
		stop() // Вместо log.Fatal отменяем контекст
		return // И выходим из функции после выполнения всех defer
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Генерация сущностей остановлена")
				return
			case <-ticker.C:
				log.Println("Запланированная генерация статических сущностей")
				err := entityService.GenerateAndSaveEntities(ctx)
				if err != nil {
					log.Printf("Ошибка при генерации и сохранении сущностей: %v", err)
				}
			}
		}
	}()

	// Создание маршрутизатора
	mux := http.NewServeMux()

	// Регистрация маршрутов
	mux.HandleFunc("GET /api/users", userHandler.GetAllUsers)

	// Конфигурация сервера
	server := &http.Server{
		Addr:    ":8484",
		Handler: mux,
	}

	// Запуск сервера в отдельной горутине
	serverError := make(chan error, 1)
	go func() {
		log.Println("Запуск сервера на порту 8484...")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		log.Printf("Ошибка запуска сервера: %v", err)
		return
	case <-ctx.Done():
		// Получен сигнал для завершения работы
		log.Println("Начинаем graceful shutdown HTTP сервера...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Ошибка при graceful shutdown сервера: %v", err)
		}

		log.Println("HTTP сервер остановлен")
	}

}
