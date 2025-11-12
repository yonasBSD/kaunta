package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/seuros/kaunta/internal/database"
)

// HandleWebsites returns list of all websites with pagination
func HandleWebsites(c fiber.Ctx) error {
	// Parse pagination parameters
	pagination := ParsePaginationParams(c)

	// Query with COUNT and pagination
	rows, err := database.DB.Query(`
		WITH total AS (
			SELECT COUNT(*)::BIGINT as count FROM website
		)
		SELECT w.website_id, w.domain, w.name, t.count as total_count
		FROM website w
		CROSS JOIN total t
		ORDER BY w.name, w.domain
		LIMIT $1 OFFSET $2
	`, pagination.Per, pagination.Offset)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to query websites",
		})
	}
	defer func() { _ = rows.Close() }()

	var websites []Website
	var totalCount int64
	for rows.Next() {
		var website Website
		var name *string
		var rowTotal int64
		if err := rows.Scan(&website.ID, &website.Domain, &name, &rowTotal); err != nil {
			continue
		}
		totalCount = rowTotal // Capture total count
		if name != nil {
			website.Name = *name
		} else {
			website.Name = website.Domain
		}
		websites = append(websites, website)
	}

	// Return paginated response
	return c.JSON(NewPaginatedResponse(websites, pagination, totalCount))
}
