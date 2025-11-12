package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleTopPages_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT * FROM get_top_pages(",
			columns: []string{"path", "views", "unique_visitors", "avg_engagement_time", "total_count"},
			rows: [][]interface{}{
				{"/", int64(42), int64(30), 45.2, int64(2)},
				{"/docs", int64(21), int64(15), 32.5, int64(2)},
			},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/pages/:website_id", HandleTopPages, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/pages/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	assert.Equal(t, http.StatusOK, resp.StatusCode, string(bodyBytes))

	var paginatedResp PaginatedResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&paginatedResp))

	pagesJSON, err := json.Marshal(paginatedResp.Data)
	require.NoError(t, err)
	var pages []TopPage
	require.NoError(t, json.Unmarshal(pagesJSON, &pages))

	assert.Len(t, pages, 2)
	assert.Equal(t, "/", pages[0].Path)
	assert.Equal(t, 42, pages[0].Views)
	assert.Equal(t, int64(2), paginatedResp.Pagination.Total)

	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopPages_WithFilters(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT * FROM get_top_pages(",
			columns: []string{"path", "views", "unique_visitors", "avg_engagement_time", "total_count"},
			rows: [][]interface{}{
				{"/docs", int64(12), int64(10), 40.0, int64(1)},
			},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/pages/:website_id", HandleTopPages, responses)
	defer cleanup()

	url := "/api/dashboard/pages/" + websiteID.String() + "?per=5&country=US&browser=Chrome&device=mobile&page=/docs"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paginatedResp PaginatedResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&paginatedResp))

	pagesJSON, err := json.Marshal(paginatedResp.Data)
	require.NoError(t, err)
	var pages []TopPage
	require.NoError(t, json.Unmarshal(pagesJSON, &pages))

	assert.Len(t, pages, 1)
	assert.Equal(t, "/docs", pages[0].Path)
	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopPages_InvalidWebsiteID(t *testing.T) {
	app := fiber.New()
	app.Get("/api/dashboard/pages/:website_id", HandleTopPages)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/pages/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleTopPages_QueryError(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match: "SELECT * FROM get_top_pages(",
			err:   assert.AnError,
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/pages/:website_id", HandleTopPages, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/pages/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}
