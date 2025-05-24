package service

import (
	"context"
	"errors"
	"mpm/internal/models"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveEntities(entities []models.Entity) error {
	args := m.Called(entities)
	return args.Error(0)
}

func (m *MockRepository) SaveEntity(entity models.Entity) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRepository) PersistData() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) LoadData() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) GetAllPhotos() []models.Photo {
	args := m.Called()
	return args.Get(0).([]models.Photo)
}

func (m *MockRepository) GetAllAlbums(ctx context.Context) ([]models.Album, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Album), args.Error(1)
}

func (m *MockRepository) GetAllTags() []models.Tag {
	args := m.Called()
	return args.Get(0).([]models.Tag)
}

func (m *MockRepository) GetEntitiesCounts() (photoCount, albumCount, tagCount int) {
	args := m.Called()
	return args.Get(0).(int), args.Get(1).(int), args.Get(2).(int)
}

func (m *MockRepository) GetNewEntities() (newPhotos []models.Photo, newAlbums []models.Album, newTags []models.Tag) {
	args := m.Called()
	return args.Get(0).([]models.Photo), args.Get(1).([]models.Album), args.Get(2).([]models.Tag)
}

func (m *MockRepository) FindPhotoByID(id int) (models.Photo, error) {
	args := m.Called(id)
	return args.Get(0).(models.Photo), args.Error(1)
}

func (m *MockRepository) FindAlbumByID(ctx context.Context, id int) (models.Album, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Album), args.Error(1)
}

func (m *MockRepository) FindTagByID(id int) (models.Tag, error) {
	args := m.Called(id)
	return args.Get(0).(models.Tag), args.Error(1)
}

func (m *MockRepository) AddAlbum(ctx context.Context, album models.Album) (int, error) {
	args := m.Called(ctx, album)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) UpdateAlbum(ctx context.Context, id int, updatedAlbum models.Album) error {
	args := m.Called(ctx, id, updatedAlbum)
	return args.Error(0)
}

func (m *MockRepository) DeleteAlbum(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) InitStorage(ctx context.Context) {
	m.Called(ctx)
}

func TestNewEntityService(t *testing.T) {
	mockRepo := &MockRepository{}

	service := NewEntityService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestEntityService_GenerateAndSaveEntities(t *testing.T) {
	t.Run("context cancelled before start", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := service.GenerateAndSaveEntities(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("successful generation", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			if len(entities) == 0 {
				return false
			}
			switch entities[0].(type) {
			case models.Album:
				return true
			case models.Tag:
				return true
			default:
				return false
			}
		})).Return(nil).Maybe()
		mockRepo.On("GetEntitiesCounts").Return(0, 1, 15)

		ctx := context.Background()
		err := service.GenerateAndSaveEntities(ctx)

		assert.NoError(t, err)
	})

	t.Run("save entities error", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		mockRepo.On("SaveEntities", mock.AnythingOfType("[]models.Entity")).Return(errors.New("save error")).Maybe()
		mockRepo.On("GetEntitiesCounts").Return(0, 0, 0)

		ctx := context.Background()
		err := service.GenerateAndSaveEntities(ctx)

		assert.NoError(t, err)
	})
}

func TestEntityService_StartMonitoring(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewEntityService(mockRepo)

	mockRepo.On("GetEntitiesCounts").Return(0, 0, 0).Maybe()
	mockRepo.On("GetNewEntities").Return([]models.Photo{}, []models.Album{}, []models.Tag{}).Maybe()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		service.StartMonitoring(ctx)
		close(done)
	}()

	<-ctx.Done()
	<-done
}

func TestEntityService_monitorEntities(t *testing.T) {
	t.Run("monitoring stops on context cancellation", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		mockRepo.On("GetEntitiesCounts").Return(0, 0, 0).Maybe()
		mockRepo.On("GetNewEntities").Return([]models.Photo{}, []models.Album{}, []models.Tag{}).Maybe()

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			service.monitorEntities(ctx)
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()
		<-done
	})

	t.Run("detects new entities", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		first := true
		mockRepo.On("GetEntitiesCounts").Return(func() (int, int, int) {
			if first {
				first = false
				return 0, 0, 0
			}
			return 1, 1, 1
		}()).Maybe()

		newPhoto := models.Photo{ID: 1, Name: "test.jpg"}
		newAlbum := models.Album{ID: 1, Name: "Test Album"}
		newTag := models.Tag{ID: 1, Name: "test"}

		mockRepo.On("GetNewEntities").Return([]models.Photo{newPhoto}, []models.Album{newAlbum}, []models.Tag{newTag}).Maybe()

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		done := make(chan struct{})
		go func() {
			service.monitorEntities(ctx)
			close(done)
		}()

		time.Sleep(250 * time.Millisecond)
		<-done
	})
}

