package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		HTTP     HTTPConfig
		Auth     AuthConfig
		Limiter  LimiterConfig
		Jwt      JWTConfig
		Postgres PostgresConfig
		Redis    RedisConfig
	}

	HTTPConfig struct {
		Host               string        `env:"APP_HOST, default=0.0.0.0"`
		Port               string        `env:"APP_PORT, default=8080"`
		ReadTimeout        time.Duration `env:"APP_READ_TIMEOUT, default=10s"`
		WriteTimeout       time.Duration `env:"APP_WRITE_TIMEOUT, default=10s"`
		MaxHeaderMegabytes int           `env:"APP_MAX_HEADER_MEGABYTES, default=1"`
	}

	AuthConfig struct {
		JWT                    JWTConfig
		PasswordSalt           string
		VerificationCodeLength int `env:"AUTH_VERIFICATION_CODE_LENGTH, default=6"`
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration `env:"accessTokenTTL, default=1h"`
		RefreshTokenTTL time.Duration `env:"refreshTokenTTL, default=72h"`
		SigningKey      string        `env:"signingKey, default=secret"`
	}

	LimiterConfig struct {
		RPS   int           `env:"RPS, default=1000"`
		Burst int           `env:"BURST, default=1000"`
		TTL   time.Duration `env:"TTL, default=1h"`
	}

	PostgresConfig struct {
		Host     string `env:"POSTGRES_HOST, default=localhost"`
		Port     string `env:"POSTGRES_PORT, default=5432"`
		User     string `env:"POSTGRES_USER, default=postgres"`
		Password string `env:"POSTGRES_PASSWORD, default="`
		DBName   string `env:"POSTGRES_DB, default=postgres"`
	}

	RedisConfig struct {
		Host     string `env:"REDIS_HOST, default=localhost"`
		Port     string `env:"REDIS_PORT, default=6379"`
		Password string `env:"REDIS_PASSWORD, default="`
	}

	SMTPConfig struct {
		Host string `env:"SMTP_HOST, default=smtp.gmail.com"`
		Port int    `env:"SMTP_PORT, default=587"`
		From string `env:"SMTP_FROM, default="`
		Pass string `env:"SMTP_PASS, default="`
	}
)

func Init() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	return &cfg, nil
}
