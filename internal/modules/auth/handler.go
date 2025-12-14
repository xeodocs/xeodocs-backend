package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	service *Service
	cfg     *config.Config
}

func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", h.Logout)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	user, token, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// Determine cookie configuration based on environment
	var domain string
	var secure bool
	var sameSite http.SameSite

	switch h.cfg.Environment {
	case "production":
		domain = "admin.xeodocs.com"
		secure = true
		sameSite = http.SameSiteLaxMode
	case "staging":
		domain = "staging-admin.xeodocs.com"
		secure = true
		sameSite = http.SameSiteLaxMode
	default: // development
		domain = "" // Host-only cookie for local development
		secure = false
		sameSite = http.SameSiteLaxMode
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    token,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"message": "Login successful",
			"user":    user,
		},
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Determine cookie configuration based on environment
	var domain string
	var secure bool
	var sameSite http.SameSite

	switch h.cfg.Environment {
	case "production":
		domain = "admin.xeodocs.com"
		secure = true
		sameSite = http.SameSiteLaxMode
	case "staging":
		domain = "staging-admin.xeodocs.com"
		secure = true
		sameSite = http.SameSiteLaxMode
	default: // development
		domain = "" // Host-only cookie for local development
		secure = false
		sameSite = http.SameSiteLaxMode
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/",
		MaxAge:   -1,
	})

	response.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Middleware

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := ""

		// Check Cookie
		cookie, err := r.Cookie("session_id")
		if err == nil {
			tokenString = cookie.Value
		}

		// Check Authorization Header if no cookie
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Check API Key Header if no token
		if tokenString == "" {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				// Validate API Key directly against DB
				// For simplicity/performance, we might cache this or add a method in service
				// But here let's quickly query the DB or use the service
				user, err := h.service.ValidateAPIKey(apiKey)
				if err != nil {
					response.Error(w, http.StatusUnauthorized, "Invalid API Key", nil)
					return
				}
				ctx := context.WithValue(r.Context(), userIDKey, user.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		if tokenString == "" {
			response.Error(w, http.StatusUnauthorized, "Missing authentication", nil)
			return
		}

		// Validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(h.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "Invalid token", nil)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID, ok := claims["sub"].(string)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Invalid token claims", nil)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			response.Error(w, http.StatusUnauthorized, "Invalid token claims", nil)
		}
	})
}
