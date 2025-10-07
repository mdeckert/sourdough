package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdeckert/sourdough/internal/models"
	"github.com/mdeckert/sourdough/internal/storage"
)

var (
	serverURL = getEnv("SOURDOUGH_SERVER_URL", "http://localhost:8080")
	dataDir   = getEnv("SOURDOUGH_DATA_DIR", "./data")
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		handleStart()
	case "log":
		handleLog()
	case "temp":
		handleTemp()
	case "status":
		handleStatus()
	case "complete":
		handleComplete()
	case "history":
		handleHistory()
	case "review":
		handleReview()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Sourdough Bread Logger")
	fmt.Println("\nUsage:")
	fmt.Println("  sourdough start                    Start a new bake")
	fmt.Println("  sourdough log <event> [options]    Log an event")
	fmt.Println("  sourdough temp <value>             Log temperature")
	fmt.Println("  sourdough status                   Show current bake status")
	fmt.Println("  sourdough complete                 Complete bake with assessment")
	fmt.Println("  sourdough history [n]              Show recent bakes (default: 10)")
	fmt.Println("  sourdough review <date>            Review a specific bake")
	fmt.Println("\nEvents:")
	fmt.Println("  starter-out, fed, levain-ready, mixed, fold, shaped,")
	fmt.Println("  fridge-in, fridge-out, oven-in, bake-complete")
	fmt.Println("\nExamples:")
	fmt.Println("  sourdough start")
	fmt.Println("  sourdough log mixed")
	fmt.Println("  sourdough log fold")
	fmt.Println("  sourdough temp 76")
	fmt.Println("  sourdough status")
	fmt.Println("  sourdough complete")
	fmt.Println("  sourdough history 5")
	fmt.Println("  sourdough review 2025-10-07")
}

func handleStart() {
	resp, err := http.Post(serverURL+"/bake/start", "application/json", nil)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		fmt.Printf("Make sure the server is running on %s\n", serverURL)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	fmt.Println("✓ Bake started!")
	fmt.Printf("Date: %s\n", time.Now().Format("2006-01-02"))
}

func handleLog() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Event type required")
		fmt.Println("Usage: sourdough log <event>")
		os.Exit(1)
	}

	event := os.Args[2]

	// Build URL
	url := fmt.Sprintf("%s/log/%s", serverURL, event)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("✓ Logged: %s\n", event)
	fmt.Printf("Time: %s\n", time.Now().Format("15:04"))
}

func handleTemp() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Temperature value required")
		fmt.Println("Usage: sourdough temp <value>")
		os.Exit(1)
	}

	temp := os.Args[2]

	// Validate temperature
	if _, err := strconv.ParseFloat(temp, 64); err != nil {
		fmt.Printf("Error: Invalid temperature value: %s\n", temp)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/log/temp/%s", serverURL, temp)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	fmt.Printf("✓ Temperature logged: %s°F\n", temp)
	fmt.Printf("Time: %s\n", time.Now().Format("15:04"))
}

func handleStatus() {
	resp, err := http.Get(serverURL + "/status")
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var bake models.Bake
	if err := json.NewDecoder(resp.Body).Decode(&bake); err != nil {
		fmt.Printf("Error: Failed to decode response: %v\n", err)
		os.Exit(1)
	}

	if len(bake.Events) == 0 {
		fmt.Println("No bake in progress today.")
		fmt.Println("Run 'sourdough start' to begin a new bake.")
		return
	}

	fmt.Printf("Bake Status - %s\n", bake.Date)
	fmt.Println(strings.Repeat("=", 50))

	for i, event := range bake.Events {
		timestamp := event.Timestamp.Format("15:04")

		// Calculate duration from previous event
		duration := ""
		if i > 0 {
			elapsed := event.Timestamp.Sub(bake.Events[i-1].Timestamp)
			duration = fmt.Sprintf(" (+%s)", formatDuration(elapsed))
		}

		// Format event info
		info := ""
		if event.TempF != nil {
			info = fmt.Sprintf(" [%.1f°F]", *event.TempF)
		}
		if event.DoughTempF != nil {
			info += fmt.Sprintf(" [dough: %.1f°F]", *event.DoughTempF)
		}
		if event.FoldCount != nil {
			info += fmt.Sprintf(" #%d", *event.FoldCount)
		}
		if event.Note != "" {
			info += fmt.Sprintf(" - %s", event.Note)
		}

		fmt.Printf("%s  %-15s%s%s\n", timestamp, event.Event, info, duration)
	}

	// Show elapsed time from start
	if len(bake.Events) > 0 {
		totalElapsed := time.Since(bake.Events[0].Timestamp)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("Total elapsed: %s\n", formatDuration(totalElapsed))
	}
}

