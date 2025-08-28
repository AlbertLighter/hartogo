package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlbertLighter/hartogo/internal/converter"
)

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Path to the input HAR file")
	outputDir := flag.String("output-dir", "", "Directory to save generated files (default: <input_filename>_req)")

	flag.Parse()

	if *inputFile == "" {
		log.Fatal("Input file path is required. Use the -input flag.")
	}

	// Determine the output directory
	finalOutputDir := *outputDir
	if finalOutputDir == "" {
		base := filepath.Base(*inputFile)
		ext := filepath.Ext(base)
		finalOutputDir = strings.TrimSuffix(base, ext) + "_req"
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Read and parse the HAR file
	httpArchive, err := converter.ReadHARFromFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading HAR file: %v", err)
	}

	if len(httpArchive.Log.Entries) == 0 {
		log.Fatal("No entries found in HAR file")
	}

	log.Printf("Found %d entries in %s. Starting conversion...", len(httpArchive.Log.Entries), *inputFile)

	// Process each entry
	for i, entry := range httpArchive.Log.Entries {
		// Generate filename
		parsedURL, err := url.Parse(entry.Request.URL)
		if err != nil {
			log.Printf("Skipping entry %d: Could not parse URL '%s': %v", i, entry.Request.URL, err)
			continue
		}
		path := strings.ReplaceAll(strings.Trim(parsedURL.Path, "/"), "/", "_")
		if path == "" {
			path = "root"
		}
		baseFilename := fmt.Sprintf("%s_%s_%s.go", entry.Request.Method, parsedURL.Host, path)

		// Generate function name
		funcName := converter.ToCamelCase(strings.TrimSuffix(baseFilename, ".go"))

		// Prepare data for the template
		data := converter.TemplateData{
			Request:      entry.Request,
			Response:     entry.Response,
			FunctionName: funcName,
		}

		// Generate the Go code
		generatedCode, err := converter.GenerateCode(data)
		if err != nil {
			log.Printf("Skipping entry %d: Could not generate code: %v", i, err)
			continue
		}

		// Write the generated code to the output file
		outputPath := filepath.Join(finalOutputDir, baseFilename)
		err = os.WriteFile(outputPath, []byte(generatedCode), 0644)
		if err != nil {
			log.Printf("Failed to write file %s: %v", outputPath, err)
			continue
		}

		log.Printf("Successfully created %s", outputPath)
	}

	log.Println("Conversion complete.")
}
