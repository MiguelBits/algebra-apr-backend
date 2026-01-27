package config

import (
	"algebra-apr-backend/internal/logger"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port             string    `mapstructure:"port"`
	Database         DBConfig  `mapstructure:"database"`
	LogLevel         string    `mapstructure:"log_level"`
	Networks         []Network `mapstructure:"networks"`
	APRUpdateMinutes int       `mapstructure:"apr_update_minutes"`
}

type DBConfig struct {
	Host        string `mapstructure:"host"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Name        string `mapstructure:"name"`
	Port        int    `mapstructure:"port"`
	DatabaseURL string `mapstructure:"database_url"` // For Heroku DATABASE_URL
}

type Network struct {
	Title                string `mapstructure:"title"`
	AnalyticsSubgraphURL string `mapstructure:"analytics_subgraph_url"`
	FarmingSubgraphURL   string `mapstructure:"subgraph_farming_url"`
	APIKey               string `mapstructure:"api_key"`
}

func (db *DBConfig) GetDSN() string {
	// Check for Heroku-style DATABASE_URL first
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return parseDatabaseURL(databaseURL)
	}
	if db.DatabaseURL != "" {
		return parseDatabaseURL(db.DatabaseURL)
	}
	
	// Fall back to individual env vars / config
	sslMode := "disable"
	if os.Getenv("DB_SSLMODE") != "" {
		sslMode = os.Getenv("DB_SSLMODE")
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		db.Host, db.User, db.Password, db.Name, db.Port, sslMode)
}

// parseDatabaseURL converts Heroku's DATABASE_URL to a GORM-compatible DSN
func parseDatabaseURL(databaseURL string) string {
	// Heroku format: postgres://user:password@host:port/database
	u, err := url.Parse(databaseURL)
	if err != nil {
		return databaseURL
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	user := u.User.Username()
	password, _ := u.User.Password()
	dbName := strings.TrimPrefix(u.Path, "/")

	// Heroku Postgres requires SSL
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		host, user, password, dbName, port)
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("json")

	// Allow environment variables with specific binding
	viper.AutomaticEnv()
	viper.BindEnv("port", "PORT") // Heroku assigns PORT dynamically
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.database_url", "DATABASE_URL") // Heroku Postgres URL

	if err := viper.ReadInConfig(); err != nil {
		logger.Logger.Warn("Config file not found, using defaults and environment variables", zap.Error(err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Override port from environment if set (Heroku requirement)
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	// Default port if still empty
	if config.Port == "" {
		config.Port = "8080"
	}

	return &config, nil
}
