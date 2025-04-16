package handlers

import (
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

func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос POST /api/albums")

}

func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен запрос PUT /api/albums/{id}")

}
