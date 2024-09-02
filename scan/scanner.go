package scan

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	butils "github.com/bang9ming9/go-hardhat/bms/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Command = &cli.Command{
	Name:  "scanner",
	Flags: []cli.Flag{flags.ConfigFlag},
	Action: func(ctx *cli.Context) error {
		// logger 설정
		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:      true,
			DisableColors:    false, // Ensure colors are not disabled
			DisableTimestamp: false,
			TimestampFormat:  "2006-01-02 15:04:05",
		})
		// TODO logger 설정 flag 및 Config 로 변경 가능할수 있도록 수정
		logger.SetLevel(logrus.TraceLevel)
		logger.SetOutput(os.Stdout)

		logger.Info("Read Config...")
		config, err := flags.ReadConfig[Config](ctx)
		if err != nil {
			return err
		}

		logger.Info("Dial Blockchain Node...")
		client, err := ethclient.DialContext(ctx.Context, config.EndPoint.URL)
		if err != nil {
			return err
		}

		logger.Info("Connect Database...")
		db, err := gorm.Open(postgres.Open(config.GetPostgreDns()), &gorm.Config{})
		if err != nil {
			return err
		}

		logger.Info("Set Scanners...")
		erc20Scanner, err := NewERC20Scanner(config.Contracts.ERC20, logger)
		if err != nil {
			return err
		}

		erc1155Scanner, err := NewERC1155Scanner(config.Contracts.ERC1155, logger)
		if err != nil {
			return err
		}

		governorScanner, err := NewGovernorScanner(config.Contracts.Governance, logger)
		if err != nil {
			return err
		}

		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

		logger.Info("Start Scan!")
		return Scan(ctx.Context, logger, stopCh, client, db, erc20Scanner, erc1155Scanner, governorScanner)
	},
	Subcommands: []*cli.Command{
		{
			Name:  "init",
			Flags: []cli.Flag{flags.ConfigFlag},
			Action: func(ctx *cli.Context) error {
				config, err := flags.ReadConfig[Config](ctx)
				if err != nil {
					return err
				}

				db, err := gorm.Open(postgres.Open(config.GetPostgreDns()), &gorm.Config{})
				if err != nil {
					return err
				}

				err = db.AutoMigrate(
					&dbtypes.ERC20Transfer{},
					&dbtypes.ERC1155Transfer{},
					&dbtypes.GovernorProposalCreated{},
					&dbtypes.GovernorProposalCanceled{},
				)
				if err == nil {
					fmt.Println("Scanner Init Successed!")
				}
				return err
			},
		},
	},
}

type Backend interface {
	butils.Backend
	ethereum.BlockNumberReader
}

func Scan(ctx context.Context, logger *logrus.Logger, stop chan os.Signal,
	client Backend,
	db *gorm.DB,
	scanners ...IScanner,
) error {
	scannerMap, filterQuery, blockNumber := prepareFilter(db, scanners...)

	latestBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-stop:
			logger.Info("Stop Scanner")
			return nil
		default:
			// 현재 블럭으로 따라왔다면, 1초씩 대기하면서 블럭이 생성되기를 기다린다.
			if latestBlockNumber < blockNumber {
				time.Sleep(1e9)
				latestBlockNumber, err = client.BlockNumber(ctx)
				if err != nil {
					return err
				}
				break
			}
			// 한 블럭씩 증가
			blockNumber++
			logger.WithFields(logrus.Fields{
				"blockNumber": blockNumber,
			}).Trace("Scan")
			filterQuery.FromBlock.SetUint64(blockNumber)
			filterQuery.ToBlock.SetUint64(blockNumber)

			// 1. 이벤트 로그를 읽어온다.
			logs, err := client.FilterLogs(ctx, filterQuery)
			if err != nil {
				logger.Error("Fail to call filter logs")
				return err
			}

			// 2. 발생한 이벤트를 DB 에 저장한다.
			// 같은 블럭 넘버의 이벤트를 한번에 저장하기 위해 트랜잭션&커밋으로 동작
			// Scanner 재 기동시, 마지막으로 저장된 블럭넘버 이후부터 읽을 수 있다.
			// 일괄로 저장하려다 보니 저장 과정에서 에러가 발생한다면 리턴(종료) 유도.
			tx := db.Begin()
			for _, l := range logs {
				scanner, ok := scannerMap[l.Address]
				if ok {
					if err := scanner.Save(tx, l); err != nil {
						logger.Error("Fail to save logs")
						return err
					}
				}
			}
			if err := tx.Commit().Error; err != nil {
				logger.Error("Fail to commit logs")
				return err
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
	Save(db *gorm.DB, log types.Log) error
}

func parse(log types.Log, outType reflect.Type, aBI *abi.ABI) (dbtypes.ICreate, error) {
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
	return out.(dbtypes.ICreate), abi.ParseTopics(out, indexed, log.Topics[1:])
}
