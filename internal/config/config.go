package config

import (
	"log/slog"
	"os"
)

// envs
const (
	LogLevel = "LOG_LEVEL"

	Socket  = "SOCKET"
	NatsUrl = "NATS_URL"

	InitDB             = "INIT_DB"
	EnableHealthCheck  = "HEALTH_CHECK"
	EnableSaveConsumer = "SAVE_CONSUMER"

	ProcessorDefault  = "PROCESSOR_DEFAULT"
	ProcessorFallback = "PROCESSOR_FALLBACK"
)

type Config struct {
	Socket             string
	NatsUrl            string
	InitDB             bool
	EnableHealthCheck  bool
	EnableSaveConsumer bool
	DefaultUrl         string
	FallbackUrl        string
}

func New() *Config {
	cfg := &Config{
		Socket:      os.Getenv(Socket),
		NatsUrl:     os.Getenv(NatsUrl),
		DefaultUrl:  os.Getenv(ProcessorDefault),
		FallbackUrl: os.Getenv(ProcessorFallback),
	}

	cfg.InitDB = parseBoolConfig(os.Getenv(InitDB))
	cfg.EnableHealthCheck = parseBoolConfig(os.Getenv(EnableHealthCheck))
	cfg.EnableSaveConsumer = parseBoolConfig(os.Getenv(EnableSaveConsumer))

	logLevel := slog.LevelError
	logLevel.UnmarshalText([]byte(os.Getenv(LogLevel)))
	slog.SetLogLoggerLevel(logLevel)

	return cfg
}

func parseBoolConfig(value string) bool {
	switch value {
	case "ON":
		return true
	case "OFF":
		return false
	default:
		return false
	}
}
