package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

// Config represents the structure of the configuration file.
type Config struct {
	Environment      string                 `yaml:"environment"`
	ClusterName      string                 `yaml:"clusterName"`
	AppName          string                 `yaml:"appName"`
	AppVersion       string                 `yaml:"appVersion"`
	AppType          string                 `yaml:"appType,omitempty"`
	HelmfileTemplate string                 `yaml:"helmfileTemplate,omitempty"`
	ValuesTemplate   string                 `yaml:"valuesTemplate,omitempty"`
	Variables        map[string]interface{} `yaml:"variables"`
}

// validateConfig checks if the essential fields in the Config are set.
func validateConfig(config *Config) error {
	if config.Environment == "" {
		return errors.New("environment must be set")
	}
	if config.ClusterName == "" {
		return errors.New("clusterName must be set")
	}
	if config.AppName == "" {
		return errors.New("appName must be set")
	}
	if config.AppVersion == "" {
		return errors.New("appVersion must be set")
	}
	// AppType is conditionally mandatory:
	// If HelmfileTemplate is not explicitly set, AppType must be provided to derive the template path.
	if config.AppType == "" && config.HelmfileTemplate == "" {
		return errors.New("AppType must be specified when HelmfileTemplate is not explicitly set")
	}
	// A similar check could be added for ValuesTemplate if it's considered independently mandatory.
	// For now, HelmfileTemplate implies the need for ValuesTemplate as well.
	return nil
}

// Hardcoded choices for interactive mode
var (
	environments = []string{"dev", "stg", "prd"}
	clustersByEnvironment = map[string][]string{
		"dev": {"dev-cluster-a", "dev-cluster-b"},
		"stg": {"stg-cluster-a"},
		"prd": {"prd-cluster-a", "prd-cluster-b", "prd-cluster-c"},
	}
	appTypes = []string{"deployment", "job", "cronjob", "service"}
)

// loadConfig reads and parses the YAML configuration file.
func loadConfig(configFile string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file %s: %w", configFile, err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML from %s: %w", configFile, err)
	}

	return &config, nil
}

// createOutputDirectory creates the directory structure based on the config.
func createOutputDirectory(config *Config) (string, error) {
	dirPath := filepath.Join("environments", config.Environment, config.ClusterName, config.AppName)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating directory %s: %w", dirPath, err)
	}
	return dirPath, nil
}

// generateFileFromTemplate reads a template file, parses it, and executes it with the given data,
// writing the output to the outputPath.
func generateFileFromTemplate(templatePath string, outputPath string, data *Config) error {
	// Read the template file content
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("error reading template file %s: %w", templatePath, err)
	}

	// Create a new template name from the base of the template path
	templateName := filepath.Base(templatePath)

	// Parse the template content
	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	// Create the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Execute the template with the data and write to the file
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing template %s: %w", templateName, err)
	}

	return nil
}

