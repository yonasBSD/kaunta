package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/seuros/kaunta/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWebsites_Success(t *testing.T) {
	responses := []mockResponse{
		{
			match:   "SELECT w.website_id, w.domain, w.name, t.count as total_count",
			columns: []string{"website_id", "domain", "name", "total_count"},
			rows: [][]interface{}{
				{"id-1", "example.com", "Example", int64(2)},
				{"id-2", "demo.com", nil, int64(2)},
			},
		},
	}

	queue := newMockQueue(responses)
	driverName, err := registerMockDriver(queue)
	require.NoError(t, err)

	db, err := sql.Open(driverName, "")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	app := fiber.New()
	app.Get("/api/websites", HandleWebsites)

	req := httptest.NewRequest(http.MethodGet, "/api/websites", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedResp PaginatedResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&paginatedResp))

	// Extract websites from data field
	websitesJSON, err := json.Marshal(paginatedResp.Data)
	require.NoError(t, err)
	var websites []Website
	require.NoError(t, json.Unmarshal(websitesJSON, &websites))

	assert.Len(t, websites, 2)
	assert.Equal(t, "Example", websites[0].Name)
	assert.Equal(t, "demo.com", websites[1].Name) // falls back to domain

	// Check pagination metadata
	assert.Equal(t, int64(2), paginatedResp.Pagination.Total)
	assert.Equal(t, 1, paginatedResp.Pagination.Page)
	assert.Equal(t, 10, paginatedResp.Pagination.Per)
	assert.False(t, paginatedResp.Pagination.HasMore)

	require.NoError(t, queue.expectationsMet())
}

func TestHandleWebsites_QueryError(t *testing.T) {
	responses := []mockResponse{
		{
			match: "SELECT w.website_id, w.domain, w.name, t.count as total_count",
			err:   assert.AnError,
		},
	}

	queue := newMockQueue(responses)
	driverName, err := registerMockDriver(queue)
	require.NoError(t, err)

	db, err := sql.Open(driverName, "")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	app := fiber.New()
	app.Get("/api/websites", HandleWebsites)

	req := httptest.NewRequest(http.MethodGet, "/api/websites", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}
