package config

// NewTestConfig creates a mock configuration for testing purposes.
func NewTestConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "Test API",
			Version:     "1.0.0",
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
		Health: HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: true,
		},
	}
}
