package engine

import (
	"testing"
)

func TestSanitizeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean command",
			input:    "cargo build --release",
			expected: "cargo build --release",
		},
		{
			name:     "token flag",
			input:    "npm publish --token ghp_123456789012345678901234567890123456",
			expected: "npm publish --token [REDACTED]",
		},
		{
			name:     "standalone github token",
			input:    "export GITHUB_TOKEN=ghp_123456789012345678901234567890123456",
			expected: "export GITHUB_TOKEN=[REDACTED_GH_TOKEN]",
		},
		{
			name:     "password flag with equals",
			input:    "mysql -u root --password=SuperSecretPassword123 mydb",
			expected: "mysql -u root --password=[REDACTED] mydb",
		},
		{
			name:     "short password flag",
			input:    "mysql -u root -p MySecretPassword",
			expected: "mysql -u root -p [REDACTED]",
		},
		{
			name:     "openai api key in curl",
			input:    "curl https://api.openai.com -H 'Authorization: Bearer sk-1234567890abcdef1234567890'",
			expected: "curl https://api.openai.com -H 'Authorization: Bearer [REDACTED_AUTH]'",
		},
		{
			name:     "standalone openai key",
			input:    "python script.py sk-1234567890abcdef1234567890",
			expected: "python script.py [REDACTED_KEY]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCommand(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
