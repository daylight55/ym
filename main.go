package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	templateFile string
	promptsFile  string
	outputFile   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yaml-generator",
	Short: "A CLI tool to generate YAML files from templates and user prompts.",
	Long: `yaml-generator is a flexible tool that takes a template file
and a prompts configuration file to interactively collect user input
and generate a final YAML output.`,
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a YAML file based on a template and prompts.",
	Long: `The generate command processes a template file and a prompts configuration file.
It will ask the user for input based on the prompts and then render the template
with the provided answers, outputting the result to a file or stdout.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Prompts Configuration
		promptsData, err := os.ReadFile(promptsFile)
		if err != nil {
			log.Fatalf("Failed to read prompts file '%s': %v", promptsFile, err)
		}

		var promptsConfig PromptsConfig
		err = yaml.Unmarshal(promptsData, &promptsConfig)
		if err != nil {
			log.Fatalf("Failed to unmarshal prompts YAML from '%s': %v", promptsFile, err)
		}

		// 2. Run Prompts
		answers, err := RunPrompts(promptsConfig.Prompts)
		if err != nil {
			log.Fatalf("Error running prompts: %v", err)
		}

		// 3. Load Template File
		templateContentBytes, err := LoadTemplateFile(templateFile)
		if err != nil {
			// LoadTemplateFile already provides a good error message
			log.Fatalf("Error loading template file: %v", err)
		}
		templateContentString := string(templateContentBytes)

		// 4. Render Template
		renderedYAML, err := RenderTemplate(templateContentString, answers)
		if err != nil {
			// RenderTemplate also provides a good error message
			log.Fatalf("Error rendering template: %v", err)
		}

		// 5. Write Output
		if outputFile != "" {
			err = os.WriteFile(outputFile, renderedYAML, 0644) // Standard file permissions
			if err != nil {
				log.Fatalf("Failed to write output to file '%s': %v", outputFile, err)
			}
			fmt.Printf("Successfully generated YAML file: %s\n", outputFile)
		} else {
			// Print to stdout
			_, err = fmt.Println(string(renderedYAML))
			if err != nil {
				log.Fatalf("Failed to print output to stdout: %v", err)
			}
		}
	},
}

func init() {
	// Add generateCmd as a subcommand of rootCmd
	rootCmd.AddCommand(generateCmd)

	// Define persistent flags for generateCmd
	generateCmd.PersistentFlags().StringVarP(&templateFile, "templateFile", "t", "", "Path to the input YAML template (required)")
	generateCmd.MarkPersistentFlagRequired("templateFile") // Mark as required

	generateCmd.PersistentFlags().StringVarP(&promptsFile, "promptsFile", "p", "", "Path to the YAML file defining the prompts (required)")
	generateCmd.MarkPersistentFlagRequired("promptsFile") // Mark as required

	generateCmd.PersistentFlags().StringVarP(&outputFile, "outputFile", "o", "", "Path to the output generated YAML file (if empty, output to stdout)")
}

func main() {
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints errors to stderr, so os.Exit(1) is often sufficient
		// log.Fatal() would print the error again.
		os.Exit(1)
	}
}
