package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdeckert/sourdough/internal/ecobee"
	"github.com/mdeckert/sourdough/internal/models"
	"github.com/mdeckert/sourdough/internal/storage"
)

// Server handles HTTP requests
type Server struct {
	storage *storage.Storage
	ecobee  *ecobee.Client
	port    string
}

// New creates a new Server instance
func New(storage *storage.Storage, ecobeeClient *ecobee.Client, port string) *Server {
	return &Server{
		storage: storage,
		ecobee:  ecobeeClient,
		port:    port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/loaf/start", s.handleLoafStart)
	mux.HandleFunc("/bake/start", s.handleLoafStart) // Legacy support
	mux.HandleFunc("/log/", s.handleLog)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/view/status", s.handleViewStatus)
	mux.HandleFunc("/view/history", s.handleViewHistory)
	mux.HandleFunc("/api/bake/current", s.handleAPICurrentBake)
	mux.HandleFunc("/api/bake/", s.handleAPIBake)
	mux.HandleFunc("/api/bakes", s.handleAPIBakesList)
	mux.HandleFunc("/temp", s.handleTempPage)
	mux.HandleFunc("/notes", s.handleNotesPage)
	mux.HandleFunc("/complete", s.handleCompletePage)
	mux.HandleFunc("/qrcodes.pdf", s.handleQRCodePDF)

	// Start automatic temperature logging if Ecobee is enabled
	if s.ecobee.IsEnabled() {
		go s.autoLogTemperature()
	}

	// Wrap mux with logging middleware
	handler := s.loggingMiddleware(mux)

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// autoLogTemperature logs kitchen temperature every 4 hours if there's an active bake
func (s *Server) autoLogTemperature() {
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	log.Printf("Automatic temperature logging enabled (every 4 hours)")

	for range ticker.C {
		// Check if there's an active bake
		hasBake, err := s.storage.HasCurrentBake()
		if err != nil {
			log.Printf("Warning: Failed to check for active bake: %v", err)
			continue
		}

		if !hasBake {
			// No active bake, skip logging
			continue
		}

		// Fetch temperature from Ecobee
		temp, err := s.ecobee.GetTemperature()
		if err != nil {
			log.Printf("Warning: Failed to auto-log temperature: %v", err)
			continue
		}

		if temp <= 0 {
			log.Printf("Warning: Invalid temperature from Ecobee: %.1f", temp)
			continue
		}

		// Create temperature event
		event := models.NewEvent(models.EventTemperature).WithTemp(temp)

		// Save event
		if err := s.storage.AppendEvent(event); err != nil {
			log.Printf("Error: Failed to save auto-logged temperature: %v", err)
			continue
		}

		log.Printf("Auto-logged kitchen temperature: %.1f°F", temp)
	}
}

// loggingMiddleware logs all requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// handleHealth returns a simple health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleLoafStart starts a new loaf
func (s *Server) handleLoafStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if there's already an active loaf
	hasBake, err := s.storage.HasCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking current loaf: %v", err), http.StatusInternalServerError)
		return
	}

	if hasBake {
		// Show nice message if accessed from browser
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Already Started</title><style>body{font-family:sans-serif;background:#f59e0b;color:white;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:20px;text-align:center;}h1{font-size:48px;margin:0 0 10px 0;}p{font-size:20px;margin:0;}</style></head><body><div><h1>⚠️</h1><h1>Already Started</h1><p>You already started a loaf!</p></div></body></html>`))
			return
		}
		http.Error(w, "Loaf already started", http.StatusBadRequest)
		return
	}

	// Create starter-out event to begin the bake
	event := models.NewEvent(models.EventStarterOut)

	// Check for temperature in query params
	if tempStr := r.URL.Query().Get("temp"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			event.WithTemp(temp)
		}
	}

	if err := s.storage.AppendEvent(event); err != nil {
		http.Error(w, fmt.Sprintf("Error starting loaf: %v", err), http.StatusInternalServerError)
		return
	}

	// Show nice success message if accessed from browser
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Loaf Started</title><style>body{font-family:sans-serif;background:#10b981;color:white;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:20px;text-align:center;}h1{font-size:48px;margin:0 0 10px 0;}p{font-size:20px;margin:0;}</style></head><body><div><h1>✅</h1><h1>Loaf Started!</h1><p>Your loaf has been logged</p></div></body></html>`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "loaf started",
	})
}

