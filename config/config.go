package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	ERPDatabase DatabaseConfig `mapstructure:"erp_database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Excel       ExcelConfig    `mapstructure:"excel"`
	Logger      LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	User     string        `mapstructure:"user"`
	Password string        `mapstructure:"password"`
	DBName   string        `mapstructure:"name"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiryHour int    `mapstructure:"expiry_hour"`
}

type ExcelConfig struct {
	DownloadPath    string `mapstructure:"download_path"`
	MaxSearchMonths int    `mapstructure:"max_search_months"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "."
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.timeout", 10)
	viper.SetDefault("jwt.expiry_hour", 24)
	viper.SetDefault("excel.max_search_months", 6)
	viper.SetDefault("excel.download_path", "public/downloads")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}

func MustConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Fatal error loading config: %s", err)
	}
	return cfg
}

// GetDSN returns SQL Server connection string
func (c *Config) GetDSN() string {
	// Format: sqlserver://username:password@host:port?database=dbname
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustServerCertificate=true",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
	)
}

func (c *Config) GetERPDatabaseDSN() string {
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustServerCertificate=true&connection timeout=%d",
		c.ERPDatabase.User,
		c.ERPDatabase.Password,
		c.ERPDatabase.Host,
		c.ERPDatabase.Port,
		c.ERPDatabase.DBName,
		c.ERPDatabase.Timeout,
	)
}

// GetJWTExpiry returns JWT expiry duration
func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWT.ExpiryHour) * time.Hour
}
