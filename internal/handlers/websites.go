package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seuros/kaunta/internal/database"
)

// HandleWebsites returns list of all websites
func HandleWebsites(c *fiber.Ctx) error {
	rows, err := database.DB.Query(`
		SELECT website_id, domain, name
		FROM website
		ORDER BY name, domain
	`)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to query websites",
		})
	}
	defer func() { _ = rows.Close() }()

	var websites []Website
	for rows.Next() {
		var website Website
		var name *string
		if err := rows.Scan(&website.ID, &website.Domain, &name); err != nil {
			continue
		}
		if name != nil {
			website.Name = *name
		} else {
			website.Name = website.Domain
		}
		websites = append(websites, website)
	}

	return c.JSON(websites)
}
