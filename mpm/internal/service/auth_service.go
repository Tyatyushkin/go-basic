package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"log"
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

func NewAuthService(userStorage UserStorageInterface) *AuthService {
	// По умолчанию используем секрет из переменной окружения или дефолтный
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Генерируем 32 байта случайных данных
		secretBytes := make([]byte, 32)
		_, err := rand.Read(secretBytes)
		if err != nil {
			// Если не удалось сгенерировать, используем запасной вариант
			jwtSecret = "your_default_secret_key_please_change_in_production"
		} else {
			// Кодируем в base64 для удобства хранения/отображения
			jwtSecret = base64.StdEncoding.EncodeToString(secretBytes)
			// Рекомендация: сохранить сгенерированный секрет
			log.Printf("Сгенерирован новый JWT секрет: %s. Рекомендуется сохранить его в переменной окружения JWT_SECRET", jwtSecret)
		}
	}

	// Время жизни токена - 24 часа по умолчанию
	tokenTTL := 24 * time.Hour

	return &AuthService{
		jwtSecret:   []byte(jwtSecret),
		tokenTTL:    tokenTTL,
		userStorage: userStorage,
	}
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
