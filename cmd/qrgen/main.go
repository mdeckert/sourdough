package main

import (
	"fmt"
	"os"

	"github.com/mdeckert/sourdough/internal/qr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: qrgen <server-url>")
		fmt.Println("Example: qrgen http://192.168.1.100:8080")
		os.Exit(1)
	}

	serverURL := os.Args[1]
	outputDir := "./qrcodes"

	fmt.Printf("Generating QR codes for server: %s\n", serverURL)
	fmt.Printf("Output directory: %s\n\n", outputDir)

	if err := qr.GenerateAll(serverURL, outputDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ QR code generation complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Print qrcodes/sheet.png")
	fmt.Println("2. Cut out QR codes and label them")
	fmt.Println("3. Stick on fridge for easy access")
	fmt.Println("\nQR codes to include:")
	fmt.Println("  - Event codes: starter-out, fed, mixed, fold, shaped, etc.")
	fmt.Println("  - Temperature codes: 70°F, 72°F, 74°F, 76°F, 78°F")
	fmt.Println("\nTip: Test each QR code with your phone to ensure it works!")
}
