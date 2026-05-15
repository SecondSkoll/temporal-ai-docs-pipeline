package controlplane

import "os"

// Config holds runtime configuration for the control-plane service.
// Values are read from environment variables at startup; defaults apply where safe.
type Config struct {
	// TemporalHostPort is the gRPC address of the Temporal frontend.
	// Defaults to "localhost:7233".
	TemporalHostPort string

	// TemporalNamespace is the Temporal namespace to use.
	// Defaults to "default".
	TemporalNamespace string

	// PostgresDSN is the libpq-style connection string for the run ledger.
	// Example: "postgres://user:pass@localhost:5432/docs_pipeline"
	PostgresDSN string

	// HTTPAddr is the TCP address the HTTP server listens on.
	// Defaults to ":8080".
	HTTPAddr string
}

// LoadConfig reads configuration from environment variables.
// Missing required variables cause the program to start with empty strings;
// callers are responsible for validating before use.
func LoadConfig() Config {
	return Config{
		TemporalHostPort:  envOrDefault("TEMPORAL_HOST_PORT", "localhost:7233"),
		TemporalNamespace: envOrDefault("TEMPORAL_NAMESPACE", "default"),
		PostgresDSN:       os.Getenv("POSTGRES_DSN"),
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
