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
	mux.HandleFunc("/log/oven-in", s.handleOvenInLog) // Must be before /log/
	mux.HandleFunc("/log/remove-lid", s.handleRemoveLidLog) // Must be before /log/
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
	mux.HandleFunc("/images/", s.handleImage)

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

// autoLogTemperature logs kitchen temperature on a fixed schedule (12am, 4am, 8am, 12pm, 4pm, 8pm)
func (s *Server) autoLogTemperature() {
	log.Printf("Automatic temperature logging enabled (every 4 hours on fixed schedule)")

	// Calculate time until next scheduled log
	now := time.Now()
	nextLog := getNextLogTime(now)
	duration := nextLog.Sub(now)

	log.Printf("Next auto-log scheduled for: %s (in %s)", nextLog.Format("3:04 PM"), duration.Round(time.Minute))

	// Initial delay to sync with schedule
	time.Sleep(duration)

	// Log immediately at the scheduled time
	s.logTemperature()

	// Then log every 4 hours
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.logTemperature()
	}
}

// getNextLogTime returns the next scheduled log time (00:00, 04:00, 08:00, 12:00, 16:00, 20:00)
func getNextLogTime(now time.Time) time.Time {
	// Round to next 4-hour boundary
	hour := now.Hour()
	nextHour := ((hour / 4) + 1) * 4
	if nextHour >= 24 {
		nextHour = 0
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), nextHour, 0, 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// logTemperature performs the actual temperature logging
func (s *Server) logTemperature() {
	// Check if there's an active bake
	hasBake, err := s.storage.HasCurrentBake()
	if err != nil {
		log.Printf("Warning: Failed to check for active bake: %v", err)
		return
	}

	if !hasBake {
		log.Printf("Skipping auto-log: no active bake")
		return
	}

	// Fetch temperature from Ecobee
	temp, err := s.ecobee.GetTemperature()
	if err != nil {
		log.Printf("Warning: Failed to auto-log temperature: %v", err)
		return
	}

	if temp <= 0 {
		log.Printf("Warning: Invalid temperature from Ecobee: %.1f", temp)
		return
	}

	// Create temperature event
	event := models.NewEvent(models.EventTemperature).WithTemp(temp)

	// Save event
	if err := s.storage.AppendEvent(event); err != nil {
		log.Printf("Error: Failed to save auto-logged temperature: %v", err)
		return
	}

	log.Printf("Auto-logged kitchen temperature: %.1f°F", temp)
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

	// Handle note logging: /log/note (expects multipart form or JSON body)
	if parts[0] == "note" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var noteText string
		var imageFilename string
		var doughTemp *float64

		// Check if this is multipart form data (with possible image)
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Parse multipart form (max 10MB)
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				http.Error(w, "Failed to parse form data", http.StatusBadRequest)
				return
			}

			noteText = r.FormValue("note")

			// Handle optional dough temperature
			if doughTempStr := r.FormValue("dough_temp"); doughTempStr != "" {
				if temp, err := strconv.ParseFloat(doughTempStr, 64); err == nil {
					doughTemp = &temp
				}
			}

			// Handle image upload if present
			file, header, err := r.FormFile("image")
			if err == nil {
				defer file.Close()

				// Validate image type
				if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
					http.Error(w, "Invalid image type", http.StatusBadRequest)
					return
				}

				// Save image with timestamp-based filename
				imageFilename = fmt.Sprintf("%d.jpg", time.Now().UnixMilli())
				if err := s.storage.SaveImage(imageFilename, file); err != nil {
					http.Error(w, fmt.Sprintf("Failed to save image: %v", err), http.StatusInternalServerError)
					return
				}
			}

			// Require either note text or image
			if noteText == "" && imageFilename == "" {
				http.Error(w, "Note text or image required", http.StatusBadRequest)
				return
			}
		} else {
			// JSON body (backward compatibility)
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

			noteText = noteData.Note
		}

		event = models.NewEvent(models.EventNote)
		event.WithNote(noteText)
		if imageFilename != "" {
			event.WithImage(imageFilename)
		}
		if doughTemp != nil {
			event.WithDoughTemp(*doughTemp)
		}
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
	// Skip for temperature events (to avoid overwriting manual temps) and notes (not relevant)
	if s.ecobee.IsEnabled() && event.Event != models.EventTemperature && event.Event != models.EventNote && event.TempF == nil {
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

// handleOvenInPage serves the oven-in temperature selection web UI
func (s *Server) handleOvenInPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(ovenInPageHTML))
}

// handleRemoveLidPage serves the remove-lid temperature selection web UI
func (s *Server) handleRemoveLidPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(removeLidPageHTML))
}

// handleOvenInLog handles oven-in event logging with temperature selection
func (s *Server) handleOvenInLog(w http.ResponseWriter, r *http.Request) {
	// GET request: show the UI page
	if r.Method == http.MethodGet {
		s.handleOvenInPage(w, r)
		return
	}

	// POST request: log the event with temperature
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get temp from query parameter
	tempStr := r.URL.Query().Get("temp")
	if tempStr == "" {
		http.Error(w, "Missing temp parameter", http.StatusBadRequest)
		return
	}

	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil || temp < 0 || temp > 600 {
		http.Error(w, "Invalid temperature", http.StatusBadRequest)
		return
	}

	// Check if there's an active bake
	hasBake, err := s.storage.HasCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking current bake: %v", err), http.StatusInternalServerError)
		return
	}
	if !hasBake {
		http.Error(w, "No active bake. Start a new loaf first", http.StatusBadRequest)
		return
	}

	// Create event with oven temperature
	event := models.NewEvent(models.EventOvenIn).WithTemp(temp)

	// Append event to current bake
	if err := s.storage.AppendEvent(event); err != nil {
		http.Error(w, fmt.Sprintf("Failed to log event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"event":  event,
	})
}

// handleRemoveLidLog handles remove-lid event logging with temperature selection
func (s *Server) handleRemoveLidLog(w http.ResponseWriter, r *http.Request) {
	// GET request: show the UI page
	if r.Method == http.MethodGet {
		s.handleRemoveLidPage(w, r)
		return
	}

	// POST request: log the event with temperature
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get temp from query parameter
	tempStr := r.URL.Query().Get("temp")
	if tempStr == "" {
		http.Error(w, "Missing temp parameter", http.StatusBadRequest)
		return
	}

	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil || temp < 0 || temp > 600 {
		http.Error(w, "Invalid temperature", http.StatusBadRequest)
		return
	}

	// Check if there's an active bake
	hasBake, err := s.storage.HasCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking current bake: %v", err), http.StatusInternalServerError)
		return
	}
	if !hasBake {
		http.Error(w, "No active bake. Start a new loaf first", http.StatusBadRequest)
		return
	}

	// Create event with oven temperature
	event := models.NewEvent(models.EventRemoveLid).WithTemp(temp)

	// Append event to current bake
	if err := s.storage.AppendEvent(event); err != nil {
		http.Error(w, fmt.Sprintf("Failed to log event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"event":  event,
	})
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

// handleImage serves image files for bakes
// URL format: /images/bake_YYYY-MM-DD_HH-MM/filename.jpg
func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /images/bake_2025-10-07_19-13-49/1696721234567.jpg
	path := strings.TrimPrefix(r.URL.Path, "/images/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) != 2 {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	bakeName := parts[0]
	filename := parts[1]

	// Extract date from bake name (remove "bake_" prefix)
	bakeDate := strings.TrimPrefix(bakeName, "bake_")

	// Get image path from storage
	imagePath := s.storage.GetImagePath(bakeDate, filename)

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, imagePath)
}
