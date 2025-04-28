package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

const Version = "0.1.0"

// Stats tracks statistics about the processing
type Stats struct {
	FilesProcessed     int
	FilesModified      int
	MovedBlocksRemoved int
	StartTime          time.Time
	EndTime            time.Time
	DryRun             bool
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
	
	// Apply formatting to all files, not just those with moved blocks
	// Write modified content back to file only if not in dry run mode
	if !stats.DryRun {
		// Format the file content
		formattedContent := hclwrite.Format(file.Bytes())
		
		if fileModified || !bytes.Equal(formattedContent, content) {
			stats.FilesModified++
			
			if fileModified {
				stats.MovedBlocksRemoved += movedBlocksCount
			}
			
			err = os.WriteFile(filePath, formattedContent, 0644)
			if err != nil {
				return fmt.Errorf("error writing file %s: %w", filePath, err)
			}
		}
	} else if fileModified {
		// In dry run mode, just update stats for moved blocks
		stats.FilesModified++
		stats.MovedBlocksRemoved += movedBlocksCount
	}

	return nil
}

// printUsage prints the usage information for the script
func printUsage() {
	fmt.Println("Terraform Moved Directive Remover")
	fmt.Println("--------------------------------")
	fmt.Println("This tool recursively scans Terraform files, removes all 'moved' blocks,")
	fmt.Println("and applies standard Terraform formatting to the files.")
	fmt.Println()
	fmt.Println("Usage: terraform-moved-remover [options] <directory>")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
}

func main() {
	helpFlag := flag.Bool("help", false, "Display help information")
	versionFlag := flag.Bool("version", false, "Display version information")
	dryRunFlag := flag.Bool("dry-run", false, "Run without modifying files")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Usage = printUsage
	
	flag.Parse()
	
	if *helpFlag {
		printUsage()
		os.Exit(0)
	}
	
	if *versionFlag {
		fmt.Printf("Terraform Moved Directive Remover v%s\n", Version)
		os.Exit(0)
	}
	
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: No directory specified.")
		printUsage()
		os.Exit(1)
	}
	
	rootDir := args[0]
	
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
		DryRun:    *dryRunFlag,
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
		if *verboseFlag {
			fmt.Printf("Processing: %s\n", file)
		}
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
	if stats.DryRun {
		fmt.Println("DRY RUN MODE: No files were modified")
	}
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files modified: %d\n", stats.FilesModified)
	fmt.Printf("Moved blocks removed: %d\n", stats.MovedBlocksRemoved)
	fmt.Printf("Processing time: %v\n", duration)
}
