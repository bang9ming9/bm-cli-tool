package scan

// import (
// 	"reflect"

// 	gov "github.com/bang9ming9/bm-governance/abis"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"gorm.io/gorm"
// )

// type ERC1155Scanner struct {
// 	address common.Address
// 	abi     *abi.ABI
// 	topics  map[common.Hash]struct {
// 		name string
// 		out  reflect.Type
// 	}
// }

// func NewERC1155Scanner(address common.Address) (*ERC1155Scanner, error) {
// 	aBI, err := gov.BmErc1155MetaData.GetAbi()
// 	if err != nil {
// 		return nil, err
// 	}
// 	topics := map[common.Hash]struct {
// 		name string
// 		out  reflect.Type
// 	}{
// 		aBI.Events["TransferBatch"].ID: {"TransferBatch", reflect.TypeOf(BmErc1155TransferBatch{})},
// 	}
// 	return &ERC1155Scanner{address, aBI, topics}, nil
// }

// func (s *ERC1155Scanner) Address() common.Address {
// 	return s.address
// }

// func (s *ERC1155Scanner) Topics() []common.Hash {
// 	ids := make([]common.Hash, len(s.topics))
// 	index := 0
// 	for id := range s.topics {
// 		ids[index] = id
// 		index++
// 	}
// 	return ids
// }

// func (s *ERC1155Scanner) Parse(log types.Log) (ICreate, error) {
// 	// Anonymous events are not supported.
// 	if len(log.Topics) == 0 {
// 		return nil, ErrNoEventSignature
// 	}
// 	topic, ok := s.topics[log.Topics[0]]
// 	if !ok {
// 		return nil, ErrEventSignatureMismatch
// 	}
// 	out := reflect.New(topic.out).Interface()
// 	event := topic.name
// 	if len(log.Data) > 0 {
// 		if err := s.abi.UnpackIntoInterface(&out, event, log.Data); err != nil {
// 			return nil, err
// 		}
// 	}
// 	var indexed abi.Arguments
// 	for _, arg := range s.abi.Events[event].Inputs {
// 		if arg.Indexed {
// 			indexed = append(indexed, arg)
// 		}
// 	}
// 	return out.(ICreate), abi.ParseTopics(out, indexed, log.Topics[1:])
// }

// type BmErc1155TransferBatch gov.BmErc1155TransferBatch

// func (event *BmErc1155TransferBatch) Create(db *gorm.DB) error {
// 	record := &ERC1155Transfer{
// 		Raw: Raw{
// 			TxHash: event.Raw.TxHash,
// 			Block:  event.Raw.BlockNumber,
// 		},
// 		From:   event.From,
// 		To:     event.To,
// 		Ids:    event.Ids,
// 		Values: event.Values,
// 	}
// 	return db.Create(record).Error
// }
