package scan

import (
	"context"
	"math/big"
	"os"
	"reflect"
	"sync"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	butils "github.com/bang9ming9/go-hardhat/bms/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ethClient interface {
	butils.Backend
}

func Scan(ctx context.Context, config ContractConfig, logger *logrus.Logger, stop chan os.Signal, db *gorm.DB, client butils.Backend, ticker <-chan uint64) error {
	logger.Info("Set Scanners...")

	scanners, err := func() ([]IScanner, error) {
		scanners := []IScanner{}
		zero := common.Address{}
		if config.ERC20 != zero {
			scanner, err := NewERC20Scanner(config.ERC20, logger)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.ERC1155 != zero {
			scanner, err := NewERC1155Scanner(config.ERC1155, logger)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.Faucet != zero {
			scanner, err := NewFaucetScanner(config.Faucet, logger)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.Governance != zero {
			scanner, err := NewGovernorScanner(config.Governance, logger)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		return scanners, nil
	}()
	if err != nil {
		return err
	}

	scannerMap, filterQuery, blockNumber := prepareFilter(db, scanners...)
	blockNumber = max(blockNumber+1, config.FromBlock)
	dblock := sync.Mutex{}

	logger.Info("Start Scan!")
	for {
		select {
		case <-stop:
			logger.Info("Stop Scanner")
			return nil
		case currentBlockNumber := <-ticker:
			if dblock.TryLock() {
				go func() {
					defer dblock.Unlock()
					for blockNumber < currentBlockNumber {
						logger.WithFields(logrus.Fields{
							"blockNumber": blockNumber,
						}).Trace("Scan")
						filterQuery.FromBlock.SetUint64(blockNumber)
						filterQuery.ToBlock.SetUint64(blockNumber)

						// 1. 이벤트 로그를 읽어온다.
						logs, err := client.FilterLogs(ctx, filterQuery)
						if err != nil {
							logger.WithField("message", err.Error()).Error("Fail to call filter logs")
							return
						}

						// 2. 발생한 이벤트를 DB 에 저장한다.
						// 같은 블럭 넘버의 이벤트를 한번에 저장하기 위해 트랜잭션&커밋으로 동작
						// Scanner 재 기동시, 마지막으로 저장된 블럭넘버 이후부터 읽을 수 있다.
						// 일괄로 저장하려다 보니 저장 과정에서 에러가 발생한다면 리턴(종료) 유도.
						tx := db.Begin()
						for _, l := range logs {
							scanner, ok := scannerMap[l.Address]
							if ok {
								if err = scanner.Work(tx, l); err != nil {
									logger.WithField("message", err.Error()).Error("Fail to save logs")
									break
								}
							}
						}
						if err == nil {
							err = tx.Commit().Error
						}
						if err != nil {
							tx.Rollback()
							logger.WithField("message", err.Error()).Error("Fail to commit logs")
							return
						}
						blockNumber++
					}
				}()
			}
		}
	}
}

func prepareFilter(db *gorm.DB, scanners ...IScanner) (map[common.Address]IScanner, ethereum.FilterQuery, uint64) {
	addresses := make([]common.Address, len(scanners))
	topics0 := make([]common.Hash, 0)

	scannerMap := make(map[common.Address]IScanner)
	for i, scanner := range scanners {
		addresses[i] = scanner.Address()
		scannerMap[scanner.Address()] = scanner
		topics0 = append(topics0, scanner.Topics()...)
	}

	// 데이터베이스에서 가장 마지막에 저장된 블럭 넘버를 가져온다.
	fromBlock := uint64(0)
	for _, table := range dbtypes.AllTables {
		var val uint64
		// 발생할 수 있는 에러 무시함
		db.Model(table).Select("MAX(block_number)").Scan(&val)
		fromBlock = max(fromBlock, val)
	}

	return scannerMap, ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    [][]common.Hash{topics0},
		FromBlock: new(big.Int),
		ToBlock:   new(big.Int),
	}, fromBlock
}

// /////////
// Common //
// /////////

var (
	ErrNoEventSignature       = errors.New("no event signature")
	ErrInvalidEventID         = errors.New("invalid eventID exists")
	ErrEventSignatureMismatch = errors.New("event signature mismatch")
	ErrNonTargetedEvent       = errors.New("non-targeted event")
)

type IScanner interface {
	Address() common.Address
	Topics() []common.Hash
	Work(db *gorm.DB, log types.Log) error
}

func parse(log types.Log, outType reflect.Type, aBI *abi.ABI) (dbtypes.IRecord, error) {
	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		return nil, ErrNoEventSignature
	}
	event, err := aBI.EventByID(log.Topics[0])
	if err != nil {
		return nil, err
	}
	if outType == nil {
		return nil, errors.Wrap(ErrNonTargetedEvent, event.Name)
	}

	out := reflect.New(outType).Interface()
	if len(log.Data) > 0 {
		if err := aBI.UnpackIntoInterface(out, event.Name, log.Data); err != nil {
			return nil, errors.Wrap(err, event.Name)
		}
	}

	var indexed abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return out.(dbtypes.IRecord), abi.ParseTopics(out, indexed, log.Topics[1:])
}
