package models

import "time"

// Customer represents a customer entity
type Customer struct {
	CustomerID       int64
	ClientCustomerID int64
	InsertDate       time.Time
}

// CustomerData represents customer channel information (email, phone, etc.)
type CustomerData struct {
	CustomerChannelID int64
	CustomerID        int64
	ChannelTypeID     int16
	ChannelValue      string
	InsertDate        time.Time
}

// CustomerEvent represents a customer event
type CustomerEvent struct {
	EventID       int64
	ClientEventID int64
	InsertDate    time.Time
}

// CustomerEventData represents the details of a customer event
type CustomerEventData struct {
	EventDataID  int64
	EventID      int64
	ContentID    int32
	CustomerID   int64
	EventTypeID  int16
	EventDate    time.Time
	Quantity     int16
	InsertDate   time.Time
}

// Content represents a product/content
type Content struct {
	ContentID       int32
	ClientContentID int64
	InsertDate      time.Time
}

// ContentPrice represents the price of a content/product
type ContentPrice struct {
	ContentPriceID int32
	ContentID      int32
	Price          float64
	Currency       string
	InsertDate     time.Time
}

// ChannelType represents the type of communication channel
type ChannelType struct {
	ChannelTypeID int16
	Name          string
}

// EventType represents the type of event
type EventType struct {
	EventTypeID int16
	Name        string
}

// CustomerRevenue holds the calculated revenue for a customer
type CustomerRevenue struct {
	CustomerID int64
	Email      string
	Revenue    float64
}

// QuantileStats holds statistics for a revenue quantile
type QuantileStats struct {
	QuantileIndex int
	CustomerCount int
	MaxRevenue    float64
	MinRevenue    float64
}