package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	MaxConns    int
	MinConns    int
	DatabaseURL string // ✅ NEW: Production support (Neon)
}

type AWSConfig struct {
	Region          string
	UseIAMRole      bool
	AccessKeyID     string
	SecretAccessKey string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type StorageConfig struct {
	BucketName       string
	PresignedExpiry  time.Duration
	MaxFileSize      int64
	AllowedMimeTypes []string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	RedisURL string // Production support
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8081"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ShutdownTimeout: parseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s")),
			ReadTimeout:     parseDuration(getEnv("READ_TIMEOUT", "15s")),
			WriteTimeout:    parseDuration(getEnv("WRITE_TIMEOUT", "15s")),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			DBName:      getEnv("DB_NAME", "file_storage"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			MaxConns:    parseInt(getEnv("DB_MAX_CONNS", "25")),
			MinConns:    parseInt(getEnv("DB_MIN_CONNS", "5")),
			DatabaseURL: getEnv("DATABASE_URL", ""), // ✅ NEW
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			UseIAMRole:      parseBool(getEnv("USE_IAM_ROLE", "false")),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
			RefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "7d")),
			Issuer:        getEnv("JWT_ISSUER", "file-storage-api"),
		},
		Storage: StorageConfig{
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
			PresignedExpiry: parseDuration(getEnv("PRESIGNED_URL_EXPIRY", "15m")),
			MaxFileSize:     parseInt64(getEnv("MAX_FILE_SIZE", "104857600")),
			AllowedMimeTypes: []string{
				"image/jpeg", "image/png", "image/gif", "image/webp",
				"application/pdf",
				"application/msword",
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				"application/vnd.ms-excel",
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				"text/plain", "text/csv",
				"audio/mpeg", "audio/mp3", "audio/wav", "audio/ogg", "audio/aac",
				"video/mp4", "video/webm", "video/quicktime", "video/x-msvideo",
			},
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       parseInt(getEnv("REDIS_DB", "0")),
			RedisURL: getEnv("REDIS_URL", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if c.Storage.BucketName == "" {
		return fmt.Errorf("S3_BUCKET_NAME is required")
	}
	if !c.AWS.UseIAMRole && (c.AWS.AccessKeyID == "" || c.AWS.SecretAccessKey == "") {
		return fmt.Errorf("AWS credentials required when not using IAM role")
	}
	return nil
}

// ✅ UPDATED: Supports DATABASE_URL for production
func (c *Config) GetDSN() string {
	if c.Database.DatabaseURL != "" {
		return c.Database.DatabaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	if c.Redis.RedisURL != "" {
		return c.Redis.RedisURL
	}
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func parseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}