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

func TestHandleTopReferrers_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COALESCE(e.referrer_domain, 'Direct / None') as name",
			args:    []interface{}{websiteID, 10},
			columns: []string{"name", "count"},
			rows:    [][]interface{}{{"example.com", 12}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/referrers/:website_id", HandleTopReferrers, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/referrers/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var items []BreakdownItem
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&items))
	assert.Len(t, items, 1)
	assert.Equal(t, "example.com", items[0].Name)

	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopReferrers_Filtered(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COALESCE(e.referrer_domain, 'Direct / None') as name",
			args:    []interface{}{websiteID, "US", "Chrome", "mobile", 5},
			columns: []string{"name", "count"},
			rows:    [][]interface{}{{"example.com", 5}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/referrers/:website_id", HandleTopReferrers, responses)
	defer cleanup()

	url := "/api/dashboard/referrers/" + websiteID.String() + "?limit=5&country=US&browser=Chrome&device=mobile"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopBrowsers_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COALESCE(s.browser, 'Unknown') as name",
			args:    []interface{}{websiteID, 10},
			columns: []string{"name", "count"},
			rows:    [][]interface{}{{"Chrome", 20}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/browsers/:website_id", HandleTopBrowsers, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/browsers/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopDevices_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COALESCE(s.device, 'Unknown') as name",
			args:    []interface{}{websiteID, 10},
			columns: []string{"name", "count"},
			rows:    [][]interface{}{{"mobile", 8}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/devices/:website_id", HandleTopDevices, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/devices/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}

func TestHandleTopCountries_Success(t *testing.T) {
	websiteID := uuid.New()
	responses := []mockResponse{
		{
			match:   "SELECT COALESCE(s.country, 'Unknown') as name",
			args:    []interface{}{websiteID, 10},
			columns: []string{"name", "count"},
			rows:    [][]interface{}{{"United States", 15}},
		},
	}

	app, queue, cleanup := setupFiberTest(t, "/api/dashboard/countries/:website_id", HandleTopCountries, responses)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/countries/"+websiteID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, queue.expectationsMet())
}

func TestBreakdownHandlers_InvalidWebsiteID(t *testing.T) {
	type invalidCase struct {
		route   string
		handler fiber.Handler
	}
	cases := []invalidCase{
		{"/api/dashboard/referrers/:website_id", HandleTopReferrers},
		{"/api/dashboard/browsers/:website_id", HandleTopBrowsers},
		{"/api/dashboard/devices/:website_id", HandleTopDevices},
		{"/api/dashboard/countries/:website_id", HandleTopCountries},
	}

	for _, tc := range cases {
		app := fiber.New()
		app.Get(tc.route, tc.handler)

		req := httptest.NewRequest(http.MethodGet, tc.route[:len(tc.route)-len(":website_id")]+"invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_ = resp.Body.Close()
	}
}
