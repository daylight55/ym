package main

import (
	"strings"
	"testing"
)

func TestRenderTemplate_Simple(t *testing.T) {
	templateContent := "name: {{ .name }}"
	answers := map[string]interface{}{"name": "test"}
	expectedOutput := "name: test"

	output, err := RenderTemplate(templateContent, answers)
	if err != nil {
		t.Fatalf("RenderTemplate returned an unexpected error: %v", err)
	}

	if string(output) != expectedOutput {
		t.Errorf("RenderTemplate output mismatch. Got '%s', want '%s'", string(output), expectedOutput)
	}
}

func TestRenderTemplate_MultiSelect(t *testing.T) {
	templateContent := `
items:
{{- range .items }}
  - {{ . }}
{{- end }}
`
	answers := map[string]interface{}{"items": []string{"a", "b"}}
	// Using strings.TrimSpace because the template might introduce leading/trailing whitespace
	// depending on how it's defined (e.g. newlines around range).
	// The key is that the YAML structure is valid and items are present.
	expectedOutputLines := []string{
		"items:",
		"  - a",
		"  - b",
	}
	expectedOutput := strings.Join(expectedOutputLines, "\n")

	output, err := RenderTemplate(templateContent, answers)
	if err != nil {
		t.Fatalf("RenderTemplate returned an unexpected error: %v", err)
	}

	// Normalize newlines and trim whitespace for comparison
	actualOutputStr := strings.TrimSpace(strings.ReplaceAll(string(output), "\r\n", "\n"))
	expectedOutputStr := strings.TrimSpace(strings.ReplaceAll(expectedOutput, "\r\n", "\n"))


	if actualOutputStr != expectedOutputStr {
		// To aid in debugging, print with explicit markers for space issues
		t.Errorf("RenderTemplate output mismatch for multiselect.\nGot:\n---\n%s\n---\nWant:\n---\n%s\n---", actualOutputStr, expectedOutputStr)
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	templateContent := "name: {{ .name " // Missing closing brace
	answers := map[string]interface{}{"name": "test"}

	_, err := RenderTemplate(templateContent, answers)
	if err == nil {
		t.Error("RenderTemplate did not return an error for an invalid template")
	}
}