// handleLog handles logging events
// Supports: /log/fold, /log/shaped, /log/temp/76, etc.
func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the path: /log/{event} or /log/temp/{value}
	path := strings.TrimPrefix(r.URL.Path, "/log/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Event type required", http.StatusBadRequest)
		return
	}

	var event *models.Event

	// Handle note logging: /log/note (expects JSON body)
	if parts[0] == "note" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var noteData struct {
			Note string `json:"note"`
		}

		if err := json.NewDecoder(r.Body).Decode(&noteData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if noteData.Note == "" {
			http.Error(w, "Note cannot be empty", http.StatusBadRequest)
			return
		}

		event = models.NewEvent(models.EventNote)
		event.WithNote(noteData.Note)
	} else if parts[0] == "temp" {
		// Handle temperature logging: /log/temp/76
		if len(parts) < 2 {
			http.Error(w, "Temperature value required", http.StatusBadRequest)
			return
		}

		temp, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			http.Error(w, "Invalid temperature value", http.StatusBadRequest)
			return
		}

		event = models.NewEvent(models.EventTemperature)

		// Check temperature type: dough/loaf, oven, or kitchen
		tempType := r.URL.Query().Get("type")
		if tempType == "dough" {
			// Dough or loaf internal temp (both use dough_temp_f)
			event.WithDoughTemp(temp)
		} else if tempType == "oven" {
			// Oven temperature (use kitchen temp_f field for now)
			event.WithTemp(temp)
		} else {
			// Kitchen temp (manual, since auto-logged via Ecobee)
			event.WithTemp(temp)
		}
	} else {
		// Handle regular events
		eventType := models.EventType(parts[0])

		// Validate event type
		validEvents := map[models.EventType]bool{
			models.EventStarterOut:   true,
			models.EventFed:          true,
			models.EventLevainReady:  true,
			models.EventMixed:        true,
			models.EventFold:         true,
			models.EventShaped:       true,
			models.EventFridgeIn:     true,
			models.EventFridgeOut:    true,
			models.EventOvenIn:       true,
			models.EventRemoveLid:    true,
			models.EventOvenOut:      true,
			models.EventLoafComplete: true,
		}

		if !validEvents[eventType] {
			http.Error(w, fmt.Sprintf("Invalid event type: %s", eventType), http.StatusBadRequest)
			return
		}

		event = models.NewEvent(eventType)

		// Handle loaf-complete with assessment data (from web UI)
		if eventType == models.EventLoafComplete && r.Method == http.MethodPost {
			var reqData struct {
				Assessment models.Assessment `json:"assessment"`
			}
			if err := json.NewDecoder(r.Body).Decode(&reqData); err == nil {
				if event.Data == nil {
					event.Data = make(map[string]interface{})
				}
				event.Data["assessment"] = reqData.Assessment
			}
		}

		// Handle fold count
		if eventType == models.EventFold {
			// Try to get fold count from last event
			lastEvent, _ := s.storage.GetLastEvent()
			foldCount := 1
			if lastEvent != nil && lastEvent.Event == models.EventFold && lastEvent.FoldCount != nil {
				foldCount = *lastEvent.FoldCount + 1
			}
			event.WithFoldCount(foldCount)
		}

		// Check for temperature in query params
		if tempStr := r.URL.Query().Get("temp"); tempStr != "" {
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
				event.WithTemp(temp)
			}
		}

		if doughTempStr := r.URL.Query().Get("dough_temp"); doughTempStr != "" {
			if temp, err := strconv.ParseFloat(doughTempStr, 64); err == nil {
				event.WithDoughTemp(temp)
			}
		}

		// Check for note
		if note := r.URL.Query().Get("note"); note != "" {
			event.WithNote(note)
		}
	}

	// Auto-fetch kitchen temp from Ecobee if enabled and no temp already set
	// Only fetch for non-temperature events to avoid overwriting manual temps
	if s.ecobee.IsEnabled() && event.Event != models.EventTemperature && event.TempF == nil {
		if temp, err := s.ecobee.GetTemperature(); err == nil && temp > 0 {
			event.WithTemp(temp)
			log.Printf("Auto-fetched kitchen temp from Ecobee: %.1f°F", temp)
		} else if err != nil {
			log.Printf("Warning: Failed to fetch Ecobee temp: %v", err)
		}
	}

	// Save event
	if err := s.storage.AppendEvent(event); err != nil {
		http.Error(w, fmt.Sprintf("Error logging event: %v", err), http.StatusInternalServerError)
		return
	}

	// Show nice success message if accessed from browser (GET request)
	if r.Method == http.MethodGet {
		eventName := string(event.Event)
		w.Header().Set("Content-Type", "text/html")
		successHTML := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Event Logged</title><style>body{font-family:sans-serif;background:#10b981;color:white;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:20px;text-align:center;}h1{font-size:48px;margin:0 0 10px 0;}p{font-size:20px;margin:0;}.time{opacity:0.8;font-size:16px;margin-top:10px;}</style></head><body><div><h1>✅</h1><h1>%s</h1><p>Event logged successfully</p><p class="time">%s</p></div></body></html>`, eventName, event.Timestamp.Format("3:04 PM"))
		w.Write([]byte(successHTML))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "logged",
		"event":  event,
	})
}

