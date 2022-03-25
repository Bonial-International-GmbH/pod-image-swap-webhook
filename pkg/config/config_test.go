package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	t.Run("sample config is always valid", func(t *testing.T) {
		cfg, err := Load("../../config.sample.yaml")
		require.NoError(t, err)
		assert.Greater(t, len(cfg.Exclude), 0)
		assert.Greater(t, len(cfg.Replace), 0)
	})

	t.Run("invalid regexp causes error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		content := "exclude:\n- regexp: '(]invalid'"
		require.NoError(t, os.WriteFile(path, []byte(content), 0664))

		_, err := Load(path)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error parsing regexp")
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
				{
					Pattern: Pattern{Prefix: "someregistry.org/some-namespace"},
				},
			},
			Replace: []ReplacementRule{
				{
					Pattern:     Pattern{Prefix: "someregistry.org"},
					Replacement: "otherregistry.org/someregistry.org",
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("exclusion rule prefix must not be empty", func(t *testing.T) {
		cfg := &Config{
			Exclude: []ExclusionRule{
				{
					Pattern: Pattern{Prefix: ""},
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("replacement rule prefix must not be empty", func(t *testing.T) {
		cfg := &Config{
			Replace: []ReplacementRule{
				{
					Pattern:     Pattern{Prefix: ""},
					Replacement: "otherregistry.org/someregistry.org",
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("replacement rule replacement must not be empty", func(t *testing.T) {
		cfg := &Config{
			Replace: []ReplacementRule{
				{
					Pattern:     Pattern{Prefix: "someregistry.org"},
					Replacement: "",
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("only one of prefix and regexp must be set at the same time", func(t *testing.T) {
		cfg := &Config{
			Replace: []ReplacementRule{
				{
					Pattern: Pattern{
						Prefix: "someregistry.org",
						Regexp: MustCompileRegexp("^someregistry.org"),
					},
					Replacement: "otherregistry.org",
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})
}

var patternTestCases = []struct {
	pattern     Pattern
	image       string
	match       bool
	replacement string
	expected    string
}{
	{
		pattern:     Pattern{Prefix: "busybox"},
		image:       "busybox",
		replacement: "lazybox",
		expected:    "lazybox",
		match:       true,
	},
	{
		pattern:     Pattern{Prefix: "busybox"},
		image:       "busybox:latest",
		replacement: "lazybox",
		expected:    "lazybox:latest",
		match:       true,
	},
	{
		pattern:     Pattern{Prefix: "busybox"},
		image:       "someregistry.org/library/busybox:latest",
		replacement: "lazybox",
		expected:    "someregistry.org/library/busybox:latest",
		match:       false,
	},
	{
		pattern:     Pattern{Regexp: MustCompileRegexp("^busybox$")},
		image:       "busybox",
		replacement: "foo",
		expected:    "foo",
		match:       true,
	},
	{
		pattern:     Pattern{Regexp: MustCompileRegexp("^busybox$")},
		image:       "busybox:latest",
		replacement: "foo",
		expected:    "busybox:latest",
		match:       false,
	},
	{
		pattern:     Pattern{Regexp: MustCompileRegexp("^busybox$")},
		image:       "someregistry.org/library/busybox:latest",
		replacement: "foo",
		expected:    "someregistry.org/library/busybox:latest",
		match:       false,
	},
	{
		pattern:     Pattern{Regexp: MustCompileRegexp("^busy")},
		image:       "busybox",
		replacement: "lazy",
		expected:    "lazybox",
		match:       true,
	},
	{
		pattern:     Pattern{Regexp: MustCompileRegexp("^(?P<registry>[^/]+)/(?P<image>[^:]+):(?P<tag>.+)$")},
		image:       "docker.io/library/nginx:latest",
		replacement: "myregistry.org/${image}:1.0.0",
		expected:    "myregistry.org/library/nginx:1.0.0",
		match:       true,
	},
}

func TestPattern(t *testing.T) {
	t.Parallel()

	for i, tc := range patternTestCases {
		tc := tc

		t.Run(fmt.Sprintf("case #%d", i), func(t *testing.T) {
			assert.Equal(t, tc.match, tc.pattern.MatchImage(tc.image))
			assert.Equal(t, tc.expected, tc.pattern.ReplaceImage(tc.image, tc.replacement))
		})
	}
}
