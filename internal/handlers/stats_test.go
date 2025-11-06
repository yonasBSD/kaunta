package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDashboardStats_Success(t *testing.T) {
	websiteID := uuid.New()

	responses := []mockResponse{
		{
			match:   "SELECT COUNT(DISTINCT e.session_id)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{3}},
		},
		{
			match:   "SELECT COUNT(*)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{12}},
		},
		{
			match:   "SELECT COUNT(DISTINCT e.session_id)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{6}},
		},
		{
			match:   "SELECT COUNT(*)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{2}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/stats/:website_id", HandleDashboardStats, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats DashboardStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.Equal(t, 3, stats.CurrentVisitors)
	assert.Equal(t, 12, stats.TodayPageviews)
	assert.Equal(t, 6, stats.TodayVisitors)
	assert.Equal(t, "33.3%", stats.TodayBounceRate)

	require.NoError(t, queue.expectationsMet())
}

func TestHandleDashboardStats_InvalidWebsiteID(t *testing.T) {
	app := fiber.New()
	app.Get("/api/dashboard/stats/:website_id", HandleDashboardStats)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleDashboardStats_QueryErrors(t *testing.T) {
	websiteID := uuid.New()

	responses := []mockResponse{
		{
			match: "SELECT COUNT(DISTINCT e.session_id)",
			args:  []interface{}{websiteID},
			err:   assert.AnError,
		},
		{
			match:   "SELECT COUNT(*)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{0}},
		},
		{
			match:   "SELECT COUNT(DISTINCT e.session_id)",
			args:    []interface{}{websiteID},
			columns: []string{"count"},
			rows:    [][]interface{}{{0}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/stats/:website_id", HandleDashboardStats, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats DashboardStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.Equal(t, 0, stats.CurrentVisitors)
	assert.Equal(t, 0, stats.TodayVisitors)

	require.NoError(t, queue.expectationsMet())
}
