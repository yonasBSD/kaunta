package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// HandleDashboard renders the dashboard UI
func HandleDashboard(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Render("dashboard", fiber.Map{
		"Title": "Kaunta Dashboard",
	})
}

// RenderDashboardHTML is used when dashboard template is embedded
func RenderDashboardHTML(templateHTML string) string {
	// Replace template variable if needed
	result := strings.ReplaceAll(templateHTML, "{{.Title}}", "Kaunta Dashboard")
	return result
}
