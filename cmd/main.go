package main

import (
	"log"
	"time"

	"quanticfy-test/internal/config"
	"quanticfy-test/internal/database"
	"quanticfy-test/internal/exporter"
	"quanticfy-test/internal/loader"
	"quanticfy-test/internal/processor"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[INFO] ")

	log.Println("========================================")
	log.Println("Quanticfy Data Processing - Starting")
	log.Println("========================================")

	startTime := time.Now()

	log.Println("Step 1/5: Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully (DB: %s@%s:%s/%s, Quantile: %.1f%%)",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.Quantile*100)

	log.Println("\nStep 2/5: Connecting to database...")
	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
	}

	conn, err := database.NewConnection(dbConfig)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.SetPrefix("[INFO] ")
		log.Println("\nClosing database connection...")
		if err := conn.Close(); err != nil {
			log.SetPrefix("[WARNING] ")
			log.Printf("Error closing database: %v", err)
		} else {
			log.SetPrefix("[INFO] ")
			log.Println("Database connection closed successfully")
		}
	}()

	if err := conn.HealthCheck(); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connection established successfully")

	var version string
	err = conn.DB.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.SetPrefix("[WARNING] ")
		log.Printf("Could not query MySQL version: %v", err)
	} else {
		log.SetPrefix("[INFO] ")
		log.Printf("Connected to MySQL version: %s", version)
	}

	log.Println("\n========================================")
	log.Println("Step 3/5: LOAD Phase")
	log.Println("========================================")
	loadStartTime := time.Now()

	dataLoader := loader.NewLoader(conn.DB)

	customerEmails, err := dataLoader.LoadCustomerEmails()
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to load customer emails: %v", err)
	}

	contentPrices, err := dataLoader.LoadContentPrices()
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to load content prices: %v", err)
	}

	sinceDate := time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC)
	purchaseEvents, err := dataLoader.LoadPurchaseEvents(sinceDate)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to load purchase events: %v", err)
	}

	log.SetPrefix("[INFO] ")
	log.Printf("LOAD Phase completed in %v", time.Since(loadStartTime))

	log.Println("\n========================================")
	log.Println("Step 4/5: COMPUTE Phase")
	log.Println("========================================")
	computeStartTime := time.Now()

	proc := processor.NewProcessor(cfg.Quantile)

	revenueMap, err := proc.CalculateCustomerRevenue(purchaseEvents, contentPrices, customerEmails)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to calculate customer revenue: %v", err)
	}

	topCustomers, err := proc.GetTopQuantileCustomers(revenueMap)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to get top customers: %v", err)
	}

	quantileStats, err := proc.CalculateQuantileStats(revenueMap)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to calculate quantile stats: %v", err)
	}
	_ = quantileStats

	log.SetPrefix("[INFO] ")
	log.Printf("COMPUTE Phase completed in %v", time.Since(computeStartTime))

	log.Println("\n========================================")
	log.Println("Step 5/5: EXPORT Phase")
	log.Println("========================================")
	exportStartTime := time.Now()

	exp := exporter.NewExporter(conn.DB)

	err = exp.ExportTopCustomers(topCustomers)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to export top customers: %v", err)
	}

	tableName := time.Now().Format("20060102")
	err = exp.GetExportStats("test_export_" + tableName)
	if err != nil {
		log.SetPrefix("[WARNING] ")
		log.Printf("Could not get export stats: %v", err)
	}

	log.SetPrefix("[INFO] ")
	log.Printf("EXPORT Phase completed in %v", time.Since(exportStartTime))

	duration := time.Since(startTime)
	log.Println("\n========================================")
	log.Println("Summary")
	log.Println("========================================")
	log.Printf("Total customers processed: %d", len(revenueMap))
	log.Printf("Top customers (%.1f%%): %d", cfg.Quantile*100, len(topCustomers))
	log.Printf("Total execution time: %v", duration)
	log.Println("========================================")
	log.Println("Process completed successfully!")
	log.Println("========================================")
}