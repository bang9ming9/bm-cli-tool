package dbtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
)

type BigInt big.Int

func (b *BigInt) Get() *big.Int {
	return (*big.Int)(b)
}

func (b *BigInt) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BigInt", src)
	}
	(*big.Int)(b).SetBytes(bytes)
	return nil
}

func (b *BigInt) Value() (driver.Value, error) {
	return b.Get().Bytes(), nil
}

func (b *BigInt) UnmarshalJSON(input []byte) error {
	var val string
	json.Unmarshal(input, &val)
	value, err := hexutil.DecodeBig(val)
	if value != nil {
		(*big.Int)(b).Set(value)
	}
	return err
}

func (b *BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(hexutil.EncodeBig((*big.Int)(b)))
}

type BigIntList []*big.Int

func (list *BigIntList) Get() []*big.Int {
	return ([]*big.Int)(*list)
}

func (list *BigIntList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BigIntList", src)
	}
	return rlp.DecodeBytes(bytes, &list)
}

func (list *BigIntList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(list.Get())
}

func (list *BigIntList) UnmarshalJSON(input []byte) error {
	var vals []string
	if err := json.Unmarshal(input, &vals); err != nil {
		return err
	}

	*list = make(BigIntList, len(vals))

	// Convert each string to *big.Int and store it in the list
	for i, str := range vals {
		bigInt, err := hexutil.DecodeBig(str)
		if err != nil {
			return err
		}
		(*list)[i] = bigInt
	}
	return nil
}

func (list *BigIntList) MarshalJSON() ([]byte, error) {
	val := make([]string, len(*list))
	for i, b := range *list {
		val[i] = hexutil.EncodeBig(b)
	}
	return json.Marshal(val)
}

type AddressList []common.Address

func (list *AddressList) Get() []common.Address {
	return ([]common.Address)(*list)
}

func (list *AddressList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into AddressList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *AddressList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(list.Get())
}

type HashList []common.Hash

func (list *HashList) Get() []common.Hash {
	return ([]common.Hash)(*list)
}

func (list *HashList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into HashList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *HashList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(list.Get())
}

type StringList []string

func (list *StringList) Get() []string {
	return ([]string)(*list)
}

func (list *StringList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into StringList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *StringList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(list.Get())
}

type BytesList [][]byte

func (list *BytesList) Get() [][]byte {
	return ([][]byte)(*list)
}

func (list *BytesList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BytesList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *BytesList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(list.Get())
}
