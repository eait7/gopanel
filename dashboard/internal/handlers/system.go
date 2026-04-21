package handlers

import (
	"encoding/json"
	"net/http"

	"gopanel/internal/config"
	"gopanel/internal/services"
)

// SystemHandler provides system stats and service links.
type SystemHandler struct {
	sysinfo *services.SysInfoService
	cfg     *config.Config
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(sysinfo *services.SysInfoService, cfg *config.Config) *SystemHandler {
	return &SystemHandler{sysinfo: sysinfo, cfg: cfg}
}

// Stats handles GET /api/system/stats
func (h *SystemHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := h.sysinfo.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Links handles GET /api/links
func (h *SystemHandler) Links(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"filebrowser": h.cfg.FileBrowserURL,
		"portainer":   h.cfg.PortainerExternalURL,
	})
}
