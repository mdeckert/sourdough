package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdeckert/sourdough/internal/ecobee"
	"github.com/mdeckert/sourdough/internal/storage"
)

func setupTestServer(t *testing.T) (*Server, string) {
	// Create temp directory for test data
	tmpDir, err := os.MkdirTemp("", "sourdough-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create storage
	store, err := storage.New(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create server
	// Create disabled Ecobee client for tests
	ecobeeClient := &ecobee.Client{}
	server := New(store, ecobeeClient, "8080")
	return server, tmpDir
}

func cleanup(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestHealthCheck(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestBakeStart(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Test starting a bake
	req := httptest.NewRequest(http.MethodPost, "/bake/start", nil)
	w := httptest.NewRecorder()

	server.handleLoafStart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "loaf started" {
		t.Errorf("Expected status 'loaf started', got '%s'", response["status"])
	}

	// Try to start another bake (should fail)
	req2 := httptest.NewRequest(http.MethodPost, "/bake/start", nil)
	w2 := httptest.NewRecorder()

	server.handleLoafStart(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w2.Code)
	}
}

func TestLogTemperature(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Start a bake first
	server.storage.AppendEvent(nil)

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"Kitchen temp", "/log/temp/72", http.StatusOK},
		{"Dough temp", "/log/temp/76?type=dough", http.StatusOK},
		{"Invalid temp", "/log/temp/invalid", http.StatusBadRequest},
		{"Missing temp", "/log/temp/", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			server.handleLog(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestLogNote(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"Valid note", `{"note":"Good oven spring"}`, http.StatusOK},
		{"Empty note", `{"note":""}`, http.StatusBadRequest},
		{"Invalid JSON", `{invalid}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/log/note", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			server.handleLog(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestLogEvents(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	events := []string{
		"/log/fed",
		"/log/levain-ready",
		"/log/mixed",
		"/log/fold",
		"/log/shaped",
		"/log/fridge-in",
		"/log/fridge-out",
		"/log/oven-in",
	}

	for _, path := range events {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()

			server.handleLog(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Path %s: expected status 200, got %d", path, w.Code)
			}
		})
	}
}

func TestFoldCountIncrement(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Log three folds
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/log/fold", nil)
		w := httptest.NewRecorder()

		server.handleLog(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Fold %d failed: status %d", i, w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		event := response["event"].(map[string]interface{})
		foldCount := int(event["fold_count"].(float64))

		if foldCount != i {
			t.Errorf("Expected fold count %d, got %d", i, foldCount)
		}
	}
}

func TestBakeComplete(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Start a bake
	server.storage.AppendEvent(nil)

	// Complete with assessment
	assessment := map[string]interface{}{
		"assessment": map[string]interface{}{
			"proof_level":   "good",
			"crumb_quality": 8,
			"browning":      "good",
			"score":         9,
			"notes":         "Excellent loaf",
		},
	}

	body, _ := json.Marshal(assessment)
	req := httptest.NewRequest(http.MethodPost, "/log/loaf-complete", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.handleLog(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify we can start a new bake now
	req2 := httptest.NewRequest(http.MethodPost, "/bake/start", nil)
	w2 := httptest.NewRecorder()

	server.handleLoafStart(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Should be able to start new bake after completion, got status %d", w2.Code)
	}
}

func TestStatus(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Start a bake and log some events
	server.storage.AppendEvent(nil)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["events"] == nil {
		t.Error("Expected events in status response")
	}
}

func TestQRCodePDF(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	// Create qrcodes directory and dummy PDF in current directory
	qrDir := "./qrcodes"
	os.MkdirAll(qrDir, 0755)
	pdfPath := filepath.Join(qrDir, "qrcodes.pdf")
	os.WriteFile(pdfPath, []byte("%PDF-1.4\ntest"), 0644)
	defer os.RemoveAll(qrDir)

	req := httptest.NewRequest(http.MethodGet, "/qrcodes.pdf", nil)
	w := httptest.NewRecorder()

	server.handleQRCodePDF(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("Expected PDF content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestWebUIPages(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	pages := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/temp", server.handleTempPage},
		{"/notes", server.handleNotesPage},
		{"/complete", server.handleCompletePage},
	}

	for _, page := range pages {
		t.Run(page.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, page.path, nil)
			w := httptest.NewRecorder()

			page.handler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if w.Header().Get("Content-Type") != "text/html" {
				t.Errorf("Expected HTML content type, got %s", w.Header().Get("Content-Type"))
			}

			if w.Body.Len() == 0 {
				t.Error("Expected non-empty response body")
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	server, tmpDir := setupTestServer(t)
	defer cleanup(tmpDir)

	tests := []struct {
		path   string
		method string
	}{
		{"/health", http.MethodPost},
		{"/bake/start", http.MethodPut},
		{"/status", http.MethodPost},
	}

	for _, tt := range tests {
		t.Run(tt.path+" "+tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			switch tt.path {
			case "/health":
				server.handleHealth(w, req)
			case "/bake/start":
				server.handleLoafStart(w, req)
			case "/status":
				server.handleStatus(w, req)
			}

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405, got %d", w.Code)
			}
		})
	}
}
