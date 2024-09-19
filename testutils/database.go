package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSQLMock(t *testing.T) *gorm.DB {
	// SQLite 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
