package loader

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"quanticfy-test/internal/models"

	"github.com/schollz/progressbar/v3"
)

type Loader struct {
	db *sql.DB
}

func NewLoader(db *sql.DB) *Loader {
	return &Loader{db: db}
}

// LoadCustomerEmails loads customer emails into a map
func (l *Loader) LoadCustomerEmails() (map[int64]string, error) {
	log.Println("[INFO] Loading customer emails...")
	startTime := time.Now()

	// Query to get customer emails (ChannelTypeID = 1 for email based on the schema)
	query := `
		SELECT cd.CustomerID, cd.ChannelValue 
		FROM CustomerData cd 
		WHERE cd.ChannelTypeID = 1
	`

	rows, err := l.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying customer emails: %w", err)
	}
	defer rows.Close()

	customerEmails := make(map[int64]string)
	count := 0

	for rows.Next() {
		var customerID int64
		var email string
		if err := rows.Scan(&customerID, &email); err != nil {
			return nil, fmt.Errorf("error scanning email row: %w", err)
		}
		customerEmails[customerID] = email
		count++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email rows: %w", err)
	}

	log.Printf("[INFO] Loaded %d customer emails in %v", count, time.Since(startTime))
	return customerEmails, nil
}

// LoadContentPrices loads all content prices into a map
func (l *Loader) LoadContentPrices() (map[int32]float64, error) {
	log.Println("[INFO] Loading content prices...")
	startTime := time.Now()

	query := `SELECT ContentID, Price FROM ContentPrice`

	rows, err := l.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying content prices: %w", err)
	}
	defer rows.Close()

	prices := make(map[int32]float64)
	count := 0

	for rows.Next() {
		var contentID int32
		var price float64
		if err := rows.Scan(&contentID, &price); err != nil {
			return nil, fmt.Errorf("error scanning price row: %w", err)
		}
		prices[contentID] = price
		count++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	log.Printf("[INFO] Loaded %d content prices in %v", count, time.Since(startTime))
	return prices, nil
}


func (l *Loader) LoadPurchaseEvents(sinceDate time.Time) ([]models.CustomerEventData, error) {
	log.Printf("[INFO] Loading purchase events since %s...", sinceDate.Format("2006-01-02"))
	startTime := time.Now()

	var totalCount int
	countQuery := `
		SELECT COUNT(*) 
		FROM CustomerEventData 
		WHERE EventTypeID = 6 AND EventDate >= ?
	`
	err := l.db.QueryRow(countQuery, sinceDate).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("error counting purchase events: %w", err)
	}

	log.Printf("[INFO] Found %d purchase events to load", totalCount)

	query := `
		SELECT EventDataID, EventID, ContentID, CustomerID, EventTypeID, 
		       EventDate, Quantity, InsertDate
		FROM CustomerEventData
		WHERE EventTypeID = 6 AND EventDate >= ?
	`

	rows, err := l.db.Query(query, sinceDate)
	if err != nil {
		return nil, fmt.Errorf("error querying purchase events: %w", err)
	}
	defer rows.Close()

	events := make([]models.CustomerEventData, 0, totalCount)
	bar := progressbar.Default(int64(totalCount), "Loading purchases")

	for rows.Next() {
		var event models.CustomerEventData
		if err := rows.Scan(
			&event.EventDataID,
			&event.EventID,
			&event.ContentID,
			&event.CustomerID,
			&event.EventTypeID,
			&event.EventDate,
			&event.Quantity,
			&event.InsertDate,
		); err != nil {
			return nil, fmt.Errorf("error scanning event row: %w", err)
		}
		events = append(events, event)
		bar.Add(1)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	fmt.Println()
	log.Printf("[INFO] Loaded %d purchase events in %v", len(events), time.Since(startTime))
	return events, nil
}