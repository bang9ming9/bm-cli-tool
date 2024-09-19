package logtypes_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/bang9ming9/bm-cli-tool/eventlogger/logtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBSON(t *testing.T) {
	bytes1 := []byte("1")
	log := types.Log{
		Address:     common.BytesToAddress(bytes1),
		Topics:      []common.Hash{common.BytesToHash(bytes1)},
		Data:        bytes1,
		BlockNumber: math.MaxUint64,
		TxHash:      common.BytesToHash(bytes1),
		TxIndex:     1,
		BlockHash:   common.BytesToHash(bytes1),
		Index:       2,
		Removed:     true,
	}
	bytes, err := bson.Marshal(logtypes.LogToBson(log))
	require.NoError(t, err)

	data := bson.M{}
	require.NoError(t, bson.Unmarshal(bytes, &data))
	raw, ok := data["raw"].(bson.M)
	require.True(t, ok)
	{ // raw
		_, ok = raw["block_number_high"].(int64)
		require.True(t, ok)
		_, ok = raw["block_number_low"].(int64)
		require.True(t, ok)
		_, ok = raw["block_hash"].(primitive.Binary)
		require.True(t, ok)
		_, ok = raw["index"].(int64)
		require.True(t, ok)
		_, ok = raw["tx_hash"].(primitive.Binary)
		require.True(t, ok)
		_, ok = raw["tx_index"].(int64)
		require.True(t, ok)
	}
	_, ok = data["address"].(primitive.Binary)
	require.True(t, ok)
	topics, ok := data["topics"].(primitive.A)
	require.True(t, ok)
	for _, topic := range topics {
		_, ok := topic.(primitive.Binary)
		require.True(t, ok)
	}
	_, ok = data["data"].(primitive.Binary)
	require.True(t, ok)
	_, ok = data["removed"].(bool)
	require.True(t, ok)

	require.Equal(t, log, logtypes.LogFromBsonM(data))
	require.True(t, reflect.DeepEqual(log, logtypes.LogFromBsonM(data)))
}

func TestProtobuf(t *testing.T) {
	bytes1 := []byte("1")
	log := types.Log{
		Address:     common.BytesToAddress(bytes1),
		Topics:      []common.Hash{common.BytesToHash(bytes1)},
		Data:        bytes1,
		BlockNumber: 1,
		TxHash:      common.BytesToHash(bytes1),
		TxIndex:     1,
		BlockHash:   common.BytesToHash(bytes1),
		Index:       1,
		Removed:     true,
	}

	to := logtypes.LogToProtobuf(log)
	from := logtypes.LogFromProtobuf(to)

	require.Equal(t, log, from)
	require.True(t, reflect.DeepEqual(log, from))
}
