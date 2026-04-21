package handlers

import (
	"encoding/json"
	"net/http"

	"gopanel/internal/config"
	"gopanel/internal/middleware"
)

// AuthHandler handles login/logout/session endpoints.
type AuthHandler struct {
	cfg  *config.Config
	auth *middleware.Auth
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(cfg *config.Config, auth *middleware.Auth) *AuthHandler {
	return &AuthHandler{cfg: cfg, auth: auth}
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if creds.Username != h.cfg.Username || creds.Password != h.cfg.Password {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := h.auth.GenerateToken(creds.Username)
	http.SetCookie(w, &http.Cookie{
		Name:     "gopanel_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"username": creds.Username,
	})
}

// Session handles GET /api/auth/session
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("gopanel_session")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	username, valid := h.auth.ValidateToken(cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": valid,
		"username":      username,
	})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gopanel_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
