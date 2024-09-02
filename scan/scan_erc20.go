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

type ERC20Scanner struct {
	address common.Address
	abi     *abi.ABI
	types   map[common.Hash]reflect.Type // event.ID => EventType
	logger  *logrus.Entry
}

func NewERC20Scanner(address common.Address, logger *logrus.Logger) (*ERC20Scanner, error) {
	aBI, err := gov.BmErc20MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	types := map[common.Hash]reflect.Type{
		aBI.Events["Transfer"].ID: reflect.TypeOf(BmErc20Transfer{}),
	}
	if _, ok := types[common.Hash{}]; ok {
		return nil, errors.Wrap(ErrInvalidEventID, "ERC20Scanner")
	}

	return &ERC20Scanner{address, aBI, types, logger.WithField("scanner", "ERC20Scanner")}, nil
}

func (s *ERC20Scanner) Address() common.Address {
	return s.address
}

func (s *ERC20Scanner) Topics() []common.Hash {
	ids := make([]common.Hash, len(s.types))
	index := 0
	for id := range s.types {
		ids[index] = id
		index++
	}
	return ids
}

func (s *ERC20Scanner) Save(db *gorm.DB, log types.Log) error {
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

	return errors.Wrap(out.Create(db, log), "ERC20Scanner")
}

type BmErc20Transfer gov.BmErc20Transfer

func (event *BmErc20Transfer) Create(db *gorm.DB, log types.Log) error {
	record := &dbtypes.ERC20Transfer{
		Raw: dbtypes.Raw{
			TxHash: log.TxHash,
			Block:  log.BlockNumber,
		},
		From:  event.From,
		To:    event.To,
		Value: (*dbtypes.BigInt)(event.Value),
	}
	return db.Create(record).Error
}
