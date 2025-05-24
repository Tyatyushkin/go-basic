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
	saveInterval := 30 * time.Second

	storage := NewJSONStorage(tempDir, saveInterval)

	if storage.dataDir != tempDir {
		t.Errorf("Expected dataDir to be %s, got %s", tempDir, storage.dataDir)
	}

	if storage.saveInterval != saveInterval {
		t.Errorf("Expected saveInterval to be %v, got %v", saveInterval, storage.saveInterval)
	}

	if storage.photos == nil {
		t.Error("Expected photos slice to be initialized")
	}

	if storage.albums == nil {
		t.Error("Expected albums slice to be initialized")
	}

	if storage.tags == nil {
		t.Error("Expected tags slice to be initialized")
	}
}

func TestJSONStorage_Save(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	tests := []struct {
		name   string
		entity models.Entity
	}{
		{
			name: "Save Photo",
			entity: models.Photo{
				ID:   1,
				Name: "test.jpg",
				Path: "/test/path.jpg",
			},
		},
		{
			name: "Save Album",
			entity: models.Album{
				ID:          1,
				Name:        "Test Album",
				Description: "Test Description",
				CreatedAt:   time.Now(),
			},
		},
		{
			name: "Save Tag",
			entity: models.Tag{
				ID:   1,
				Name: "test-tag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Save(tt.entity)
			if err != nil {
				t.Errorf("Save() error = %v", err)
			}

			switch e := tt.entity.(type) {
			case models.Photo:
				if len(storage.photos) != 1 {
					t.Errorf("Expected 1 photo, got %d", len(storage.photos))
				}
				if storage.photos[0].ID != e.ID {
					t.Errorf("Expected photo ID %d, got %d", e.ID, storage.photos[0].ID)
				}
			case models.Album:
				if len(storage.albums) != 1 {
					t.Errorf("Expected 1 album, got %d", len(storage.albums))
				}
				if storage.albums[0].ID != e.ID {
					t.Errorf("Expected album ID %d, got %d", e.ID, storage.albums[0].ID)
				}
			case models.Tag:
				if len(storage.tags) != 1 {
					t.Errorf("Expected 1 tag, got %d", len(storage.tags))
				}
				if storage.tags[0].ID != e.ID {
					t.Errorf("Expected tag ID %d, got %d", e.ID, storage.tags[0].ID)
				}
			}

			if !storage.dirtyFlag {
				t.Error("Expected dirtyFlag to be true after Save")
			}

			storage.photos = nil
			storage.albums = nil
			storage.tags = nil
			storage.dirtyFlag = false
		})
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
		models.Tag{ID: 2, Name: "tag2"},
	}

	err := storage.SaveBatch(entities)
	if err != nil {
		t.Errorf("SaveBatch() error = %v", err)
	}

	if len(storage.photos) != 2 {
		t.Errorf("Expected 2 photos, got %d", len(storage.photos))
	}

	if len(storage.albums) != 1 {
		t.Errorf("Expected 1 album, got %d", len(storage.albums))
	}

	if len(storage.tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(storage.tags))
	}

	if !storage.dirtyFlag {
		t.Error("Expected dirtyFlag to be true after SaveBatch")
	}
}

func TestJSONStorage_SaveBatch_Empty(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	err := storage.SaveBatch([]models.Entity{})
	if err != nil {
		t.Errorf("SaveBatch() with empty slice error = %v", err)
	}

	if storage.dirtyFlag {
		t.Error("Expected dirtyFlag to be false after SaveBatch with empty slice")
	}
}

