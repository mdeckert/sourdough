package models

import "time"

// EventType represents the type of baking event
type EventType string

const (
	EventStarterOut    EventType = "starter-out"
	EventFed           EventType = "fed"
	EventLevainReady   EventType = "levain-ready"
	EventMixed         EventType = "mixed"
	EventKnead         EventType = "knead"
	EventFold          EventType = "fold"
	EventShaped        EventType = "shaped"
	EventFridgeIn      EventType = "fridge-in"
	EventFridgeOut     EventType = "fridge-out"
	EventOvenIn        EventType = "oven-in"
	EventRemoveLid     EventType = "remove-lid"
	EventOvenOut       EventType = "oven-out"
	EventLoafComplete  EventType = "loaf-complete"
	EventTemperature   EventType = "temperature"
	EventNote          EventType = "note"
)

// Event represents a single baking event
type Event struct {
	Timestamp   time.Time              `json:"timestamp"`
	Event       EventType              `json:"event"`
	TempF       *float64               `json:"temp_f,omitempty"`
	DoughTempF  *float64               `json:"dough_temp_f,omitempty"`
	OvenTempF   *float64               `json:"oven_temp_f,omitempty"`
	FoldCount   *int                   `json:"fold_count,omitempty"`
	Note        string                 `json:"note,omitempty"`
	Image       string                 `json:"image,omitempty"`       // Image filename (stored in data/images/BAKE_DATE/)
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ProofLevel represents how well the dough was proofed
type ProofLevel string

const (
	ProofUnder ProofLevel = "underproofed"
	ProofGood  ProofLevel = "good"
	ProofOver  ProofLevel = "overproofed"
)

// BrowningLevel represents the browning of the crust
type BrowningLevel string

const (
	BrowningNone   BrowningLevel = "none"
	BrowningSlight BrowningLevel = "slight"
	BrowningGood   BrowningLevel = "good"
	BrowningOver   BrowningLevel = "over"
)

// Assessment represents the post-bake evaluation
type Assessment struct {
	ProofLevel    ProofLevel    `json:"proof_level"`
	CrumbQuality  int           `json:"crumb_quality"` // 1-10 scale
	Browning      BrowningLevel `json:"browning"`
	Score         int           `json:"score"` // 1-10 overall
	Notes         string        `json:"notes,omitempty"`
}

// Bake represents a complete baking session
type Bake struct {
	Date       string       `json:"date"`          // Date portion only (e.g., "2025-10-07")
	Filename   string       `json:"filename"`      // Full filename without extension (e.g., "bake_2025-10-07_19-13-49")
	Events     []Event      `json:"events"`
	Assessment *Assessment  `json:"assessment,omitempty"`
}

// NewEvent creates a new event with the current timestamp
func NewEvent(eventType EventType) *Event {
	return &Event{
		Timestamp: time.Now(),
		Event:     eventType,
	}
}

// WithTemp adds kitchen temperature to an event
func (e *Event) WithTemp(temp float64) *Event {
	e.TempF = &temp
	return e
}

// WithDoughTemp adds dough temperature to an event
func (e *Event) WithDoughTemp(temp float64) *Event {
	e.DoughTempF = &temp
	return e
}

// WithOvenTemp adds oven temperature to an event
func (e *Event) WithOvenTemp(temp float64) *Event {
	e.OvenTempF = &temp
	return e
}

// WithFoldCount adds fold count to an event
func (e *Event) WithFoldCount(count int) *Event {
	e.FoldCount = &count
	return e
}

// WithNote adds a note to an event
func (e *Event) WithNote(note string) *Event {
	e.Note = note
	return e
}

// WithImage adds an image filename to an event
func (e *Event) WithImage(imageFilename string) *Event {
	e.Image = imageFilename
	return e
}