func main() {
	// Define command-line flags
	configFile := flag.String("configFile", "config.yaml", "Path to the input configuration YAML file.")
	interactive := flag.Bool("interactive", false, "Enable interactive mode to input configuration.")

	// Parse the flags
	flag.Parse()

	// Print a message indicating which config file will be used
	// This message will be conditional later based on interactive mode
	// fmt.Printf("Using config file: %s\n", *configFile)

	// Load the configuration
	var config *Config
	var err error

	// Check if interactive mode should be triggered
	_, statErr := os.Stat(*configFile)
	if *interactive || (isFlagPassed("configFile") == false && os.IsNotExist(statErr)) {
		fmt.Println("Running in interactive mode.")
		config, err = getConfigFromInteractiveMode()
		if err == nil { // Validate only if interactive mode succeeded
			err = validateConfig(config)
		}
	} else {
		fmt.Printf("Using config file: %s\n", *configFile)
		config, err = loadConfig(*configFile)
		if err == nil { // Validate only if loading succeeded
			err = validateConfig(config)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Print the loaded configuration
	fmt.Printf("Loaded config: %+v\n", config)

	// Create the output directory
	dirPath, err := createOutputDirectory(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Output directory ensured: %s\n", dirPath)

	// --- Determine template paths with AppType and overrides ---
	var defaultHelmfileTmplName, defaultValuesTmplName string

	if config.AppType != "" {
		defaultHelmfileTmplName = filepath.Join(config.AppType, "helmfile.yaml.gotmpl")
		defaultValuesTmplName = filepath.Join(config.AppType, "values.yaml.gotmpl")
	} else {
		// Fallback to root templates if AppType is not set
		defaultHelmfileTmplName = "helmfile.yaml.gotmpl"
		defaultValuesTmplName = "values.yaml.gotmpl"
	}

	helmfileTemplateName := defaultHelmfileTmplName
	if config.HelmfileTemplate != "" {
		helmfileTemplateName = config.HelmfileTemplate
	}

	valuesTemplateName := defaultValuesTmplName
	if config.ValuesTemplate != "" {
		valuesTemplateName = config.ValuesTemplate
	}

	// Construct full template paths
	helmfileTemplatePath := filepath.Join("templates", helmfileTemplateName)
	valuesTemplatePath := filepath.Join("templates", valuesTemplateName)

	// Check if the selected template files exist
	if _, err := os.Stat(helmfileTemplatePath); os.IsNotExist(err) {
		// Specific error if it was an AppType derived path that's missing
		if config.AppType != "" && helmfileTemplateName == filepath.Join(config.AppType, "helmfile.yaml.gotmpl") {
			fmt.Fprintf(os.Stderr, "Error: Helmfile template for AppType '%s' not found at %s\n", config.AppType, helmfileTemplatePath)
		} else {
			fmt.Fprintf(os.Stderr, "Error: Specified Helmfile template not found at %s\n", helmfileTemplatePath)
		}
		os.Exit(1)
	}

	if _, err := os.Stat(valuesTemplatePath); os.IsNotExist(err) {
		// Specific error if it was an AppType derived path that's missing
		if config.AppType != "" && valuesTemplateName == filepath.Join(config.AppType, "values.yaml.gotmpl") {
			fmt.Fprintf(os.Stderr, "Error: Values template for AppType '%s' not found at %s\n", config.AppType, valuesTemplatePath)
		} else {
			fmt.Fprintf(os.Stderr, "Error: Specified Values template not found at %s\n", valuesTemplatePath)
		}
		os.Exit(1)
	}
	// --- End of template path determination ---

	// Construct output file paths
	helmfileOutputPath := filepath.Join(dirPath, "helmfile.yaml")
	valuesOutputPath := filepath.Join(dirPath, "values.yaml")

	// Generate helmfile.yaml
	err = generateFileFromTemplate(helmfileTemplatePath, helmfileOutputPath, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating helmfile.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated %s\n", helmfileOutputPath)

	// Generate values.yaml
	err = generateFileFromTemplate(valuesTemplatePath, valuesOutputPath, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating values.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated %s\n", valuesOutputPath)
}

// promptUser displays a prompt and reads a line of input from the user.
func promptUser(reader *bufio.Reader, promptText string) (string, error) {
	fmt.Print(promptText + " ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

// promptSelection displays a list of options and prompts the user to select one.
func promptSelection(reader *bufio.Reader, promptText string, options []string) (string, error) {
	fmt.Println(promptText)
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	for { // Loop until valid input is received
		fmt.Print("Enter number: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading selection: %w", err)
		}
		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if choice < 1 || choice > len(options) {
			fmt.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(options))
			continue
		}
		return options[choice-1], nil
	}
}

// getConfigFromInteractiveMode prompts the user for configuration values.
func getConfigFromInteractiveMode() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	var err error
	config := &Config{}

	// Prompt for Environment
	config.Environment, err = promptSelection(reader, "Select Environment:", environments)
	if err != nil {
		return nil, err
	}

	// Prompt for ClusterName based on selected Environment
	availableClusters, ok := clustersByEnvironment[config.Environment]
	if !ok || len(availableClusters) == 0 {
		// Fallback to free-text if no clusters are defined or env is missing (should not happen with selection)
		fmt.Printf("No predefined clusters for %s. Please enter manually.\n", config.Environment)
		config.ClusterName, err = promptUser(reader, "Enter Cluster Name:")
		if err != nil {
			return nil, err
		}
	} else {
		config.ClusterName, err = promptSelection(reader, "Select Cluster Name:", availableClusters)
		if err != nil {
			return nil, err
		}
	}

	// Prompt for AppName (free-text)
	config.AppName, err = promptUser(reader, "Enter App Name:")
	if err != nil {
		return nil, err
	}

	// Prompt for AppType
	config.AppType, err = promptSelection(reader, "Select App Type:", appTypes)
	if err != nil {
		return nil, err
	}

	// Prompt for AppVersion (free-text)
	config.AppVersion, err = promptUser(reader, "Enter App Version:")
	if err != nil {
		return nil, err
	}

	config.Variables = make(map[string]interface{})
	fmt.Println("Enter variables (key-value pairs):")
	for {
		key, err := promptUser(reader, "Enter variable key (or leave empty to finish variables):")
		if err != nil {
			return nil, err
		}
		if key == "" {
			break
		}

		value, err := promptUser(reader, fmt.Sprintf("Enter variable value for %s:", key))
		if err != nil {
			return nil, err
		}
		config.Variables[key] = value
	}

	// HelmfileTemplate and ValuesTemplate are left empty as they are optional
	// and not easily collected in a simple interactive CLI without more complex logic.

	return config, nil
}

// isFlagPassed checks if a command-line flag was explicitly set by the user.
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
