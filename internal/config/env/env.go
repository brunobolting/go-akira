package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

var (
	APP_NAME                         string
	ENVIRONMENT                      string
	ISPROD                           bool
	PORT                             string
	DATABASE_DSN                     string
	SESSION_SECRET                   string
	LOGGER_TYPE                      string
	LOGGER_SENTRY_DSN                string
	LOGGER_SENTRY_TRACES_SAMPLE_RATE float64
	LOGGER_SENTRY_DEBUG              bool
	TURNSTILE_SITE_KEY               string
	TURNSTILE_SECRET_KEY             string
)

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	setenvs()
	return nil
}

func str(s string) string {
	return s
}

func num(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func flt64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func boolean(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func environment(s string) string {
	if s == DEV || s == PROD {
		return s
	}
	return DEV
}

func setenvs() {
	APP_NAME = getenv("APP_NAME", "akira", str)
	ENVIRONMENT = getenv("ENVIRONMENT", DEV, environment)
	ISPROD = ENVIRONMENT == PROD
	PORT = getenv("PORT", "8080", str)
	DATABASE_DSN = getenv("DATABASE_DSN", "db/app.db", str)
	LOGGER_TYPE = getenv("LOGGER_TYPE", "slog", str)
	LOGGER_SENTRY_DSN = getenv("LOGGER_SENTRY_DSN", "", str)
	LOGGER_SENTRY_TRACES_SAMPLE_RATE = getenv("LOGGER_SENTRY_TRACES_SAMPLE_RATE", 0.0, flt64)
	LOGGER_SENTRY_DEBUG = getenv("LOGGER_SENTRY_DEBUG", false, boolean)
	TURNSTILE_SITE_KEY = getenv("TURNSTILE_SITE_KEY", "", str)
	TURNSTILE_SECRET_KEY = getenv("TURNSTILE_SECRET_KEY", "", str)
	SESSION_SECRET = getenv("SESSION_SECRET", "Uy@!DNv3@8iikzWNBqb24bFCWgi!FaBY", str)
}

func getenv[T any](key string, defaultValue T, parser func(string) T) T {
	if value, ok := os.LookupEnv(key); ok {
		return parser(value)
	}
	return defaultValue
}
