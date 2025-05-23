package main

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderTemplate processes a template string with a map of answers and returns the rendered output.
//
// Args:
//   templateContent (string): The raw template string to be processed.
//   answers (map[string]interface{}): A map containing the data to be injected into the template.
//                                      Keys are variable names used in the template.
//
// Returns:
//   []byte: The rendered content (intended to be YAML).
//   error: An error if template parsing or execution fails.
func RenderTemplate(templateContent string, answers map[string]interface{}) ([]byte, error) {
	// Create a new template instance.
	// The name "yamlTemplate" is arbitrary and used for error reporting.
	tmpl, err := template.New("yamlTemplate").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Create a buffer to store the executed template output.
	var buffer bytes.Buffer

	// Execute the template, injecting the answers.
	// The 'answers' map will be the data context for the template.
	// For example, if answers has {"name": "John"}, {{.name}} in the template will render "John".
	// Slices within the answers map, when used with `{{range}}`, should render correctly for YAML lists.
	// e.g., if answers has {"items": ["one", "two"]}, and template is:
	// items:
	// {{range .items}}  - {{.}}
	// {{end}}
	// It will render as:
	// items:
	//   - one
	//   - two
	err = tmpl.Execute(&buffer, answers)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Return the rendered content from the buffer.
	return buffer.Bytes(), nil
}

// Example usage (can be in main.go or a test file)
/*
func main() {
	templateStr := `
name: {{.projectName}}
module: {{.selectedModule}}
features:
{{range .selectedFeatures}}  - {{.}}
{{end}}
replicas: {{.replicaCount}}
`
	answers := map[string]interface{}{
		"projectName":      "WebApp1",
		"selectedModule":   "nginx",
		"selectedFeatures": []string{"logging", "monitoring"},
		"replicaCount":     3,
	}

	renderedYAML, err := RenderTemplate(templateStr, answers)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	fmt.Println("Rendered YAML:\n", string(renderedYAML))
}
*/
