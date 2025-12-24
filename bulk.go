package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

// BulkConfig represents the jails.json structure
type BulkConfig struct {
	Jails []map[string]interface{} `json:"jails"`
}

// RenderTemplate renders a template with the given jail data
func RenderTemplate(tmpl *template.Template, jail map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, jail); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// CompareJailConf compares the template output with an existing config file
func CompareJailConf(tmpl *template.Template, jail map[string]interface{}, confPath string) (bool, error) {
	// Render template
	rendered, err := RenderTemplate(tmpl, jail)
	if err != nil {
		return false, err
	}

	// Read existing file
	existing, err := os.ReadFile(confPath)
	if err != nil {
		return false, fmt.Errorf("failed to read existing config: %w", err)
	}

	// Byte comparison
	return bytes.Equal(rendered, existing), nil
}

// LoadConfig loads and parses jails.json
func LoadConfig(path string) (*BulkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config BulkConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// ValidateJail validates required fields in a jail entry
func ValidateJail(jail map[string]interface{}) error {
	required := []string{"name", "number", "version"}
	for _, field := range required {
		if _, ok := jail[field]; !ok {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// LoadTemplate loads and parses the jail.conf template
func LoadTemplate(path string) (*template.Template, error) {
	return template.ParseFiles(path)
}

// GetJailName extracts the jail name from a jail entry
func GetJailName(jail map[string]interface{}) (string, error) {
	name, ok := jail["name"].(string)
	if !ok {
		return "", fmt.Errorf("'name' must be a string")
	}
	return name, nil
}

// GetJailNumber extracts the jail number from a jail entry
func GetJailNumber(jail map[string]interface{}) (int, error) {
	// JSON numbers are float64
	num, ok := jail["number"].(float64)
	if !ok {
		return 0, fmt.Errorf("'number' must be a number")
	}
	return int(num), nil
}

// GetJailVersion extracts the version from a jail entry
func GetJailVersion(jail map[string]interface{}) (string, error) {
	version, ok := jail["version"].(string)
	if !ok {
		return "", fmt.Errorf("'version' must be a string")
	}
	return version, nil
}

// FilterJails filters jails by name (if specified)
func FilterJails(jails []map[string]interface{}, name string) []map[string]interface{} {
	if name == "" {
		return jails
	}

	for _, jail := range jails {
		if jailName, _ := GetJailName(jail); jailName == name {
			return []map[string]interface{}{jail}
		}
	}

	return nil
}
