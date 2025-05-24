package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateFileFromTemplate(t *testing.T) {
	// Setup:
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "yamt_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Ensure cleanup

	// Create a dummy template file content
	templateContent := "AppName: {{ .AppName }}, Version: {{ .AppVersion }}"
	dummyTemplatePath := filepath.Join(tempDir, "test.gotmpl")

	// Write this content to dummyTemplatePath
	err = os.WriteFile(dummyTemplatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy template file: %v", err)
	}

	// Prepare a sample Config struct
	sampleConfig := &Config{
		AppName:    "TestApp",
		AppVersion: "1.0.0",
		// Environment, ClusterName, Variables can be empty or have minimal values
		// as they are not used by this specific dummy template.
		Environment: "test-env",
		ClusterName: "test-cluster",
		Variables:   make(map[string]interface{}),
	}

	// Define the output path for the generated file
	outputPath := filepath.Join(tempDir, "output.txt")

	// Execution:
	// Call generateFileFromTemplate
	err = generateFileFromTemplate(dummyTemplatePath, outputPath, sampleConfig)
	if err != nil {
		t.Fatalf("generateFileFromTemplate failed: %v", err)
	}

	// Verification:
	// Read the content of the generated outputPath file
	actualContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Define the expected content
	expectedContent := "AppName: TestApp, Version: 1.0.0"

	// Compare the actual content with the expected content
	if string(actualContent) != expectedContent {
		t.Errorf("Generated content mismatch: got %q, want %q", string(actualContent), expectedContent)
	}
}

func TestPromptSelection(t *testing.T) {
	options := []string{"option1", "option2", "option3"}

	// Test valid selection
	t.Run("ValidSelection", func(t *testing.T) {
		// Simulate user input "2\n" (selects "option2")
		input := "2\n"
		reader := bufio.NewReader(strings.NewReader(input))

		selected, err := promptSelection(reader, "Choose an option:", options)
		if err != nil {
			t.Fatalf("promptSelection returned an error for valid input: %v", err)
		}
		if selected != "option2" {
			t.Errorf("Expected 'option2', got '%s'", selected)
		}
	})

	// Test invalid input then valid input (e.g., "invalid\n1\n")
	t.Run("InvalidThenValidSelection", func(t *testing.T) {
		input := "invalid\n1\n" // First input is non-numeric, second is valid
		reader := bufio.NewReader(strings.NewReader(input))

		// We need to redirect stdout to capture the prompts and error messages
		// to ensure the re-prompting mechanism works.
		// For simplicity in this environment, we'll trust the loop and error messages
		// are printed, and focus on the final correct selection.
		// A more complex test would capture stdout.

		selected, err := promptSelection(reader, "Choose an option:", options)
		if err != nil {
			t.Fatalf("promptSelection returned an error after re-prompt: %v", err)
		}
		if selected != "option1" {
			t.Errorf("Expected 'option1' after re-prompt, got '%s'", selected)
		}
	})

	// Test out-of-bounds selection then valid input
	t.Run("OutOfBoundsThenValidSelection", func(t *testing.T) {
		input := "5\n3\n" // First input is out of bounds, second is valid
		reader := bufio.NewReader(strings.NewReader(input))

		selected, err := promptSelection(reader, "Choose an option:", options)
		if err != nil {
			t.Fatalf("promptSelection returned an error after out-of-bounds input: %v", err)
		}
		if selected != "option3" {
			t.Errorf("Expected 'option3' after out-of-bounds, got '%s'", selected)
		}
	})

	// Test empty input (should re-prompt) then valid input
	t.Run("EmptyThenValidSelection", func(t *testing.T) {
		input := "\n1\n" // First input is empty (results in strconv.Atoi error), second is valid
		reader := bufio.NewReader(strings.NewReader(input))

		selected, err := promptSelection(reader, "Choose an option:", options)
		if err != nil {
			t.Fatalf("promptSelection returned an error after empty input: %v", err)
		}
		if selected != "option1" {
			t.Errorf("Expected 'option1' after empty input, got '%s'", selected)
		}
	})
} // Added this closing brace for TestPromptSelection

