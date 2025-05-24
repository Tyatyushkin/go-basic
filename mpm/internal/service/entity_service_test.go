package service

import (
	"context"
	"mpm/internal/models"
	"mpm/internal/repository"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "service_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func TestNewEntityService(t *testing.T) {
	tempDir := createTempDir(t)
	repo := repository.NewRepository("json", tempDir, 0)

	service := NewEntityService(repo)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
}

func TestEntityService_GenerateAndSaveEntities(t *testing.T) {
	t.Run("context cancelled before start", func(t *testing.T) {
		tempDir := createTempDir(t)
		repo := repository.NewRepository("json", tempDir, 0)
		service := NewEntityService(repo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := service.GenerateAndSaveEntities(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("successful generation", func(t *testing.T) {
		tempDir := createTempDir(t)
		repo := repository.NewRepository("json", tempDir, time.Second)
		repo.InitStorage(context.Background())
		service := NewEntityService(repo)

		ctx := context.Background()
		err := service.GenerateAndSaveEntities(ctx)

		assert.NoError(t, err)

		photoCount, albumCount, tagCount := repo.GetEntitiesCounts()
		assert.Equal(t, 1, albumCount)
		assert.Equal(t, 15, tagCount)
		assert.Equal(t, 0, photoCount)
	})
}

func TestEntityService_StartMonitoring(t *testing.T) {
	tempDir := createTempDir(t)
	repo := repository.NewRepository("json", tempDir, time.Second)
	repo.InitStorage(context.Background())
	service := NewEntityService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.StartMonitoring(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestEntityService_monitorEntities(t *testing.T) {
	t.Run("monitoring stops on context cancellation", func(t *testing.T) {
		tempDir := createTempDir(t)
		repo := repository.NewRepository("json", tempDir, time.Second)
		repo.InitStorage(context.Background())
		service := NewEntityService(repo)

		ctx, cancel := context.WithCancel(context.Background())

		go service.monitorEntities(ctx)

		time.Sleep(50 * time.Millisecond)
		cancel()
		time.Sleep(50 * time.Millisecond)
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

func TestEntityService_saveEntities_UnknownType(t *testing.T) {
	tempDir := createTempDir(t)
	repo := repository.NewRepository("json", tempDir, time.Second)
	repo.InitStorage(context.Background())
	service := NewEntityService(repo)

	entityChannel := make(chan EntityJob, 1)
	var wg sync.WaitGroup

	entityChannel <- EntityJob{
		Entity: models.Album{ID: 1, Name: "Test"},
		Type:   "unknown",
	}
	close(entityChannel)

	wg.Add(1)
	service.saveEntities(entityChannel, &wg, 0)
}
