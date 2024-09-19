package scan

import (
	"reflect"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/abis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FaucetScanner struct {
	Scanner
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

	return &FaucetScanner{Scanner{
		address, aBI, types, logger.WithField("scanner", "FaucetScanner"),
	}}, nil
}

type FaucetClaimed gov.FaucetClaimed

func (event *FaucetClaimed) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		record := &dbtypes.FaucetClaimed{
			Raw: dbtypes.Raw{
				TxHash: log.TxHash,
				Block:  log.BlockNumber,
			},
			Account: event.Account,
		}
		return errors.Wrap(db.Create(record).Error, "FaucetClaimed")
	}
}
