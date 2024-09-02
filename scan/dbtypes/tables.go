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
	TxHash common.Hash `gorm:"type:char(32);primaryKey;column:tx_hash"`
	Block  uint64      `gorm:"column:block_number"`
}

var (
	AllTables = []interface{}{&ERC20Transfer{}, &ERC1155Transfer{}, &GovernorProposalCreated{}, &GovernorProposalCanceled{}}
)

type ERC20Transfer struct {
	Raw
	From  common.Address `gorm:"type:char(20)"`
	To    common.Address `gorm:"type:char(20)"`
	Value *BigInt        `gorm:"type:char(32)"`
}

type ERC1155Transfer struct {
	Raw
	Index    int            `gorm:"primaryKey"`
	Operator common.Address `gorm:"type:char(20)"`
	From     common.Address `gorm:"type:char(20)"`
	To       common.Address `gorm:"type:char(20)"`
	Id       *BigInt        `gorm:"type:char(32)"`
	Value    *BigInt        `gorm:"type:char(32)"`
}

type GovernorProposalCreated struct {
	Raw
	ProposalId  *BigInt        `gorm:"type:char(32)"`
	Proposer    common.Address `gorm:"type:char(20)"`
	Targets     *AddressList
	Values      *BigIntList
	Signatures  *StringList
	Calldatas   *BytesList
	VoteStart   *BigInt `gorm:"type:char(32)"`
	VoteEnd     *BigInt `gorm:"type:char(32)"`
	Description string
}

type GovernorProposalCanceled struct {
	Raw
	ProposalId *BigInt `gorm:"type:char(32)"`
}
