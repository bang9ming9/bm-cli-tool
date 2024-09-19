package scan

import (
	"fmt"
	"reflect"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/abis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ERC1155Scanner struct {
	Scanner
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

	return &ERC1155Scanner{Scanner{
		address, aBI, types, logger.WithField("scanner", "ERC1155Scanner"),
	}}, nil
}

type BmErc1155TransferSingle gov.BmErc1155TransferSingle

func (event *BmErc1155TransferSingle) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
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
		return errors.Wrap(db.Create(record).Error, "BmErc1155TransferSingle")
	}
}

type BmErc1155TransferBatch gov.BmErc1155TransferBatch

func (event *BmErc1155TransferBatch) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
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
				return errors.Wrap(db.Create(record).Error, fmt.Sprintf("BmErc1155TransferBatch[%d]", i))
			}
		}
		return nil
	}
}
