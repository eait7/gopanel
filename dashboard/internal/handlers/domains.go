package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopanel/internal/services"
)

// DomainsHandler manages domain/site CRUD via Caddy's API.
type DomainsHandler struct {
	caddy  *services.CaddyService
	docker *services.DockerService
}

// NewDomainsHandler creates a new domains handler.
func NewDomainsHandler(caddy *services.CaddyService, docker *services.DockerService) *DomainsHandler {
	return &DomainsHandler{caddy: caddy, docker: docker}
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

// Restore handles POST /api/domains/{id}/restore mapping domains dynamically onto internal Docker daemon references safely!
func (h *DomainsHandler) Restore(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, `{"error":"invalid path mapping"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, `{"error":"invalid domain index"}`, http.StatusBadRequest)
		return
	}

	domains, err := h.caddy.ListDomains()
	if err != nil || id < 0 || id >= len(domains) {
		http.Error(w, `{"error":"domain not found payload"}`, http.StatusNotFound)
		return
	}
	
	domain := domains[id]
	if domain.Type != "reverse_proxy" || domain.Upstream == "" {
		http.Error(w, `{"error":"domain must strictly map into a proxy"}`, http.StatusBadRequest)
		return
	}

	upstreamParts := strings.Split(domain.Upstream, ":")
	if len(upstreamParts) != 2 {
		http.Error(w, `{"error":"malformed upstream"}`, http.StatusBadRequest)
		return
	}
	
	targetPortInt, _ := strconv.Atoi(upstreamParts[1])
	targetPort := uint16(targetPortInt)

	if h.docker == nil {
		http.Error(w, `{"error":"docker orchestration offline"}`, http.StatusInternalServerError)
		return
	}

	containers, err := h.docker.ListContainers()
	if err != nil {
		http.Error(w, `{"error":"cannot resolve backend instances"}`, http.StatusInternalServerError)
		return
	}

	var targetContainerID string
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.PublicPort == targetPort {
				targetContainerID = c.ID
				break
			}
		}
		if targetContainerID != "" {
			break
		}
	}

	if targetContainerID == "" {
		http.Error(w, `{"error":"failed to resolve organic site container over native port"}`, http.StatusNotFound)
		return
	}

	// 1. Limit form payload size appropriately
	if err := r.ParseMultipartForm(250 << 20); err != nil {
		http.Error(w, `{"error":"payload too large"}`, http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, `{"error":"failed to grab backup file"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 2. Safely dump payload natively inside Go container over isolated /tmp allocation
	tmpDir := filepath.Join("/tmp", "restore_"+targetContainerID)
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		http.Error(w, `{"error":"failed creating tmp boundary"}`, http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join("/tmp", targetContainerID+"_payload.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, `{"error":"failed locking tmp zip"}`, http.StatusInternalServerError)
		return
	}
	io.Copy(out, file)
	out.Close()
	defer os.Remove(zipPath)

	// 3. Extract the ZIP standardly over Golang core library natively preventing zip-slip vulnerabilities securely!
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		http.Error(w, `{"error":"invalid zip formatting"}`, http.StatusBadRequest)
		return
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(tmpDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(tmpDir)+string(os.PathSeparator)) {
			continue // Zip Slip Mitigation structurally seamlessly
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			continue
		}

		outFile, extractErr := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if extractErr != nil {
			continue
		}

		rc, rcErr := f.Open()
		if rcErr != nil {
			outFile.Close()
			continue
		}

		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}

	// 4. Overwrite running application memory actively utilizing Host daemon streams safely 
	exec.Command("docker", "cp", filepath.Join(tmpDir, "data")+"/.", targetContainerID+":/app/data/").Run()
	exec.Command("docker", "cp", filepath.Join(tmpDir, "uploads")+"/.", targetContainerID+":/app/uploads/").Run()
	exec.Command("docker", "cp", filepath.Join(tmpDir, "sqlite.db"), targetContainerID+":/app/data/sqlite.db").Run()

	// 5. Hard reboot natively releasing internal SQLLite locks dynamically!
	h.docker.RestartContainer(targetContainerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
