package main

import (
	"log"
	"mpm/internal/handlers"
	"mpm/internal/storage"
	"net/http"
)

func main() {
	// Инициализация хранилища пользователей
	userStorage := storage.NewUserStorage()

	// Создание обработчика для пользователей
	userHandler := handlers.NewUserHandler(userStorage)

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
