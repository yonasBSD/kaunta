package database

import (
	"fmt"
	"log"
	"time"
)

// PartitionScheduler manages automatic partition creation and cleanup
type PartitionScheduler struct {
	databaseURL string
	stopChan    chan struct{}
}

// NewPartitionScheduler creates a new partition scheduler
func NewPartitionScheduler(databaseURL string) *PartitionScheduler {
	return &PartitionScheduler{
		databaseURL: databaseURL,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the partition management tasks
func (ps *PartitionScheduler) Start() {
	log.Println("🗓️  Starting partition scheduler...")

	// Create future partitions daily at 2 AM
	go ps.schedulePartitionCreation()

	// Clean up old partitions weekly
	go ps.schedulePartitionCleanup()
}

// Stop gracefully stops the scheduler
func (ps *PartitionScheduler) Stop() {
	close(ps.stopChan)
}

// schedulePartitionCreation creates partitions 30 days in advance
func (ps *PartitionScheduler) schedulePartitionCreation() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run immediately on start
	ps.createFuturePartitions()

	for {
		select {
		case <-ticker.C:
			ps.createFuturePartitions()
		case <-ps.stopChan:
			return
		}
	}
}

// createFuturePartitions creates partitions for the next 30 days
func (ps *PartitionScheduler) createFuturePartitions() {
	log.Println("Creating future partitions...")

	for i := 1; i <= 30; i++ {
		date := time.Now().AddDate(0, 0, i)
		partitionName := fmt.Sprintf("website_event_%s", date.Format("2006_01_02"))
		startDate := date.Format("2006-01-02")
		endDate := date.AddDate(0, 0, 1).Format("2006-01-02")

		query := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s
			PARTITION OF website_event
			FOR VALUES FROM ('%s') TO ('%s')
		`, partitionName, startDate, endDate)

		_, err := DB.Exec(query)
		if err != nil {
			log.Printf("⚠️  Failed to create partition %s: %v", partitionName, err)
			continue
		}

		log.Printf("✓ Created partition: %s", partitionName)
	}
}

// schedulePartitionCleanup removes partitions older than 90 days
func (ps *PartitionScheduler) schedulePartitionCleanup() {
	ticker := time.NewTicker(7 * 24 * time.Hour) // Weekly
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ps.cleanupOldPartitions()
		case <-ps.stopChan:
			return
		}
	}
}

// cleanupOldPartitions drops partitions older than retention period
func (ps *PartitionScheduler) cleanupOldPartitions() {
	retentionDays := 90 // Keep 90 days of data
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	log.Printf("Cleaning up partitions older than %s...", cutoffDate.Format("2006-01-02"))

	// Find old partitions
	rows, err := DB.Query(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename LIKE 'website_event_%'
		  AND tablename < $1
		ORDER BY tablename
	`, fmt.Sprintf("website_event_%s", cutoffDate.Format("2006_01_02")))

	if err != nil {
		log.Printf("⚠️  Failed to query old partitions: %v", err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	droppedCount := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// Drop old partition
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		_, err := DB.Exec(query)
		if err != nil {
			log.Printf("⚠️  Failed to drop partition %s: %v", tableName, err)
			continue
		}

		log.Printf("🗑️  Dropped old partition: %s", tableName)
		droppedCount++
	}

	if droppedCount > 0 {
		log.Printf("✓ Cleaned up %d old partitions", droppedCount)
	}
}

// MaterializedViewScheduler manages concurrent refreshes
type MaterializedViewScheduler struct {
	stopChan chan struct{}
}

// NewMaterializedViewScheduler creates a new refresh scheduler
func NewMaterializedViewScheduler() *MaterializedViewScheduler {
	return &MaterializedViewScheduler{
		stopChan: make(chan struct{}),
	}
}

// Start begins the refresh tasks
func (mvs *MaterializedViewScheduler) Start() {
	log.Println("🔄 Starting materialized view refresh scheduler...")

	// Real-time stats: every minute
	go mvs.scheduleRefresh("realtime_website_stats", 1*time.Minute)

	// Hourly stats: every 5 minutes
	go mvs.scheduleRefresh("hourly_website_stats", 5*time.Minute)

	// Daily stats: every hour
	go mvs.scheduleRefresh("daily_website_stats", 1*time.Hour)
}

// Stop gracefully stops the scheduler
func (mvs *MaterializedViewScheduler) Stop() {
	close(mvs.stopChan)
}

// scheduleRefresh refreshes a materialized view at the specified interval
func (mvs *MaterializedViewScheduler) scheduleRefresh(viewName string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial refresh on startup
	mvs.refreshView(viewName)

	for {
		select {
		case <-ticker.C:
			mvs.refreshView(viewName)
		case <-mvs.stopChan:
			return
		}
	}
}

// refreshView performs a concurrent refresh of the materialized view
func (mvs *MaterializedViewScheduler) refreshView(viewName string) {
	start := time.Now()

	query := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)
	_, err := DB.Exec(query)

	duration := time.Since(start)

	if err != nil {
		log.Printf("⚠️  Failed to refresh %s: %v", viewName, err)
		return
	}

	log.Printf("✓ Refreshed %s in %v", viewName, duration)
}

// GetMaterializedViewStats returns refresh statistics
func GetMaterializedViewStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Query view sizes
	rows, err := DB.Query(`
		SELECT
			schemaname || '.' || matviewname as view_name,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size,
			last_refresh
		FROM pg_matviews
		WHERE schemaname = 'public'
		ORDER BY matviewname
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query matview stats: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	views := []map[string]interface{}{}
	for rows.Next() {
		var viewName, size string
		var lastRefresh *time.Time

		if err := rows.Scan(&viewName, &size, &lastRefresh); err != nil {
			continue
		}

		viewInfo := map[string]interface{}{
			"name": viewName,
			"size": size,
		}

		if lastRefresh != nil {
			viewInfo["last_refresh"] = lastRefresh.Format(time.RFC3339)
			viewInfo["age"] = time.Since(*lastRefresh).String()
		}

		views = append(views, viewInfo)
	}

	stats["views"] = views
	return stats, nil
}
