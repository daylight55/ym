# Yamt - Go CLI for Kubernetes YAML Generation

`yamt` is a command-line tool written in Go to help automate the generation of configuration files, such as `helmfile.yaml` and `values.yaml`, for Kubernetes applications managed via GitOps. It uses Go's `text/template` engine.

## Current Status

This tool supports:
*   Reading configuration from a YAML file (`config.yaml` by default).
*   **Interactive mode** for configuration input if no config file is specified or `-interactive` flag is used.
*   Creating a structured output directory (`environments/ENV/CLUSTER/APP/`).
*   Generating `helmfile.yaml` and `values.yaml` from:
    *   **AppType-specific templates** (e.g., `templates/deployment/helmfile.yaml.gotmpl`).
    *   Root templates (`templates/helmfile.yaml.gotmpl`) as a fallback.
    *   User-specified custom template paths.
*   Basic input validation.

## Prerequisites

*   Go (version 1.18 or later recommended)

## Setup & Dependencies

The project uses Go modules. The primary external dependency is:
*   `gopkg.in/yaml.v2`: For parsing YAML input files.

To fetch dependencies, navigate to the project root and run:
```bash
go mod tidy
```
(Or simply let `go run` or `go build` handle it automatically).

## Usage

`yamt` can be run in two primary modes: file-based or interactive.

### 1. File-Based Mode

*   **Create a configuration file.** By default, `yamt` looks for `config.yaml` in the current directory. You can specify a different path using the `-configFile` flag:
    ```bash
    go run main.go -configFile=path/to/your/config.yaml
    ```

    **Example `config.yaml`:**
    ```yaml
    environment: development
    clusterName: staging-cluster
    appName: my-awesome-app
    appVersion: "1.2.3"
    appType: deployment # Used to find templates in templates/deployment/
    variables:
      key1: "value1"
      anotherKey: 123
      isProduction: false
    # helmfileTemplate: "custom-helmfile.yaml.gotmpl" # Optional: overrides AppType-based path
    # valuesTemplate: "custom-values.yaml.gotmpl"   # Optional: overrides AppType-based path
    ```
    *   The `appType` field is used to look for templates in a subdirectory under `templates/` (e.g., `templates/deployment/`). If `appType` is omitted, `yamt` looks for templates in the root `templates/` directory.
    *   `helmfileTemplate` and `valuesTemplate` can be used to specify exact template paths (relative to the `templates/` directory), overriding any `appType`-based conventions.

### 2. Interactive Mode

If no `-configFile` is provided and `config.yaml` is not found, or if the `-interactive` flag is used, `yamt` will enter interactive mode.
```bash
go run main.go
# or
go run main.go -interactive
```
The tool will guide you with prompts for the following information:
*   **Select Environment:** From a predefined list (e.g., dev, stg, prd).
*   **Select Cluster Name:** From a list filtered by the chosen environment.
*   **Enter Application Name:** Free-text input.
*   **Select Application Type (AppType):** From a predefined list (e.g., deployment, job, cronjob, service). This selection influences which templates are used by default.
*   **Enter Application Version:** Free-text input.
*   **Enter custom variables:** Key-value pairs, ending with an empty key.

### Output

Regardless of the mode, the tool will generate `helmfile.yaml` and `values.yaml` in the directory `environments/<environment>/<clusterName>/<appName>/`.

## Template Organization

`yamt` uses the following structure for finding and using templates:

1.  **AppType-Specific Templates (Default):**
    *   If `appType` is specified in the configuration (either in `config.yaml` or via interactive mode), `yamt` looks for templates in a subdirectory named after the `appType` within the `templates/` directory.
    *   Example: If `appType: deployment`, it will look for:
        *   `templates/deployment/helmfile.yaml.gotmpl`
        *   `templates/deployment/values.yaml.gotmpl`
    *   Sample `deployment` and `job` AppType templates are included.

2.  **Root Templates (Fallback):**
    *   If `appType` is *not* specified, `yamt` falls back to looking for templates directly in the root `templates/` directory:
        *   `templates/helmfile.yaml.gotmpl`
        *   `templates/values.yaml.gotmpl`

3.  **Custom Template Overrides:**
    *   The `helmfileTemplate` and `valuesTemplate` fields in `config.yaml` can be used to specify exact template paths (relative to the `templates/` directory). These paths will override any `appType`-based or root template selections.
    *   Example: If `helmfileTemplate: "my_custom_templates/special_helmfile.yaml.gotmpl"` is in `config.yaml`, that exact path will be used.

## Current Static Choices (Interactive Mode)

The following choices are currently hardcoded for interactive mode. These may become configurable in the future.

*   **Environments:**
    *   `dev`
    *   `stg`
    *   `prd`
*   **Clusters (by Environment):**
    *   `dev`: `dev-cluster-a`, `dev-cluster-b`
    *   `stg`: `stg-cluster-a`
    *   `prd`: `prd-cluster-a`, `prd-cluster-b`, `prd-cluster-c`
*   **Application Types (AppTypes):**
    *   `deployment`
    *   `job`
    *   `cronjob`
    *   `service`

## Development

*   Run tests: `go test -v ./...`

## Next Steps (Planned Features)
*   More flexible template selection (e.g., via flags).
*   Enhanced input validation.
*   More comprehensive test coverage, especially for interactive mode scenarios.
*   Allow configuration of interactive mode choices (environments, clusters, appTypes).
```
