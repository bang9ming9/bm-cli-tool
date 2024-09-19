package dbtypes_test

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/bang9ming9/bm-cli-tool/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBigInt(t *testing.T) {
	type BigIntStruct struct {
		gorm.Model
		A *dbtypes.BigInt
		B *dbtypes.BigInt
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&BigIntStruct{})
	require.NoError(t, db.AutoMigrate(&BigIntStruct{}))

	uint256Max := new(big.Int).SetBytes(common.MaxHash[:])
	createData := &BigIntStruct{A: (*dbtypes.BigInt)(common.Big0), B: (*dbtypes.BigInt)(uint256Max)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(BigIntStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	require.Equal(t, createData.A, firstData.A)
	require.Equal(t, createData.B, firstData.B)
}
func TestBigIntJSON(t *testing.T) {
	type BigIntStruct struct {
		A *dbtypes.BigInt
		B *dbtypes.BigInt
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&BigIntStruct{})
	require.NoError(t, db.AutoMigrate(&BigIntStruct{}))

	uint256Max := new(big.Int).SetBytes(common.MaxHash[:])
	createData := &BigIntStruct{A: (*dbtypes.BigInt)(common.Big0), B: (*dbtypes.BigInt)(uint256Max)}

	bytes, err := json.Marshal(createData)
	require.NoError(t, err)

	copyData := new(BigIntStruct)
	require.NoError(t, json.Unmarshal(bytes, copyData))
	require.True(t, reflect.DeepEqual(createData, copyData))
}

func TestBigIntList(t *testing.T) {
	type BigIntListStruct struct {
		gorm.Model
		List *dbtypes.BigIntList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&BigIntListStruct{})
	require.NoError(t, db.AutoMigrate(&BigIntListStruct{}))
	uint256Max := new(big.Int).SetBytes(common.MaxHash[:])

	list := []*big.Int{common.Big0, common.Big1, common.Big2, common.Big3, uint256Max}
	createData := &BigIntListStruct{List: (*dbtypes.BigIntList)(&list)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(BigIntListStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	results := firstData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.True(t, results[i].Cmp(l) == 0)
	}
}
func TestBigIntListJSON(t *testing.T) {
	type BigIntListStruct struct {
		gorm.Model
		List *dbtypes.BigIntList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&BigIntListStruct{})
	require.NoError(t, db.AutoMigrate(&BigIntListStruct{}))

	uint256Max := new(big.Int).SetBytes(common.MaxHash[:])
	list := []*big.Int{common.Big0, common.Big1, common.Big2, common.Big3, uint256Max}
	createData := &BigIntListStruct{List: (*dbtypes.BigIntList)(&list)}

	bytes, err := json.Marshal(createData)
	require.NoError(t, err)

	copyData := new(BigIntListStruct)
	json.Unmarshal(bytes, copyData)

	results := copyData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.True(t, results[i].Cmp(l) == 0)
	}
}

func TestAddressList(t *testing.T) {
	type AddressListStruct struct {
		gorm.Model
		List *dbtypes.AddressList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&AddressListStruct{})
	require.NoError(t, db.AutoMigrate(&AddressListStruct{}))

	list := []common.Address{common.HexToAddress("0x1"), common.HexToAddress("0x2"), common.HexToAddress("0x3")}
	createData := &AddressListStruct{List: (*dbtypes.AddressList)(&list)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(AddressListStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	results := firstData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.True(t, results[i].Cmp(l) == 0)
	}
}

func TestHashList(t *testing.T) {
	type HashListStruct struct {
		gorm.Model
		List *dbtypes.HashList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&HashListStruct{})
	require.NoError(t, db.AutoMigrate(&HashListStruct{}))

	list := []common.Hash{common.HexToHash("0x1"), common.HexToHash("0x2"), common.HexToHash("0x3")}
	createData := &HashListStruct{List: (*dbtypes.HashList)(&list)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(HashListStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	results := firstData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.True(t, results[i].Cmp(l) == 0)
	}
}

func TestStringList(t *testing.T) {
	type StringListStruct struct {
		gorm.Model
		List *dbtypes.StringList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&StringListStruct{})
	require.NoError(t, db.AutoMigrate(&StringListStruct{}))

	list := []string{"", "1", "0xabcd", "hello"}
	createData := &StringListStruct{List: (*dbtypes.StringList)(&list)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(StringListStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	results := firstData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.True(t, results[i] == l)
	}
}

func TestBytesList(t *testing.T) {
	type BytesListStruct struct {
		gorm.Model
		List *dbtypes.BytesList
	}

	db := testutils.NewSQLMock(t)
	defer db.Migrator().DropTable(&BytesListStruct{})
	require.NoError(t, db.AutoMigrate(&BytesListStruct{}))

	list := [][]byte{{}, []byte(""), []byte("1"), []byte("0xabcd"), []byte("hello")}
	createData := &BytesListStruct{List: (*dbtypes.BytesList)(&list)}
	require.NoError(t, db.Create(createData).Error)

	firstData := new(BytesListStruct)
	require.NoError(t, db.First(&firstData, createData.ID).Error)

	results := firstData.List.Get()
	require.Equal(t, len(list), len(results))
	for i, l := range list {
		require.Equal(t, results[i], l, i)
	}
}
