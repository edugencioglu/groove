package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Labels maps worktree paths to user-assigned labels.
type Labels map[string]string

// configDir returns the config directory, respecting WT_CONFIG_DIR for testing.
func configDir() string {
	if d := os.Getenv("WT_CONFIG_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "groove")
	}
	return filepath.Join(home, ".config", "groove")
}

func labelsPath() string {
	return filepath.Join(configDir(), "labels.json")
}

// LoadLabels reads labels from disk. Returns an empty map if the file doesn't exist.
func LoadLabels() (Labels, error) {
	data, err := os.ReadFile(labelsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Labels{}, nil
		}
		return nil, err
	}

	var labels Labels
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

// SaveLabels writes labels to disk, creating the config directory if needed.
func SaveLabels(labels Labels) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(labels, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(labelsPath(), data, 0o644)
}
