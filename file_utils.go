package main

import (
	"fmt"
	"os"
)

// LoadTemplateFile reads the content of the file specified by filePath.
// It returns the file content as a byte slice and an error.
// It handles potential errors, such as the file not existing or not being readable,
// by returning an appropriate error.
func LoadTemplateFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file '%s': %w", filePath, err)
	}
	return content, nil
}
