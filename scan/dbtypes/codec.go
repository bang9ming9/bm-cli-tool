package dbtypes

import (
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type BigInt big.Int

func (b *BigInt) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BigInt", src)
	}
	(*big.Int)(b).SetBytes(bytes)
	return nil
}

func (b *BigInt) Value() (driver.Value, error) {
	return (*big.Int)(b).Bytes(), nil
}

func (b *BigInt) Get() *big.Int {
	return (*big.Int)(b)
}

type BigIntList []*big.Int

func (list *BigIntList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BigIntList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *BigIntList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(([]*big.Int)(*list))
}

func (list *BigIntList) Get() []*big.Int {
	return ([]*big.Int)(*list)
}

type AddressList []common.Address

func (list *AddressList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into AddressList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *AddressList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(([]common.Address)(*list))
}

func (list *AddressList) Get() []common.Address {
	return ([]common.Address)(*list)
}

type HashList []common.Hash

func (list *HashList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into HashList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *HashList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(([]common.Hash)(*list))
}

func (list *HashList) Get() []common.Hash {
	return ([]common.Hash)(*list)
}

type StringList []string

func (list *StringList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into StringList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *StringList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(([]string)(*list))
}

func (list *StringList) Get() []string {
	return ([]string)(*list)
}

type BytesList [][]byte

func (list *BytesList) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into BytesList", src)
	}
	rlp.DecodeBytes(bytes, &list)
	return nil
}

func (list *BytesList) Value() (driver.Value, error) {
	return rlp.EncodeToBytes(([][]byte)(*list))
}

func (list *BytesList) Get() [][]byte {
	return ([][]byte)(*list)
}
