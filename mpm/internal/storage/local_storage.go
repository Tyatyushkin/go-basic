package storage

import (
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
)

// LocalStorage реализует работу с локальной файловой системой
type LocalStorage struct {
	BasePath string // Базовый путь для хранения файлов
	BaseURL  string // Базовый URL для доступа к файлам
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// Создаем директорию, если не существует
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		os.MkdirAll(basePath, 0755)
	}

	return &LocalStorage{
		BasePath: basePath,
		BaseURL:  baseURL,
	}
}

func (ls *LocalStorage) Save(file multipart.File, filename string) (string, error) {
	// Создаем полный путь для сохранения файла
	fullPath := filepath.Join(ls.BasePath, filename)

	// Создаем директорию для файла, если она не существует
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	// Создаем файл
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Копируем содержимое входного файла в новый файл
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Возвращаем относительный путь к файлу
	return filename, nil
}

func (ls *LocalStorage) Get(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(ls.BasePath, path))
}

func (ls *LocalStorage) GetReader(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(ls.BasePath, path))
}

func (ls *LocalStorage) Delete(path string) error {
	return os.Remove(filepath.Join(ls.BasePath, path))
}

func (ls *LocalStorage) GetPublicURL(path string) string {
	return ls.BaseURL + "/" + path
}
