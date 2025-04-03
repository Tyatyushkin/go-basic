package main

import (
	"log"
	"mpm/internal/handlers"
	"mpm/internal/repository"
	"mpm/internal/service"
	"mpm/internal/storage"
	"net/http"
	"time"
)

func main() {
	// Создаем репозиторий
	repo := repository.NewRepository()
	// Инициализация хранилища пользователей
	userStorage := storage.NewUserStorage()

	// Создание обработчика для пользователей
	userHandler := handlers.NewUserHandler(userStorage)

	entityService := service.NewEntityService(repo)

	// Вызываем функцию генерации и сохранения сущностей сразу
	err := entityService.GenerateAndSaveEntities()
	if err != nil {
		log.Fatalf("Ошибка при генерации и сохранении сущностей: %v", err)
	}

	// Запускаем планировщик для периодической генерации сущностей
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Запланированная генерация статических сущностей")
			// Используем экземпляр сервиса вместо глобальной функции:
			err := entityService.GenerateAndSaveEntities()
			if err != nil {
				log.Printf("Ошибка при генерации и сохранении сущностей: %v", err)
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

	// Запуск сервера
	log.Println("Запуск сервера на порту 8484...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

}
