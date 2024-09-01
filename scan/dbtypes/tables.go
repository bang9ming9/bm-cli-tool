package dbtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type ICreate interface {
	Create(db *gorm.DB) error
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
	Operator common.Address
	From     common.Address
	To       common.Address
	Ids      []*BigInt
	Values   []*BigInt
}

type GovernorProposalCreated struct {
	Raw
	ProposalId  *BigInt `gorm:"primaryKey"`
	Proposer    common.Address
	Targets     []common.Address
	Values      []*BigInt
	Signatures  []string
	Calldatas   [][]byte
	VoteStart   *BigInt
	VoteEnd     *BigInt
	Description string
}

type GovernorProposalCancel struct {
	Raw
	TxHash  common.Hash
	Block   uint64
	Targets []common.Address
	Values  []*BigInt
}
