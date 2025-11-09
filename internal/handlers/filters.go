package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// buildFilterClause creates SQL WHERE conditions and args from query params
func buildFilterClause(c fiber.Ctx, baseArgs []interface{}) (string, []interface{}) {
	var conditions []string
	args := baseArgs

	// Helper function to add filter
	addFilter := func(columnName, value string) {
		if value != "" {
			argNum := len(args) + 1
			conditions = append(conditions, fmt.Sprintf("%s = $%d", columnName, argNum))
			args = append(args, value)
		}
	}

	// Add filters
	addFilter("s.country", c.Query("country"))
	addFilter("s.browser", c.Query("browser"))
	addFilter("s.device", c.Query("device"))
	addFilter("e.url_path", c.Query("page"))

	clause := ""
	if len(conditions) > 0 {
		clause = " AND " + strings.Join(conditions, " AND ")
	}

	return clause, args
}
