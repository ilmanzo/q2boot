package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load reads and JSON-decodes a VMConfig from path. Fields absent from the
// file keep their DefaultConfig values. The returned error wraps os.ErrNotExist
// when the file is missing so callers can distinguish "first run".
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

// Save JSON-encodes cfg to path with 0600 perms (config may contain
// host-port choices etc. — not secret, but no reason to be world-readable).
func Save(path string, cfg *VMConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
