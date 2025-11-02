package db

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

func TestNewSQLiteDB(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "in-memory database",
			dbPath:  ":memory:",
			wantErr: false,
		},
		{
			name:    "file database",
			dbPath:  "test.db",
			wantErr: false,
		},
		{
			name:    "invalid sqlite path",
			dbPath:  "/nonexistent/path/to/file.db",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if tt.dbPath != ":memory:" && tt.dbPath != "/nonexistent/path/to/file.db" {
					_ = os.Remove(tt.dbPath)
				}
			}()

			db, err := NewSQLiteDB(tt.dbPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)

				sqlDB, err := db.DB()
				assert.NoError(t, err)
				assert.NoError(t, sqlDB.Ping())
			}
		})
	}
}

func TestNewPostgresDB(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config but no database server (expected failure)",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to database",
		},
		{
			name: "invalid host",
			config: Config{
				Host:     "invalid-host-that-does-not-exist",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to database",
		},
		{
			name: "invalid port",
			config: Config{
				Host:     "localhost",
				Port:     99999,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDB(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestNewPostgresDBFromDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  config.DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config but no database server (expected failure)",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to postgres database",
		},
		{
			name: "invalid database name",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to postgres database",
		},
		{
			name: "empty host",
			config: config.DatabaseConfig{
				Host:     "",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			wantErr: true,
			errMsg:  "failed to connect to postgres database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDBFromDatabaseConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	originalValues := map[string]string{
		"database.host":     viper.GetString("database.host"),
		"database.port":     viper.GetString("database.port"),
		"database.user":     viper.GetString("database.user"),
		"database.password": viper.GetString("database.password"),
		"database.name":     viper.GetString("database.name"),
		"database.sslmode":  viper.GetString("database.sslmode"),
	}

	defer func() {
		for key, value := range originalValues {
			viper.Set(key, value)
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]interface{}
		expected Config
	}{
		{
			name: "default configuration",
			envVars: map[string]interface{}{
				"database.host":     "localhost",
				"database.port":     5432,
				"database.user":     "postgres",
				"database.password": "password",
				"database.name":     "testdb",
				"database.sslmode":  "disable",
			},
			expected: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Name:     "testdb",
				SSLMode:  "disable",
			},
		},
		{
			name: "production-like configuration",
			envVars: map[string]interface{}{
				"database.host":     "prod-db.example.com",
				"database.port":     5432,
				"database.user":     "produser",
				"database.password": "securepassword",
				"database.name":     "proddb",
				"database.sslmode":  "require",
			},
			expected: Config{
				Host:     "prod-db.example.com",
				Port:     5432,
				User:     "produser",
				Password: "securepassword",
				Name:     "proddb",
				SSLMode:  "require",
			},
		},
		{
			name: "empty values",
			envVars: map[string]interface{}{
				"database.host":     "",
				"database.port":     0,
				"database.user":     "",
				"database.password": "",
				"database.name":     "",
				"database.sslmode":  "",
			},
			expected: Config{
				Host:     "",
				Port:     0,
				User:     "",
				Password: "",
				Name:     "",
				SSLMode:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				viper.Set(key, value)
			}

			cfg := LoadConfigFromEnv()
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		isValid bool
	}{
		{
			name: "valid production config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Name:     "mydb",
				SSLMode:  "disable",
			},
			isValid: true,
		},
		{
			name: "missing required fields",
			config: Config{
				Host:     "",
				Port:     0,
				User:     "",
				Password: "",
				Name:     "",
				SSLMode:  "",
			},
			isValid: false,
		},
		{
			name: "invalid port range",
			config: Config{
				Host:     "localhost",
				Port:     70000, // Invalid port
				User:     "postgres",
				Password: "password",
				Name:     "mydb",
				SSLMode:  "disable",
			},
			isValid: false,
		},
		{
			name: "negative port",
			config: Config{
				Host:     "localhost",
				Port:     -1,
				User:     "postgres",
				Password: "password",
				Name:     "mydb",
				SSLMode:  "disable",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Host != "" &&
				tt.config.Port > 0 &&
				tt.config.Port <= 65535 &&
				tt.config.User != "" &&
				tt.config.Name != ""

			assert.Equal(t, tt.isValid, isValid)
		})
	}
}
