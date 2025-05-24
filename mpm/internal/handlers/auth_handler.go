package handlers

import (
	"encoding/json"
	"log"
	"mpm/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Авторизация пользователя и получение JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginRequest true "Учетные данные пользователя"
// @Success 200 {object} loginResponse
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 401 {string} string "Неверные учетные данные"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	token, err := h.authService.GenerateToken(req.Username, req.Password)
	if err != nil {
		log.Printf("Ошибка при генерации токена: %v", err)
		http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{Token: token})
}
