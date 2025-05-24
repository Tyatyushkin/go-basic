package repository

import (
	"context"
	"mpm/internal/models"
	"testing"
	"time"
)

func TestNewRepository(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		storageType  string
		dataDir      string
		saveInterval time.Duration
	}{
		{
			name:         "JSON storage",
			storageType:  "json",
			dataDir:      tempDir,
			saveInterval: 30 * time.Second,
		},
		{
			name:         "Default storage (empty type)",
			storageType:  "",
			dataDir:      tempDir,
			saveInterval: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewRepository(tt.storageType, tt.dataDir, tt.saveInterval)

			if repo == nil {
				t.Error("Expected repository to be created")
			}

			if repo.storage == nil {
				t.Error("Expected storage to be initialized")
			}
		})
	}
}

func TestRepository_SaveEntity(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	photo := models.Photo{
		ID:   1,
		Name: "test.jpg",
		Path: "/test/path.jpg",
	}

	err := repo.SaveEntity(photo)
	if err != nil {
		t.Errorf("SaveEntity() error = %v", err)
	}

	photos := repo.GetAllPhotos()
	if len(photos) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(photos))
	}

	if photos[0].ID != photo.ID {
		t.Errorf("Expected photo ID %d, got %d", photo.ID, photos[0].ID)
	}
}

func TestRepository_SaveEntities(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	entities := []models.Entity{
		models.Photo{ID: 1, Name: "photo1.jpg"},
		models.Album{ID: 1, Name: "Album 1"},
		models.Tag{ID: 1, Name: "tag1"},
	}

	err := repo.SaveEntities(entities)
	if err != nil {
		t.Errorf("SaveEntities() error = %v", err)
	}

	photos := repo.GetAllPhotos()
	if len(photos) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(photos))
	}

	ctx := context.Background()
	albums, err := repo.GetAllAlbums(ctx)
	if err != nil {
		t.Errorf("GetAllAlbums() error = %v", err)
	}
	if len(albums) < 1 {
		t.Errorf("Expected at least 1 album, got %d", len(albums))
	}

	tags := repo.GetAllTags()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
}

func TestRepository_GetAllPhotos(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	photo1 := models.Photo{ID: 1, Name: "photo1.jpg"}
	photo2 := models.Photo{ID: 2, Name: "photo2.jpg"}

	repo.SaveEntity(photo1)
	repo.SaveEntity(photo2)

	photos := repo.GetAllPhotos()
	if len(photos) != 2 {
		t.Errorf("Expected 2 photos, got %d", len(photos))
	}
}

func TestRepository_GetAllAlbums(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	album1 := models.Album{ID: 1, Name: "Album 1"}
	album2 := models.Album{ID: 2, Name: "Album 2"}
	defaultAlbum := models.Album{ID: 0, Name: "Default", Description: "Альбом по умолчанию для всех фотографий"}

	repo.SaveEntity(album1)
	repo.SaveEntity(album2)
	repo.SaveEntity(defaultAlbum)

	albums, err := repo.GetAllAlbums(ctx)
	if err != nil {
		t.Errorf("GetAllAlbums() error = %v", err)
	}

	if len(albums) != 3 {
		t.Errorf("Expected 3 albums, got %d", len(albums))
	}

	defaultFound := false
	for _, album := range albums {
		if album.Name == "Default" && album.ID == 0 {
			defaultFound = true
			break
		}
	}

	if !defaultFound {
		t.Error("Expected default album with ID 0")
	}
}

