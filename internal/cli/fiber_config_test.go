package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFiberConfig(t *testing.T) {
	appName := "Test App"
	config := createFiberConfig(appName, nil)

	// AppName should always be set correctly
	assert.Equal(t, appName, config.AppName, "AppName should match input")
}

func TestCreateFiberConfigAppNameFormat(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		expected string
	}{
		{
			name:     "simple name",
			appName:  "Kaunta",
			expected: "Kaunta",
		},
		{
			name:     "name with version",
			appName:  "Kaunta v1.0.0",
			expected: "Kaunta v1.0.0",
		},
		{
			name:     "empty name",
			appName:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createFiberConfig(tt.appName, nil)
			assert.Equal(t, tt.expected, config.AppName)
		})
	}
}
