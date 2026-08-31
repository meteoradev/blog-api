package handler

import (
	"encoding/json"
	"net/http"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/port"
	"github.com/meteoradev/BlogApi/internal/service"
)

type AuthController struct {
	authService port.AuthService
}

type AuthRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

// AuthResponse represents the response body for login
type AuthResponse struct {
	Token        string      `json:"token"`
	TokenPayload interface{} `json:"token_payload"`
}

func NewAuthResponse(res *domain.AuthResult) *AuthResponse {
	return &AuthResponse{Token: res.Token, TokenPayload: res.TokenPayload}
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAuthController(authService port.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Login handles user login and returns a JWT token
// @Summary      Login to the blog
// @Description  Authenticates a user by email and password and returns a JWT access token (valid for 1 hour)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      AuthRequest  true  "Login credentials"
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse  "invalid JSON"
// @Failure      401      {object}  ErrorResponse  "wrong password / can not login"
// @Failure      404      {object}  ErrorResponse  "user not found"
// @Failure      429      {object}  ErrorResponse  "rate limit exceeded"
// @Failure      500      {object}  ErrorResponse  "internal server error"
// @Router       /login [post]
func (a *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var authReq AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}
	token, err := a.authService.Login(r.Context(), authReq.Email, authReq.Password)
	if err != nil {
		switch err {
		case service.ErrWrongPassword:
			http.Error(w, `{"error": "wrong password"}`, http.StatusUnauthorized)
			return
		case service.ErrCanNotLogin:
			http.Error(w, `{"error": "can not login"}`, http.StatusUnauthorized)
			return
		case service.ErrUserNotFound:
			http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
			return
		default:
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NewAuthResponse(token))
}
