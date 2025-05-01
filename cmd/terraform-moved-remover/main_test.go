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

// TestConsecutiveMovedBlocks tests formatting when multiple consecutive moved blocks are removed
func TestConsecutiveMovedBlocks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-consecutive-moved-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with consecutive moved blocks
	testFile := filepath.Join(tempDir, "consecutive_moved.tf")
	content := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

moved {
  from = aws_instance.old1
  to   = aws_instance.web
}

moved {
  from = aws_instance.old2
  to   = aws_instance.web
}

moved {
  from = aws_instance.old3
  to   = aws_instance.web
}

resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
}
`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process the file with whitespace normalization enabled (default)
	stats := Stats{
		StartTime:           time.Now(),
		NormalizeWhitespace: true,
	}
	err = processFile(testFile, &stats)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Read the modified file
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	t.Logf("Modified content: %s", string(modifiedContent))

	// Check that moved blocks are removed
	if strings.Contains(string(modifiedContent), "moved {") {
		t.Errorf("File still contains moved blocks after processing")
	}

	// Check that there are no excessive empty lines
	lines := strings.Split(string(modifiedContent), "\n")
	consecutiveEmptyLines := 0
	maxConsecutiveEmptyLines := 0
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveEmptyLines++
		} else {
			if consecutiveEmptyLines > maxConsecutiveEmptyLines {
				maxConsecutiveEmptyLines = consecutiveEmptyLines
			}
			consecutiveEmptyLines = 0
		}
	}
	
	if maxConsecutiveEmptyLines > 2 {
		t.Errorf("File contains %d consecutive empty lines, expected at most 1", maxConsecutiveEmptyLines-1)
	}
}

func TestWhitespaceNormalizationFlag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-whitespace-flag-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test file content with consecutive moved blocks
	content := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

moved {
  from = aws_instance.old1
  to   = aws_instance.web
}

moved {
  from = aws_instance.old2
  to   = aws_instance.web
}

moved {
  from = aws_instance.old3
  to   = aws_instance.web
}

resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
}
`

	// Test with normalization disabled
	testFileDisabled := filepath.Join(tempDir, "normalization_disabled.tf")
	err = os.WriteFile(testFileDisabled, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	statsDisabled := Stats{
		StartTime:           time.Now(),
		NormalizeWhitespace: false,
	}
	err = processFile(testFileDisabled, &statsDisabled)
	if err != nil {
		t.Fatalf("processFile failed with normalization disabled: %v", err)
	}

	// Read the modified file
	disabledContent, err := os.ReadFile(testFileDisabled)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	t.Logf("Content with normalization disabled: %s", string(disabledContent))

	// Test with normalization enabled
	testFileEnabled := filepath.Join(tempDir, "normalization_enabled.tf")
	err = os.WriteFile(testFileEnabled, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	statsEnabled := Stats{
		StartTime:           time.Now(),
		NormalizeWhitespace: true,
	}
	err = processFile(testFileEnabled, &statsEnabled)
	if err != nil {
		t.Fatalf("processFile failed with normalization enabled: %v", err)
	}

	// Read the modified file
	enabledContent, err := os.ReadFile(testFileEnabled)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	t.Logf("Content with normalization enabled: %s", string(enabledContent))

	disabledLines := strings.Split(string(disabledContent), "\n")
	disabledConsecutiveEmptyLines := 0
	disabledMaxConsecutiveEmptyLines := 0
	
	for _, line := range disabledLines {
		if strings.TrimSpace(line) == "" {
			disabledConsecutiveEmptyLines++
		} else {
			if disabledConsecutiveEmptyLines > disabledMaxConsecutiveEmptyLines {
				disabledMaxConsecutiveEmptyLines = disabledConsecutiveEmptyLines
			}
			disabledConsecutiveEmptyLines = 0
		}
	}
	
	enabledLines := strings.Split(string(enabledContent), "\n")
	enabledConsecutiveEmptyLines := 0
	enabledMaxConsecutiveEmptyLines := 0
	
	for _, line := range enabledLines {
		if strings.TrimSpace(line) == "" {
			enabledConsecutiveEmptyLines++
		} else {
			if enabledConsecutiveEmptyLines > enabledMaxConsecutiveEmptyLines {
				enabledMaxConsecutiveEmptyLines = enabledConsecutiveEmptyLines
			}
			enabledConsecutiveEmptyLines = 0
		}
	}
	
	// With normalization disabled, we expect more consecutive empty lines
	if disabledMaxConsecutiveEmptyLines <= enabledMaxConsecutiveEmptyLines {
		t.Errorf("Expected more consecutive empty lines with normalization disabled, but got %d (disabled) vs %d (enabled)",
			disabledMaxConsecutiveEmptyLines, enabledMaxConsecutiveEmptyLines)
	}
	
	if enabledMaxConsecutiveEmptyLines > 2 {
		t.Errorf("With normalization enabled, file contains %d consecutive empty lines, expected at most 1", 
			enabledMaxConsecutiveEmptyLines-1)
	}
}
