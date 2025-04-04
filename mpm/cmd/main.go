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
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Создаем репозиторий
	repo := repository.NewRepository()
	// Создаем контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для получения сигналов от ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Получен сигнал: %v, начинаем graceful shutdown", sig)
		cancel() // Отменяем контекст для корректного завершения всех горутин
	}()

	// Инициализация хранилища пользователей
	userStorage := storage.NewUserStorage()

	// Создание обработчика для пользователей
	userHandler := handlers.NewUserHandler(userStorage)

	entityService := service.NewEntityService(repo)

	// Запускаем мониторинг с контекстом
	entityService.StartMonitoring(ctx)

	// Вызываем функцию генерации и сохранения сущностей сразу
	err := entityService.GenerateAndSaveEntities()
	if err != nil {
		log.Fatalf("Ошибка при генерации и сохранении сущностей: %v", err)
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
				err := entityService.GenerateAndSaveEntities()
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
		log.Fatalf("Ошибка запуска сервера: %v", err)
	case <-ctx.Done():
		// Получен сигнал для завершения работы
		log.Println("Начинаем graceful shutdown HTTP сервера...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Ошибка при graceful shutdown сервера: %v", err)
		}

		log.Println("HTTP сервер остановлен")
	}

}
