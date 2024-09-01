package dbtypes_test

import (
	"math/big"
	"testing"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBigInt(t *testing.T) {
	type BigIntStruct struct {
		gorm.Model
		A *dbtypes.BigInt
		B *dbtypes.BigInt
	}

	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&BigIntStruct{}))

	uint256Max := new(big.Int).SetBytes(common.MaxHash[:])
	createData := &BigIntStruct{A: (*dbtypes.BigInt)(common.Big0), B: (*dbtypes.BigInt)(uint256Max)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(BigIntStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	require.Equal(t, createData.A, firstData.A)
	require.Equal(t, createData.B, firstData.B)
	t.Log("A", (*big.Int)(firstData.A).String())
	t.Log("B", (*big.Int)(firstData.B).String())
}

func testDB(t *testing.T) *gorm.DB {
	// SQLite 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
