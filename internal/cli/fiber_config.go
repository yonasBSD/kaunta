//go:build !docker

package cli

import "github.com/gofiber/fiber/v3"

// createFiberConfig returns Fiber configuration for bare metal deployments.
func createFiberConfig(appName string) fiber.Config {
	return fiber.Config{
		AppName: appName,
	}
}
