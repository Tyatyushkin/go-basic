package repository

import (
	"mpm/internal/models"
	"sync"
	"testing"
	"time"
)

func TestNewPhotoStore(t *testing.T) {
	store := NewPhotoStore()

	if store == nil {
		t.Error("Expected PhotoStore to be created")
		return
	}

	if store.items == nil {
		t.Error("Expected items slice to be initialized")
		return
	}

	if len(store.items) != 0 {
		t.Errorf("Expected empty items slice, got %d items", len(store.items))
	}

	if store.lastIndex != 0 {
		t.Errorf("Expected lastIndex to be 0, got %d", store.lastIndex)
	}
}

func TestPhotoStore_Add(t *testing.T) {
	store := NewPhotoStore()

	photo := models.Photo{
		ID:   1,
		Name: "test.jpg",
		Path: "/test/path.jpg",
	}

	err := store.Add(photo)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	if len(store.items) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(store.items))
	}

	if store.items[0].ID != photo.ID {
		t.Errorf("Expected photo ID %d, got %d", photo.ID, store.items[0].ID)
	}

	if store.items[0].Name != photo.Name {
		t.Errorf("Expected photo name %s, got %s", photo.Name, store.items[0].Name)
	}
}

func TestPhotoStore_AddEntity(t *testing.T) {
	store := NewPhotoStore()

	photo := models.Photo{ID: 1, Name: "test.jpg"}

	err := store.AddEntity(photo)
	if err != nil {
		t.Errorf("AddEntity() error = %v", err)
	}

	if len(store.items) != 1 {
		t.Errorf("Expected 1 photo, got %d", len(store.items))
	}

	wrongType := models.Album{ID: 1, Name: "Album"}
	err = store.AddEntity(wrongType)
	if err == nil {
		t.Error("Expected error when adding wrong entity type")
		return
	}

	expectedError := "неверный тип сущности для PhotoStore"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPhotoStore_GetAll(t *testing.T) {
	store := NewPhotoStore()

	photo1 := models.Photo{ID: 1, Name: "photo1.jpg"}
	photo2 := models.Photo{ID: 2, Name: "photo2.jpg"}

	_ = store.Add(photo1)
	_ = store.Add(photo2)

	photos := store.GetAll()

	if len(photos) != 2 {
		t.Errorf("Expected 2 photos, got %d", len(photos))
	}

	photos[0].Name = "modified"
	if store.items[0].Name == "modified" {
		t.Error("GetAll() should return a copy, not reference to original slice")
	}
}

func TestPhotoStore_GetAllEntities(t *testing.T) {
	store := NewPhotoStore()

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	_ = store.Add(photo)

	entities := store.GetAllEntities()

	if len(entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(entities))
	}

	photoEntity, ok := entities[0].(models.Photo)
	if !ok {
		t.Errorf("Expected Photo entity, got %T", entities[0])
	}

	if photoEntity.ID != photo.ID {
		t.Errorf("Expected photo ID %d, got %d", photo.ID, photoEntity.ID)
	}
}

func TestPhotoStore_GetNew(t *testing.T) {
	store := NewPhotoStore()

	photo1 := models.Photo{ID: 1, Name: "photo1.jpg"}
	photo2 := models.Photo{ID: 2, Name: "photo2.jpg"}

	_ = store.Add(photo1)

	newPhotos := store.GetNew()
	if len(newPhotos) != 1 {
		t.Errorf("Expected 1 new photo, got %d", len(newPhotos))
	}

	if newPhotos[0].ID != photo1.ID {
		t.Errorf("Expected new photo ID %d, got %d", photo1.ID, newPhotos[0].ID)
	}

	newPhotos2 := store.GetNew()
	if len(newPhotos2) != 0 {
		t.Errorf("Expected 0 new photos on second call, got %d", len(newPhotos2))
	}

	_ = store.Add(photo2)

	newPhotos3 := store.GetNew()
	if len(newPhotos3) != 1 {
		t.Errorf("Expected 1 new photo after adding another, got %d", len(newPhotos3))
	}

	if newPhotos3[0].ID != photo2.ID {
		t.Errorf("Expected new photo ID %d, got %d", photo2.ID, newPhotos3[0].ID)
	}
}

func TestPhotoStore_GetNewEntities(t *testing.T) {
	store := NewPhotoStore()

	photo := models.Photo{ID: 1, Name: "test.jpg"}
	_ = store.Add(photo)

	newEntities := store.GetNewEntities()

	if len(newEntities) != 1 {
		t.Errorf("Expected 1 new entity, got %d", len(newEntities))
	}

	photoEntity, ok := newEntities[0].(models.Photo)
	if !ok {
		t.Errorf("Expected Photo entity, got %T", newEntities[0])
	}

	if photoEntity.ID != photo.ID {
		t.Errorf("Expected photo ID %d, got %d", photo.ID, photoEntity.ID)
	}
}

func TestPhotoStore_Count(t *testing.T) {
	store := NewPhotoStore()

	if store.Count() != 0 {
		t.Errorf("Expected count 0, got %d", store.Count())
	}

	_ = store.Add(models.Photo{ID: 1, Name: "photo1.jpg"})
	_ = store.Add(models.Photo{ID: 2, Name: "photo2.jpg"})

	if store.Count() != 2 {
		t.Errorf("Expected count 2, got %d", store.Count())
	}
}

