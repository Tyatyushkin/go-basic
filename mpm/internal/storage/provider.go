package storage

import (
	"io"
	"mime/multipart"
)

// Provider определяет общий интерфейс для работы с разными хранилищами файлов
type Provider interface {
	// Save сохраняет файл в хранилище и возвращает путь для доступа к нему
	Save(file multipart.File, filename string) (string, error)

	// Get извлекает файл по указанному пути
	Get(path string) ([]byte, error)

	// GetReader возвращает reader для потоковой передачи файла
	GetReader(path string) (io.ReadCloser, error)

	// Delete удаляет файл из хранилища
	Delete(path string) error

	// GetPublicURL возвращает публичную ссылку для доступа к файлу
	GetPublicURL(path string) string
}
