package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type Connection struct {
	DB *sql.DB
}


func NewConnection(config DBConfig) (*Connection, error) {
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	db.SetMaxOpenConns(25)                 
	db.SetMaxIdleConns(5)                  
	db.SetConnMaxLifetime(5 * time.Minute) 

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &Connection{DB: db}, nil
}

func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

func (c *Connection) HealthCheck() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return c.DB.Ping()
}