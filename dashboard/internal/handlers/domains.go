package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gopanel/internal/services"
)

// DomainsHandler manages domain/site CRUD via Caddy's API.
type DomainsHandler struct {
	caddy *services.CaddyService
}

// NewDomainsHandler creates a new domains handler.
func NewDomainsHandler(caddy *services.CaddyService) *DomainsHandler {
	return &DomainsHandler{caddy: caddy}
}

// List handles GET /api/domains
func (h *DomainsHandler) List(w http.ResponseWriter, r *http.Request) {
	domains, err := h.caddy.ListDomains()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"domains": []interface{}{},
			"warning": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domains": domains,
	})
}

// Add handles POST /api/domains
func (h *DomainsHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain   string `json:"domain"`
		Upstream string `json:"upstream"`
		Type     string `json:"type"` // "reverse_proxy" or "file_server"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.Upstream == "" {
		http.Error(w, `{"error":"domain and upstream are required"}`, http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "reverse_proxy"
	}

	if err := h.caddy.AddSite(req.Domain, req.Upstream, req.Type); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Domain " + req.Domain + " added successfully",
	})
}

// Delete handles DELETE /api/domains/{id}
func (h *DomainsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract route index from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"missing domain id"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, `{"error":"invalid domain id"}`, http.StatusBadRequest)
		return
	}

	if err := h.caddy.RemoveSite(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Domain removed successfully",
	})
}

// Update handles PUT /api/domains/{id}
func (h *DomainsHandler) Update(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"missing domain id"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, `{"error":"invalid domain id"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Domain   string `json:"domain"`
		Upstream string `json:"upstream"`
		Type     string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.caddy.UpdateSite(id, req.Domain, req.Upstream, req.Type); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Domain updated successfully",
	})
}

// CaddyConfig handles GET /api/caddy/config
func (h *DomainsHandler) CaddyConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.caddy.GetConfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
