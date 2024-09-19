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

type GovernorScanner struct {
	Scanner
}

func NewGovernorScanner(address common.Address, logger *logrus.Logger) (*GovernorScanner, error) {
	aBI, err := gov.BmGovernorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	types := map[common.Hash]reflect.Type{
		aBI.Events["ProposalCreated"].ID:  reflect.TypeOf(BmGovernorProposalCreated{}),
		aBI.Events["ProposalCanceled"].ID: reflect.TypeOf(BmGovernorProposalCanceled{}),
		aBI.Events["ProposalExecuted"].ID: reflect.TypeOf(BmGovernorProposalExecuted{}),
	}
	if _, ok := types[common.Hash{}]; ok {
		return nil, errors.Wrap(ErrInvalidEventID, "GovernorScanner")
	}

	return &GovernorScanner{Scanner{
		address, aBI, types, logger.WithField("scanner", "GovernorScanner"),
	}}, nil
}

type BmGovernorProposalCreated gov.BmGovernorProposalCreated

func (event *BmGovernorProposalCreated) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		record := &dbtypes.GovernorProposal{
			Raw: dbtypes.Raw{
				TxHash: log.TxHash,
				Block:  log.BlockNumber,
			},
			Active:      true,
			ProposalId:  (*dbtypes.BigInt)(event.ProposalId),
			Proposer:    event.Proposer,
			Targets:     (*dbtypes.AddressList)(&event.Targets),
			Values:      (*dbtypes.BigIntList)(&event.Values),
			Signatures:  (*dbtypes.StringList)(&event.Signatures),
			Calldatas:   (*dbtypes.BytesList)(&event.Calldatas),
			VoteStart:   event.VoteStart.Uint64(),
			VoteEnd:     event.VoteEnd.Uint64(),
			Description: event.Description,
		}
		return errors.Wrap(db.Create(record).Error, "BmGovernorProposalCreated")
	}
}

type BmGovernorProposalCanceled gov.BmGovernorProposalCanceled

func (event *BmGovernorProposalCanceled) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		proposalID := (*dbtypes.BigInt)(event.ProposalId)
		err := db.Model(&dbtypes.GovernorProposal{}).
			Where("proposal_id = ?", proposalID).
			Update("active", false).
			Error
		return errors.Wrap(err, "BmGovernorProposalCanceled")
	}
}

type BmGovernorProposalExecuted gov.BmGovernorProposalExecuted

func (event *BmGovernorProposalExecuted) Do(log types.Log) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		proposalID := (*dbtypes.BigInt)(event.ProposalId)
		err := db.Model(&dbtypes.GovernorProposal{}).
			Where("proposal_id = ?", proposalID).
			Update("active", false).
			Error
		return errors.Wrap(err, "BmGovernorProposalExecuted")
	}
}
