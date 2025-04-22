package main

import (
	"context"
	"errors"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	_ "mpm/docs"
	"mpm/internal/handlers"
	"mpm/internal/repository"
	"mpm/internal/service"
	"mpm/internal/storage"
	"mpm/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title MPM API
// @version 1.0
// @description API для Masterplan Photo Manager
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://tyatyushkin.ru
// @contact.email maxim.tyatyushkin@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host tyatyushkin.ru:8484
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}

func main() {
	// Выводим информацию о доступных переменных окружения
	log.Println("Доступные переменные окружения для настройки хранилища:")
	log.Println("MPM_STORAGE_TYPE - тип хранилища (json или postgres, по умолчанию json)")
	log.Println("MPM_DATA_PATH - путь к директории с данными для JSON-хранилища")

	// Получаем настройки из переменных окружения
	storageType := os.Getenv("MPM_STORAGE_TYPE")
	dataDir := os.Getenv("MPM_DATA_PATH")
	if dataDir == "" {
		dataDir = "/opt/mpm/data" // Значение по умолчанию
	}
	saveInterval := 30 * time.Second
	if intervalStr := os.Getenv("MPM_SAVE_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			saveInterval = interval
		}
	}

	// Создаем репозиторий
	repo := repository.NewRepository(storageType, dataDir, saveInterval)

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

	// Создание обработчика для альбомов
	albumHandler := handlers.NewAlbumHandler(repo)
	entityService := service.NewEntityService(repo)

	// Создание сервиса аутентификации
	authService := service.NewAuthService(userStorage)
	authHandler := handlers.NewAuthHandler(authService)

	// Middlewares
	authMiddleware := middleware.AuthMiddleware(authService)

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

	// Защищенные маршруты (с аутентификацией)
	// Оберните группу защищенных маршрутов
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /api/users", userHandler.GetAllUsers)
	authMux.HandleFunc("POST /api/albums", albumHandler.CreateAlbum)
	authMux.HandleFunc("PUT /api/albums/{id}", albumHandler.UpdateAlbum)
	authMux.HandleFunc("GET /api/albums", albumHandler.GetAllAlbums)
	authMux.HandleFunc("GET /api/albums/{id}", albumHandler.GetAlbumByID)
	authMux.HandleFunc("DELETE /api/albums/{id}", albumHandler.DeleteAlbum)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
	))

	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// Регистрация защищенных маршрутов с middleware
	mux.Handle("/api/", authMiddleware(authMux))

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
