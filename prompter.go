package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
)

// RunPrompts executes a series of prompts based on the provided configurations.
// It handles dynamic option generation for prompts and uses the survey library
// to interact with the user.
//
// Args:
//   prompts: A slice of Prompt structs, each defining a question to ask the user.
//
// Returns:
//   A map where keys are prompt names and values are the user's answers.
//   An error if any issue occurs during the prompting process (e.g., directory
//   not found for dynamic prompts, user interruption).
func RunPrompts(prompts []Prompt) (map[string]interface{}, error) {
	answers := make(map[string]interface{})

	for i := range prompts {
		p := &prompts[i] // Use a pointer to modify the original prompt in the slice

		// Handle Dynamic Options
		if p.Dynamic != nil && p.Dynamic.Source == "directory" {
			err := loadDynamicOptions(p)
			if err != nil {
				// Decide if this error should halt all prompts or just skip this one.
				// For now, let's return the error, halting further prompts.
				// A warning could also be logged and the prompt skipped.
				return nil, fmt.Errorf("error loading dynamic options for prompt '%s': %w", p.Name, err)
			}
		}

		// Use survey library for prompting
		var answer interface{} // Use interface{} to hold different answer types

		switch p.Type {
		case "input":
			qs := &survey.Input{
				Message: p.Message,
				Default: p.DefaultValue,
			}
			var inputAnswer string
			err := survey.AskOne(qs, &inputAnswer)
			if err != nil {
				return nil, fmt.Errorf("prompt failed for '%s': %w", p.Name, err)
			}
			answer = inputAnswer
		case "select":
			if len(p.Options) == 0 { // Check options after dynamic loading attempt
				// This implies dynamic loading failed to find options or static options were not provided.
				fmt.Printf("Warning: Prompt '%s' of type 'select' has no options. Skipping this prompt.\n", p.Name)
				// Optionally, store a specific value or nil, or return an error
				// For now, we'll skip and not add it to answers.
				continue
			}
			qs := &survey.Select{
				Message: p.Message,
				Options: p.Options,
				Default: p.DefaultValue,
			}
			var selectAnswer string
			err := survey.AskOne(qs, &selectAnswer)
			if err != nil {
				return nil, fmt.Errorf("prompt failed for '%s': %w", p.Name, err)
			}
			answer = selectAnswer
		case "multiselect":
			if len(p.Options) == 0 { // Check options after dynamic loading attempt
				fmt.Printf("Warning: Prompt '%s' of type 'multiselect' has no options. Skipping this prompt.\n", p.Name)
				continue
			}
			qs := &survey.MultiSelect{
				Message: p.Message,
				Options: p.Options,
				Default: p.DefaultValue,
			}
			var multiSelectAnswer []string
			err := survey.AskOne(qs, &multiSelectAnswer)
			if err != nil {
				return nil, fmt.Errorf("prompt failed for '%s': %w", p.Name, err)
			}
			answer = multiSelectAnswer
		default:
			return nil, fmt.Errorf("unsupported prompt type '%s' for prompt '%s'", p.Type, p.Name)
		}
		answers[p.Name] = answer
	}

	return answers, nil
}

// loadDynamicOptions populates the Options field of a Prompt struct
// if it's configured for dynamic options from a directory.
func loadDynamicOptions(p *Prompt) error {
	if p.Dynamic == nil || p.Dynamic.Source != "directory" {
		// Not a dynamic directory prompt, or misconfigured.
		// This function should only be called if p.Dynamic.Source == "directory".
		return nil // Or an error if called inappropriately
	}

	if p.Dynamic.Path == "" {
		return fmt.Errorf("dynamic prompt '%s' has source 'directory' but no path is specified", p.Name)
	}

	files, err := os.ReadDir(p.Dynamic.Path)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s' for dynamic prompt '%s': %w", p.Dynamic.Path, p.Name, err)
	}

	p.Options = []string{} // Clear existing options
	for _, file := range files {
		p.Options = append(p.Options, file.Name())
	}

	if len(p.Options) == 0 {
		// It's not necessarily an error for a directory to be empty.
		// The calling function (RunPrompts) can decide how to handle a prompt with no options.
		// We can print a warning here as it was done before.
		fmt.Printf("Warning: No items found in directory '%s' for dynamic prompt '%s'. Prompt will have no options.\n", p.Dynamic.Path, p.Name)
	}
	return nil
}

// Helper function to check if a path is a directory (not directly used in RunPrompts above but good for reference)
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

// Example usage (can be in main.go or a test file)
/*
func main() {
	// Create a dummy directory for dynamic prompt testing
	os.MkdirAll("./test_modules/moduleA", 0755)
	os.MkdirAll("./test_modules/moduleB", 0755)
	defer os.RemoveAll("./test_modules")

	prompts := []Prompt{
		{
			Name:    "projectName",
			Message: "Enter the project name:",
			Type:    "input",
			DefaultValue: "MyAwesomeProject",
		},
		{
			Name:    "module",
			Message: "Select a module:",
			Type:    "select",
			Dynamic: &DynamicPromptConfig{
				Source: "directory",
				Path:   "./test_modules",
			},
		},
		{
			Name:    "features",
			Message: "Select features:",
			Type:    "multiselect",
			Options: []string{"FeatureA", "FeatureB", "FeatureC"},
		},
	}

	answers, err := RunPrompts(prompts)
	if err != nil {
		fmt.Printf("Error running prompts: %v\n", err)
		return
	}

	fmt.Printf("Answers: %+v\n", answers)

	// Example of accessing a specific answer
	// if projectName, ok := answers["projectName"].(string); ok {
	// 	fmt.Printf("Project Name: %s\n", projectName)
	// }
}
*/
