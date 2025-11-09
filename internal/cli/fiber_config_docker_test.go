//go:build docker

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFiberConfigPreforkDocker(t *testing.T) {
	config := createFiberConfig("Test App")

	// Prefork should be disabled for Docker builds
	assert.False(t, config.Prefork, "Prefork should be disabled for Docker deployments")
}
