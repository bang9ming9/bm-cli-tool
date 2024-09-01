package dbtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type ICreate interface {
	Create(db *gorm.DB, log types.Log) error
}

type Raw struct {
	TxHash common.Hash `gorm:"primaryKey;column:tx_hash"`
	Block  uint64      `gorm:"column:block_number"`
}

type ERC20Transfer struct {
	Raw
	From  common.Address
	To    common.Address
	Value *BigInt
}

type ERC1155Transfer struct {
	Raw
	Index    int `gorm:"primaryKey"`
	Operator common.Address
	From     common.Address
	To       common.Address
	Id       *BigInt
	Value    *BigInt
}

type GovernorProposalCreated struct {
	Raw
	ProposalId  *BigInt
	Proposer    common.Address
	Targets     *AddressList
	Values      *BigIntList
	Signatures  *StringList
	Calldatas   *BytesList
	VoteStart   *BigInt
	VoteEnd     *BigInt
	Description string
}

type GovernorProposalCanceled struct {
	Raw
	ProposalId *BigInt
}
