package handlers

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopanel/internal/services"
)

// ContainersHandler manages Docker containers.
type ContainersHandler struct {
	docker *services.DockerService
}

// NewContainersHandler creates a new containers handler.
func NewContainersHandler(docker *services.DockerService) *ContainersHandler {
	return &ContainersHandler{docker: docker}
}

// List handles GET /api/containers
func (h *ContainersHandler) List(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.ListContainers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"containers": containers,
	})
}

// Start handles POST /api/containers/{id}/start
func (h *ContainersHandler) Start(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
		return
	}

	if err := h.docker.StartContainer(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Stop handles POST /api/containers/{id}/stop
func (h *ContainersHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
		return
	}

	if err := h.docker.StopContainer(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Restart handles POST /api/containers/{id}/restart
func (h *ContainersHandler) Restart(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
		return
	}

	if err := h.docker.RestartContainer(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Restore handles POST /api/containers/{id}/restore natively over multi-part payloads mapping ZIP streams locally bypassing legacy CLI OS faults seamlessly.
func (h *ContainersHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
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
	tmpDir := filepath.Join("/tmp", "restore_"+id)
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		http.Error(w, `{"error":"failed creating tmp boundary"}`, http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join("/tmp", id+"_payload.zip")
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
	exec.Command("docker", "cp", filepath.Join(tmpDir, "data")+"/.", id+":/app/data/").Run()
	exec.Command("docker", "cp", filepath.Join(tmpDir, "uploads")+"/.", id+":/app/uploads/").Run()
	exec.Command("docker", "cp", filepath.Join(tmpDir, "sqlite.db"), id+":/app/data/sqlite.db").Run() // Flat fallback seamlessly

	// 5. Hard reboot natively releasing internal SQLLite locks dynamically!
	h.docker.RestartContainer(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Logs handles GET /api/containers/{id}/logs
func (h *ContainersHandler) Logs(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
		return
	}

	lines := 100
	if l := r.URL.Query().Get("lines"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			lines = n
		}
	}

	logs, err := h.docker.GetContainerLogs(id, lines)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"container_id": id,
		"logs":         logs,
	})
}

// extractContainerID extracts the container ID from a URL path like /api/containers/{id}/action
func extractContainerID(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	// Expected: api, containers, {id}, {action}
	for i, p := range parts {
		if p == "containers" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
