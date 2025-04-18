package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Stats tracks statistics about the processing
type Stats struct {
	FilesProcessed     int
	FilesModified      int
	MovedBlocksRemoved int
	StartTime          time.Time
	EndTime            time.Time
}

// findTerraformFiles recursively finds all .tf files in the given directory
func findTerraformFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// processFile processes a single Terraform file to remove moved blocks
// It returns true if the file was modified, false otherwise
func processFile(filePath string, stats *Stats) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Parse HCL file
	file, diags := hclwrite.ParseConfig(content, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("error parsing %s: %s", filePath, diags.Error())
	}

	// Track if file was modified
	fileModified := false
	movedBlocksCount := 0

	// Find and remove moved blocks
	body := file.Body()
	for _, block := range body.Blocks() {
		if block.Type() == "moved" {
			body.RemoveBlock(block)
			movedBlocksCount++
			fileModified = true
		}
	}

	// Update statistics
	stats.FilesProcessed++
	if fileModified {
		stats.FilesModified++
		stats.MovedBlocksRemoved += movedBlocksCount

		// Write modified content back to file
		err = os.WriteFile(filePath, file.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("error writing file %s: %w", filePath, err)
		}
	}

	return nil
}

// printUsage prints the usage information for the script
func printUsage() {
	fmt.Println("Terraform Moved Directive Remover")
	fmt.Println("--------------------------------")
	fmt.Println("This tool recursively scans Terraform files and removes all 'moved' blocks.")
	fmt.Println()
	fmt.Println("Usage: terraform-moved-remover <directory>")
	fmt.Println()
	fmt.Println("Example: terraform-moved-remover ./terraform")
	fmt.Println()
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	rootDir := os.Args[1]

	// Verify directory exists
	info, err := os.Stat(rootDir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", rootDir)
		os.Exit(1)
	}

	// Initialize statistics
	stats := Stats{
		StartTime: time.Now(),
	}

	// Find all Terraform files
	fmt.Printf("Scanning directory: %s\n", rootDir)
	files, err := findTerraformFiles(rootDir)
	if err != nil {
		fmt.Printf("Error finding Terraform files: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d Terraform files\n", len(files))

	// Process each file
	for _, file := range files {
		err := processFile(file, &stats)
		if err != nil {
			fmt.Printf("Error processing %s: %s\n", file, err)
		}
	}

	// Record end time
	stats.EndTime = time.Now()
	duration := stats.EndTime.Sub(stats.StartTime)

	// Print statistics
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files modified: %d\n", stats.FilesModified)
	fmt.Printf("Moved blocks removed: %d\n", stats.MovedBlocksRemoved)
	fmt.Printf("Processing time: %v\n", duration)
}
