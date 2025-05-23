package main

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadPromptsConfig(t *testing.T) {
	yamlContent := `
prompts:
  - name: "projectName"
    message: "Enter the project name:"
    type: "input"
    defaultValue: "MyNewProject"
  - name: "projectType"
    message: "Select the project type:"
    type: "select"
    options:
      - "Web Application"
      - "Mobile Application"
    dynamic:
      source: "directory"
      path: "./modules"
`
	var config PromptsConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Verify number of prompts
	if len(config.Prompts) != 2 {
		t.Fatalf("Expected 2 prompts, got %d", len(config.Prompts))
	}

	// Verify details of the first prompt
	prompt1 := config.Prompts[0]
	if prompt1.Name != "projectName" {
		t.Errorf("Expected prompt1 name 'projectName', got '%s'", prompt1.Name)
	}
	if prompt1.Type != "input" {
		t.Errorf("Expected prompt1 type 'input', got '%s'", prompt1.Type)
	}
	if prompt1.DefaultValue != "MyNewProject" {
		t.Errorf("Expected prompt1 defaultValue 'MyNewProject', got '%s'", prompt1.DefaultValue)
	}
	if prompt1.Dynamic != nil {
		t.Errorf("Expected prompt1 to have no dynamic config, but it did")
	}

	// Verify details of the second prompt
	prompt2 := config.Prompts[1]
	if prompt2.Name != "projectType" {
		t.Errorf("Expected prompt2 name 'projectType', got '%s'", prompt2.Name)
	}
	if prompt2.Type != "select" {
		t.Errorf("Expected prompt2 type 'select', got '%s'", prompt2.Type)
	}
	expectedOptions := []string{"Web Application", "Mobile Application"}
	if !reflect.DeepEqual(prompt2.Options, expectedOptions) {
		t.Errorf("Expected prompt2 options %v, got %v", expectedOptions, prompt2.Options)
	}
	if prompt2.Dynamic == nil {
		t.Fatalf("Expected prompt2 to have dynamic config, but it didn't")
	}
	if prompt2.Dynamic.Source != "directory" {
		t.Errorf("Expected prompt2 dynamic source 'directory', got '%s'", prompt2.Dynamic.Source)
	}
	if prompt2.Dynamic.Path != "./modules" {
		t.Errorf("Expected prompt2 dynamic path './modules', got '%s'", prompt2.Dynamic.Path)
	}
}
