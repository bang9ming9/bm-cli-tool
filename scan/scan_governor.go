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

type GovernorScanner struct {
	address common.Address
	abi     *abi.ABI
	types   map[common.Hash]reflect.Type // event.ID => EventType
	logger  *logrus.Entry
}

func NewGovernorScanner(address common.Address, logger *logrus.Logger) (*GovernorScanner, error) {
	aBI, err := gov.BmGovernorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	types := map[common.Hash]reflect.Type{
		aBI.Events["ProposalCreated"].ID:  reflect.TypeOf(BmGovernorProposalCreated{}),
		aBI.Events["ProposalCanceled"].ID: reflect.TypeOf(BmGovernorProposalCanceled{}),
	}
	if _, ok := types[common.Hash{}]; ok {
		return nil, errors.Wrap(ErrInvalidEventID, "GovernorScanner")
	}

	return &GovernorScanner{address, aBI, types, logger.WithField("scanner", "GovernorScanner")}, nil
}

func (s *GovernorScanner) Address() common.Address {
	return s.address
}

func (s *GovernorScanner) Topics() []common.Hash {
	ids := make([]common.Hash, len(s.types))
	index := 0
	for id := range s.types {
		ids[index] = id
		index++
	}
	return ids
}

func (s *GovernorScanner) Save(db *gorm.DB, log types.Log) error {
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

	return errors.Wrap(out.Create(db, log), "GovernorScanner")
}

type BmGovernorProposalCreated gov.BmGovernorProposalCreated

func (event *BmGovernorProposalCreated) Create(db *gorm.DB, log types.Log) error {
	record := &dbtypes.GovernorProposalCreated{
		Raw: dbtypes.Raw{
			TxHash: log.TxHash,
			Block:  log.BlockNumber,
		},
		ProposalId:  (*dbtypes.BigInt)(event.ProposalId),
		Proposer:    event.Proposer,
		Targets:     (*dbtypes.AddressList)(&event.Targets),
		Values:      (*dbtypes.BigIntList)(&event.Values),
		Signatures:  (*dbtypes.StringList)(&event.Signatures),
		Calldatas:   (*dbtypes.BytesList)(&event.Calldatas),
		VoteStart:   (*dbtypes.BigInt)(event.VoteStart),
		VoteEnd:     (*dbtypes.BigInt)(event.VoteEnd),
		Description: event.Description,
	}
	return db.Create(record).Error
}

type BmGovernorProposalCanceled gov.BmGovernorProposalCanceled

func (event *BmGovernorProposalCanceled) Create(db *gorm.DB, log types.Log) error {
	record := &dbtypes.GovernorProposalCanceled{
		Raw: dbtypes.Raw{
			TxHash: log.TxHash,
			Block:  log.BlockNumber,
		},
		ProposalId: (*dbtypes.BigInt)(event.ProposalId),
	}
	return db.Create(record).Error
}
