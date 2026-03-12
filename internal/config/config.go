// Package config handles loading and validating the migration JSON configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the top-level migration configuration.
type Config struct {
	Schema     string    `json:"$schema,omitempty"`
	Version    string    `json:"version"`
	Metadata   Metadata  `json:"metadata"`
	BackupRoot string    `json:"backup_root"`
	Apps       []App     `json:"apps"`
	Profiles   []Profile `json:"profiles,omitempty"`
}

// Metadata holds project-level information about this config.
type Metadata struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description,omitempty"`
	Repository  string `json:"repository,omitempty"`
}

// App represents a single application whose state can be migrated.
type App struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Items    []Item `json:"items"`
}

// Item is a single file or directory copy operation within an App.
type Item struct {
	Label       string `json:"label"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Strategy    string `json:"strategy,omitempty"` // "mirror" (default) | "file" | "selective"
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// Profile is a named preset that selects a subset of items by ID.
type Profile struct {
	Name string   `json:"name"`
	IDs  []string `json:"ids"`
}

// Load reads and parses a migration config from the given JSON file path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic structural checks on the loaded config.
func validate(cfg *Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("missing required field: version")
	}
	if cfg.BackupRoot == "" {
		return fmt.Errorf("missing required field: backup_root")
	}
	if len(cfg.Apps) == 0 {
		return fmt.Errorf("no apps defined in config")
	}

	seen := make(map[string]bool)
	for _, app := range cfg.Apps {
		if app.Name == "" {
			return fmt.Errorf("app missing required field: name")
		}
		if app.Category == "" {
			return fmt.Errorf("app %q missing required field: category", app.Name)
		}
		for _, item := range app.Items {
			if item.Label == "" {
				return fmt.Errorf("app %q has item with missing label", app.Name)
			}
			if item.Source == "" {
				return fmt.Errorf("app %q item %q missing source path", app.Name, item.Label)
			}
			if item.Target == "" {
				return fmt.Errorf("app %q item %q missing target path", app.Name, item.Label)
			}
			id := app.Name + ":" + item.Label
			if seen[id] {
				return fmt.Errorf("duplicate item ID: %s", id)
			}
			seen[id] = true
		}
	}

	return nil
}
