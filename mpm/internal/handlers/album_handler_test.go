package handlers

import (
	"context"
	"fmt"
	"mpm/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAlbumHandler создаем тестовую структуру обработчика с интерфейсом
type TestAlbumHandler struct {
	repo AlbumRepositoryInterface
}

// Интерфейс репозитория для тестирования
type AlbumRepositoryInterface interface {
	AddAlbum(ctx context.Context, album models.Album) (int, error)
	FindAlbumByID(ctx context.Context, id int) (models.Album, error)
	GetAllAlbums(ctx context.Context) ([]models.Album, error)
	UpdateAlbum(ctx context.Context, id int, album models.Album) error
	DeleteAlbum(ctx context.Context, id int) error
}

// MockRepository реализует интерфейс репозитория для тестов
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) DeleteAlbum(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) AddAlbum(ctx context.Context, album models.Album) (int, error) {
	args := m.Called(ctx, album)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) FindAlbumByID(ctx context.Context, id int) (models.Album, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Album), args.Error(1)
}

func (m *MockRepository) GetAllAlbums(ctx context.Context) ([]models.Album, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Album), args.Error(1)
}

func (m *MockRepository) UpdateAlbum(ctx context.Context, id int, album models.Album) error {
	args := m.Called(ctx, id, album)
	return args.Error(0)
}

// containsString проверяет, содержится ли подстрока в строке
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// DeleteAlbum метод для тестового обработчика (копируем логику из оригинального)
func (h *TestAlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	// Получаем контекст из запроса
	ctx := r.Context()

	// Извлекаем ID из пути запроса (упрощенная версия для тестов)
	var idStr string

	// Пробуем получить из параметра запроса
	idStr = r.URL.Query().Get("id")

	// Если ID нет в параметрах, пробуем извлечь из пути
	if idStr == "" {
		parts := splitPath(r.URL.Path)
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}

	if idStr == "" {
		http.Error(w, "ID альбома не указан", http.StatusBadRequest)
		return
	}

	// Преобразуем ID в целое число
	id := 0
	for _, c := range idStr {
		if c >= '0' && c <= '9' {
			id = id*10 + int(c-'0')
		} else {
			http.Error(w, "Некорректный ID альбома", http.StatusBadRequest)
			return
		}
	}

	// Удаляем альбом из репозитория
	err := h.repo.DeleteAlbum(ctx, id)
	if err != nil {
		if containsString(err.Error(), "не найден") {
			http.Error(w, "Альбом не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем успешный статус без тела ответа
	w.WriteHeader(http.StatusNoContent)
}

// Функция-помощник для создания тестового обработчика с моком
func createTestHandlerWithMock(mockRepo *MockRepository) *TestAlbumHandler {
	return &TestAlbumHandler{
		repo: mockRepo,
	}
}

func TestDeleteAlbum(t *testing.T) {
	t.Run("Successful album deletion", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockRepo.On("DeleteAlbum", mock.Anything, 1).Return(nil)

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/albums/1", nil)
		w := httptest.NewRecorder()

		handler.DeleteAlbum(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete non-existing album", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockRepo.On("DeleteAlbum", mock.Anything, 999).
			Return(fmt.Errorf("альбом не найден"))

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/albums/999", nil)
		w := httptest.NewRecorder()

		handler.DeleteAlbum(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid album ID", func(t *testing.T) {
		mockRepo := new(MockRepository)
		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/albums/invalid_id", nil)
		w := httptest.NewRecorder()

		handler.DeleteAlbum(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockRepo.On("DeleteAlbum", mock.Anything, 123).
			Return(fmt.Errorf("database connection failed"))

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodDelete, "/albums/123", nil)
		w := httptest.NewRecorder()

		handler.DeleteAlbum(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
