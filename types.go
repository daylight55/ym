package main

// DynamicPromptConfig holds the configuration for dynamically generating prompt options.
type DynamicPromptConfig struct {
	// Source specifies the source of the dynamic options, e.g., "directory".
	Source string `yaml:"source"`
	// Path specifies the path to the source, e.g., "./modules".
	Path string `yaml:"path"`
}

// Prompt represents a single prompt configuration.
type Prompt struct {
	// Name is the variable name for the template.
	Name string `yaml:"name"`
	// Message is the text to display to the user.
	Message string `yaml:"message"`
	// Type is the type of the prompt, e.g., "input", "select", "multiselect".
	Type string `yaml:"type"`
	// Options is a list of options for select/multiselect prompts.
	Options []string `yaml:"options"`
	// DefaultValue is the default value for the prompt.
	DefaultValue string `yaml:"defaultValue"`
	// Dynamic holds the configuration for dynamic option generation.
	// This field is a pointer to allow for optional dynamic configuration.
	Dynamic *DynamicPromptConfig `yaml:"dynamic,omitempty"`
}

// PromptsConfig represents a list of prompt configurations.
type PromptsConfig struct {
	// Prompts is a slice to hold multiple prompt configurations.
	Prompts []Prompt `yaml:"prompts"`
}
