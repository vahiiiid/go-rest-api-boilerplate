package db

import (
"testing"
"github.com/stretchr/testify/assert"
)

func TestNewSQLiteDB(t *testing.T) {
db, err := NewSQLiteDB(":memory:")
assert.NoError(t, err)
assert.NotNil(t, db)
}

func TestLoadConfigFromEnv(t *testing.T) {
cfg := LoadConfigFromEnv()
assert.NotNil(t, cfg)
}
