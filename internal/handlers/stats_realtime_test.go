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

func TestHandleCurrentVisitors_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COUNT(DISTINCT session_id)",
			args:    []interface{}{websiteID},
			columns: []string{"value"},
			rows: [][]interface{}{
				{7},
			},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/stats/realtime/:website_id", HandleCurrentVisitors, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/stats/realtime/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var payload map[string]int
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	assert.Equal(t, 7, payload["value"])

	require.NoError(t, queue.expectationsMet())
}

func TestHandleCurrentVisitors_InvalidWebsiteID(t *testing.T) {
	app := fiber.New()
	app.Get("/api/stats/realtime/:website_id", HandleCurrentVisitors)

	req := httptest.NewRequest(http.MethodGet, "/api/stats/realtime/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleCurrentVisitors_QueryError(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match: "SELECT COUNT(DISTINCT session_id)",
			args:  []interface{}{websiteID},
			err:   assert.AnError,
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/stats/realtime/:website_id", HandleCurrentVisitors, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/stats/realtime/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}