// handleStatus returns the current bake status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bake, err := s.storage.ReadCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading bake: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bake)
}

// handleTempPage serves the temperature logging web UI
func (s *Server) handleTempPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tempPageHTML))
}

// handleNotesPage serves the notes logging web UI
func (s *Server) handleNotesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(notesPageHTML))
}

// handleCompletePage serves the bake completion assessment web UI
func (s *Server) handleCompletePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(completePageHTML))
}

// handleQRCodePDF serves the QR codes PDF file
func (s *Server) handleQRCodePDF(w http.ResponseWriter, r *http.Request) {
	pdfPath := "./qrcodes/qrcodes.pdf"

	// Check if file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		http.Error(w, "QR codes PDF not found. Generate it with: ./bin/qrgen http://YOUR_SERVER_URL:8080", http.StatusNotFound)
		return
	}

	// Read the PDF file
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, "Error reading PDF file", http.StatusInternalServerError)
		return
	}

	// Serve the PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=sourdough-qrcodes.pdf")
	w.Write(pdfData)
}

// handleViewStatus serves the current/recent bake status page with graphs
func (s *Server) handleViewStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(statusViewPageHTML))
}

// handleViewHistory serves the bake history list page
func (s *Server) handleViewHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(historyViewPageHTML))
}

// handleAPICurrentBake returns the current or most recent bake as JSON
func (s *Server) handleAPICurrentBake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to get current bake
	bake, err := s.storage.ReadCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading bake: %v", err), http.StatusInternalServerError)
		return
	}

	// If current bake is empty or completed, get the most recent one
	if len(bake.Events) == 0 || (len(bake.Events) > 0 && bake.Events[len(bake.Events)-1].Event == models.EventLoafComplete) {
		dates, err := s.storage.ListBakes()
		if err == nil && len(dates) > 0 {
			// Get most recent bake
			bake, err = s.storage.ReadBake(dates[0])
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading recent bake: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bake)
}

// handleAPIBake returns a specific bake by date/timestamp
func (s *Server) handleAPIBake(w http.ResponseWriter, r *http.Request) {
	// Extract date from path: /api/bake/2025-10-07_19-06
	path := strings.TrimPrefix(r.URL.Path, "/api/bake/")
	if path == "" {
		http.Error(w, "Bake date/timestamp required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		bake, err := s.storage.ReadBake(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading bake: %v", err), http.StatusInternalServerError)
			return
		}

		if len(bake.Events) == 0 {
			http.Error(w, "Bake not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bake)

	case http.MethodDelete:
		err := s.storage.DeleteBake(path)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "date": path})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPIBakesList returns list of all bakes with summary info
func (s *Server) handleAPIBakesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dates, err := s.storage.ListBakes()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing bakes: %v", err), http.StatusInternalServerError)
		return
	}

	type BakeSummary struct {
		Date       string              `json:"date"`
		StartTime  string              `json:"start_time"`
		EndTime    string              `json:"end_time,omitempty"`
		EventCount int                 `json:"event_count"`
		Completed  bool                `json:"completed"`
		Assessment *models.Assessment  `json:"assessment,omitempty"`
	}

	summaries := make([]BakeSummary, 0, len(dates))

	for _, date := range dates {
		bake, err := s.storage.ReadBake(date)
		if err != nil || len(bake.Events) == 0 {
			continue
		}

		summary := BakeSummary{
			Date:       date,
			StartTime:  bake.Events[0].Timestamp.Format("2006-01-02 15:04"),
			EventCount: len(bake.Events),
			Assessment: bake.Assessment,
		}

		// Check if completed
		lastEvent := bake.Events[len(bake.Events)-1]
		if lastEvent.Event == models.EventLoafComplete {
			summary.Completed = true
			summary.EndTime = lastEvent.Timestamp.Format("2006-01-02 15:04")
		}

		summaries = append(summaries, summary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}
