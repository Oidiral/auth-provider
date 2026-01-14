package config

import (
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	EnvLocal = "local"
	Prod     = "prod"
)

type (
	Config struct {
		Environment string
		Tracing     TracingConfig
		AppName     string `env:"APP_NAME" envDefault:"auth-provider"`
		HTTP        HTTPConfig
		Auth        AuthConfig
		Limiter     LimiterConfig
		Jwt         JWTConfig
		Postgres    PostgresConfig
		Redis       RedisConfig
	}

	HTTPConfig struct {
		Host               string        `env:"APP_HOST" envDefault:"0.0.0.0"`
		Port               string        `env:"APP_PORT" envDefault:"8080"`
		ReadTimeout        time.Duration `env:"APP_READ_TIMEOUT" envDefault:"10s"`
		WriteTimeout       time.Duration `env:"APP_WRITE_TIMEOUT" envDefault:"10s"`
		MaxHeaderMegabytes int           `env:"APP_MAX_HEADER_MEGABYTES" envDefault:"1"`
	}

	AuthConfig struct {
		JWT                    JWTConfig
		PasswordSalt           string        `env:"AUTH_PASSWORD_SALT" envDefault:"salt"`
		VerificationCodeLength int           `env:"AUTH_VERIFICATION_CODE_LENGTH" envDefault:"6"`
		ExpiresSessionTTL      time.Duration `env:"AUTH_EXPIRES_SESSION_TTL" envDefault:"1h"`
		ExpiresOtpTTL          time.Duration `env:"AUTH_EXPIRES_OTP_TTL" envDefault:"1h"`
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TOKEN_TTL" envDefault:"1h"`
		RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TOKEN_TTL" envDefault:"72h"`
		SigningKey      string        `env:"JWT_SIGNING_KEY" envDefault:"secret"`
	}

	LimiterConfig struct {
		RPS   int           `env:"LIMITER_RPS" envDefault:"1000"`
		Burst int           `env:"LIMITER_BURST" envDefault:"1000"`
		TTL   time.Duration `env:"LIMITER_TTL" envDefault:"1h"`
	}

	PostgresConfig struct {
		Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
		Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
		User     string `env:"POSTGRES_USER" envDefault:"postgres"`
		Password string `env:"POSTGRES_PASSWORD" envDefault:""`
		DBName   string `env:"POSTGRES_DB" envDefault:"postgres"`
	}

	RedisConfig struct {
		Host     string `env:"REDIS_HOST" envDefault:"localhost"`
		Port     string `env:"REDIS_PORT" envDefault:"6379"`
		Password string `env:"REDIS_PASSWORD" envDefault:""`
	}

	SMTPConfig struct {
		Host string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
		Port int    `env:"SMTP_PORT" envDefault:"587"`
		From string `env:"SMTP_FROM" envDefault:""`
		Pass string `env:"SMTP_PASS" envDefault:""`
	}

	TracingConfig struct {
		Endpoint string `env:"TRACING_ENDPOINT" envDefault:"localhost:4317"`
	}
)

func Init(log logger.Logger) (*Config, error) {
	log.Info("loading configuration")

	if err := godotenv.Load(); err != nil {
		log.Warn(".env file not found or failed to load", logger.Field{Key: "error", Value: err.Error()})
	}

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Error("failed to parse environment variables", err)
		return nil, err
	}

	log.Info("configuration loaded successfully")
	return &cfg, nil
}
