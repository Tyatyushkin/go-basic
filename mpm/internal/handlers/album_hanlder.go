package handlers

import (
	"encoding/json"
	"log"
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

//func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
//	log.Println("Получен запрос POST /api/albums")
//
//}

//func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
//	log.Println("Получен запрос PUT /api/albums/{id}")
//
//}

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

//func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
//	log.Println("Получен запрос DELETE /api/albums/{id}")
//}
