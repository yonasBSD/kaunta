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
			match:   "SELECT website_id, domain, name",
			columns: []string{"website_id", "domain", "name"},
			rows: [][]interface{}{
				{"id-1", "example.com", "Example"},
				{"id-2", "demo.com", nil},
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

	var websites []Website
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&websites))
	assert.Len(t, websites, 2)
	assert.Equal(t, "Example", websites[0].Name)
	assert.Equal(t, "demo.com", websites[1].Name) // falls back to domain

	require.NoError(t, queue.expectationsMet())
}

func TestHandleWebsites_QueryError(t *testing.T) {
	responses := []mockResponse{
		{
			match: "SELECT website_id, domain, name",
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
