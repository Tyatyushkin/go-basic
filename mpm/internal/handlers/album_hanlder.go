package handlers

import (
	"encoding/json"
	"log"
	"mpm/internal/models"
	"mpm/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

type AlbumHandler struct {
	repo *repository.Repository
}

func NewAlbumHandler(repo *repository.Repository) *AlbumHandler {
	return &AlbumHandler{
		repo: repo,
	}
}

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос POST /api/albums")

	// Декодируем тело запроса в структуру альбома
	var album models.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Добавление альбома через репозиторий
	id, err := h.repo.AddAlbum(album)
	if err != nil {
		log.Printf("Ошибка при создании альбома: %v", err)
		http.Error(w, "Ошибка при создании альбома", http.StatusInternalServerError)
		return
	}

	// Получаем альбом с присвоенным ID
	newAlbum, err := h.repo.FindAlbumByID(id)
	if err != nil {
		log.Printf("Ошибка при получении созданного альбома: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type и статус
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Сериализуем созданный альбом в JSON и отправляем клиенту
	if err := json.NewEncoder(w).Encode(newAlbum); err != nil {
		log.Printf("Ошибка при сериализации альбома: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Успешно создан альбом с ID=%d", newAlbum.ID)
}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос PUT /api/albums/{id}")

	// Извлекаем ID из пути запроса
	var idStr string
	idStr = r.URL.Query().Get("id")
	if idStr == "" {
		parts := splitPath(r.URL.Path)
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}

	if idStr == "" {
		http.Error(w, "ID альбома не указан", http.StatusBadRequest)
		return
	}

	// Преобразуем ID в целое число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса в структуру альбома
	var updatedAlbum models.Album
	if err := json.NewDecoder(r.Body).Decode(&updatedAlbum); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Обновляем альбом через репозиторий
	if err := h.repo.UpdateAlbum(id, updatedAlbum); err != nil {
		if strings.Contains(err.Error(), "не найден") {
			http.Error(w, "Альбом не найден", http.StatusNotFound)
		} else {
			log.Printf("Ошибка при обновлении альбома: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем успешный статус
	w.WriteHeader(http.StatusOK)
	log.Printf("Успешно обновлен альбом с ID=%d", id)
}

func (h *AlbumHandler) GetAllAlbums(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос GET /api/albums")

	// Получаем все альбомы из репозитория
	albums := h.repo.GetAllAlbums()

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Сериализуем альбомы в JSON и отправляем клиенту
	if err := json.NewEncoder(w).Encode(albums); err != nil {
		log.Printf("Ошибка при сериализации альбомов: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Успешно отправлены данные о %d альбомах", len(albums))

}

func (h *AlbumHandler) GetAlbumByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос GET /api/albums/{id}")

	// Извлекаем ID из пути запроса
	// Предполагаем, что ID передается как последний сегмент пути или как параметр запроса
	var idStr string

	// Пробуем получить из параметра запроса
	idStr = r.URL.Query().Get("id")

	// Если ID нет в параметрах, пробуем извлечь из пути
	if idStr == "" {
		parts := splitPath(r.URL.Path)
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}

	if idStr == "" {
		http.Error(w, "ID альбома не указан", http.StatusBadRequest)
		return
	}

	// Преобразуем ID в целое число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	// Ищем альбом в репозитории
	album, err := h.repo.FindAlbumByID(id)
	if err != nil {
		http.Error(w, "Альбом не найден", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Сериализуем альбом в JSON и отправляем клиенту
	if err := json.NewEncoder(w).Encode(album); err != nil {
		log.Printf("Ошибка при сериализации альбома: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	log.Printf("Успешно отправлены данные об альбоме с ID=%d", id)
}

// splitPath разделяет URL-путь на сегменты
func splitPath(path string) []string {
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос DELETE /api/albums/{id}")

	// Извлекаем ID из пути запроса
	var idStr string

	// Пробуем получить из параметра запроса
	idStr = r.URL.Query().Get("id")

	// Если ID нет в параметрах, пробуем извлечь из пути
	if idStr == "" {
		parts := splitPath(r.URL.Path)
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}

	if idStr == "" {
		http.Error(w, "ID альбома не указан", http.StatusBadRequest)
		return
	}

	// Преобразуем ID в целое число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
		return
	}

	// Удаляем альбом из репозитория
	err = h.repo.DeleteAlbum(id)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			http.Error(w, "Альбом не найден", http.StatusNotFound)
		} else {
			log.Printf("Ошибка при удалении альбома: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем успешный статус без тела ответа
	w.WriteHeader(http.StatusNoContent)
	log.Printf("Успешно удален альбом с ID=%d", id)
}
