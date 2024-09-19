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

type ERC20Scanner struct {
	Scanner
}

func NewERC20Scanner(address common.Address, logentry *logrus.Logger) (*ERC20Scanner, error) {
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

	return &ERC20Scanner{Scanner{
		address, aBI, types, logentry.WithField("scanner", "ERC20Scanner"),
	}}, nil
}

type BmErc20Transfer gov.BmErc20Transfer

func (event *BmErc20Transfer) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		record := &dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: log.TxHash,
				Block:  log.BlockNumber,
			},
			From:  event.From,
			To:    event.To,
			Value: (*dbtypes.BigInt)(event.Value),
		}
		return errors.Wrap(db.Create(record).Error, "BmErc20Transfer")
	}
}
