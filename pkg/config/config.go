// Package config holds the configuration schema for image replacement rules.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = log.Log.WithName("config")

// Config represents the schema of the webhook's configuration.
type Config struct {
	Exclude []ExclusionRule   `json:"exclude"`
	Replace []ReplacementRule `json:"replace"`
}

// ExclusionRule represents a rule for an image prefix that should explicitly
// be excluded from any replacements.
type ExclusionRule struct {
	Prefix string `json:"prefix"`
}

// ReplacementRule represents a rule that matches an image prefix and replaces
// it with the provided replacement.
type ReplacementRule struct {
	Prefix      string `json:"prefix"`
	Replacement string `json:"replacement"`
}

// Load loads the configuration for the webhook from the given path.
func Load(path string) (*Config, error) {
	logger.V(1).Info("loading configuration", "path", path)

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	for i, rule := range c.Exclude {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid exclusion rule #%d: %w", i, err)
		}
	}

	for i, rule := range c.Replace {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid replacement rule #%d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the exclusion rule.
func (r *ExclusionRule) Validate() error {
	if r.Prefix == "" {
		return errors.New("prefix must not be empty")
	}

	return nil
}

// Validate validates the replacement rule.
func (r *ReplacementRule) Validate() error {
	if r.Prefix == "" {
		return errors.New("prefix must not be empty")
	}

	if r.Replacement == "" {
		return errors.New("replacement must not be empty")
	}

	return nil
}
