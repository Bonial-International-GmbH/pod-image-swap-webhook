// Package config holds the configuration schema for image replacement rules.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	Pattern `yaml:",inline" json:",inline"`
}

// ReplacementRule represents a rule that matches an image prefix and replaces
// it with the provided replacement.
type ReplacementRule struct {
	Pattern     `yaml:",inline" json:",inline"`
	Replacement string `json:"replacement"`
}

// Pattern holds the different possible match pattern for exclusion and
// replacement rules.
type Pattern struct {
	Prefix string  `json:"prefix"`
	Regexp *Regexp `json:"regexp"`
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

// Validate validates the replacement rule.
func (r *ReplacementRule) Validate() error {
	if err := r.Pattern.Validate(); err != nil {
		return err
	}

	if r.Replacement == "" {
		return errors.New("replacement must not be empty")
	}

	return nil
}

// Validate validates the match pattern.
func (p *Pattern) Validate() error {
	if p.Prefix == "" && p.Regexp.IsEmpty() {
		return errors.New("one of `prefix` and `regexp` must be non-empty")
	}

	if p.Prefix != "" && !p.Regexp.IsEmpty() {
		return errors.New("only one of `prefix` and `regexp` must be set")
	}

	return nil
}

// ReplaceImage replaces the provides image based on the kind of pattern
// used. Returns the image unchanged if no replacement occurred.
func (p *Pattern) ReplaceImage(image string, replacement string) string {
	if p.matchPrefix(image) {
		return strings.Replace(image, p.Prefix, replacement, 1)
	}

	if p.matchRegexp(image) {
		return p.Regexp.ReplaceAllString(image, replacement)
	}

	return image
}

// MatchImage returns true if the pattern matches on the provided image.
func (p *Pattern) MatchImage(image string) bool {
	return p.matchPrefix(image) || p.matchRegexp(image)
}

func (p *Pattern) matchPrefix(image string) bool {
	return p.Prefix != "" && strings.HasPrefix(image, p.Prefix)
}

func (p *Pattern) matchRegexp(image string) bool {
	return !p.Regexp.IsEmpty() && p.Regexp.MatchString(image)
}
