package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load returns os.ErrNotExist verbatim so callers can detect "first run".
// Fields absent from the file keep their DefaultConfig values.
func Load(path string) (*VMConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg *VMConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