func TestRepository_GetAllAlbums_WithDuplicateDefaults(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	defaultAlbum1 := models.Album{ID: 1, Name: "Default", Description: "Альбом по умолчанию для всех фотографий"}
	defaultAlbum2 := models.Album{ID: 2, Name: "Default", Description: "Альбом по умолчанию для всех фотографий"}
	regularAlbum := models.Album{ID: 3, Name: "Regular Album"}

	repo.SaveEntity(defaultAlbum1)
	repo.SaveEntity(defaultAlbum2)
	repo.SaveEntity(regularAlbum)

	albums, err := repo.GetAllAlbums(ctx)
	if err != nil {
		t.Errorf("GetAllAlbums() error = %v", err)
	}

	defaultCount := 0
	for _, album := range albums {
		if album.Name == "Default" && album.Description == "Альбом по умолчанию для всех фотографий" {
			defaultCount++
		}
	}

	if defaultCount != 1 {
		t.Errorf("Expected exactly 1 default album, got %d", defaultCount)
	}
}

func TestRepository_GetAllAlbums_WithContext(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetAllAlbums(ctx)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestRepository_GetAllTags(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	tag1 := models.Tag{ID: 1, Name: "tag1"}
	tag2 := models.Tag{ID: 2, Name: "tag2"}

	repo.SaveEntity(tag1)
	repo.SaveEntity(tag2)

	tags := repo.GetAllTags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}

func TestRepository_GetEntitiesCounts(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	repo.SaveEntity(models.Photo{ID: 1, Name: "photo1.jpg"})
	repo.SaveEntity(models.Photo{ID: 2, Name: "photo2.jpg"})
	repo.SaveEntity(models.Album{ID: 1, Name: "Album 1"})
	repo.SaveEntity(models.Tag{ID: 1, Name: "tag1"})

	photoCount, albumCount, tagCount := repo.GetEntitiesCounts()

	if photoCount != 2 {
		t.Errorf("Expected 2 photos, got %d", photoCount)
	}

	if albumCount != 1 {
		t.Errorf("Expected 1 album, got %d", albumCount)
	}

	if tagCount != 1 {
		t.Errorf("Expected 1 tag, got %d", tagCount)
	}
}

func TestRepository_GetNewEntities(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	repo.SaveEntity(models.Photo{ID: 1, Name: "photo1.jpg"})
	repo.SaveEntity(models.Album{ID: 1, Name: "Album 1"})
	repo.SaveEntity(models.Tag{ID: 1, Name: "tag1"})

	newPhotos, newAlbums, newTags := repo.GetNewEntities()

	if len(newPhotos) != 1 {
		t.Errorf("Expected 1 new photo, got %d", len(newPhotos))
	}

	if len(newAlbums) != 1 {
		t.Errorf("Expected 1 new album, got %d", len(newAlbums))
	}

	if len(newTags) != 1 {
		t.Errorf("Expected 1 new tag, got %d", len(newTags))
	}

	newPhotos2, newAlbums2, newTags2 := repo.GetNewEntities()

	if len(newPhotos2) != 0 {
		t.Errorf("Expected 0 new photos on second call, got %d", len(newPhotos2))
	}

	if len(newAlbums2) != 0 {
		t.Errorf("Expected 0 new albums on second call, got %d", len(newAlbums2))
	}

	if len(newTags2) != 0 {
		t.Errorf("Expected 0 new tags on second call, got %d", len(newTags2))
	}
}

func TestRepository_FindPhotoByID(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	repo.SaveEntity(photo)

	foundPhoto, err := repo.FindPhotoByID(1)
	if err != nil {
		t.Errorf("FindPhotoByID() error = %v", err)
	}

	if foundPhoto.ID != photo.ID {
		t.Errorf("Expected photo ID %d, got %d", photo.ID, foundPhoto.ID)
	}

	_, err = repo.FindPhotoByID(999)
	if err == nil {
		t.Error("Expected error when finding non-existent photo")
	}
}

func TestRepository_FindAlbumByID(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	album := models.Album{ID: 1, Name: "Test Album"}
	repo.SaveEntity(album)

	foundAlbum, err := repo.FindAlbumByID(ctx, 1)
	if err != nil {
		t.Errorf("FindAlbumByID() error = %v", err)
	}

	if foundAlbum.ID != album.ID {
		t.Errorf("Expected album ID %d, got %d", album.ID, foundAlbum.ID)
	}

	_, err = repo.FindAlbumByID(ctx, 999)
	if err == nil {
		t.Error("Expected error when finding non-existent album")
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.FindAlbumByID(canceledCtx, 1)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestRepository_FindTagByID(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	tag := models.Tag{ID: 1, Name: "test-tag"}
	repo.SaveEntity(tag)

	foundTag, err := repo.FindTagByID(1)
	if err != nil {
		t.Errorf("FindTagByID() error = %v", err)
	}

	if foundTag.ID != tag.ID {
		t.Errorf("Expected tag ID %d, got %d", tag.ID, foundTag.ID)
	}

	_, err = repo.FindTagByID(999)
	if err == nil {
		t.Error("Expected error when finding non-existent tag")
	}
}

func TestRepository_AddAlbum(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	album := models.Album{
		Name:        "New Album",
		Description: "Test Description",
	}

	newID, err := repo.AddAlbum(ctx, album)
	if err != nil {
		t.Errorf("AddAlbum() error = %v", err)
	}

	if newID <= 0 {
		t.Errorf("Expected positive ID, got %d", newID)
	}

	foundAlbum, err := repo.FindAlbumByID(ctx, newID)
	if err != nil {
		t.Errorf("Failed to find added album: %v", err)
	}

	if foundAlbum.Name != album.Name {
		t.Errorf("Expected album name %s, got %s", album.Name, foundAlbum.Name)
	}

	if foundAlbum.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.AddAlbum(canceledCtx, album)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestRepository_UpdateAlbum(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	album := models.Album{ID: 1, Name: "Original Album", CreatedAt: time.Now()}
	repo.SaveEntity(album)

	updatedAlbum := models.Album{
		Name:        "Updated Album",
		Description: "Updated Description",
	}

	err := repo.UpdateAlbum(ctx, 1, updatedAlbum)
	if err != nil {
		t.Errorf("UpdateAlbum() error = %v", err)
	}

	foundAlbum, err := repo.FindAlbumByID(ctx, 1)
	if err != nil {
		t.Errorf("Failed to find updated album: %v", err)
	}

	if foundAlbum.Name != updatedAlbum.Name {
		t.Errorf("Expected updated name %s, got %s", updatedAlbum.Name, foundAlbum.Name)
	}

	if foundAlbum.ID != 1 {
		t.Errorf("Expected ID to remain 1, got %d", foundAlbum.ID)
	}

	if foundAlbum.CreatedAt != album.CreatedAt {
		t.Error("Expected CreatedAt to be preserved")
	}

	err = repo.UpdateAlbum(ctx, 999, updatedAlbum)
	if err == nil {
		t.Error("Expected error when updating non-existent album")
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.UpdateAlbum(canceledCtx, 1, updatedAlbum)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestRepository_DeleteAlbum(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)
	ctx := context.Background()

	album := models.Album{ID: 1, Name: "Test Album"}
	repo.SaveEntity(album)

	err := repo.DeleteAlbum(ctx, 1)
	if err != nil {
		t.Errorf("DeleteAlbum() error = %v", err)
	}

	_, err = repo.FindAlbumByID(ctx, 1)
	if err == nil {
		t.Error("Expected error when finding deleted album")
	}

	err = repo.DeleteAlbum(ctx, 999)
	if err == nil {
		t.Error("Expected error when deleting non-existent album")
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.DeleteAlbum(canceledCtx, 1)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestRepository_PersistData(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	repo.SaveEntity(photo)

	err := repo.PersistData()
	if err != nil {
		t.Errorf("PersistData() error = %v", err)
	}
}

func TestRepository_LoadData(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, time.Hour)

	err := repo.LoadData()
	if err != nil {
		t.Errorf("LoadData() error = %v", err)
	}
}

func TestRepository_InitStorage(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewRepository("json", tempDir, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	repo.InitStorage(ctx)

	time.Sleep(150 * time.Millisecond)
}
