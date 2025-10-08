package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

// EventQR represents a QR code for an event
type EventQR struct {
	Event string
	Label string
	URL   string
}

// GenerateAll generates QR codes for all common events
func GenerateAll(serverURL, outputDir string) error {
	// Validate server URL - reject localhost addresses
	if isLocalhostURL(serverURL) {
		return fmt.Errorf("server URL cannot be localhost/127.0.0.1 - QR codes must be accessible from mobile devices. Use your server's IP address (e.g., http://192.168.1.50:8080)")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	events := []EventQR{
		{"start", "START LOAF", fmt.Sprintf("%s/loaf/start", serverURL)},
		{"fed", "Fed", fmt.Sprintf("%s/log/fed", serverURL)},
		{"levain-ready", "Levain Ready", fmt.Sprintf("%s/log/levain-ready", serverURL)},
		{"mixed", "Mixed", fmt.Sprintf("%s/log/mixed", serverURL)},
		{"fold", "Fold", fmt.Sprintf("%s/log/fold", serverURL)},
		{"shaped", "Shaped", fmt.Sprintf("%s/log/shaped", serverURL)},
		{"fridge-in", "Fridge In", fmt.Sprintf("%s/log/fridge-in", serverURL)},
		{"oven-in", "Oven In", fmt.Sprintf("%s/log/oven-in", serverURL)},
		{"remove-lid", "Remove Lid", fmt.Sprintf("%s/log/remove-lid", serverURL)},
		{"oven-out", "Oven Out", fmt.Sprintf("%s/log/oven-out", serverURL)},
		{"temp", "LOG TEMP", fmt.Sprintf("%s/temp", serverURL)},
		{"notes", "ADD NOTE", fmt.Sprintf("%s/notes", serverURL)},
		{"complete", "COMPLETE", fmt.Sprintf("%s/complete", serverURL)},
		{"status", "VIEW STATUS", fmt.Sprintf("%s/view/status", serverURL)},
		{"history", "VIEW HISTORY", fmt.Sprintf("%s/view/history", serverURL)},
		{"qr-pdf", "GET QR CODES", fmt.Sprintf("%s/qrcodes.pdf", serverURL)},
	}

	// Generate individual QR codes
	for _, event := range events {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s.png", event.Event))
		if err := generateQRCode(event.URL, filename); err != nil {
			return fmt.Errorf("failed to generate QR code for %s: %w", event.Event, err)
		}
		fmt.Printf("Generated: %s\n", filename)
	}

	// Generate printable sheet (PNG)
	sheetPath := filepath.Join(outputDir, "sheet.png")
	if err := generateSheet(events, outputDir, sheetPath); err != nil {
		return fmt.Errorf("failed to generate sheet: %w", err)
	}
	fmt.Printf("Generated: %s\n", sheetPath)

	// Generate PDF
	pdfPath := filepath.Join(outputDir, "qrcodes.pdf")
	if err := generatePDF(events, outputDir, pdfPath); err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}
	fmt.Printf("Generated: %s\n", pdfPath)

	return nil
}

// generateQRCode generates a single QR code
func generateQRCode(url, filename string) error {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return err
	}

	qr.DisableBorder = false

	return qr.WriteFile(256, filename)
}

// generateSheet generates a printable sheet with all QR codes
func generateSheet(events []EventQR, qrDir, outputPath string) error {
	const (
		qrSize    = 200
		padding   = 20
		labelH    = 30
		cellW     = qrSize + padding*2
		cellH     = qrSize + labelH + padding*2
		cols      = 3
	)

	rows := (len(events) + cols - 1) / cols
	width := cols * cellW
	height := rows * cellH

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill white background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Add each QR code
	for i, event := range events {
		row := i / cols
		col := i % cols

		x := col * cellW
		y := row * cellH

		// Load QR code
		qrPath := filepath.Join(qrDir, fmt.Sprintf("%s.png", event.Event))
		qrFile, err := os.Open(qrPath)
		if err != nil {
			continue
		}

		qrImg, err := png.Decode(qrFile)
		qrFile.Close()
		if err != nil {
			continue
		}

		// Draw QR code
		qrRect := image.Rect(x+padding, y+padding, x+padding+qrSize, y+padding+qrSize)
		draw.Draw(img, qrRect, qrImg, image.Point{}, draw.Src)

		// Draw label (simple text rendering would require a font library,
		// so we'll skip it for now - users can add labels manually)
	}

	// Save sheet
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// generatePDF generates a printable PDF with all QR codes and labels
func generatePDF(events []EventQR, qrDir, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Add title and workflow synopsis
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Sourdough Bread Logger - QR Codes")
	pdf.Ln(10)

	// Workflow synopsis
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)

	workflow := []string{
		"WORKFLOW: START LOAF -> Fed -> Levain Ready -> Mixed -> Fold (3-4x) -> Shaped -> Fridge In -> Oven In -> Oven Out -> COMPLETE",
		"",
		"LOG TEMP: Scan anytime to log kitchen/dough temperature (critical for timing)",
		"ADD NOTE: Scan anytime to add observations (crumb, taste, process notes)",
		"COMPLETE: Final assessment (proof level, crumb, browning, score)",
		"VIEW STATUS: See current bake timeline with temperature graphs",
		"VIEW HISTORY: Browse all past bakes with scores and statistics",
		"GET QR CODES: Download this PDF to your phone",
	}

	for _, line := range workflow {
		if line == "" {
			pdf.Ln(3)
		} else {
			pdf.MultiCell(0, 4, line, "", "L", false)
		}
	}

	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)

	const (
		qrSize  = 35.0 // QR code size in mm (smaller to fit more)
		spacing = 5.0  // Spacing between codes
		cols    = 4    // 4 columns to fit more per page
	)

	x := 10.0
	y := pdf.GetY()
	col := 0

	for _, event := range events {
		// Check if we need a new page
		if y+qrSize+15 > 265 { // Letter height is ~279mm
			pdf.AddPage()
			y = 10.0
			x = 10.0
			col = 0
		}

		// Calculate position
		x = 10.0 + float64(col)*(qrSize+spacing)

		// Add QR code image
		qrPath := filepath.Join(qrDir, fmt.Sprintf("%s.png", event.Event))
		if _, err := os.Stat(qrPath); err == nil {
			pdf.Image(qrPath, x, y, qrSize, qrSize, false, "", 0, "")
		}

		// Add label below QR code
		pdf.SetFont("Arial", "B", 8)
		pdf.SetXY(x, y+qrSize+1)
		pdf.CellFormat(qrSize, 4, event.Label, "0", 0, "C", false, 0, "")

		// Move to next position
		col++
		if col >= cols {
			col = 0
			y += qrSize + spacing + 8
		}
	}

	return pdf.OutputFileAndClose(outputPath)
}

// isLocalhostURL checks if a URL contains localhost or 127.0.0.1
func isLocalhostURL(url string) bool {
	url = strings.ToLower(url)
	return strings.Contains(url, "localhost") ||
		strings.Contains(url, "127.0.0.1") ||
		strings.Contains(url, "0.0.0.0") ||
		strings.Contains(url, "[::1]")
}