func TestJSONStorage_LoadAndPersist(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	album := models.Album{ID: 1, Name: "Test Album"}
	tag := models.Tag{ID: 1, Name: "test-tag"}

	err := storage.Save(photo)
	if err != nil {
		t.Errorf("Save photo error = %v", err)
	}

	err = storage.Save(album)
	if err != nil {
		t.Errorf("Save album error = %v", err)
	}

	err = storage.Save(tag)
	if err != nil {
		t.Errorf("Save tag error = %v", err)
	}

	err = storage.Persist()
	if err != nil {
		t.Errorf("Persist() error = %v", err)
	}

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

	newStorage := NewJSONStorage(tempDir, time.Hour)
	err = newStorage.Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if len(newStorage.photos) != 1 {
		t.Errorf("Expected 1 photo after load, got %d", len(newStorage.photos))
	}

	if len(newStorage.albums) != 1 {
		t.Errorf("Expected 1 album after load, got %d", len(newStorage.albums))
	}

	if len(newStorage.tags) != 1 {
		t.Errorf("Expected 1 tag after load, got %d", len(newStorage.tags))
	}

	loadedPhoto := newStorage.photos[0]
	if loadedPhoto.ID != photo.ID || loadedPhoto.Name != photo.Name {
		t.Errorf("Loaded photo doesn't match saved photo")
	}
}

func TestJSONStorage_GetMethods(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	photo1 := models.Photo{ID: 1, Name: "photo1.jpg"}
	photo2 := models.Photo{ID: 2, Name: "photo2.jpg"}
	album1 := models.Album{ID: 1, Name: "Album 1"}
	tag1 := models.Tag{ID: 1, Name: "tag1"}

	storage.Save(photo1)
	storage.Save(album1)
	storage.Save(tag1)

	photos := storage.GetPhotos()
	if len(photos) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(photos))
	}

	albums := storage.GetAlbums()
	if len(albums) != 1 {
		t.Errorf("Expected 1 album, got %d", len(albums))
	}

	tags := storage.GetTags()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	photosCount, albumsCount, tagsCount := storage.GetCounts()
	if photosCount != 1 || albumsCount != 1 || tagsCount != 1 {
		t.Errorf("Expected counts (1,1,1), got (%d,%d,%d)", photosCount, albumsCount, tagsCount)
	}

	storage.Save(photo2)

	newPhotos := storage.GetNewPhotos()
	if len(newPhotos) != 1 {
		t.Errorf("Expected 1 new photo, got %d", len(newPhotos))
	}

	if newPhotos[0].ID != photo2.ID {
		t.Errorf("Expected new photo ID %d, got %d", photo2.ID, newPhotos[0].ID)
	}

	newPhotosAgain := storage.GetNewPhotos()
	if len(newPhotosAgain) != 0 {
		t.Errorf("Expected 0 new photos on second call, got %d", len(newPhotosAgain))
	}
}

func TestJSONStorage_StartAutoSave(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	storage.Save(photo)

	storage.StartAutoSave(ctx)

	time.Sleep(200 * time.Millisecond)

	photosFile := filepath.Join(tempDir, "photos.json")
	if _, err := os.Stat(photosFile); os.IsNotExist(err) {
		t.Error("Auto-save did not create photos.json file")
	}
}

func TestJSONStorage_loadFile_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	var photos []models.Photo
	err := storage.loadFile(filepath.Join(tempDir, "nonexistent.json"), &photos)
	if err != nil {
		t.Errorf("loadFile() with non-existent file should not error, got %v", err)
	}

	if len(photos) != 0 {
		t.Errorf("Expected empty slice for non-existent file, got %d items", len(photos))
	}
}

func TestJSONStorage_loadFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	emptyFile := filepath.Join(tempDir, "empty.json")
	err := os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	var photos []models.Photo
	err = storage.loadFile(emptyFile, &photos)
	if err != nil {
		t.Errorf("loadFile() with empty file should not error, got %v", err)
	}

	if len(photos) != 0 {
		t.Errorf("Expected empty slice for empty file, got %d items", len(photos))
	}
}

type UnknownEntity struct {
	ID int
}

func (u UnknownEntity) GetID() int {
	return u.ID
}

func (u UnknownEntity) GetType() string {
	return "unknown"
}

func TestJSONStorage_Save_UnknownType(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	unknown := UnknownEntity{ID: 1}
	err := storage.Save(unknown)
	if err == nil {
		t.Error("Expected error when saving unknown entity type")
	}

	expectedError := "неизвестный тип сущности"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONStorage_persistData_NotDirty(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewJSONStorage(tempDir, time.Hour)

	storage.dirtyFlag = false

	err := storage.persistData()
	if err != nil {
		t.Errorf("persistData() with clean state should not error, got %v", err)
	}
}
