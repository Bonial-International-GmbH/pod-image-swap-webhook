package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
	t.Run("sample config is always valid", func(t *testing.T) {
		cfg, err := Load("../../config.sample.yaml")
		assert.NoError(t, err)
		assert.Greater(t, len(cfg.Exclude), 0)
		assert.Greater(t, len(cfg.Replace), 0)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("empty config is valid", func(t *testing.T) {
		cfg := &Config{}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Exclude: []ExclusionRule{
				{Prefix: "someregistry.org/some-namespace"},
			},
			Replace: []ReplacementRule{
				{
					Prefix:      "someregistry.org",
					Replacement: "otherregistry.org/someregistry.org",
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("exclusion rule prefix must not be empty", func(t *testing.T) {
		cfg := &Config{
			Exclude: []ExclusionRule{
				{Prefix: ""},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("replacement rule prefix must not be empty", func(t *testing.T) {
		cfg := &Config{
			Replace: []ReplacementRule{
				{Prefix: "", Replacement: "otherregistry.org/someregistry.org"},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("replacement rule replacement must not be empty", func(t *testing.T) {
		cfg := &Config{
			Replace: []ReplacementRule{
				{Prefix: "someregistry.org", Replacement: ""},
			},
		}
		assert.Error(t, cfg.Validate())
	})
}