func TestEntityService_generateEntities(t *testing.T) {
	service := &EntityService{}
	entityChannel := make(chan EntityJob, 100)
	var wg sync.WaitGroup

	wg.Add(1)
	go service.generateEntities(entityChannel, &wg)
	wg.Wait()

	var albumCount, tagCount int
	for job := range entityChannel {
		switch job.Type {
		case "album":
			albumCount++
			album, ok := job.Entity.(models.Album)
			assert.True(t, ok)
			assert.Equal(t, "Default", album.Name)
		case "tag":
			tagCount++
			tag, ok := job.Entity.(models.Tag)
			assert.True(t, ok)
			assert.NotEmpty(t, tag.Name)
		}
	}

	assert.Equal(t, 1, albumCount, "Should generate exactly 1 album")
	assert.Equal(t, 15, tagCount, "Should generate exactly 15 tags")
}

func TestEntityService_saveEntities(t *testing.T) {
	t.Run("save albums successfully", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 1)
		var wg sync.WaitGroup

		album := models.Album{ID: 1, Name: "Test Album"}
		entityChannel <- EntityJob{
			Entity: album,
			Type:   "album",
		}
		close(entityChannel)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Album).Name == "Test Album"
		})).Return(nil)

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("save photos successfully", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 1)
		var wg sync.WaitGroup

		photo := models.Photo{ID: 1, Name: "test.jpg"}
		entityChannel <- EntityJob{
			Entity: photo,
			Type:   "photo",
		}
		close(entityChannel)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Photo).Name == "test.jpg"
		})).Return(nil)

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("save tags successfully", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 1)
		var wg sync.WaitGroup

		tag := models.Tag{ID: 1, Name: "test-tag"}
		entityChannel <- EntityJob{
			Entity: tag,
			Type:   "tag",
		}
		close(entityChannel)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Tag).Name == "test-tag"
		})).Return(nil)

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("unknown entity type", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 1)
		var wg sync.WaitGroup

		entityChannel <- EntityJob{
			Entity: models.Album{ID: 1, Name: "Test"},
			Type:   "unknown",
		}
		close(entityChannel)

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)
	})

	t.Run("save error handled gracefully", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 1)
		var wg sync.WaitGroup

		album := models.Album{ID: 1, Name: "Test Album"}
		entityChannel <- EntityJob{
			Entity: album,
			Type:   "album",
		}
		close(entityChannel)

		mockRepo.On("SaveEntities", mock.AnythingOfType("[]models.Entity")).Return(errors.New("save error"))

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("multiple entity types", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewEntityService(mockRepo)

		entityChannel := make(chan EntityJob, 3)
		var wg sync.WaitGroup

		photo := models.Photo{ID: 1, Name: "test.jpg"}
		album := models.Album{ID: 1, Name: "Test Album"}
		tag := models.Tag{ID: 1, Name: "test-tag"}

		entityChannel <- EntityJob{Entity: photo, Type: "photo"}
		entityChannel <- EntityJob{Entity: album, Type: "album"}
		entityChannel <- EntityJob{Entity: tag, Type: "tag"}
		close(entityChannel)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Photo).Name == "test.jpg"
		})).Return(nil)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Album).Name == "Test Album"
		})).Return(nil)

		mockRepo.On("SaveEntities", mock.MatchedBy(func(entities []models.Entity) bool {
			return len(entities) == 1 && entities[0].(models.Tag).Name == "test-tag"
		})).Return(nil)

		wg.Add(1)
		service.saveEntities(entityChannel, &wg, 0)

		mockRepo.AssertExpectations(t)
	})
}
