package scan

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"reflect"
	"syscall"

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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Command = &cli.Command{
	Name:  "scan",
	Flags: []cli.Flag{flags.ConfigFlag},
	Action: func(ctx *cli.Context) error {
		config, err := flags.ReadConfig[Config](ctx)
		if err != nil {
			return err
		}

		client, err := ethclient.DialContext(ctx.Context, config.EndPoint.URL)
		if err != nil {
			return err
		}

		// erc20, err := gov.NewBmErc20Filterer(config.Contracts.ERC20, client)
		// if err != nil {
		// 	return err
		// }

		// erc1155, err := gov.NewBmErc1155Filterer(config.Contracts.ERC1155, client)
		// if err != nil {
		// 	return err
		// }

		// governance, err := gov.NewBmGovernorFilterer(config.Contracts.Governance, client)
		// if err != nil {
		// 	return err
		// }

		dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}

		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:      true,
			DisableColors:    false, // Ensure colors are not disabled
			DisableTimestamp: false,
			TimestampFormat:  "2006-01-02 15:04:05",
		})

		erc20Scanner, err := NewERC20Scanner(config.Contracts.ERC20, logger)
		if err != nil {
			return err
		}

		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

		return Scan(ctx.Context, logger, stopCh, client, db, erc20Scanner)
	},
}

func Scan(ctx context.Context, logger *logrus.Logger,
	stop chan os.Signal,
	client butils.Backend, db *gorm.DB, scanners ...IScanner) error {
	scannerMap, filterQuery, blockNumber := prepareFilter(db, scanners...)

	for {
		select {
		case <-stop:
			return nil
		default:
			blockNumber++
			logger.WithFields(logrus.Fields{
				"blockNumber": blockNumber,
			}).Trace("Scan")

			// 한 블럭씩 증가
			filterQuery.FromBlock.SetUint64(blockNumber)
			filterQuery.ToBlock.SetUint64(blockNumber)
			// 1. 이벤트 로그를 읽어온다.
			logs, err := client.FilterLogs(ctx, filterQuery)

			if err != nil {
				return err
			}

			// 2. 발생한 이벤트를 DB 에 저장한다.
			// 같은 블럭 넘버의 이벤트를 한번에 저장하기 위해 트랜잭션&커밋으로 동작
			// Scanner 재 기동시, 마지막으로 저장된 블럭넘버 이후부터 읽을 수 있다.
			// 일괄로 저장하려다 보니 저장 과정에서 에러가 발생한다면 리턴.
			tx := db.Begin()
			for _, l := range logs {
				scanner, ok := scannerMap[l.Address]
				if ok {
					if err := scanner.Save(tx, l); err != nil {
						return err
					}
				}
			}
			if err := tx.Commit().Error; err != nil {
				return err
			}
		}
	}
}

func prepareFilter(_ *gorm.DB, scanners ...IScanner) (map[common.Address]IScanner, ethereum.FilterQuery, uint64) {
	// TODO 데이터베이스에서 가장 마지막에 저장된 블럭 넘버를 가져온다.
	// var fromBlock uint64
	// query := `SELECT MAX(Block) AS largest_value
	//     FROM (
	//         SELECT MAX(common_field) AS Block FROM table1
	//         UNION
	//         SELECT MAX(common_field) AS Block FROM table2
	//         UNION
	//         SELECT MAX(common_field) AS Block FROM table3
	//     ) AS Blocks;
	// `
	addresses := make([]common.Address, len(scanners))
	topics0 := make([]common.Hash, 0)

	scannerMap := make(map[common.Address]IScanner)
	for i, scanner := range scanners {
		addresses[i] = scanner.Address()
		scannerMap[scanner.Address()] = scanner
		topics0 = append(topics0, scanner.Topics()...)
	}

	return scannerMap, ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    [][]common.Hash{topics0},
		FromBlock: new(big.Int),
		ToBlock:   new(big.Int),
	}, 0
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
