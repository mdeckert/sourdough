package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
// An active bake is one that hasn't been completed (no loaf-complete event)
func (s *Storage) getCurrentBakeFile() string {
	// First, check if there's an active bake (most recent file without loaf-complete)
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

	// Check most recent files for active bake (no loaf-complete event)
	for _, bf := range bakeFiles {
		filePath := filepath.Join(s.dataDir, bf.name)
		if !s.isCompleted(filePath) {
			return filePath
		}
	}

	// No active bake found, create new one with current timestamp (including seconds to prevent collisions)
	date := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(s.dataDir, fmt.Sprintf("bake_%s.jsonl", date))
}

// isCompleted checks if a bake file ENDS with a loaf-complete event (no events after)
func (s *Storage) isCompleted(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
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

	// File is completed only if it ends with loaf-complete (no events after)
	if len(events) == 0 {
		return false
	}

	lastEvent := events[len(events)-1]
	return lastEvent.Event == models.EventLoafComplete
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

	// Find the last loaf-complete event index
	// If there are events after it, return only those (new bake started in same file)
	lastCompleteIdx := -1
	for i, event := range events {
		if event.Event == models.EventLoafComplete {
			lastCompleteIdx = i
		}
	}

	// If there's a loaf-complete and events after it, return only events after completion
	if lastCompleteIdx >= 0 && lastCompleteIdx < len(events)-1 {
		events = events[lastCompleteIdx+1:]
	} else if lastCompleteIdx >= 0 && lastCompleteIdx == len(events)-1 {
		// File ends with loaf-complete, extract assessment but return empty events
		var assessment *models.Assessment
		lastEvent := events[len(events)-1]
		if lastEvent.Data != nil {
			if assessmentData, ok := lastEvent.Data["assessment"]; ok {
				assessmentJSON, _ := json.Marshal(assessmentData)
				var a models.Assessment
				if json.Unmarshal(assessmentJSON, &a) == nil {
					assessment = &a
				}
			}
		}

		return &models.Bake{
			Date:       time.Now().Format("2006-01-02"),
			Events:     []models.Event{},
			Assessment: assessment,
		}, nil
	}

	// Extract date from first event or use today
	date := time.Now().Format("2006-01-02")
	if len(events) > 0 {
		date = events[0].Timestamp.Format("2006-01-02")
	}

	// Extract filename without extension
	filename := strings.TrimSuffix(filepath.Base(filePath), ".jsonl")

	bake := &models.Bake{
		Date:     date,
		Filename: filename,
		Events:   events,
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

	// Extract filename without extension
	filename := strings.TrimSuffix(filepath.Base(filePath), ".jsonl")

	bake := &models.Bake{
		Date:     date,
		Filename: filename,
		Events:   events,
	}

	// Check if last event is an assessment (loaf-complete with assessment data)
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		if lastEvent.Event == models.EventLoafComplete && lastEvent.Data != nil {
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
		if event.Event == models.EventLoafComplete {
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

// DeleteBake moves a bake file to the trash directory
func (s *Storage) DeleteBake(date string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the source file path
	srcPath := s.getBakeFile(date)

	// Check if file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("bake not found: %s", date)
	}

	// Create trash directory if it doesn't exist
	trashDir := filepath.Join(s.dataDir, "trash")
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return fmt.Errorf("failed to create trash directory: %w", err)
	}

	// Destination path in trash
	dstPath := filepath.Join(trashDir, fmt.Sprintf("bake_%s.jsonl", date))

	// Move file to trash
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move bake to trash: %w", err)
	}

	return nil
}

// SaveImage saves an uploaded image file for the current bake
func (s *Storage) SaveImage(filename string, data io.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current bake date from filename
	bakeFile := s.getCurrentBakeFile()
	bakeName := strings.TrimSuffix(filepath.Base(bakeFile), ".jsonl")

	// Create images directory for this bake
	imageDir := filepath.Join(s.dataDir, "images", bakeName)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Save image file
	imagePath := filepath.Join(imageDir, filename)
	outFile, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, data); err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// GetImagePath returns the full path to an image file for a given bake
func (s *Storage) GetImagePath(bakeDate, filename string) string {
	bakeName := fmt.Sprintf("bake_%s", bakeDate)
	return filepath.Join(s.dataDir, "images", bakeName, filename)
}

// DeleteEvent removes an event from the current bake by index and timestamp
func (s *Storage) DeleteEvent(index int, timestamp string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current bake file
	bakeFile := s.getCurrentBakeFile()

	// Read all events
	file, err := os.Open(bakeFile)
	if err != nil {
		return fmt.Errorf("failed to open bake file: %w", err)
	}

	var events []models.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event models.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			file.Close()
			return fmt.Errorf("failed to parse event: %w", err)
		}
		events = append(events, event)
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading bake file: %w", err)
	}

	// Validate index
	if index < 0 || index >= len(events) {
		return fmt.Errorf("invalid event index: %d", index)
	}

	// Verify timestamp matches as extra safety
	if events[index].Timestamp.Format(time.RFC3339Nano) != timestamp {
		return fmt.Errorf("timestamp mismatch - event may have changed")
	}

	// Remove the event at the specified index
	events = append(events[:index], events[index+1:]...)

	// Write back all events
	tempFile := bakeFile + ".tmp"
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			f.Close()
			os.Remove(tempFile)
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		if _, err := f.Write(data); err != nil {
			f.Close()
			os.Remove(tempFile)
			return fmt.Errorf("failed to write event: %w", err)
		}

		if _, err := f.Write([]byte("\n")); err != nil {
			f.Close()
			os.Remove(tempFile)
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	if err := f.Close(); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically replace original file with temp file
	if err := os.Rename(tempFile, bakeFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to replace bake file: %w", err)
	}

	return nil
}
