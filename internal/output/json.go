package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteJSON saves the result to a JSON file.
func WriteJSON(result *Result, pathPattern string) (string, error) {
	path := expandPattern(pathPattern, result)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory %q: %w", dir, err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write result to %q: %w", path, err)
	}

	return path, nil
}

// DefaultJSONPath returns the default output path for JSON results.
func DefaultJSONPath(scenarioName string) string {
	ts := time.Now().Format("20060102-150405")
	safe := sanitizeName(scenarioName)
	return filepath.Join("results", fmt.Sprintf("result-%s-%s.json", ts, safe))
}

func expandPattern(pattern string, result *Result) string {
	ts := result.StartedAt.Format("20060102-150405")
	pattern = strings.ReplaceAll(pattern, "{{timestamp}}", ts)
	pattern = strings.ReplaceAll(pattern, "{{name}}", sanitizeName(result.ScenarioName))
	return pattern
}

func sanitizeName(name string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-")
	return replacer.Replace(strings.ToLower(name))
}
