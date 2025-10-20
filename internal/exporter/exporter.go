package exporter

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"quanticfy-test/internal/models"

	"github.com/schollz/progressbar/v3"
)

type Exporter struct {
	db *sql.DB
}

func NewExporter(db *sql.DB) *Exporter {
	return &Exporter{db: db}
}

// ExportTopCustomers exports top customers to a date-specific table
// Table structure: CustomerID # Email # CA
func (e *Exporter) ExportTopCustomers(
	topCustomers map[int64]*models.CustomerRevenue,
) error {

	log.Println("[INFO] Exporting top customers to database...")
	startTime := time.Now()

	// Generate table name with current date: test_export_YYYYMMDD
	tableName := fmt.Sprintf("test_export_%s", time.Now().Format("20060102"))

	// Create or verify table exists
	if err := e.createExportTable(tableName); err != nil {
		return fmt.Errorf("error creating export table: %w", err)
	}

	// Convert map to slice for processing
	customers := make([]*models.CustomerRevenue, 0, len(topCustomers))
	for _, rev := range topCustomers {
		customers = append(customers, rev)
	}

	if len(customers) == 0 {
		log.Println("[WARNING] No customers to export")
		return nil
	}

	// Mass insert using batch INSERT statements
	if err := e.massInsertCustomers(tableName, customers); err != nil {
		return fmt.Errorf("error inserting customers: %w", err)
	}

	log.Printf("[INFO] Successfully exported %d customers to table '%s' in %v",
		len(customers), tableName, time.Since(startTime))

	return nil
}

// createExportTable creates the export table if it doesn't exist
func (e *Exporter) createExportTable(tableName string) error {
	log.Printf("[INFO] Creating/verifying table '%s'...", tableName)

	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			CustomerID BIGINT UNSIGNED NOT NULL,
			Email VARCHAR(600) NOT NULL,
			CA DECIMAL(12,2) NOT NULL,
			InsertDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UpdateDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (CustomerID),
			INDEX idx_ca (CA DESC)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, tableName)

	_, err := e.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	log.Printf("[INFO] Table '%s' ready", tableName)
	return nil
}

// massInsertCustomers performs batch insert with ON DUPLICATE KEY UPDATE
func (e *Exporter) massInsertCustomers(
	tableName string,
	customers []*models.CustomerRevenue,
) error {

	batchSize := 1000
	totalBatches := (len(customers) + batchSize - 1) / batchSize

	log.Printf("[INFO] Inserting %d customers in %d batches...", len(customers), totalBatches)
	bar := progressbar.Default(int64(len(customers)), "Exporting")

	for i := 0; i < len(customers); i += batchSize {
		end := i + batchSize
		if end > len(customers) {
			end = len(customers)
		}

		batch := customers[i:end]

		// Build VALUES clause
		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*3)

		for _, customer := range batch {
			valueStrings = append(valueStrings, "(?, ?, ?)")
			valueArgs = append(valueArgs, customer.CustomerID, customer.Email, customer.Revenue)
		}

		// Build INSERT statement with ON DUPLICATE KEY UPDATE
		query := fmt.Sprintf(`
			INSERT INTO %s (CustomerID, Email, CA)
			VALUES %s
			ON DUPLICATE KEY UPDATE
				CA = VALUES(CA),
				Email = VALUES(Email)
		`, tableName, strings.Join(valueStrings, ","))

		// Execute batch insert
		_, err := e.db.Exec(query, valueArgs...)
		if err != nil {
			return fmt.Errorf("error executing batch insert: %w", err)
		}

		bar.Add(len(batch))
	}

	fmt.Println()
	return nil
}

// GetExportStats returns statistics about the exported data
func (e *Exporter) GetExportStats(tableName string) error {
	log.Printf("[INFO] Export statistics for table '%s':", tableName)

	var count int
	var totalRevenue float64
	var avgRevenue float64
	var maxRevenue float64
	var minRevenue float64

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as count,
			SUM(CA) as total_revenue,
			AVG(CA) as avg_revenue,
			MAX(CA) as max_revenue,
			MIN(CA) as min_revenue
		FROM %s
	`, tableName)

	err := e.db.QueryRow(query).Scan(&count, &totalRevenue, &avgRevenue, &maxRevenue, &minRevenue)
	if err != nil {
		return fmt.Errorf("error getting export stats: %w", err)
	}

	log.Printf("  Total Customers: %d", count)
	log.Printf("  Total Revenue: %.2f", totalRevenue)
	log.Printf("  Average Revenue: %.2f", avgRevenue)
	log.Printf("  Max Revenue: %.2f", maxRevenue)
	log.Printf("  Min Revenue: %.2f", minRevenue)

	return nil
}