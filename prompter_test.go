package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestLoadDynamicOptions_Success(t *testing.T) {
	// 1. Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test_dynamic_options")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// 2. Create a few files/subdirectories inside it
	expectedOptions := []string{"item1", "item2", "subDir"}
	for _, item := range expectedOptions {
		if item == "subDir" {
			err = os.Mkdir(filepath.Join(tempDir, item), 0755)
		} else {
			_, err = os.Create(filepath.Join(tempDir, item))
		}
		if err != nil {
			t.Fatalf("Failed to create test item '%s': %v", item, err)
		}
	}

	// 3. Define a Prompt struct
	prompt := &Prompt{
		Name: "testDynamicPrompt",
		Dynamic: &DynamicPromptConfig{
			Source: "directory",
			Path:   tempDir,
		},
	}

	// 4. Call loadDynamicOptions
	err = loadDynamicOptions(prompt)
	if err != nil {
		t.Fatalf("loadDynamicOptions returned an unexpected error: %v", err)
	}

	// 5. Assert that prompt.Options is populated correctly
	// Sort both slices because ReadDir does not guarantee order
	sort.Strings(prompt.Options)
	sort.Strings(expectedOptions)

	if !reflect.DeepEqual(prompt.Options, expectedOptions) {
		t.Errorf("loadDynamicOptions populated options incorrectly. Got %v, want %v", prompt.Options, expectedOptions)
	}
}

func TestLoadDynamicOptions_DirectoryNotFound(t *testing.T) {
	prompt := &Prompt{
		Name: "testNonExistentDir",
		Dynamic: &DynamicPromptConfig{
			Source: "directory",
			Path:   "./non_existent_directory_for_test",
		},
	}

	err := loadDynamicOptions(prompt)
	if err == nil {
		t.Error("loadDynamicOptions did not return an error for a non-existent directory")
	}
}

func TestLoadDynamicOptions_EmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_empty_dynamic_options")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	prompt := &Prompt{
		Name: "testEmptyDir",
		Dynamic: &DynamicPromptConfig{
			Source: "directory",
			Path:   tempDir,
		},
	}

	err = loadDynamicOptions(prompt)
	if err != nil {
		t.Fatalf("loadDynamicOptions returned an unexpected error for an empty directory: %v", err)
	}

	if len(prompt.Options) != 0 {
		t.Errorf("loadDynamicOptions should result in empty options for an empty directory, got %v", prompt.Options)
	}
	// The function currently prints a warning for empty directories, which is acceptable.
	// This test ensures it doesn't error and that options are indeed empty.
}

func TestLoadDynamicOptions_NoDynamicConfig(t *testing.T) {
	prompt := &Prompt{
		Name:    "testNoDynamicConfig",
		Dynamic: nil, // No dynamic configuration
	}
	originalOptions := []string{"static1"}
	prompt.Options = originalOptions

	err := loadDynamicOptions(prompt)
	if err != nil {
		t.Fatalf("loadDynamicOptions returned an unexpected error when no dynamic config is present: %v", err)
	}

	if !reflect.DeepEqual(prompt.Options, originalOptions) {
		t.Errorf("loadDynamicOptions modified options when no dynamic config was present. Got %v, want %v", prompt.Options, originalOptions)
	}
}

func TestLoadDynamicOptions_DynamicConfigNotDirectorySource(t *testing.T) {
	prompt := &Prompt{
		Name: "testWrongSource",
		Dynamic: &DynamicPromptConfig{
			Source: "other_source", // Not "directory"
			Path:   "./some_path",
		},
	}
	originalOptions := []string{"static2"}
	prompt.Options = originalOptions

	err := loadDynamicOptions(prompt)
	if err != nil {
		t.Fatalf("loadDynamicOptions returned an unexpected error for non-directory dynamic source: %v", err)
	}

	if !reflect.DeepEqual(prompt.Options, originalOptions) {
		t.Errorf("loadDynamicOptions modified options for non-directory dynamic source. Got %v, want %v", prompt.Options, originalOptions)
	}
}

func TestLoadDynamicOptions_DynamicPathNotSet(t *testing.T) {
	prompt := &Prompt{
		Name: "testDynamicPathNotSet",
		Dynamic: &DynamicPromptConfig{
			Source: "directory",
			Path:   "", // Path not set
		},
	}

	err := loadDynamicOptions(prompt)
	if err == nil {
		t.Error("loadDynamicOptions did not return an error when dynamic path is not set")
	}
}
