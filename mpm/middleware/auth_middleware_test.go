package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mpm/internal/models"
	"mpm/internal/service"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUserByCredentials(username, password string) (*models.User, error) {
	args := m.Called(username, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStorage) GetUserByID(id int) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)

	user := models.User{
		ID:       1,
		Username: "testuser",
		Password: "hashedpassword",
	}

	mockStorage.On("GetUserByCredentials", "testuser", "testpass").Return(&user, nil)
	mockStorage.On("GetUserByID", 1).Return(&user, nil)

	// Генерируем токен
	token, err := authService.GenerateToken(user.Username, "testpass")
	assert.NoError(t, err)

	// Создаем middleware
	middleware := AuthMiddleware(authService)

	// Создаем тестовый handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userFromContext := r.Context().Value(UserContextKey)
		assert.NotNil(t, userFromContext)

		contextUser, ok := userFromContext.(*models.User)
		assert.True(t, ok)
		assert.Equal(t, user.ID, contextUser.ID)
		assert.Equal(t, user.Username, contextUser.Username)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Оборачиваем handler в middleware
	handler := middleware(testHandler)

	// Создаем запрос с валидным токеном
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())

	mockStorage.AssertExpectations(t)
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)
	middleware := AuthMiddleware(authService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when auth header is missing")
	})

	handler := middleware(testHandler)

	// Создаем запрос без заголовка Authorization
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Отсутствует заголовок авторизации")
}

func TestAuthMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)
	middleware := AuthMiddleware(authService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid auth header")
	})

	handler := middleware(testHandler)

	testCases := []struct {
		name   string
		header string
	}{
		{"Only token", "sometoken"},
		{"Wrong prefix", "Basic sometoken"},
		{"Empty Bearer", "Bearer"},
		{"Multiple spaces", "Bearer  token  extra"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "Неверный формат заголовка авторизации")
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)
	middleware := AuthMiddleware(authService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid token")
	})

	handler := middleware(testHandler)

	// Создаем запрос с невалидным токеном
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Неверный токен")
}

func TestAuthMiddleware_UserContextKey(t *testing.T) {
	// Проверяем, что константа правильно определена
	assert.Equal(t, contextKey("user"), UserContextKey)
}

func TestAuthMiddleware_ChainedMiddleware(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)

	user := models.User{
		ID:       1,
		Username: "testuser",
		Password: "hashedpassword",
	}

	mockStorage.On("GetUserByCredentials", "testuser", "testpass").Return(&user, nil)
	mockStorage.On("GetUserByID", 1).Return(&user, nil)

	// Генерируем токен
	token, err := authService.GenerateToken(user.Username, "testpass")
	assert.NoError(t, err)

	// Создаем цепочку middleware
	authMiddleware := AuthMiddleware(authService)

	// Дополнительный middleware для проверки цепочки
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что пользователь есть в контексте
			userFromContext := r.Context().Value(UserContextKey)
			assert.NotNil(t, userFromContext)

			// Добавляем заголовок для проверки
			w.Header().Set("X-Middleware-Chain", "passed")
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("final"))
	})

	// Создаем цепочку: auth -> log -> final
	handler := authMiddleware(logMiddleware(finalHandler))

	// Создаем запрос
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "final", rr.Body.String())
	assert.Equal(t, "passed", rr.Header().Get("X-Middleware-Chain"))

	mockStorage.AssertExpectations(t)
}

func TestAuthMiddleware_ContextPropagation(t *testing.T) {
	// Setup
	mockStorage := &MockUserStorage{}
	authService := service.NewAuthService(mockStorage)

	user := models.User{
		ID:       1,
		Username: "testuser",
		Password: "hashedpassword",
	}

	mockStorage.On("GetUserByCredentials", "testuser", "testpass").Return(&user, nil)
	mockStorage.On("GetUserByID", 1).Return(&user, nil)

	// Генерируем токен
	token, err := authService.GenerateToken(user.Username, "testpass")
	assert.NoError(t, err)

	middleware := AuthMiddleware(authService)

	// Handler, который проверяет контекст более детально
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем исходный контекст
		originalCtx := r.Context()
		assert.NotNil(t, originalCtx)

		// Проверяем, что пользователь добавлен в контекст
		userFromContext := originalCtx.Value(UserContextKey)
		assert.NotNil(t, userFromContext)

		// Проверяем тип и данные пользователя
		contextUser, ok := userFromContext.(*models.User)
		assert.True(t, ok)
		assert.Equal(t, user.ID, contextUser.ID)
		assert.Equal(t, user.Username, contextUser.Username)

		// Проверяем, что можем создать дочерний контекст
		childCtx, cancel := context.WithCancel(originalCtx)
		defer cancel()

		// Проверяем, что пользователь доступен и в дочернем контексте
		userFromChild := childCtx.Value(UserContextKey)
		assert.NotNil(t, userFromChild)
		assert.Equal(t, user.ID, userFromChild.(*models.User).ID)

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	// Создаем запрос
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)

	mockStorage.AssertExpectations(t)
}
