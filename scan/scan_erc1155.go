package scan

import (
	"reflect"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/abis"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ERC1155Scanner struct {
	address common.Address
	abi     *abi.ABI
	types   map[common.Hash]reflect.Type // event.ID => EventType
	logger  *logrus.Entry
}

func NewERC1155Scanner(address common.Address, logger *logrus.Logger) (*ERC1155Scanner, error) {
	aBI, err := gov.BmErc1155MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	types := map[common.Hash]reflect.Type{
		aBI.Events["TransferSingle"].ID: reflect.TypeOf(BmErc1155TransferSingle{}),
		aBI.Events["TransferBatch"].ID:  reflect.TypeOf(BmErc1155TransferBatch{}),
	}
	if _, ok := types[common.Hash{}]; ok {
		return nil, errors.Wrap(ErrInvalidEventID, "ERC1155Scanner")
	}

	return &ERC1155Scanner{address, aBI, types, logger.WithField("scanner", "ERC1155Scanner")}, nil
}

func (s *ERC1155Scanner) Address() common.Address {
	return s.address
}

func (s *ERC1155Scanner) Topics() []common.Hash {
	ids := make([]common.Hash, len(s.types))
	index := 0
	for id := range s.types {
		ids[index] = id
		index++
	}
	return ids
}

func (s *ERC1155Scanner) Save(db *gorm.DB, log types.Log) error {
	logger := s.logger.WithField("log", log)
	logger.Debug("Save")

	logger = logger.WithField("method", "Save")
	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		logger.Error(ErrNoEventSignature)
		return nil
	}
	out, err := parse(log, s.types[log.Topics[0]], s.abi)
	if err != nil {
		logger.Error(err)
		return nil
	}

	return errors.Wrap(out.Create(db, log), "ERC1155Scanner")
}

type BmErc1155TransferSingle gov.BmErc1155TransferSingle

func (event *BmErc1155TransferSingle) Create(db *gorm.DB, log types.Log) error {
	record := &dbtypes.ERC1155Transfer{
		Raw: dbtypes.Raw{
			TxHash: log.TxHash,
			Block:  log.BlockNumber,
		},
		Index:    0,
		Operator: event.Operator,
		From:     event.From,
		To:       event.To,
		Id:       (*dbtypes.BigInt)(event.Id),
		Value:    (*dbtypes.BigInt)(event.Value),
	}
	return db.Create(record).Error
}

type BmErc1155TransferBatch gov.BmErc1155TransferBatch

func (event *BmErc1155TransferBatch) Create(db *gorm.DB, log types.Log) error {
	length := len(event.Ids)
	for i := 0; i < length; i++ {
		record := &dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: log.TxHash,
				Block:  log.BlockNumber,
			},
			Index:    i,
			Operator: event.Operator,
			From:     event.From,
			To:       event.To,
			Id:       (*dbtypes.BigInt)(event.Ids[i]),
			Value:    (*dbtypes.BigInt)(event.Values[i]),
		}
		if err := db.Create(record).Error; err != nil {
			return err
		}
	}
	return nil
}
