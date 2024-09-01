package dbtypes

import (
	"database/sql/driver"
	"fmt"
	"math/big"
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

func (b *BigInt) Big() *big.Int {
	return (*big.Int)(b)
}
