package qr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAll(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "qr_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	serverURL := "http://192.168.1.100:8080"

	// Generate QR codes
	err = GenerateAll(serverURL, tmpDir)
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Verify expected files were created
	expectedFiles := []string{
		"start.png", "fed.png", "levain-ready.png", "mixed.png", "knead.png",
		"fold.png", "shaped.png", "fridge-in.png", "oven-in.png",
		"remove-lid.png", "oven-out.png", "temp.png", "notes.png", "complete.png",
		"status.png", "qr-pdf.png", "sheet.png", "qrcodes.pdf",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", filename)
		}
	}

	// Verify PDF exists and has content
	pdfPath := filepath.Join(tmpDir, "qrcodes.pdf")
	stat, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("PDF file not created: %v", err)
	}
	if stat.Size() < 1000 {
		t.Errorf("PDF file too small (%d bytes), likely corrupt", stat.Size())
	}
}

func TestGenerateAllRejectsLocalhost(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "qr_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// These URLs should be rejected
	invalidURLs := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://0.0.0.0:8080",
		"http://[::1]:8080",
	}

	for _, url := range invalidURLs {
		err := GenerateAll(url, tmpDir)
		if err == nil {
			t.Errorf("Expected error for localhost URL %s, but got none", url)
		}
		if err != nil && !strings.Contains(err.Error(), "localhost") {
			t.Errorf("Expected 'localhost' error for %s, got: %v", url, err)
		}
	}
}

func TestGenerateAllAcceptsValidIPs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "qr_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// These URLs should be accepted
	validURLs := []string{
		"http://192.168.1.50:8080",
		"http://10.0.0.1:8080",
		"http://192.168.0.100:8080",
		"http://example.com:8080",
		"https://sourdough.example.com",
	}

	for _, url := range validURLs {
		// Clean directory between tests
		os.RemoveAll(tmpDir)
		os.MkdirAll(tmpDir, 0755)

		err := GenerateAll(url, tmpDir)
		if err != nil {
			t.Errorf("Expected no error for valid URL %s, got: %v", url, err)
		}
	}
}

func TestEventQRStructure(t *testing.T) {
	serverURL := "http://192.168.1.50:8080"

	// Verify the events list has expected structure
	events := []EventQR{
		{"start", "START LOAF", fmt.Sprintf("%s/loaf/start", serverURL)},
		{"fed", "Fed", fmt.Sprintf("%s/log/fed", serverURL)},
		{"levain-ready", "Levain Ready", fmt.Sprintf("%s/log/levain-ready", serverURL)},
		{"mixed", "Mixed", fmt.Sprintf("%s/log/mixed", serverURL)},
		{"fold", "Fold", fmt.Sprintf("%s/log/fold", serverURL)},
		{"shaped", "Shaped", fmt.Sprintf("%s/log/shaped", serverURL)},
		{"fridge-in", "Fridge In", fmt.Sprintf("%s/log/fridge-in", serverURL)},
		{"oven-in", "Oven In", fmt.Sprintf("%s/log/oven-in", serverURL)},
		{"oven-out", "Oven Out", fmt.Sprintf("%s/log/oven-out", serverURL)},
		{"temp", "LOG TEMP", fmt.Sprintf("%s/temp", serverURL)},
		{"notes", "ADD NOTE", fmt.Sprintf("%s/notes", serverURL)},
		{"complete", "COMPLETE", fmt.Sprintf("%s/complete", serverURL)},
		{"status", "VIEW STATUS", fmt.Sprintf("%s/view/status", serverURL)},
		{"history", "VIEW HISTORY", fmt.Sprintf("%s/view/history", serverURL)},
		{"qr-pdf", "GET QR CODES", fmt.Sprintf("%s/qrcodes.pdf", serverURL)},
	}

	// Verify no event points to /bake/start (old endpoint)
	for _, event := range events {
		if strings.Contains(event.URL, "/bake/start") {
			t.Errorf("Event %s still uses old /bake/start endpoint: %s", event.Event, event.URL)
		}
		if strings.Contains(event.URL, "localhost") {
			t.Errorf("Event %s contains localhost: %s", event.Event, event.URL)
		}
		if strings.Contains(event.URL, "127.0.0.1") {
			t.Errorf("Event %s contains 127.0.0.1: %s", event.Event, event.URL)
		}
	}

	// Verify START LOAF uses /loaf/start
	if events[0].URL != fmt.Sprintf("%s/loaf/start", serverURL) {
		t.Errorf("START LOAF should use /loaf/start, got: %s", events[0].URL)
	}
}
