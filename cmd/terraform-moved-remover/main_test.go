package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFindTerraformFiles tests the findTerraformFiles function
func TestFindTerraformFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		filepath.Join(tempDir, "main.tf"),
		filepath.Join(tempDir, "variables.tf"),
		filepath.Join(tempDir, "nested", "module.tf"),
		filepath.Join(tempDir, "nested", "deep", "resource.tf"),
		filepath.Join(tempDir, "not-terraform.txt"),
	}

	// Create directories
	if err := os.MkdirAll(filepath.Join(tempDir, "nested", "deep"), 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create files
	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", file, err)
		}
	}

	// Test finding files
	files, err := findTerraformFiles(tempDir)
	if err != nil {
		t.Fatalf("findTerraformFiles failed: %v", err)
	}

	// We should find 4 .tf files
	if len(files) != 4 {
		t.Errorf("Expected to find 4 .tf files, but found %d", len(files))
	}

	// Test with non-existent directory
	_, err = findTerraformFiles("/non-existent-dir")
	if err == nil {
		t.Errorf("Expected error for non-existent directory, but got nil")
	}
}

// TestProcessFile tests the processFile function
func TestProcessFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with moved blocks
	testFile := filepath.Join(tempDir, "test.tf")
	content := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

moved {
  from = aws_instance.old
  to   = aws_instance.web
}

resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
}

moved {
  from = aws_s3_bucket.logs
  to   = aws_s3_bucket.data
}
`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process the file
	stats := Stats{
		StartTime: time.Now(),
	}
	err = processFile(testFile, &stats)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Check statistics
	if stats.FilesProcessed != 1 {
		t.Errorf("Expected FilesProcessed to be 1, but got %d", stats.FilesProcessed)
	}
	if stats.FilesModified != 1 {
		t.Errorf("Expected FilesModified to be 1, but got %d", stats.FilesModified)
	}
	if stats.MovedBlocksRemoved != 2 {
		t.Errorf("Expected MovedBlocksRemoved to be 2, but got %d", stats.MovedBlocksRemoved)
	}

	// Read the modified file
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	// Check that moved blocks are removed
	if string(modifiedContent) == content {
		t.Errorf("File content was not modified")
	}

	// Test with non-existent file
	err = processFile("/non-existent-file.tf", &stats)
	if err == nil {
		t.Errorf("Expected error for non-existent file, but got nil")
	}

	// Test with invalid HCL
	invalidFile := filepath.Join(tempDir, "invalid.tf")
	err = os.WriteFile(invalidFile, []byte("this is not valid HCL"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	err = processFile(invalidFile, &stats)
	if err == nil {
		t.Errorf("Expected error for invalid HCL, but got nil")
	}
	
	unformattedFile := filepath.Join(tempDir, "unformatted.tf")
	unformattedContent := `
resource "aws_instance" "web" {
ami = "ami-123456"
  instance_type   =     "t2.micro"
}
`
	err = os.WriteFile(unformattedFile, []byte(unformattedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write unformatted file: %v", err)
	}

	// Process the file (should format it)
	err = processFile(unformattedFile, &stats)
	if err != nil {
		t.Fatalf("processFile failed for formatting test: %v", err)
	}

	// Read the formatted file
	formattedContent, err := os.ReadFile(unformattedFile)
	if err != nil {
		t.Fatalf("Failed to read formatted file: %v", err)
	}

	// Check that the file was formatted (should have consistent indentation)
	if string(formattedContent) == unformattedContent {
		t.Errorf("File was not formatted")
	}

	formattedString := string(formattedContent)
	t.Logf("Formatted content: %s", formattedString)
	
	if !strings.Contains(formattedString, "  ami") {
		t.Errorf("Formatting did not properly indent attributes")
	}
}

// TestMainFunction tests the main function indirectly
func TestMainFunction(t *testing.T) {
	// Save original os.Args and os.Exit
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile := filepath.Join(tempDir, "main.tf")
	content := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

moved {
  from = aws_instance.old
  to   = aws_instance.web
}
`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with valid directory
	os.Args = []string{"cmd", "-dry-run=false", tempDir}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset flags for testing
	
	// We can't directly test main() because it calls os.Exit
	// Instead, we'll test the individual components that main calls
	stats := Stats{
		StartTime: time.Now(),
	}
	
	files, err := findTerraformFiles(tempDir)
	if err != nil {
		t.Fatalf("findTerraformFiles failed: %v", err)
	}
	
	if len(files) != 1 {
		t.Errorf("Expected to find 1 .tf file, but found %d", len(files))
	}
	
	err = processFile(files[0], &stats)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}
	
	if stats.MovedBlocksRemoved != 1 {
		t.Errorf("Expected MovedBlocksRemoved to be 1, but got %d", stats.MovedBlocksRemoved)
	}
}

func TestFlagHandling(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() { 
		os.Args = oldArgs 
		flag.CommandLine = oldFlagCommandLine
	}()
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-flag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file with moved blocks
	testFile := filepath.Join(tempDir, "test.tf")
	content := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

moved {
  from = aws_instance.old
  to   = aws_instance.web
}
`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	os.Args = []string{"cmd", "-dry-run", tempDir}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	
	// Instead of calling main(), create Stats and test processFile with DryRun=true
	stats := Stats{
		StartTime: time.Now(),
		DryRun:    true,
	}
	
	err = processFile(testFile, &stats)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}
	
	// Read the file after processing - it should remain unchanged due to dry run
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after dry run: %v", err)
	}
	
	if string(modifiedContent) != content {
		t.Errorf("Dry run mode modified the file, but it shouldn't have")
	}
}
