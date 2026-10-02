package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App   AppConfig   `env-prefix:"APP_"`
	HTTP  HTTPConfig  `env-prefix:"HTTP_"`
	DB    DBConfig    `env-prefix:"DB_"`
	Auth  AuthConfig  `env-prefix:"AUTH_"`
	Media MediaConfig `env-prefix:"MEDIA_"`
}

type AppConfig struct {
	Env string `env:"ENV" env-default:"prod"`
}

type HTTPConfig struct {
	Address           string        `env:"ADDRESS" env-default:"0.0.0.0"`
	Port              string        `env:"PORT" env-default:"8081"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" env-default:"15s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" env-default:"15s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" env-default:"5s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     string `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-default:"tutorina"`
	Password string `env:"PASSWORD" env-default:"tutorina"`
	Name     string `env:"NAME" env-default:"tutorina"`
	SSLMode  string `env:"SSL_MODE" env-default:"disable"`
}

type AuthConfig struct {
	SessionTTL   time.Duration `env:"SESSION_TTL" env-default:"24h"`
	CookieSecure bool          `env:"COOKIE_SECURE" env-default:"true"`
}

type MediaConfig struct {
	TeacherPhotosPath string `env:"TEACHER_PHOTOS_PATH" env-default:"./data/teacher-photos"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
