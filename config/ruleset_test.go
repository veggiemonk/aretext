package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigForPath(t *testing.T) {
	testCases := []struct {
		name           string
		ruleSet        RuleSet
		path           string
		expectedConfig Config
	}{
		{
			name:    "no rules, default config",
			ruleSet: nil,
			path:    "test.go",
			expectedConfig: Config{
				SyntaxLanguage: DefaultSyntaxLanguage,
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				MenuCommands:   []MenuCommandConfig{},
			},
		},
		{
			name: "rule matches, set syntax language",
			ruleSet: []Rule{
				Rule{
					Name:    "json",
					Pattern: "**/*.json",
					Config: map[string]interface{}{
						"syntaxLanguage": "json",
					},
				},
				Rule{
					Name:    "mismatched rule",
					Pattern: "**/*.txt",
					Config: map[string]interface{}{
						"syntaxLanguage": "undefined",
					},
				},
			},
			path: "test.json",
			expectedConfig: Config{
				SyntaxLanguage: "json",
				TabSize:        DefaultTabSize,
				TabExpand:      DefaultTabExpand,
				AutoIndent:     DefaultAutoIndent,
				MenuCommands:   []MenuCommandConfig{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.ruleSet.ConfigForPath(tc.path)
			assert.Equal(t, tc.expectedConfig, c)
		})
	}
}
