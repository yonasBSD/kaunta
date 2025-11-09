//go:build docker

package cli

import "github.com/gofiber/fiber/v3"

// createFiberConfig returns Fiber configuration for Docker deployments.
// Prefork is disabled to maintain single-process behavior required for
// container orchestration, health checks, and proper signal handling.
func createFiberConfig(appName string) fiber.Config {
	return fiber.Config{
		AppName: appName,
		Prefork: false, // Disable multi-process mode for Docker
	}
}
