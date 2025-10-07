package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mdeckert/sourdough/internal/models"
)

// Storage handles reading and writing bake data
type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

// New creates a new Storage instance
func New(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// getCurrentBakeFile returns the path to the current active bake file
// An active bake is one that hasn't been completed (no bake-complete event)
func (s *Storage) getCurrentBakeFile() string {
	// First, check if there's an active bake (most recent file without bake-complete)
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		// If we can't read dir, fall back to today's date
		date := time.Now().Format("2006-01-02")
		return filepath.Join(s.dataDir, fmt.Sprintf("bake_%s.jsonl", date))
	}

	// Sort files by modification time (most recent first)
	var bakeFiles []struct {
		name    string
		modTime time.Time
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "bake_") || !strings.HasSuffix(file.Name(), ".jsonl") {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		bakeFiles = append(bakeFiles, struct {
			name    string
			modTime time.Time
		}{file.Name(), info.ModTime()})
	}

	sort.Slice(bakeFiles, func(i, j int) bool {
		return bakeFiles[i].modTime.After(bakeFiles[j].modTime)
	})

	// Check most recent files for active bake (no bake-complete event)
	for _, bf := range bakeFiles {
		filePath := filepath.Join(s.dataDir, bf.name)
		if !s.isCompleted(filePath) {
			return filePath
		}
	}

	// No active bake found, create new one with current timestamp
	date := time.Now().Format("2006-01-02_15-04")
	return filepath.Join(s.dataDir, fmt.Sprintf("bake_%s.jsonl", date))
}

// isCompleted checks if a bake file has a bake-complete event
func (s *Storage) isCompleted(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event models.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Event == models.EventBakeComplete {
			return true
		}
	}
	return false
}

// getBakeFile returns the path to a specific bake file by date
func (s *Storage) getBakeFile(date string) string {
	return filepath.Join(s.dataDir, fmt.Sprintf("bake_%s.jsonl", date))
}

// AppendEvent appends an event to the current bake file
func (s *Storage) AppendEvent(event *models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := s.getCurrentBakeFile()

	// Open file in append mode, create if doesn't exist
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open bake file: %w", err)
	}
	defer f.Close()

	// Encode event as JSON and write
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(event); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// ReadCurrentBake reads all events from the current active bake
func (s *Storage) ReadCurrentBake() (*models.Bake, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := s.getCurrentBakeFile()

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &models.Bake{
			Date:   time.Now().Format("2006-01-02"),
			Events: []models.Event{},
		}, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bake file: %w", err)
	}
	defer f.Close()

	var events []models.Event
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		var event models.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading bake file: %w", err)
	}

	// Extract date from first event or use today
	date := time.Now().Format("2006-01-02")
	if len(events) > 0 {
		date = events[0].Timestamp.Format("2006-01-02")
	}

	bake := &models.Bake{
		Date:   date,
		Events: events,
	}

	// Check if last event is bake-complete with assessment
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		if lastEvent.Event == models.EventBakeComplete && lastEvent.Data != nil {
			if assessmentData, ok := lastEvent.Data["assessment"]; ok {
				assessmentJSON, _ := json.Marshal(assessmentData)
				var assessment models.Assessment
				if json.Unmarshal(assessmentJSON, &assessment) == nil {
					bake.Assessment = &assessment
				}
			}
		}
	}

	return bake, nil
}

// ReadBake reads all events from a specific bake by date
func (s *Storage) ReadBake(date string) (*models.Bake, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := s.getBakeFile(date)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &models.Bake{
			Date:   date,
			Events: []models.Event{},
		}, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bake file: %w", err)
	}
	defer f.Close()

	var events []models.Event
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		var event models.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			// Skip malformed lines
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading bake file: %w", err)
	}

	bake := &models.Bake{
		Date:   date,
		Events: events,
	}

	// Check if last event is an assessment (bake-complete with assessment data)
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		if lastEvent.Event == models.EventBakeComplete && lastEvent.Data != nil {
			// Try to extract assessment from data
			if assessmentData, ok := lastEvent.Data["assessment"]; ok {
				assessmentJSON, _ := json.Marshal(assessmentData)
				var assessment models.Assessment
				if json.Unmarshal(assessmentJSON, &assessment) == nil {
					bake.Assessment = &assessment
				}
			}
		}
	}

	return bake, nil
}

// ListBakes returns a list of all bake dates in descending order
func (s *Storage) ListBakes() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var dates []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, "bake_") && strings.HasSuffix(name, ".jsonl") {
			// Extract date from filename: bake_2025-10-07.jsonl -> 2025-10-07
			date := strings.TrimSuffix(strings.TrimPrefix(name, "bake_"), ".jsonl")
			dates = append(dates, date)
		}
	}

	// Sort in descending order (most recent first)
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	return dates, nil
}

// HasCurrentBake checks if there's an active (uncompleted) bake
func (s *Storage) HasCurrentBake() (bool, error) {
	filePath := s.getCurrentBakeFile()

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	}

	// Check if it has events and isn't completed
	bake, err := s.ReadCurrentBake()
	if err != nil {
		return false, err
	}

	if len(bake.Events) == 0 {
		return false, nil
	}

	// Check if the bake is completed
	for _, event := range bake.Events {
		if event.Event == models.EventBakeComplete {
			return false, nil
		}
	}

	return true, nil
}

// GetLastEvent returns the most recent event from the current bake
func (s *Storage) GetLastEvent() (*models.Event, error) {
	bake, err := s.ReadCurrentBake()
	if err != nil {
		return nil, err
	}

	if len(bake.Events) == 0 {
		return nil, nil
	}

	return &bake.Events[len(bake.Events)-1], nil
}