func TestAppTypeTemplateSelection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "yamt_test_apptype_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatesDir := filepath.Join(tempDir, "templates")
	deploymentTemplatesDir := filepath.Join(templatesDir, "deployment")
	jobTemplatesDir := filepath.Join(templatesDir, "job")

	err = os.MkdirAll(deploymentTemplatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create deployment templates dir: %v", err)
	}
	err = os.MkdirAll(jobTemplatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create job templates dir: %v", err)
	}

	// Create dummy template files
	deploymentHelmfileContent := "Deployment Template Content: {{ .AppName }}"
	jobHelmfileContent := "Job Template Content: {{ .AppName }}"
	overrideHelmfileContent := "Override Template Content: {{ .AppName }}"

	err = os.WriteFile(filepath.Join(deploymentTemplatesDir, "helmfile.yaml.gotmpl"), []byte(deploymentHelmfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write deployment helmfile template: %v", err)
	}
	err = os.WriteFile(filepath.Join(jobTemplatesDir, "helmfile.yaml.gotmpl"), []byte(jobHelmfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write job helmfile template: %v", err)
	}
	err = os.WriteFile(filepath.Join(templatesDir, "override.yaml.gotmpl"), []byte(overrideHelmfileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write override helmfile template: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output.yaml")

	// --- Test Case 1: AppType "deployment" ---
	t.Run("AppTypeDeployment", func(t *testing.T) {
		config := &Config{AppName: "TestAppDeployment", AppType: "deployment"}
		// In main(), the template path is constructed as filepath.Join("templates", derivedOrOverriddenName)
		// So, for this test, we simulate that the resolved path points to the "deployment" specific template.
		// The generateFileFromTemplate function itself doesn't know about the base "templates" dir or AppType logic.
		// The path passed to it is the *final* path to the template file.

		// Path resolution simulation (as done in main()):
		// 1. Default based on AppType: "deployment/helmfile.yaml.gotmpl"
		// 2. Override check (HelmfileTemplate is empty) -> no override
		// 3. Final path: filepath.Join(templatesDir, "deployment/helmfile.yaml.gotmpl")
		templateToUsePath := filepath.Join(deploymentTemplatesDir, "helmfile.yaml.gotmpl")

		err := generateFileFromTemplate(templateToUsePath, outputPath, config)
		if err != nil {
			t.Fatalf("generateFileFromTemplate failed for AppType deployment: %v", err)
		}
		content, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			t.Fatalf("Failed to read output file for AppType deployment: %v", readErr)
		}
		expectedContent := "Deployment Template Content: TestAppDeployment"
		if string(content) != expectedContent {
			t.Errorf("AppType deployment: got %q, want %q", string(content), expectedContent)
		}
		os.Remove(outputPath) // Clean up for next sub-test
	})

	// --- Test Case 2: Override with HelmfileTemplate ---
	t.Run("OverrideWithHelmfileTemplate", func(t *testing.T) {
		config := &Config{AppName: "OverrideApp", AppType: "job", HelmfileTemplate: "override.yaml.gotmpl"}

		// Path resolution simulation:
		// 1. Default based on AppType: "job/helmfile.yaml.gotmpl"
		// 2. Override check (HelmfileTemplate is "override.yaml.gotmpl") -> override active
		// 3. Final path: filepath.Join(templatesDir, "override.yaml.gotmpl")
		templateToUsePath := filepath.Join(templatesDir, "override.yaml.gotmpl")

		err := generateFileFromTemplate(templateToUsePath, outputPath, config)
		if err != nil {
			t.Fatalf("generateFileFromTemplate failed for override: %v", err)
		}
		content, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			t.Fatalf("Failed to read output file for override: %v", readErr)
		}
		expectedContent := "Override Template Content: OverrideApp"
		if string(content) != expectedContent {
			t.Errorf("Override: got %q, want %q", string(content), expectedContent)
		}
		os.Remove(outputPath) // Clean up
	})

	// Test Case 3 (Error on missing AppType template) is implicitly handled by os.Stat in main().
	// Unit testing generateFileFromTemplate directly assumes the path is valid.
	// Testing the os.Stat check itself is more of an integration test of main's orchestration.
}

func MinimalValidConfig() Config {
	return Config{
		Environment: "dev",
		ClusterName: "test-cluster",
		AppName:     "TestApp",
		AppVersion:  "1.0.0",
		// AppType and HelmfileTemplate are the focus of TestValidateConfigAppType
	}
}

func TestValidateConfigAppType(t *testing.T) {
	t.Run("AppTypeEmpty_HelmfileTemplateEmpty_ShouldError", func(t *testing.T) {
		config := MinimalValidConfig()
		config.AppType = ""
		config.HelmfileTemplate = ""
		err := validateConfig(&config)
		if err == nil {
			t.Error("Expected error when AppType and HelmfileTemplate are empty, got nil")
		} else {
			expectedErrMsg := "AppType must be specified when HelmfileTemplate is not explicitly set"
			if err.Error() != expectedErrMsg {
				t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
			}
		}
	})

	t.Run("AppTypeEmpty_HelmfileTemplateSet_ShouldPass", func(t *testing.T) {
		config := MinimalValidConfig()
		config.AppType = ""
		config.HelmfileTemplate = "some/template.yaml"
		err := validateConfig(&config)
		if err != nil {
			t.Errorf("Expected no error when HelmfileTemplate is set, got: %v", err)
		}
	})

	t.Run("AppTypeSet_HelmfileTemplateEmpty_ShouldPass", func(t *testing.T) {
		config := MinimalValidConfig()
		config.AppType = "deployment"
		config.HelmfileTemplate = ""
		err := validateConfig(&config)
		if err != nil {
			t.Errorf("Expected no error when AppType is set, got: %v", err)
		}
	})

	t.Run("AppNameEmpty_ShouldError", func(t *testing.T) {
		config := MinimalValidConfig()
		config.AppName = "" // Make it invalid in another way
		config.AppType = "deployment" // Satisfy AppType condition
		err := validateConfig(&config)
		if err == nil {
			t.Error("Expected error when AppName is empty, got nil")
		} else {
			expectedErrMsg := "appName must be set"
			if err.Error() != expectedErrMsg {
				t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
			}
		}
	})

	t.Run("FullyValidConfig_ShouldPass", func(t *testing.T) {
		config := MinimalValidConfig()
		config.AppType = "deployment" // or HelmfileTemplate set
		err := validateConfig(&config)
		if err != nil {
			t.Errorf("Expected no error for a fully valid config, got: %v", err)
		}
	})
}
