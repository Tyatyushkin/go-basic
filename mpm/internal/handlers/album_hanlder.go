package handlers

import (
	"encoding/json"
	"log"
	"mpm/internal/repository"
	"net/http"
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

//func (h *AlbumHandler) GetAlbumByID(w http.ResponseWriter, r *http.Request) {
//	log.Println("Получен запрос GET /api/albums/{id}")
//}

//func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
//	log.Println("Получен запрос DELETE /api/albums/{id}")
//}
