//go:build !docker

package cli

import "github.com/gofiber/fiber/v2"

// createFiberConfig returns Fiber configuration for bare metal deployments.
// Prefork is enabled to spawn multiple processes (one per CPU core) for
// improved performance on dedicated servers and VPS environments.
func createFiberConfig(appName string) fiber.Config {
	return fiber.Config{
		AppName: appName,
		Prefork: true, // Enable multi-process mode for bare metal
	}
}
