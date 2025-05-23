package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplateFile_Success(t *testing.T) {
	// Create a temporary directory for the test file
	tempDir, err := os.MkdirTemp("", "test_loadtemplate")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a temporary test file with some content
	filePath := filepath.Join(tempDir, "test_template.txt")
	expectedContent := []byte("Hello, World!")
	err = os.WriteFile(filePath, expectedContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Call LoadTemplateFile with its path
	actualContent, err := LoadTemplateFile(filePath)
	if err != nil {
		t.Fatalf("LoadTemplateFile returned an unexpected error: %v", err)
	}

	// Verify the returned content matches what was written
	if string(actualContent) != string(expectedContent) {
		t.Errorf("LoadTemplateFile returned wrong content. Got %s, want %s", string(actualContent), string(expectedContent))
	}
}

func TestLoadTemplateFile_NotFound(t *testing.T) {
	// Call LoadTemplateFile with a non-existent file path
	_, err := LoadTemplateFile("non_existent_file.txt")

	// Verify an error is returned
	if err == nil {
		t.Error("LoadTemplateFile did not return an error for a non-existent file")
	}
}
