package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"github.com/seuros/kaunta/internal/database"
	"github.com/seuros/kaunta/internal/geoip"
	"github.com/seuros/kaunta/internal/handlers"
)

//go:embed VERSION
var versionFile string

//go:embed assets/kaunta.min.js
var trackerScript []byte

//go:embed assets/vendor/alpine.min.js
var alpineJS []byte

//go:embed assets/vendor/chart.min.js
var chartJS []byte

//go:embed assets/vendor/leaflet-1.9.4.min.js
var leafletJS []byte

//go:embed assets/vendor/leaflet-1.9.4.min.css
var leafletCSS []byte

//go:embed assets/vendor/topojson-client-3.1.0.min.js
var topojsonJS []byte

//go:embed assets/data/countries-110m.json
var countriesGeoJSON []byte

//go:embed dashboard.html
var dashboardTemplate string

// Version of Kaunta (read from VERSION file at build time)
var Version string

func init() {
	Version = strings.TrimSpace(versionFile)
}

func main() {
	// Parse CLI flags
	version := flag.Bool("v", false, "Show version and exit")
	vLong := flag.Bool("version", false, "Show version and exit")
	help := flag.Bool("h", false, "Show help and exit")
	helpLong := flag.Bool("help", false, "Show help and exit")
	flag.Parse()

	// Handle version flag
	if *version || *vLong {
		fmt.Printf("Kaunta v%s\n", Version)
		fmt.Println("Analytics without bloat")
		os.Exit(0)
	}

	// Handle help flag
	if *help || *helpLong {
		fmt.Printf("Kaunta v%s - Analytics without bloat\n\n", Version)
		fmt.Println("Usage: kaunta [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  DATABASE_URL  PostgreSQL connection string (required)")
		fmt.Println("  PORT          Server port (default: 3000)")
		fmt.Println("  DATA_DIR      GeoIP database directory (default: ./data)")
		os.Exit(0)
	}

	// Get database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Run migrations
	log.Println("Running database migrations...")
	if err := database.RunMigrations(databaseURL); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	} else {
		log.Println("✓ Migrations completed")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize GeoIP database (downloads if missing)
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := geoip.Init(dataDir); err != nil {
		log.Fatalf("GeoIP initialization failed: %v", err)
	}
	defer func() {
		if err := geoip.Close(); err != nil {
			log.Printf("Error closing GeoIP: %v", err)
		}
	}()

	// Initialize template engine (using embedded template)
	// All templates are embedded in binary via //go:embed directives above
	// Fiber's html engine requires a directory, but we don't use it since we handle templates manually
	engine := html.New(".", ".html")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Kaunta - Analytics without bloat",
		Views:   engine,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, OPTIONS",
	}))

	// Add version header to all responses
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Kaunta-Version", Version)
		return c.Next()
	})

	// Static assets - serve embedded JS/CSS files
	app.Get("/assets/vendor/:filename<*>", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		// Strip query string if present
		if idx := strings.Index(filename, "?"); idx > -1 {
			filename = filename[:idx]
		}

		c.Set("Cache-Control", "public, max-age=31536000, immutable")
		c.Set("CF-Cache-Tag", "kaunta-assets")

		switch filename {
		case "alpine.min.js":
			c.Set("Content-Type", "application/javascript; charset=utf-8")
			return c.Send(alpineJS)
		case "chart.min.js":
			c.Set("Content-Type", "application/javascript; charset=utf-8")
			return c.Send(chartJS)
		case "leaflet-1.9.4.min.js":
			c.Set("Content-Type", "application/javascript; charset=utf-8")
			return c.Send(leafletJS)
		case "leaflet-1.9.4.min.css":
			c.Set("Content-Type", "text/css; charset=utf-8")
			return c.Send(leafletCSS)
		case "topojson-client-3.1.0.min.js":
			c.Set("Content-Type", "application/javascript; charset=utf-8")
			return c.Send(topojsonJS)
		default:
			return c.Status(404).SendString("Not found")
		}
	})

	// Static data files
	app.Get("/assets/data/:filename<*>", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		// Strip query string if present
		if idx := strings.Index(filename, "?"); idx > -1 {
			filename = filename[:idx]
		}

		c.Set("Cache-Control", "public, max-age=86400, immutable")
		c.Set("CF-Cache-Tag", "kaunta-data")

		switch filename {
		case "countries-110m.json":
			c.Set("Content-Type", "application/json; charset=utf-8")
			return c.Send(countriesGeoJSON)
		default:
			return c.Status(404).SendString("Not found")
		}
	})

	// Routes
	app.Get("/", handleIndex)
	app.Get("/health", handleHealth)
	app.Get("/up", handleUp) // Docker health check
	app.Get("/api/version", handleVersion)

	// Tracker script
	app.Get("/k.js", handleTrackerScript)
	app.Get("/kaunta.js", handleTrackerScript) // Long form
	app.Get("/script.js", handleTrackerScript) // Umami-compatible alias

	// Tracking API (Umami-compatible)
	app.Options("/api/send", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/api/send", handlers.HandleTracking)

	// Stats API (Plausible-inspired)
	app.Get("/api/stats/realtime/:website_id", handlers.HandleCurrentVisitors)

	// Dashboard UI (Alpine.js)
	app.Get("/dashboard", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		// Replace Go template variables in embedded HTML
		html := strings.ReplaceAll(dashboardTemplate, "{{.Title}}", "Kaunta Dashboard")
		html = strings.ReplaceAll(html, "{{.Version}}", Version)
		return c.SendString(html)
	})

	// Dashboard API endpoints
	app.Get("/api/websites", handlers.HandleWebsites)
	app.Get("/api/dashboard/stats/:website_id", handlers.HandleDashboardStats)
	app.Get("/api/dashboard/pages/:website_id", handlers.HandleTopPages)
	app.Get("/api/dashboard/timeseries/:website_id", handlers.HandleTimeSeries)
	app.Get("/api/dashboard/referrers/:website_id", handlers.HandleTopReferrers)
	app.Get("/api/dashboard/browsers/:website_id", handlers.HandleTopBrowsers)
	app.Get("/api/dashboard/devices/:website_id", handlers.HandleTopDevices)
	app.Get("/api/dashboard/countries/:website_id", handlers.HandleTopCountries)
	app.Get("/api/dashboard/cities/:website_id", handlers.HandleTopCities)
	app.Get("/api/dashboard/regions/:website_id", handlers.HandleTopRegions)
	app.Get("/api/dashboard/map/:website_id", handlers.HandleMapData)

	// Start server
	port := getEnv("PORT", "3000")
	log.Printf("Kaunta starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func handleIndex(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kaunta - Analytics without bloat</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            max-width: 800px;
            margin: 100px auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 { color: #2c3e50; }
        .subtitle { color: #7f8c8d; font-style: italic; }
        .status {
            background: #fff3cd;
            border: 1px solid #ffc107;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>Kaunta (カウンタ)</h1>
    <p class="subtitle">Analytics without bloat.</p>

    <div class="status">
        <strong>Status:</strong> Production<br>
        <strong>Memory:</strong> ~12MB (vs Umami's 200-500MB)<br>
        <strong>Size:</strong> Single Go binary (vs 300MB node_modules)<br>
        <strong>React:</strong> None (that's the point)
    </div>

    <h2>Features</h2>
    <ul>
        <li>Privacy-focused visitor tracking</li>
        <li>Minimal resource usage</li>
        <li>PostgreSQL backend</li>
        <li>Real-time analytics dashboard</li>
    </ul>

    <p style="text-align: center; margin-top: 40px;">
        <a href="https://github.com/seuros/kaunta">View on GitHub</a>
    </p>
</body>
</html>
	`)
}

func handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "kaunta",
		"react":   false,
		"bloat":   false,
	})
}

func handleUp(c *fiber.Ctx) error {
	// Simple Docker health check endpoint
	// Returns 200 OK if server is running and can connect to database
	if err := database.DB.Ping(); err != nil {
		return c.Status(503).SendString("database unavailable")
	}
	return c.SendStatus(200)
}

func handleVersion(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version": Version,
	})
}

func handleTrackerScript(c *fiber.Ctx) error {
	// Security headers
	c.Set("Content-Type", "application/javascript; charset=utf-8")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-XSS-Protection", "1; mode=block")

	// Cache headers (1 hour)
	c.Set("Cache-Control", "public, max-age=3600, immutable")
	c.Set("ETag", "\"kaunta-v1.0.0\"")

	// CORS headers - allow from anywhere (JS file is public)
	// Origin validation happens at /api/send endpoint
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Timing headers
	c.Set("Timing-Allow-Origin", "*")

	return c.Send(trackerScript)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
