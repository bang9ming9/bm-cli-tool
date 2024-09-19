package logtypes

import (
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func LogToBson(log types.Log) bson.D {
	bnh, bnl := SplitUint64(log.BlockNumber)
	return bson.D{
		{Key: "raw", Value: bson.D{
			{Key: "block_number_high", Value: bnh},
			{Key: "block_number_low", Value: bnl},
			{Key: "block_hash", Value: log.BlockHash},
			{Key: "index", Value: int64(log.Index)},
			{Key: "tx_hash", Value: log.TxHash},
			{Key: "tx_index", Value: int64(log.TxIndex)},
		}},
		{Key: "address", Value: log.Address},
		{Key: "topics", Value: log.Topics},
		{Key: "data", Value: log.Data},
		{Key: "removed", Value: log.Removed},
	}
}

func LogsToBson(logs []types.Log) []interface{} {
	length := len(logs)
	if length == 0 {
		return nil
	}

	list := make([]interface{}, length)
	for i, log := range logs {
		list[i] = LogToBson(log)
	}
	return list
}

func LogFromBsonM(data bson.M) types.Log {
	log := new(types.Log)

	if rawData, ok := data["raw"]; ok {
		if raw, ok := rawData.(bson.M); ok {
			var high, low int64 = 0, 0
			if data, ok := raw["block_number_high"]; ok {
				high, _ = data.(int64)
			}
			if data, ok := raw["block_number_low"]; ok {
				low, _ = data.(int64)
			}
			log.BlockNumber = JoinUint64(high, low)

			if data, ok := raw["block_hash"]; ok {
				if binary, ok := data.(primitive.Binary); ok {
					log.BlockHash = common.BytesToHash(binary.Data)
				}
			}
			if data, ok := raw["index"]; ok {
				if number, ok := data.(int64); ok {
					log.Index = uint(number)
				}
			}
			if data, ok := raw["tx_hash"]; ok {
				if binary, ok := data.(primitive.Binary); ok {
					log.TxHash = common.BytesToHash(binary.Data)
				}
			}
			if data, ok := raw["tx_index"]; ok {
				if number, ok := data.(int64); ok {
					log.TxIndex = uint(number)
				}
			}
		}
	}
	if data, ok := data["address"]; ok {
		if binary, ok := data.(primitive.Binary); ok {
			log.Address = common.BytesToAddress(binary.Data)
		}
	}
	if data, ok := data["topics"]; ok {
		if array, ok := data.(primitive.A); ok {
			topics := make([]common.Hash, len(array))
			for i, a := range array {
				if binary, ok := a.(primitive.Binary); ok {
					topics[i] = common.BytesToHash(binary.Data)
				}
			}
			log.Topics = topics
		}
	}
	if data, ok := data["data"]; ok {
		if binary, ok := data.(primitive.Binary); ok {
			log.Data = binary.Data
		}
	}
	if removed, ok := data["removed"]; ok {
		log.Removed, _ = removed.(bool)
	}
	return *log
}

func LogToProtobuf(log types.Log) *logger.Log {
	topics := make([][]byte, len(log.Topics))
	for i, topic := range log.Topics {
		topics[i] = topic[:]
	}
	return &logger.Log{
		Raw: &logger.Log_Raw{
			BlockNumber: log.BlockNumber,
			BlockHash:   log.BlockHash[:],
			Index:       uint32(log.Index),
			TxHash:      log.TxHash[:],
			TxIndex:     uint32(log.TxIndex),
		},
		Address: log.Address[:],
		Topics:  topics,
		Data:    log.Data,
		Removed: log.Removed,
	}
}

func LogFromProtobuf(log *logger.Log) types.Log {
	topics := make([]common.Hash, len(log.Topics))
	for i, topic := range log.Topics {
		topics[i] = common.BytesToHash(topic)
	}
	return types.Log{
		BlockNumber: log.Raw.BlockNumber,
		BlockHash:   common.BytesToHash(log.Raw.BlockHash),
		Index:       uint(log.Raw.Index),
		TxHash:      common.BytesToHash(log.Raw.TxHash),
		TxIndex:     uint(log.Raw.TxIndex),
		Address:     common.BytesToAddress(log.Address),
		Topics:      topics,
		Data:        log.Data,
		Removed:     log.Removed,
	}
}

func SplitUint64(x uint64) (high int64, low int64) {
	high = int64(x >> 32)
	low = int64(x & 0xFFFFFFFF)
	return
}

func JoinUint64(high, low int64) uint64 {
	return (uint64(high) << 32) | uint64(low)
}
