package dbtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type IRecord interface {
	Do(db *gorm.DB, log types.Log) error
}

type Raw struct {
	TxHash common.Hash `gorm:"primaryKey;column:tx_hash;type:char(32)"`
	Block  uint64      `gorm:"column:block_number"`
}

var (
	AllTables = []interface{}{
		&ERC20Transfer{},
		&ERC1155Transfer{},
		&FaucetClaimed{},
		&GovernorProposal{},
		&GovernorVoteCast{},
	}
)

type ERC20Transfer struct {
	Raw
	From  common.Address `gorm:"column:_from;type:char(20)"`
	To    common.Address `gorm:"column:_to;type:char(20)"`
	Value *BigInt        `gorm:"type:char(32)"`
}

type ERC1155Transfer struct {
	Raw
	Index    int            `gorm:"primaryKey"`
	Operator common.Address `gorm:"type:char(20)"`
	From     common.Address `gorm:"column:_from;type:char(20)"`
	To       common.Address `gorm:"column:_to;type:char(20)"`
	Id       *BigInt        `gorm:"type:char(32)"`
	Value    *BigInt        `gorm:"type:char(32)"`
}

type FaucetClaimed struct {
	Raw
	Account common.Address `gorm:"type:char(20)"`
}

type GovernorProposal struct {
	Raw
	Active      bool           // true:Pending,Active,Succeeded; false:Canceled,Defeated,Expired,Executed
	ProposalId  *BigInt        `gorm:"type:char(32)"`
	Proposer    common.Address `gorm:"type:char(20)"`
	Targets     *AddressList
	Values      *BigIntList
	Signatures  *StringList
	Calldatas   *BytesList
	VoteStart   uint64
	VoteEnd     uint64
	Description string `gorm:"size:2048"`
}

type GovernorVoteCast struct {
	Raw
	Voter      common.Address `gorm:"type:char(20)"`
	ProposalId *BigInt        `gorm:"type:char(32)"`
	Support    uint8          `gorm:"type:smallint"`
	Weight     *BigInt        `gorm:"type:char(32)"`
	Reason     string         `gorm:"size:1024"`
}
