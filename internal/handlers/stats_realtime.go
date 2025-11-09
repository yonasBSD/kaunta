package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/seuros/kaunta/internal/database"
)

// HandleCurrentVisitors returns count of visitors in last 5 minutes
// GET /api/stats/realtime/:website_id
func HandleCurrentVisitors(c fiber.Ctx) error {
	websiteIDStr := c.Params("website_id")
	websiteID, err := uuid.Parse(websiteIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid website ID",
		})
	}

	// Count distinct sessions from last 5 minutes
	// (Plausible uses last 5 minutes as default)
	query := `
		SELECT COUNT(DISTINCT session_id)
		FROM website_event
		WHERE website_id = $1
		  AND created_at >= NOW() - INTERVAL '5 minutes'
		  AND event_type = 1
	`

	var count int
	err = database.DB.QueryRow(query, websiteID).Scan(&count)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to query current visitors",
		})
	}

	return c.JSON(fiber.Map{
		"value": count,
	})
}
