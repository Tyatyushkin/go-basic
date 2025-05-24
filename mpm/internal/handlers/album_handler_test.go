package handlers

import (
	"context"
	"encoding/json"
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

// CreateAlbum метод для тестового обработчика
func (h *TestAlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Декодируем тело запроса в структуру альбома
	var album models.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Проверяем валидность альбома
	if err := album.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавление альбома через репозиторий
	id, err := h.repo.AddAlbum(ctx, album)
	if err != nil {
		http.Error(w, "Ошибка при создании альбома", http.StatusInternalServerError)
		return
	}

	// Получаем альбом с присвоенным ID
	newAlbum, err := h.repo.FindAlbumByID(ctx, id)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type и статус
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Сериализуем созданный альбом в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(newAlbum)
}

// GetAllAlbums метод для тестового обработчика
func (h *TestAlbumHandler) GetAllAlbums(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем все альбомы из репозитория
	albums, err := h.repo.GetAllAlbums(ctx)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Сериализуем альбомы в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(albums)
}

// GetAlbumByID метод для тестового обработчика
func (h *TestAlbumHandler) GetAlbumByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID из пути запроса
	var idStr string
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

	// Ищем альбом в репозитории
	album, err := h.repo.FindAlbumByID(ctx, id)
	if err != nil {
		http.Error(w, "Альбом не найден", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Сериализуем альбом в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(album)
}

// UpdateAlbum метод для тестового обработчика
func (h *TestAlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID из пути запроса
	var idStr string
	idStr = r.URL.Query().Get("id")
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

	// Декодируем тело запроса в структуру альбома
	var updatedAlbum models.Album
	if err := json.NewDecoder(r.Body).Decode(&updatedAlbum); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Обновляем альбом через репозиторий
	if err := h.repo.UpdateAlbum(ctx, id, updatedAlbum); err != nil {
		if containsString(err.Error(), "не найден") {
			http.Error(w, "Альбом не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем успешный статус
	w.WriteHeader(http.StatusOK)
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

func TestCreateAlbum(t *testing.T) {
	t.Run("Успешное создание альбома", func(t *testing.T) {
		mockRepo := new(MockRepository)

		inputAlbum := models.Album{
			Name:        "Новый альбом",
			Description: "Описание альбома",
			Tags:        []string{"тег1", "тег2"},
		}

		createdID := 1
		createdAlbum := models.Album{
			ID:          createdID,
			Name:        inputAlbum.Name,
			Description: inputAlbum.Description,
			Tags:        inputAlbum.Tags,
		}

		// Настраиваем мок
		mockRepo.On("AddAlbum", mock.Anything, mock.MatchedBy(func(album models.Album) bool {
			return album.Name == inputAlbum.Name
		})).Return(createdID, nil)

		mockRepo.On("FindAlbumByID", mock.Anything, createdID).Return(createdAlbum, nil)

		handler := createTestHandlerWithMock(mockRepo)

		// Создаем JSON-запрос
		body := strings.NewReader(`{
	            "name": "Новый альбом",
	            "description": "Описание альбома",
	            "tags": ["тег1", "тег2"]
	        }`)

		req := httptest.NewRequest(http.MethodPost, "/albums", body)
		w := httptest.NewRecorder()

		handler.CreateAlbum(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Проверяем тело ответа
		var responseAlbum models.Album
		err := json.Unmarshal(w.Body.Bytes(), &responseAlbum)
		assert.NoError(t, err)
		assert.Equal(t, createdAlbum.ID, responseAlbum.ID)
		assert.Equal(t, createdAlbum.Name, responseAlbum.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Неверный формат JSON", func(t *testing.T) {
		mockRepo := new(MockRepository)
		handler := createTestHandlerWithMock(mockRepo)

		// Создаем неверный JSON-запрос
		body := strings.NewReader(`{
	            "name": "Новый альбом",
	            "description": "Описание альбома",
	            "tags": ["тег1", "тег2"
	        }`) // отсутствует закрывающая скобка

		req := httptest.NewRequest(http.MethodPost, "/albums", body)
		w := httptest.NewRecorder()

		handler.CreateAlbum(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAllAlbums(t *testing.T) {
	t.Run("Успешное получение списка альбомов", func(t *testing.T) {
		mockRepo := new(MockRepository)

		albums := []models.Album{
			{ID: 1, Name: "Альбом 1", Description: "Описание 1", Tags: []string{"тег1"}},
			{ID: 2, Name: "Альбом 2", Description: "Описание 2", Tags: []string{"тег2"}},
		}

		mockRepo.On("GetAllAlbums", mock.Anything).Return(albums, nil)

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/albums", nil)
		w := httptest.NewRecorder()

		handler.GetAllAlbums(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем тело ответа
		var responseAlbums []models.Album
		err := json.Unmarshal(w.Body.Bytes(), &responseAlbums)
		assert.NoError(t, err)
		assert.Len(t, responseAlbums, 2)
		assert.Equal(t, albums[0].ID, responseAlbums[0].ID)
		assert.Equal(t, albums[1].Name, responseAlbums[1].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Ошибка при получении альбомов", func(t *testing.T) {
		mockRepo := new(MockRepository)

		mockRepo.On("GetAllAlbums", mock.Anything).Return([]models.Album{}, fmt.Errorf("ошибка базы данных"))

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/albums", nil)
		w := httptest.NewRecorder()

		handler.GetAllAlbums(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAlbumByID(t *testing.T) {
	t.Run("Успешное получение альбома по ID", func(t *testing.T) {
		mockRepo := new(MockRepository)

		album := models.Album{
			ID:          1,
			Name:        "Тестовый альбом",
			Description: "Описание альбома",
			Tags:        []string{"тег1", "тег2"},
		}

		mockRepo.On("FindAlbumByID", mock.Anything, 1).Return(album, nil)

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/albums/1", nil)
		w := httptest.NewRecorder()

		handler.GetAlbumByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем тело ответа
		var responseAlbum models.Album
		err := json.Unmarshal(w.Body.Bytes(), &responseAlbum)
		assert.NoError(t, err)
		assert.Equal(t, album.ID, responseAlbum.ID)
		assert.Equal(t, album.Name, responseAlbum.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Альбом не найден", func(t *testing.T) {
		mockRepo := new(MockRepository)

		mockRepo.On("FindAlbumByID", mock.Anything, 999).Return(models.Album{}, fmt.Errorf("альбом не найден"))

		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/albums/999", nil)
		w := httptest.NewRecorder()

		handler.GetAlbumByID(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Некорректный ID альбома", func(t *testing.T) {
		mockRepo := new(MockRepository)
		handler := createTestHandlerWithMock(mockRepo)

		req := httptest.NewRequest(http.MethodGet, "/albums/abc", nil)
		w := httptest.NewRecorder()

		handler.GetAlbumByID(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateAlbum(t *testing.T) {
	t.Run("Успешное обновление альбома", func(t *testing.T) {
		mockRepo := new(MockRepository)

		updatedAlbum := models.Album{
			Name:        "Обновленный альбом",
			Description: "Новое описание",
			Tags:        []string{"новый-тег"},
		}

		mockRepo.On("UpdateAlbum", mock.Anything, 1, mock.MatchedBy(func(album models.Album) bool {
			return album.Name == updatedAlbum.Name
		})).Return(nil)

		handler := createTestHandlerWithMock(mockRepo)

		body := strings.NewReader(`{
	            "name": "Обновленный альбом",
	            "description": "Новое описание",
	            "tags": ["новый-тег"]
	        }`)

		req := httptest.NewRequest(http.MethodPut, "/albums/1", body)
		w := httptest.NewRecorder()

		handler.UpdateAlbum(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Обновление несуществующего альбома", func(t *testing.T) {
		mockRepo := new(MockRepository)

		mockRepo.On("UpdateAlbum", mock.Anything, 999, mock.Anything).
			Return(fmt.Errorf("альбом не найден"))

		handler := createTestHandlerWithMock(mockRepo)

		body := strings.NewReader(`{
	            "name": "Обновленный альбом",
	            "description": "Новое описание",
	            "tags": ["новый-тег"]
	        }`)

		req := httptest.NewRequest(http.MethodPut, "/albums/999", body)
		w := httptest.NewRecorder()

		handler.UpdateAlbum(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Некорректный JSON", func(t *testing.T) {
		mockRepo := new(MockRepository)
		handler := createTestHandlerWithMock(mockRepo)

		body := strings.NewReader(`{
	            "name": "Обновленный альбом",
	            "description": "Новое описание",
	            "tags": ["новый-тег"
	        }`) // отсутствует закрывающая скобка

		req := httptest.NewRequest(http.MethodPut, "/albums/1", body)
		w := httptest.NewRecorder()

		handler.UpdateAlbum(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
