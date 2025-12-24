package handlers

import (
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/MohamedElashri/snipo/internal/config"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
	version   string
	commit    string
	features  *config.FeatureFlags
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, version, commit string, features *config.FeatureFlags) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
		commit:    commit,
		features:  features,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Commit    string            `json:"commit,omitempty"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
	Memory    MemoryStats       `json:"memory"`
	Features  *FeatureFlags     `json:"features,omitempty"`
	Timestamp string            `json:"timestamp"`
}

// FeatureFlags represents enabled features
type FeatureFlags struct {
	PublicSnippets bool `json:"public_snippets"`
	S3Sync         bool `json:"s3_sync"`
	APITokens      bool `json:"api_tokens"`
	BackupRestore  bool `json:"backup_restore"`
}

// MemoryStats represents memory statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	status := "healthy"

	// Check database
	if err := h.db.Ping(); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := HealthResponse{
		Status:  status,
		Version: h.version,
		Commit:  h.commit,
		Uptime:  time.Since(h.startTime).Round(time.Second).String(),
		Checks:  checks,
		Memory: MemoryStats{
			Alloc:      m.Alloc / 1024 / 1024,
			TotalAlloc: m.TotalAlloc / 1024 / 1024,
			Sys:        m.Sys / 1024 / 1024,
			NumGC:      m.NumGC,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Add feature flags if available
	if h.features != nil {
		response.Features = &FeatureFlags{
			PublicSnippets: h.features.PublicSnippets,
			S3Sync:         h.features.S3Sync,
			APITokens:      h.features.APITokens,
			BackupRestore:  h.features.BackupRestore,
		}
	}

	if status == "healthy" {
		OK(w, r, response)
	} else {
		JSON(w, http.StatusServiceUnavailable, response)
	}
}

// Ping handles GET /ping - simple liveness check
func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}
