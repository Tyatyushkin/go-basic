package handlers

import (
	"encoding/json"
	"log"
	"mpm/internal/models"
	"mpm/internal/storage"
	"net/http"
	"os"
	"path/filepath"
)

type UserHandler struct {
	userStorage *storage.JSONUserStorage
}

func NewUserHandler(userStorage *storage.JSONUserStorage) *UserHandler {
	return &UserHandler{
		userStorage: userStorage,
	}
}

// GetAllUsers godoc
// @Summary Получить всех пользователей
// @Description Получить список всех зарегистрированных пользователей
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} string "Внутренняя ошибка сервера"
// @Router /users [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос GET /api/users")

	// Добавляем диагностическую информацию
	dataPath := os.Getenv("MPM_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/mpm/data/users.json"
	}

	log.Printf("Путь к файлу пользователей: %s", dataPath)

	// Проверка существования директории
	dir := filepath.Dir(dataPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("ОШИБКА: Директория %s не существует", dir)
		http.Error(w, "Ошибка конфигурации сервера", http.StatusInternalServerError)
		return
	}
	log.Printf("Директория %s существует", dir)

	// Загружаем пользователей из хранилища
	users, err := h.userStorage.LoadUsers()
	if err != nil {
		log.Printf("ОШИБКА при загрузке пользователей: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если список пустой, создаем тестового пользователя для проверки
	if len(users) == 0 {
		log.Println("Список пользователей пуст, возвращаем тестовый набор данных")
		users = []models.User{
			{ID: 0, Username: "test", Email: "test@example.com"},
		}
	}

	log.Printf("Успешно загружено %d пользователей", len(users))

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Отправляем данные клиенту
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("ОШИБКА при сериализации JSON: %v", err)
		http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
		return
	}

	log.Println("Ответ успешно отправлен")
}
