package service

import (
	"errors"
	"mpm/internal/models"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUserByCredentials(username, password string) (*models.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStorage) GetUserByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestNewAuthService(t *testing.T) {
	mockStorage := &MockUserStorage{}

	t.Run("with JWT_SECRET env var", func(t *testing.T) {
		_ = os.Setenv("JWT_SECRET", "test_secret")
		defer func() { _ = os.Unsetenv("JWT_SECRET") }()

		service := NewAuthService(mockStorage)

		assert.NotNil(t, service)
		assert.Equal(t, []byte("test_secret"), service.jwtSecret)
		assert.Equal(t, 24*time.Hour, service.tokenTTL)
		assert.Equal(t, mockStorage, service.userStorage)
	})

	t.Run("without JWT_SECRET env var", func(t *testing.T) {
		_ = os.Unsetenv("JWT_SECRET")

		service := NewAuthService(mockStorage)

		assert.NotNil(t, service)
		assert.NotEmpty(t, service.jwtSecret)
		assert.Equal(t, 24*time.Hour, service.tokenTTL)
		assert.Equal(t, mockStorage, service.userStorage)
	})
}

func TestAuthService_GenerateToken(t *testing.T) {
	mockStorage := &MockUserStorage{}
	service := &AuthService{
		jwtSecret:   []byte("test_secret"),
		tokenTTL:    24 * time.Hour,
		userStorage: mockStorage,
	}

	t.Run("successful token generation", func(t *testing.T) {
		user := &models.User{ID: 1, Username: "testuser"}
		mockStorage.On("GetUserByCredentials", "testuser", "password").Return(user, nil)

		token, err := service.GenerateToken("testuser", "password")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage := &MockUserStorage{}
		service := &AuthService{
			jwtSecret:   []byte("test_secret"),
			tokenTTL:    24 * time.Hour,
			userStorage: mockStorage,
		}

		mockStorage.On("GetUserByCredentials", "testuser", "password").Return((*models.User)(nil), errors.New("storage error"))

		token, err := service.GenerateToken("testuser", "password")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "storage error", err.Error())
		mockStorage.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockStorage := &MockUserStorage{}
		service := &AuthService{
			jwtSecret:   []byte("test_secret"),
			tokenTTL:    24 * time.Hour,
			userStorage: mockStorage,
		}

		mockStorage.On("GetUserByCredentials", "wronguser", "wrongpass").Return((*models.User)(nil), nil)

		token, err := service.GenerateToken("wronguser", "wrongpass")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "неверные учетные данные", err.Error())
		mockStorage.AssertExpectations(t)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	service := &AuthService{
		jwtSecret:   []byte("test_secret"),
		tokenTTL:    24 * time.Hour,
		userStorage: &MockUserStorage{},
	}

	t.Run("valid token", func(t *testing.T) {
		claims := TokenClaims{
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.tokenTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(service.jwtSecret)
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)

		assert.NoError(t, err)
		assert.NotNil(t, validatedClaims)
		assert.Equal(t, 1, validatedClaims.UserID)
	})

	t.Run("invalid token format", func(t *testing.T) {
		claims, err := service.ValidateToken("invalid_token")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		claims := TokenClaims{
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(service.jwtSecret)
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)

		assert.Error(t, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("wrong signing method", func(t *testing.T) {
		claims := TokenClaims{
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.tokenTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString([]byte("wrong_secret"))
		if err == nil {
			validatedClaims, err := service.ValidateToken(tokenString)

			assert.Error(t, err)
			assert.Nil(t, validatedClaims)
		}
	})
}

func TestAuthService_GetUserFromToken(t *testing.T) {
	mockStorage := &MockUserStorage{}
	service := &AuthService{
		jwtSecret:   []byte("test_secret"),
		tokenTTL:    24 * time.Hour,
		userStorage: mockStorage,
	}

	t.Run("successful user retrieval", func(t *testing.T) {
		claims := TokenClaims{
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.tokenTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(service.jwtSecret)
		assert.NoError(t, err)

		expectedUser := &models.User{ID: 1, Username: "testuser"}
		mockStorage.On("GetUserByID", 1).Return(expectedUser, nil)

		user, err := service.GetUserFromToken(tokenString)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid token", func(t *testing.T) {
		user, err := service.GetUserFromToken("invalid_token")

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("user not found in storage", func(t *testing.T) {
		claims := TokenClaims{
			UserID: 999,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.tokenTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(service.jwtSecret)
		assert.NoError(t, err)

		mockStorage.On("GetUserByID", 999).Return((*models.User)(nil), errors.New("user not found"))

		user, err := service.GetUserFromToken(tokenString)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "user not found", err.Error())
		mockStorage.AssertExpectations(t)
	})
}
