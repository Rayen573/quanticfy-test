package processor

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"quanticfy-test/internal/models"

	"github.com/schollz/progressbar/v3"
)

type Processor struct {
	quantile float64
}

func NewProcessor(quantile float64) *Processor {
	return &Processor{quantile: quantile}
}

func (p *Processor) CalculateCustomerRevenue(
	events []models.CustomerEventData,
	prices map[int32]float64,
	emails map[int64]string,
) (map[int64]*models.CustomerRevenue, error) {

	log.Println("[INFO] Calculating customer revenues...")
	startTime := time.Now()

	revenueMap := make(map[int64]*models.CustomerRevenue)
	bar := progressbar.Default(int64(len(events)), "Processing events")

	for _, event := range events {
		price, exists := prices[event.ContentID]
		if !exists {
			price = 0
		}

		eventRevenue := float64(event.Quantity) * price

		if rev, exists := revenueMap[event.CustomerID]; exists {
			rev.Revenue += eventRevenue
		} else {
			email := emails[event.CustomerID]
			if email == "" {
				email = "no-email@unknown.com"
			}
			revenueMap[event.CustomerID] = &models.CustomerRevenue{
				CustomerID: event.CustomerID,
				Email:      email,
				Revenue:    eventRevenue,
			}
		}
		bar.Add(1)
	}

	fmt.Println()
	log.Printf("[INFO] Calculated revenue for %d customers in %v", len(revenueMap), time.Since(startTime))

	p.printRandomEntries(revenueMap, 10)

	return revenueMap, nil
}

func (p *Processor) printRandomEntries(revenueMap map[int64]*models.CustomerRevenue, count int) {
	log.Printf("[INFO] Printing %d random customer revenue entries:", count)

	customers := make([]*models.CustomerRevenue, 0, len(revenueMap))
	for _, rev := range revenueMap {
		customers = append(customers, rev)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(customers), func(i, j int) {
		customers[i], customers[j] = customers[j], customers[i]
	})

	printCount := count
	if len(customers) < count {
		printCount = len(customers)
	}

	for i := 0; i < printCount; i++ {
		log.Printf("  CustomerID: %d | Email: %s | Revenue: %.2f",
			customers[i].CustomerID,
			customers[i].Email,
			customers[i].Revenue)
	}
}

func (p *Processor) GetTopQuantileCustomers(
	revenueMap map[int64]*models.CustomerRevenue,
) (map[int64]*models.CustomerRevenue, error) {

	log.Printf("[INFO] Identifying top %.1f%% customers by revenue...", p.quantile*100)
	startTime := time.Now()

	customers := make([]*models.CustomerRevenue, 0, len(revenueMap))
	for _, rev := range revenueMap {
		customers = append(customers, rev)
	}

	sort.Slice(customers, func(i, j int) bool {
		return customers[i].Revenue > customers[j].Revenue
	})

	topCount := int(float64(len(customers)) * p.quantile)
	if topCount == 0 && len(customers) > 0 {
		topCount = 1 
	}

	topCustomers := make(map[int64]*models.CustomerRevenue)
	for i := 0; i < topCount; i++ {
		topCustomers[customers[i].CustomerID] = customers[i]
	}

	log.Printf("[INFO] Found %d top customers (top %.1f%%) in %v",
		len(topCustomers), p.quantile*100, time.Since(startTime))

	return topCustomers, nil
}

func (p *Processor) CalculateQuantileStats(
	revenueMap map[int64]*models.CustomerRevenue,
) ([]models.QuantileStats, error) {

	log.Printf("[INFO] Calculating quantile statistics (quantile=%.3f)...", p.quantile)
	startTime := time.Now()

	customers := make([]*models.CustomerRevenue, 0, len(revenueMap))
	for _, rev := range revenueMap {
		customers = append(customers, rev)
	}

	sort.Slice(customers, func(i, j int) bool {
		return customers[i].Revenue > customers[j].Revenue
	})

	numQuantiles := int(1.0 / p.quantile)
	customersPerQuantile := len(customers) / numQuantiles

	stats := make([]models.QuantileStats, 0, numQuantiles)

	for q := 0; q < numQuantiles; q++ {
		startIdx := q * customersPerQuantile
		endIdx := (q + 1) * customersPerQuantile

		if q == numQuantiles-1 {
			endIdx = len(customers)
		}

		if startIdx >= len(customers) {
			break
		}

		quantileCustomers := customers[startIdx:endIdx]
		if len(quantileCustomers) == 0 {
			continue
		}

		stat := models.QuantileStats{
			QuantileIndex: q,
			CustomerCount: len(quantileCustomers),
			MaxRevenue:    quantileCustomers[0].Revenue,
			MinRevenue:    quantileCustomers[len(quantileCustomers)-1].Revenue,
		}
		stats = append(stats, stat)
	}

	log.Printf("[INFO] Calculated %d quantile statistics in %v", len(stats), time.Since(startTime))

	log.Println("[INFO] Quantile Statistics:")
	for _, stat := range stats {
		log.Printf("  Quantile %d (%.1f%%-%.1f%%): %d customers | Max Revenue: %.2f | Min Revenue: %.2f",
			stat.QuantileIndex,
			float64(stat.QuantileIndex)*p.quantile*100,
			float64(stat.QuantileIndex+1)*p.quantile*100,
			stat.CustomerCount,
			stat.MaxRevenue,
			stat.MinRevenue)
	}

	return stats, nil
}