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

type FaucetScanner struct {
	address common.Address
	abi     *abi.ABI
	types   map[common.Hash]reflect.Type // event.ID => EventType
	logger  *logrus.Entry
}

func NewFaucetScanner(address common.Address, logger *logrus.Logger) (*FaucetScanner, error) {
	aBI, err := gov.FaucetMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	types := map[common.Hash]reflect.Type{
		aBI.Events["Claimed"].ID: reflect.TypeOf(FaucetClaimed{}),
	}
	if _, ok := types[common.Hash{}]; ok {
		return nil, errors.Wrap(ErrInvalidEventID, "FaucetScanner")
	}

	return &FaucetScanner{address, aBI, types, logger.WithField("scanner", "FaucetScanner")}, nil
}

func (s *FaucetScanner) Address() common.Address {
	return s.address
}

func (s *FaucetScanner) Topics() []common.Hash {
	ids := make([]common.Hash, len(s.types))
	index := 0
	for id := range s.types {
		ids[index] = id
		index++
	}
	return ids
}

func (s *FaucetScanner) Work(db *gorm.DB, log types.Log) error {
	logger := s.logger.WithField("log", log)
	logger.Debug("Work")

	logger = logger.WithField("method", "Work")
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

	return errors.Wrap(out.Do(db, log), "FaucetScanner")
}

type FaucetClaimed gov.FaucetClaimed

func (event *FaucetClaimed) Do(db *gorm.DB, log types.Log) error {
	record := &dbtypes.FaucetClaimed{
		Raw: dbtypes.Raw{
			TxHash: log.TxHash,
			Block:  log.BlockNumber,
		},
		Account: event.Account,
	}
	return errors.Wrap(db.Create(record).Error, "FaucetClaimed")
}
