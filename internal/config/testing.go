package config

// NewTestConfig creates a mock configuration for testing purposes.
func NewTestConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "Test API",
			Environment: "test",
			Debug:       true,
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "test",
			Name:    "test_db",
			SSLMode: "disable",
		},
		JWT: JWTConfig{
			Secret:   "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
			TTLHours: 1,
		},
		Server: ServerConfig{
			Port: "8081",
		},
		Logging: LoggingConfig{
			Level: "debug",
		},
	}
}