func handleComplete() {
	// First, get current bake to ensure there is one
	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	hasBake, err := store.HasCurrentBake()
	if err != nil || !hasBake {
		fmt.Println("No bake in progress today.")
		os.Exit(1)
	}

	fmt.Println("Complete Bake Assessment")
	fmt.Println(strings.Repeat("=", 50))

	reader := bufio.NewReader(os.Stdin)

	// Proof level
	fmt.Println("\nProof level:")
	fmt.Println("1. Underproofed")
	fmt.Println("2. Good")
	fmt.Println("3. Overproofed")
	fmt.Print("Choice (1-3): ")
	proofChoice, _ := reader.ReadString('\n')
	proofChoice = strings.TrimSpace(proofChoice)

	var proofLevel models.ProofLevel
	switch proofChoice {
	case "1":
		proofLevel = models.ProofUnder
	case "2":
		proofLevel = models.ProofGood
	case "3":
		proofLevel = models.ProofOver
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	// Crumb quality
	fmt.Print("\nCrumb quality (1-10): ")
	crumbStr, _ := reader.ReadString('\n')
	crumb, err := strconv.Atoi(strings.TrimSpace(crumbStr))
	if err != nil || crumb < 1 || crumb > 10 {
		fmt.Println("Invalid crumb quality")
		os.Exit(1)
	}

	// Browning
	fmt.Println("\nBrowning:")
	fmt.Println("1. None")
	fmt.Println("2. Slight")
	fmt.Println("3. Good")
	fmt.Println("4. Over")
	fmt.Print("Choice (1-4): ")
	browningChoice, _ := reader.ReadString('\n')
	browningChoice = strings.TrimSpace(browningChoice)

	var browning models.BrowningLevel
	switch browningChoice {
	case "1":
		browning = models.BrowningNone
	case "2":
		browning = models.BrowningSlight
	case "3":
		browning = models.BrowningGood
	case "4":
		browning = models.BrowningOver
	default:
		fmt.Println("Invalid choice")
		os.Exit(1)
	}

	// Overall score
	fmt.Print("\nOverall score (1-10): ")
	scoreStr, _ := reader.ReadString('\n')
	score, err := strconv.Atoi(strings.TrimSpace(scoreStr))
	if err != nil || score < 1 || score > 10 {
		fmt.Println("Invalid score")
		os.Exit(1)
	}

	// Notes
	fmt.Print("\nNotes (optional): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	// Create assessment
	assessment := models.Assessment{
		ProofLevel:   proofLevel,
		CrumbQuality: crumb,
		Browning:     browning,
		Score:        score,
		Notes:        notes,
	}

	// Create bake-complete event with assessment
	event := models.NewEvent(models.EventBakeComplete)
	event.Data = map[string]interface{}{
		"assessment": assessment,
	}

	if err := store.AppendEvent(event); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Bake completed and assessed!")
	fmt.Printf("Proof: %s | Crumb: %d/10 | Browning: %s | Score: %d/10\n",
		proofLevel, crumb, browning, score)
}

func handleHistory() {
	limit := 10
	if len(os.Args) >= 3 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			limit = n
		}
	}

	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	dates, err := store.ListBakes()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(dates) == 0 {
		fmt.Println("No bakes found.")
		return
	}

	fmt.Println("Recent Bakes")
	fmt.Println(strings.Repeat("=", 70))

	for i, date := range dates {
		if i >= limit {
			break
		}

		bake, err := store.ReadBake(date)
		if err != nil {
			continue
		}

		// Format output
		status := "In progress"
		if bake.Assessment != nil {
			status = fmt.Sprintf("Score: %d/10 | Proof: %s | Crumb: %d/10",
				bake.Assessment.Score,
				bake.Assessment.ProofLevel,
				bake.Assessment.CrumbQuality)
		}

		eventCount := len(bake.Events)
		fmt.Printf("%s  %d events  %s\n", date, eventCount, status)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Showing %d of %d bakes\n", min(limit, len(dates)), len(dates))
}

func handleReview() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Date required")
		fmt.Println("Usage: sourdough review <date>")
		fmt.Println("Example: sourdough review 2025-10-07")
		os.Exit(1)
	}

	date := os.Args[2]

	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	bake, err := store.ReadBake(date)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(bake.Events) == 0 {
		fmt.Printf("No bake found for %s\n", date)
		return
	}

	fmt.Printf("Bake Review - %s\n", date)
	fmt.Println(strings.Repeat("=", 70))

	for i, event := range bake.Events {
		timestamp := event.Timestamp.Format("15:04")

		duration := ""
		if i > 0 {
			elapsed := event.Timestamp.Sub(bake.Events[i-1].Timestamp)
			duration = fmt.Sprintf(" (+%s)", formatDuration(elapsed))
		}

		info := ""
		if event.TempF != nil {
			info = fmt.Sprintf(" [%.1f°F]", *event.TempF)
		}
		if event.DoughTempF != nil {
			info += fmt.Sprintf(" [dough: %.1f°F]", *event.DoughTempF)
		}
		if event.FoldCount != nil {
			info += fmt.Sprintf(" #%d", *event.FoldCount)
		}
		if event.Note != "" {
			info += fmt.Sprintf(" - %s", event.Note)
		}

		fmt.Printf("%s  %-15s%s%s\n", timestamp, event.Event, info, duration)
	}

	if bake.Assessment != nil {
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println("Assessment:")
		fmt.Printf("  Proof Level:   %s\n", bake.Assessment.ProofLevel)
		fmt.Printf("  Crumb Quality: %d/10\n", bake.Assessment.CrumbQuality)
		fmt.Printf("  Browning:      %s\n", bake.Assessment.Browning)
		fmt.Printf("  Overall Score: %d/10\n", bake.Assessment.Score)
		if bake.Assessment.Notes != "" {
			fmt.Printf("  Notes:         %s\n", bake.Assessment.Notes)
		}
	}

	if len(bake.Events) > 0 {
		totalElapsed := bake.Events[len(bake.Events)-1].Timestamp.Sub(bake.Events[0].Timestamp)
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("Total time: %s\n", formatDuration(totalElapsed))
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// API response helper
func callAPI(method, path string, body interface{}) (*http.Response, error) {
	url := serverURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}
