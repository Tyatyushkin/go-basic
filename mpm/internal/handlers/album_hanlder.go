package handlers

import "mpm/internal/repository"

type AlbumHandler struct {
	repo *repository.Repository
}

func NewAlbumHandler(repo *repository.Repository) *AlbumHandler {
	return &AlbumHandler{
		repo: repo,
	}
}
