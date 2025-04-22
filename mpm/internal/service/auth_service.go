package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"mpm/internal/models"
	"time"
)

type AuthService struct {
	jwtSecret   []byte
	tokenTTL    time.Duration
	userStorage UserStorageInterface
}

type UserStorageInterface interface {
	GetUserByCredentials(username, password string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
}

type TokenClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateToken(username, password string) (string, error) {
	user, err := s.userStorage.GetUserByCredentials(username, password)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("неверные учетные данные")
	}

	claims := TokenClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи токена")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("неверный токен")
}

func (s *AuthService) GetUserFromToken(tokenString string) (*models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return s.userStorage.GetUserByID(claims.UserID)
}
