package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mdeckert/sourdough/internal/models"
	"github.com/mdeckert/sourdough/internal/storage"
)

// Server handles HTTP requests
type Server struct {
	storage *storage.Storage
	port    string
}

// New creates a new Server instance
func New(storage *storage.Storage, port string) *Server {
	return &Server{
		storage: storage,
		port:    port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/bake/start", s.handleBakeStart)
	mux.HandleFunc("/log/", s.handleLog)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/temp", s.handleTempPage)
	mux.HandleFunc("/notes", s.handleNotesPage)
	mux.HandleFunc("/complete", s.handleCompletePage)
	mux.HandleFunc("/qrcodes.pdf", s.handleQRCodePDF)

	// Wrap mux with logging middleware
	handler := s.loggingMiddleware(mux)

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, handler)
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

// handleBakeStart starts a new bake
func (s *Server) handleBakeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if there's already a bake today
	hasBake, err := s.storage.HasCurrentBake()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking current bake: %v", err), http.StatusInternalServerError)
		return
	}

	if hasBake {
		// Show nice message if accessed from browser
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Already Started</title><style>body{font-family:sans-serif;background:#f59e0b;color:white;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:20px;text-align:center;}h1{font-size:48px;margin:0 0 10px 0;}p{font-size:20px;margin:0;}</style></head><body><div><h1>⚠️</h1><h1>Already Started</h1><p>You already started a bake today!</p></div></body></html>`))
			return
		}
		http.Error(w, "Bake already started today", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Error starting bake: %v", err), http.StatusInternalServerError)
		return
	}

	// Show nice success message if accessed from browser
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Bake Started</title><style>body{font-family:sans-serif;background:#10b981;color:white;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;padding:20px;text-align:center;}h1{font-size:48px;margin:0 0 10px 0;}p{font-size:20px;margin:0;}</style></head><body><div><h1>✅</h1><h1>Bake Started!</h1><p>Your loaf has been logged</p></div></body></html>`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "bake started",
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

		// Check if this is dough temp or kitchen temp
		if r.URL.Query().Get("type") == "dough" {
			event.WithDoughTemp(temp)
		} else {
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
			models.EventBakeComplete: true,
		}

		if !validEvents[eventType] {
			http.Error(w, fmt.Sprintf("Invalid event type: %s", eventType), http.StatusBadRequest)
			return
		}

		event = models.NewEvent(eventType)

		// Handle bake-complete with assessment data (from web UI)
		if eventType == models.EventBakeComplete && r.Method == http.MethodPost {
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
