package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Quantile   float64
	SkipDB     bool
}


func LoadConfig() (*Config, error) {
	
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Note: .env file not found, using system environment variables")
	}

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "quanticfy_test"),
		Quantile:   0.025, // Default quantile value (2.5%)
		SkipDB:     getEnvBool("SKIP_DB", false),
	}

	if !config.SkipDB {
		if config.DBUser == "" {
			return nil, fmt.Errorf("DB_USER environment variable is required (check your .env file)")
		}
		if config.DBPassword == "" {
			return nil, fmt.Errorf("DB_PASSWORD environment variable is required (check your .env file)")
		}
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if val == "" {
		return defaultValue
	}
	switch val {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}
