package scenario

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a scenario YAML file.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file %q: %w", path, err)
	}

	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %q: %w", path, err)
	}

	// Apply defaults
	if s.Target.Method == "" {
		s.Target.Method = "GET"
	}
	if s.Execution.Timeout == "" {
		s.Execution.Timeout = "10s"
	}
	if s.Execution.Concurrency == 0 {
		s.Execution.Concurrency = 1
	}
	if s.Safety.Environment == "" {
		s.Safety.Environment = "staging"
	}

	return &s, nil
}