func TestPhotoStore_Concurrency(t *testing.T) {
	store := NewPhotoStore()
	var wg sync.WaitGroup

	numGoroutines := 10
	photosPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < photosPerGoroutine; j++ {
				photo := models.Photo{
					ID:   goroutineID*photosPerGoroutine + j + 1,
					Name: "photo.jpg",
				}
				_ = store.Add(photo)
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * photosPerGoroutine
	if store.Count() != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, store.Count())
	}
}

func TestNewAlbumStore(t *testing.T) {
	store := NewAlbumStore()

	if store == nil {
		t.Error("Expected AlbumStore to be created")
		return
	}

	if store.items == nil {
		t.Error("Expected items slice to be initialized")
		return
	}

	if len(store.items) != 0 {
		t.Errorf("Expected empty items slice, got %d items", len(store.items))
	}
}

func TestAlbumStore_Add(t *testing.T) {
	store := NewAlbumStore()

	album := models.Album{
		ID:          1,
		Name:        "Test Album",
		Description: "Test Description",
		CreatedAt:   time.Now(),
	}

	err := store.Add(album)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	if len(store.items) != 1 {
		t.Errorf("Expected 1 album, got %d", len(store.items))
	}

	if store.items[0].ID != album.ID {
		t.Errorf("Expected album ID %d, got %d", album.ID, store.items[0].ID)
	}
}

func TestAlbumStore_AddEntity(t *testing.T) {
	store := NewAlbumStore()

	album := models.Album{ID: 1, Name: "Test Album"}

	err := store.AddEntity(album)
	if err != nil {
		t.Errorf("AddEntity() error = %v", err)
	}

	wrongType := models.Photo{ID: 1, Name: "Photo"}
	err = store.AddEntity(wrongType)
	if err == nil {
		t.Error("Expected error when adding wrong entity type")
		return
	}
}

func TestAlbumStore_GetAll(t *testing.T) {
	store := NewAlbumStore()

	album1 := models.Album{ID: 1, Name: "Album 1"}
	album2 := models.Album{ID: 2, Name: "Album 2"}

	_ = store.Add(album1)
	_ = store.Add(album2)

	albums := store.GetAll()

	if len(albums) != 2 {
		t.Errorf("Expected 2 albums, got %d", len(albums))
	}

	albums[0].Name = "modified"
	if store.items[0].Name == "modified" {
		t.Error("GetAll() should return a copy, not reference to original slice")
	}
}

func TestNewTagStore(t *testing.T) {
	store := NewTagStore()

	if store == nil {
		t.Error("Expected TagStore to be created")
		return
	}

	if store.items == nil {
		t.Error("Expected items slice to be initialized")
		return
	}

	if len(store.items) != 0 {
		t.Errorf("Expected empty items slice, got %d items", len(store.items))
	}
}

func TestTagStore_Add(t *testing.T) {
	store := NewTagStore()

	tag := models.Tag{
		ID:   1,
		Name: "test-tag",
	}

	err := store.Add(tag)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	if len(store.items) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(store.items))
	}

	if store.items[0].ID != tag.ID {
		t.Errorf("Expected tag ID %d, got %d", tag.ID, store.items[0].ID)
	}
}

func TestTagStore_AddEntity(t *testing.T) {
	store := NewTagStore()

	tag := models.Tag{ID: 1, Name: "test-tag"}

	err := store.AddEntity(tag)
	if err != nil {
		t.Errorf("AddEntity() error = %v", err)
	}

	wrongType := models.Photo{ID: 1, Name: "Photo"}
	err = store.AddEntity(wrongType)
	if err == nil {
		t.Error("Expected error when adding wrong entity type")
		return
	}
}

func TestTagStore_GetAll(t *testing.T) {
	store := NewTagStore()

	tag1 := models.Tag{ID: 1, Name: "tag1"}
	tag2 := models.Tag{ID: 2, Name: "tag2"}

	_ = store.Add(tag1)
	_ = store.Add(tag2)

	tags := store.GetAll()

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	tags[0].Name = "modified"
	if store.items[0].Name == "modified" {
		t.Error("GetAll() should return a copy, not reference to original slice")
	}
}

func TestTagStore_GetNew(t *testing.T) {
	store := NewTagStore()

	tag1 := models.Tag{ID: 1, Name: "tag1"}
	tag2 := models.Tag{ID: 2, Name: "tag2"}

	_ = store.Add(tag1)

	newTags := store.GetNew()
	if len(newTags) != 1 {
		t.Errorf("Expected 1 new tag, got %d", len(newTags))
	}

	newTags2 := store.GetNew()
	if len(newTags2) != 0 {
		t.Errorf("Expected 0 new tags on second call, got %d", len(newTags2))
	}

	_ = store.Add(tag2)

	newTags3 := store.GetNew()
	if len(newTags3) != 1 {
		t.Errorf("Expected 1 new tag after adding another, got %d", len(newTags3))
	}
}

func TestTagStore_Count(t *testing.T) {
	store := NewTagStore()

	if store.Count() != 0 {
		t.Errorf("Expected count 0, got %d", store.Count())
	}

	_ = store.Add(models.Tag{ID: 1, Name: "tag1"})
	_ = store.Add(models.Tag{ID: 2, Name: "tag2"})

	if store.Count() != 2 {
		t.Errorf("Expected count 2, got %d", store.Count())
	}
}
