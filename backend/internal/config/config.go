package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App            AppConfig
	DB             DBConfig
	Auth           AuthConfig
	S3             S3Config
	Log            LogConfig
	CORS           CORSConfig
	BootstrapAdmin BootstrapAdminConfig
}

type AppConfig struct {
	Port string
	Env  string
	Name string
}

type DBConfig struct {
	DSN      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
	Audience      string
}

type LogConfig struct {
	Level string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type S3Config struct {
	Endpoint      string
	PublicBaseURL string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
}

type BootstrapAdminConfig struct {
	Email    string
	Login    string
	Password string
}

func Load() (Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_ENV", "local")
	v.SetDefault("APP_NAME", "history-app-backend")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "history-db")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("AUTH_ACCESS_TTL", "15m")
	v.SetDefault("AUTH_REFRESH_TTL", "168h")
	v.SetDefault("AUTH_ISSUER", "history-app-backend")
	v.SetDefault("AUTH_AUDIENCE", "history-app-users")
	v.SetDefault("S3_ENDPOINT", "localhost:9000")
	v.SetDefault("S3_PUBLIC_BASE_URL", "http://localhost:9000/history-media")
	v.SetDefault("S3_ACCESS_KEY", "minioadmin")
	v.SetDefault("S3_SECRET_KEY", "minioadmin")
	v.SetDefault("S3_BUCKET", "history-media")
	v.SetDefault("S3_USE_SSL", false)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:*,http://127.0.0.1:*")

	accessTTL, err := parseDuration(v.GetString("AUTH_ACCESS_TTL"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := parseDuration(v.GetString("AUTH_REFRESH_TTL"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_REFRESH_TTL: %w", err)
	}

	cfg := Config{
		App: AppConfig{
			Port: v.GetString("APP_PORT"),
			Env:  v.GetString("APP_ENV"),
			Name: v.GetString("APP_NAME"),
		},
		DB: DBConfig{
			DSN:      v.GetString("DATABASE_URL"),
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
		Auth: AuthConfig{
			AccessSecret:  v.GetString("AUTH_ACCESS_SECRET"),
			RefreshSecret: v.GetString("AUTH_REFRESH_SECRET"),
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
			Issuer:        v.GetString("AUTH_ISSUER"),
			Audience:      v.GetString("AUTH_AUDIENCE"),
		},
		S3: S3Config{
			Endpoint:      v.GetString("S3_ENDPOINT"),
			PublicBaseURL: strings.TrimRight(v.GetString("S3_PUBLIC_BASE_URL"), "/"),
			AccessKey:     v.GetString("S3_ACCESS_KEY"),
			SecretKey:     v.GetString("S3_SECRET_KEY"),
			Bucket:        v.GetString("S3_BUCKET"),
			UseSSL:        v.GetBool("S3_USE_SSL"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitCSV(v.GetString("CORS_ALLOWED_ORIGINS")),
		},
		BootstrapAdmin: BootstrapAdminConfig{
			Email:    v.GetString("BOOTSTRAP_ADMIN_EMAIL"),
			Login:    v.GetString("BOOTSTRAP_ADMIN_LOGIN"),
			Password: v.GetString("BOOTSTRAP_ADMIN_PASSWORD"),
		},
	}

	if cfg.Auth.AccessSecret == "" {
		return Config{}, fmt.Errorf("AUTH_ACCESS_SECRET is required")
	}
	if cfg.Auth.RefreshSecret == "" {
		return Config{}, fmt.Errorf("AUTH_REFRESH_SECRET is required")
	}

	return cfg, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (c DBConfig) ConnString() string {
	if c.DSN != "" {
		return c.DSN
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func parseDuration(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("duration is empty")
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d, nil
	}
	hours, err := time.ParseDuration(raw + "h")
	if err != nil {
		return 0, err
	}
	return hours, nil
}
