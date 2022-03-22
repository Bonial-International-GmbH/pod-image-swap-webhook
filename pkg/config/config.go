// Package config holds the configuration schema for image replacement rules.
package config

import (
	"io"
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

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
