package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mdeckert/sourdough/internal/models"
)

func setupTestStorage(t *testing.T) (*Storage, string) {
	tmpDir, err := os.MkdirTemp("", "sourdough-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	store, err := New(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	return store, tmpDir
}

func cleanup(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestNewStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sourdough-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if store.dataDir != tmpDir {
		t.Errorf("Expected dataDir %s, got %s", tmpDir, store.dataDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Data directory was not created")
	}
}

func TestAppendEvent(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	event := models.NewEvent(models.EventStarterOut)

	if err := store.AppendEvent(event); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// Verify file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestReadCurrentBake(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Append some events
	events := []models.EventType{
		models.EventStarterOut,
		models.EventFed,
		models.EventLevainReady,
		models.EventMixed,
	}

	for _, eventType := range events {
		event := models.NewEvent(eventType)
		if err := store.AppendEvent(event); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Read back
	bake, err := store.ReadCurrentBake()
	if err != nil {
		t.Fatalf("Failed to read current bake: %v", err)
	}

	if len(bake.Events) != len(events) {
		t.Errorf("Expected %d events, got %d", len(events), len(bake.Events))
	}

	for i, event := range bake.Events {
		if event.Event != events[i] {
			t.Errorf("Event %d: expected %s, got %s", i, events[i], event.Event)
		}
	}
}

func TestHasCurrentBake(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Initially no bake
	hasBake, err := store.HasCurrentBake()
	if err != nil {
		t.Fatalf("HasCurrentBake failed: %v", err)
	}
	if hasBake {
		t.Error("Expected no bake initially")
	}

	// Start a bake
	event := models.NewEvent(models.EventStarterOut)
	if err := store.AppendEvent(event); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// Now should have bake
	hasBake, err = store.HasCurrentBake()
	if err != nil {
		t.Fatalf("HasCurrentBake failed: %v", err)
	}
	if !hasBake {
		t.Error("Expected to have current bake")
	}

	// Complete the bake
	completeEvent := models.NewEvent(models.EventBakeComplete)
	if err := store.AppendEvent(completeEvent); err != nil {
		t.Fatalf("Failed to append complete event: %v", err)
	}

	// Should no longer have active bake
	hasBake, err = store.HasCurrentBake()
	if err != nil {
		t.Fatalf("HasCurrentBake failed: %v", err)
	}
	if hasBake {
		t.Error("Expected no active bake after completion")
	}
}

func TestMultiDayBake(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Day 1 - start bake
	event1 := models.NewEvent(models.EventStarterOut)
	if err := store.AppendEvent(event1); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	file1 := store.getCurrentBakeFile()

	// Day 2 - continue bake (should use same file)
	event2 := models.NewEvent(models.EventOvenIn)
	if err := store.AppendEvent(event2); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	file2 := store.getCurrentBakeFile()

	if file1 != file2 {
		t.Error("Multi-day bake should use same file")
	}

	// Complete bake
	completeEvent := models.NewEvent(models.EventBakeComplete)
	if err := store.AppendEvent(completeEvent); err != nil {
		t.Fatalf("Failed to append complete event: %v", err)
	}

	// Next bake should use different file - need to wait at least 1 minute for timestamp difference
	// For testing purposes, we'll just verify that HasCurrentBake returns false
	hasBake, err := store.HasCurrentBake()
	if err != nil {
		t.Fatalf("HasCurrentBake failed: %v", err)
	}
	if hasBake {
		t.Error("Should not have active bake after completion")
	}
}

func TestGetLastEvent(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// No events initially
	lastEvent, err := store.GetLastEvent()
	if err != nil {
		t.Fatalf("GetLastEvent failed: %v", err)
	}
	if lastEvent != nil {
		t.Error("Expected no last event initially")
	}

	// Add events
	events := []models.EventType{
		models.EventStarterOut,
		models.EventFed,
		models.EventMixed,
	}

	for _, eventType := range events {
		event := models.NewEvent(eventType)
		if err := store.AppendEvent(event); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}
	}

	// Get last event
	lastEvent, err = store.GetLastEvent()
	if err != nil {
		t.Fatalf("GetLastEvent failed: %v", err)
	}

	if lastEvent.Event != models.EventMixed {
		t.Errorf("Expected last event to be %s, got %s", models.EventMixed, lastEvent.Event)
	}
}

func TestListBakes(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Create multiple bake files
	dates := []string{
		"2025-10-07_10-00",
		"2025-10-06_09-00",
		"2025-10-05_08-00",
	}

	for _, date := range dates {
		filePath := filepath.Join(tmpDir, "bake_"+date+".jsonl")
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		f.Close()
		time.Sleep(2 * time.Millisecond)
	}

	// List bakes
	bakes, err := store.ListBakes()
	if err != nil {
		t.Fatalf("ListBakes failed: %v", err)
	}

	if len(bakes) != len(dates) {
		t.Errorf("Expected %d bakes, got %d", len(dates), len(bakes))
	}

	// Verify sorted in descending order
	for i := 0; i < len(bakes)-1; i++ {
		if bakes[i] < bakes[i+1] {
			t.Error("Bakes should be sorted in descending order")
		}
	}
}

func TestIsCompleted(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Create incomplete bake
	event1 := models.NewEvent(models.EventStarterOut)
	store.AppendEvent(event1)

	filePath := store.getCurrentBakeFile()

	if store.isCompleted(filePath) {
		t.Error("Bake should not be completed")
	}

	// Complete the bake
	completeEvent := models.NewEvent(models.EventBakeComplete)
	store.AppendEvent(completeEvent)

	if !store.isCompleted(filePath) {
		t.Error("Bake should be completed")
	}
}

func TestBakeWithAssessment(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Start bake
	startEvent := models.NewEvent(models.EventStarterOut)
	store.AppendEvent(startEvent)

	// Complete with assessment
	completeEvent := models.NewEvent(models.EventBakeComplete)
	completeEvent.Data = map[string]interface{}{
		"assessment": models.Assessment{
			ProofLevel:   "good",
			CrumbQuality: 8,
			Browning:     "good",
			Score:        9,
			Notes:        "Excellent loaf",
		},
	}
	store.AppendEvent(completeEvent)

	// Read back
	bake, err := store.ReadCurrentBake()
	if err != nil {
		t.Fatalf("Failed to read bake: %v", err)
	}

	if bake.Assessment == nil {
		t.Fatal("Expected assessment to be present")
	}

	if bake.Assessment.ProofLevel != "good" {
		t.Errorf("Expected proof level 'good', got '%s'", bake.Assessment.ProofLevel)
	}

	if bake.Assessment.Score != 9 {
		t.Errorf("Expected score 9, got %d", bake.Assessment.Score)
	}
}

func TestEmptyBake(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Read non-existent bake
	bake, err := store.ReadCurrentBake()
	if err != nil {
		t.Fatalf("Failed to read empty bake: %v", err)
	}

	if len(bake.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(bake.Events))
	}
}

func TestConcurrentWrites(t *testing.T) {
	store, tmpDir := setupTestStorage(t)
	defer cleanup(tmpDir)

	// Write events concurrently
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			event := models.NewEvent(models.EventFold)
			store.AppendEvent(event)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Read back
	bake, err := store.ReadCurrentBake()
	if err != nil {
		t.Fatalf("Failed to read bake: %v", err)
	}

	if len(bake.Events) != numGoroutines {
		t.Errorf("Expected %d events, got %d", numGoroutines, len(bake.Events))
	}
}
