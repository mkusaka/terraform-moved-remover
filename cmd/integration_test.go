package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration performs integration testing with various Terraform file structures
func TestIntegration(t *testing.T) {
	// Create a temporary test directory
	testDir, err := os.MkdirTemp("", "terraform-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create nested directory structure
	nestedDir := filepath.Join(testDir, "modules", "networking")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create test files with various moved block patterns
	testFiles := map[string]string{
		filepath.Join(testDir, "main.tf"): `
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "web_server" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  
  tags = {
    Name = "WebServer"
  }
}

moved {
  from = aws_instance.web
  to   = aws_instance.web_server
}

resource "aws_s3_bucket" "data" {
  bucket = "my-data-bucket"
}

moved {
  from = aws_s3_bucket.logs
  to   = aws_s3_bucket.data
}`,

		filepath.Join(testDir, "modules", "main.tf"): `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module "vpc" {
  source = "./networking"
}

moved {
  from = module.network
  to   = module.vpc
}`,

		filepath.Join(nestedDir, "vpc.tf"): `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "MainVPC"
  }
}

moved {
  from = aws_vpc.primary
  to   = aws_vpc.main
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

moved {
  from = aws_subnet.external
  to   = aws_subnet.public
}`,
	}

	// Write test files
	for path, content := range testFiles {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", path, err)
		}
	}

	// Run the script on the test directory
	stats := Stats{}
	
	// Find all Terraform files
	files, err := findTerraformFiles(testDir)
	if err != nil {
		t.Fatalf("findTerraformFiles failed: %v", err)
	}
	
	// Process each file
	for _, file := range files {
		err := processFile(file, &stats)
		if err != nil {
			t.Fatalf("processFile failed for %s: %v", file, err)
		}
	}

	// Verify statistics
	expectedStats := Stats{
		FilesProcessed:     3,
		FilesModified:      3,
		MovedBlocksRemoved: 5,
	}
	
	if stats.FilesProcessed != expectedStats.FilesProcessed {
		t.Errorf("Expected FilesProcessed to be %d, but got %d", expectedStats.FilesProcessed, stats.FilesProcessed)
	}
	if stats.FilesModified != expectedStats.FilesModified {
		t.Errorf("Expected FilesModified to be %d, but got %d", expectedStats.FilesModified, stats.FilesModified)
	}
	if stats.MovedBlocksRemoved != expectedStats.MovedBlocksRemoved {
		t.Errorf("Expected MovedBlocksRemoved to be %d, but got %d", expectedStats.MovedBlocksRemoved, stats.MovedBlocksRemoved)
	}

	// Verify file contents - check that moved blocks are removed
	for path := range testFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read modified file %s: %v", path, err)
		}
		
		// 実際の出力内容を確認するためにログ出力
		t.Logf("File %s content after processing: %s", path, string(content))
		
		// ファイルが空でないことを確認
		if string(content) == "" {
			t.Errorf("File %s should not be empty after processing", path)
		}
		
		// ファイルが変更されていることを確認
		if string(content) == testFiles[path] {
			t.Errorf("File %s was not modified", path)
		}
		
		// movedブロックが含まれていないことを確認
		if strings.Contains(string(content), "moved {") {
			t.Errorf("File %s still contains moved blocks after processing", path)
		}
	}
}

// TestEdgeCases tests edge cases for the script
func TestEdgeCases(t *testing.T) {
	// Create a temporary test directory
	testDir, err := os.MkdirTemp("", "terraform-edge-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Test case 1: Empty file
	emptyFile := filepath.Join(testDir, "empty.tf")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	// Test case 2: File with only moved blocks
	onlyMovedFile := filepath.Join(testDir, "only_moved.tf")
	onlyMovedContent := `
moved {
  from = aws_instance.old
  to   = aws_instance.new
}

moved {
  from = aws_s3_bucket.old
  to   = aws_s3_bucket.new
}
`
	if err := os.WriteFile(onlyMovedFile, []byte(onlyMovedContent), 0644); err != nil {
		t.Fatalf("Failed to write only moved file: %v", err)
	}

	// Test case 3: File with commented moved blocks
	commentedFile := filepath.Join(testDir, "commented.tf")
	commentedContent := `
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}

# This is a commented moved block
# moved {
#   from = aws_instance.old
#   to   = aws_instance.web
# }
`
	if err := os.WriteFile(commentedFile, []byte(commentedContent), 0644); err != nil {
		t.Fatalf("Failed to write commented file: %v", err)
	}

	// Process the files
	stats := Stats{}
	
	files, err := findTerraformFiles(testDir)
	if err != nil {
		t.Fatalf("findTerraformFiles failed: %v", err)
	}
	
	for _, file := range files {
		err := processFile(file, &stats)
		if err != nil {
			t.Fatalf("processFile failed for %s: %v", file, err)
		}
	}

	// Verify statistics
	if stats.FilesProcessed != 3 {
		t.Errorf("Expected FilesProcessed to be 3, but got %d", stats.FilesProcessed)
	}
	if stats.FilesModified != 1 {
		t.Errorf("Expected FilesModified to be 1, but got %d", stats.FilesModified)
	}
	if stats.MovedBlocksRemoved != 2 {
		t.Errorf("Expected MovedBlocksRemoved to be 2, but got %d", stats.MovedBlocksRemoved)
	}

	// Verify the content of the only_moved file - it should be empty after processing
	content, err := os.ReadFile(onlyMovedFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}
	
	// The file should be empty or contain only whitespace
	if len(content) > 0 && len(string(content)) > 0 {
		contentStr := string(content)
		for _, c := range contentStr {
			if c != ' ' && c != '\n' && c != '\t' && c != '\r' {
				t.Errorf("Expected only_moved.tf to be empty or whitespace, but got: %s", contentStr)
				break
			}
		}
	}
}
