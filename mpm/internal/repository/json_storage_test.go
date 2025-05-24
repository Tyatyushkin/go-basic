package repository

import (
	"context"
	"mpm/internal/models"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewJSONStorage(t *testing.T) {
	tempDir := t.TempDir()
	saveInterval := time.Second

	storage := NewJSONStorage(tempDir, saveInterval)

	if storage.dataDir != tempDir {
		t.Errorf("Expected dataDir %s, got %s", tempDir, storage.dataDir)
	}
	if storage.saveInterval != saveInterval {
		t.Errorf("Expected saveInterval %v, got %v", saveInterval, storage.saveInterval)
	}
	if len(storage.photos) != 0 {
		t.Errorf("Expected empty photos slice, got %d items", len(storage.photos))
	}
}

func TestJSONStorage_SavePhoto(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour) // Большой интервал чтобы не триггерить автосохранение

	photo := models.Photo{
		ID:   1,
		Name: "test.jpg",
		Path: "/path/to/test.jpg",
	}

	err := storage.Save(photo)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	photos := storage.GetPhotos()
	if len(photos) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(photos))
	}
	if photos[0].Name != "test.jpg" {
		t.Errorf("Expected name 'test.jpg', got '%s'", photos[0].Name)
	}
}

func TestJSONStorage_SaveAlbum(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	album := models.Album{
		ID:   1,
		Name: "Test Album",
	}

	err := storage.Save(album)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	albums := storage.GetAlbums()
	if len(albums) != 1 {
		t.Errorf("Expected 1 album, got %d", len(albums))
	}
	if albums[0].Name != "Test Album" {
		t.Errorf("Expected name 'Test Album', got '%s'", albums[0].Name)
	}
}

func TestJSONStorage_SaveTag(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	tag := models.Tag{
		ID:   1,
		Name: "nature",
	}

	err := storage.Save(tag)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	tags := storage.GetTags()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "nature" {
		t.Errorf("Expected name 'nature', got '%s'", tags[0].Name)
	}
}

func TestJSONStorage_SaveBatch(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	entities := []models.Entity{
		models.Photo{ID: 1, Name: "photo1.jpg"},
		models.Photo{ID: 2, Name: "photo2.jpg"},
		models.Album{ID: 1, Name: "Album 1"},
		models.Tag{ID: 1, Name: "tag1"},
	}

	err := storage.SaveBatch(entities)
	if err != nil {
		t.Fatalf("SaveBatch failed: %v", err)
	}

	photosCount, albumsCount, tagsCount := storage.GetCounts()
	if photosCount != 2 {
		t.Errorf("Expected 2 photos, got %d", photosCount)
	}
	if albumsCount != 1 {
		t.Errorf("Expected 1 album, got %d", albumsCount)
	}
	if tagsCount != 1 {
		t.Errorf("Expected 1 tag, got %d", tagsCount)
	}
}

func TestJSONStorage_PersistAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	// Добавляем тестовые данные
	photo := models.Photo{ID: 1, Name: "test.jpg"}
	album := models.Album{ID: 1, Name: "Test Album"}
	tag := models.Tag{ID: 1, Name: "nature"}

	storage.Save(photo)
	storage.Save(album)
	storage.Save(tag)

	// Принудительно сохраняем
	err := storage.Persist()
	if err != nil {
		t.Fatalf("Persist failed: %v", err)
	}

	// Проверяем что файлы созданы
	photosFile := filepath.Join(tempDir, "photos.json")
	albumsFile := filepath.Join(tempDir, "albums.json")
	tagsFile := filepath.Join(tempDir, "tags.json")

	if _, err := os.Stat(photosFile); os.IsNotExist(err) {
		t.Error("photos.json file was not created")
	}
	if _, err := os.Stat(albumsFile); os.IsNotExist(err) {
		t.Error("albums.json file was not created")
	}
	if _, err := os.Stat(tagsFile); os.IsNotExist(err) {
		t.Error("tags.json file was not created")
	}

	// Создаем новое хранилище и загружаем данные
	newStorage := NewJSONStorage(tempDir, time.Hour)
	err = newStorage.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Проверяем что данные загрузились
	photosCount, albumsCount, tagsCount := newStorage.GetCounts()
	if photosCount != 1 {
		t.Errorf("Expected 1 photo after load, got %d", photosCount)
	}
	if albumsCount != 1 {
		t.Errorf("Expected 1 album after load, got %d", albumsCount)
	}
	if tagsCount != 1 {
		t.Errorf("Expected 1 tag after load, got %d", tagsCount)
	}
}

func TestJSONStorage_GetNewEntities(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	// Добавляем первую партию
	storage.Save(models.Photo{ID: 1, Name: "photo1.jpg"})
	storage.Save(models.Album{ID: 1, Name: "Album 1"})

	// Получаем новые (должны быть все)
	newPhotos := storage.GetNewPhotos()
	newAlbums := storage.GetNewAlbums()
	newTags := storage.GetNewTags()

	if len(newPhotos) != 1 {
		t.Errorf("Expected 1 new photo, got %d", len(newPhotos))
	}
	if len(newAlbums) != 1 {
		t.Errorf("Expected 1 new album, got %d", len(newAlbums))
	}
	if len(newTags) != 0 {
		t.Errorf("Expected 0 new tags, got %d", len(newTags))
	}

	// Добавляем еще
	storage.Save(models.Photo{ID: 2, Name: "photo2.jpg"})
	storage.Save(models.Tag{ID: 1, Name: "nature"})

	// Получаем новые (должны быть только добавленные)
	newPhotos = storage.GetNewPhotos()
	newAlbums = storage.GetNewAlbums()
	newTags = storage.GetNewTags()

	if len(newPhotos) != 1 {
		t.Errorf("Expected 1 new photo, got %d", len(newPhotos))
	}
	if len(newAlbums) != 0 {
		t.Errorf("Expected 0 new albums, got %d", len(newAlbums))
	}
	if len(newTags) != 1 {
		t.Errorf("Expected 1 new tag, got %d", len(newTags))
	}
}

func TestJSONStorage_AutoSave(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Millisecond*100) // Короткий интервал для теста

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем автосохранение
	storage.StartAutoSave(ctx)

	// Добавляем данные
	photo := models.Photo{ID: 1, Name: "test.jpg"}
	storage.Save(photo)

	// Ждем автосохранения
	time.Sleep(time.Millisecond * 200)

	// Проверяем что файл создан
	photosFile := filepath.Join(tempDir, "photos.json")
	if _, err := os.Stat(photosFile); os.IsNotExist(err) {
		t.Error("AutoSave did not create photos.json file")
	}

	// Отменяем контекст
	cancel()
	time.Sleep(time.Millisecond * 50) // Даем время для завершения горутины
}
