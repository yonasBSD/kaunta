//go:build !docker

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFiberConfigPreforkBareMetal(t *testing.T) {
	config := createFiberConfig("Test App")

	// Prefork should be enabled for bare metal builds
	assert.True(t, config.Prefork, "Prefork should be enabled for bare metal deployments")
}
