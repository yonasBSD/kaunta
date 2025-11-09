package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/seuros/kaunta/internal/database"
)

// HandleDashboardStats returns aggregated stats for the dashboard
func HandleDashboardStats(c fiber.Ctx) error {
	websiteIDStr := c.Params("website_id")
	websiteID, err := uuid.Parse(websiteIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid website ID",
		})
	}

	// Build filter clause
	filterClause, filterArgs := buildFilterClause(c, []interface{}{websiteID})

	// Get current visitors (last 5 minutes)
	var currentVisitors int
	query := `
		SELECT COUNT(DISTINCT e.session_id)
		FROM website_event e
		JOIN session s ON e.session_id = s.session_id
		WHERE e.website_id = $1
		  AND e.created_at >= NOW() - INTERVAL '5 minutes'
		  AND e.event_type = 1` + filterClause

	err = database.DB.QueryRow(query, filterArgs...).Scan(&currentVisitors)
	if err != nil {
		currentVisitors = 0
	}

	// Get today's pageviews
	var todayPageviews int
	filterClause, filterArgs = buildFilterClause(c, []interface{}{websiteID})
	query = `
		SELECT COUNT(*)
		FROM website_event e
		JOIN session s ON e.session_id = s.session_id
		WHERE e.website_id = $1
		  AND e.created_at >= CURRENT_DATE
		  AND e.event_type = 1` + filterClause

	err = database.DB.QueryRow(query, filterArgs...).Scan(&todayPageviews)
	if err != nil {
		todayPageviews = 0
	}

	// Get today's unique visitors
	var todayVisitors int
	filterClause, filterArgs = buildFilterClause(c, []interface{}{websiteID})
	query = `
		SELECT COUNT(DISTINCT e.session_id)
		FROM website_event e
		JOIN session s ON e.session_id = s.session_id
		WHERE e.website_id = $1
		  AND e.created_at >= CURRENT_DATE
		  AND e.event_type = 1` + filterClause

	err = database.DB.QueryRow(query, filterArgs...).Scan(&todayVisitors)
	if err != nil {
		todayVisitors = 0
	}

	// Calculate bounce rate (simplified: sessions with only 1 pageview)
	bounceRate := "0%"
	if todayVisitors > 0 {
		var bounces int
		filterClause, filterArgs = buildFilterClause(c, []interface{}{websiteID})
		query = `
			SELECT COUNT(*)
			FROM (
				SELECT e.session_id, COUNT(*) as views
				FROM website_event e
				JOIN session s ON e.session_id = s.session_id
				WHERE e.website_id = $1
				  AND e.created_at >= CURRENT_DATE
				  AND e.event_type = 1` + filterClause + `
				GROUP BY e.session_id
				HAVING COUNT(*) = 1
			) bounced_sessions`

		_ = database.DB.QueryRow(query, filterArgs...).Scan(&bounces)

		bounceRatePercent := float64(bounces) / float64(todayVisitors) * 100
		bounceRate = fmt.Sprintf("%.1f%%", bounceRatePercent)
	}

	return c.JSON(DashboardStats{
		CurrentVisitors: currentVisitors,
		TodayPageviews:  todayPageviews,
		TodayVisitors:   todayVisitors,
		TodayBounceRate: bounceRate,
	})
}
