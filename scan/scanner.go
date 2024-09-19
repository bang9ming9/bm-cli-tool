package scan

import (
	"context"
	"os"
	"time"

	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Scan(
	ctx context.Context, stop chan os.Signal,
	config ContractConfig,
	db *gorm.DB, client logger.LoggerClient,
	log *logrus.Logger,
) error {
	log.Info("Set Scanners...")

	scanners, err := func() ([]IScanner, error) {
		scanners := []IScanner{}
		zero := common.Address{}
		if config.ERC20 != zero {
			scanner, err := NewERC20Scanner(config.ERC20, log)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.ERC1155 != zero {
			scanner, err := NewERC1155Scanner(config.ERC1155, log)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.Faucet != zero {
			scanner, err := NewFaucetScanner(config.Faucet, log)
			if err != nil {
				return nil, err
			}
			scanners = append(scanners, scanner)
		}
		if config.Governance != zero {
			scanner, err := NewGovernorScanner(config.Governance, log)
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	txCH := make(chan func(db *gorm.DB) error, 256)
	for _, scanner := range scanners {
		err := scanner.Scan(ctx, client, config.FromBlock, txCH)
		if err != nil {
			return err
		}
	}

	tick, txs := time.NewTicker(1e9), make([]func(db *gorm.DB) error, 0, 256)
	log.Info("Start Scan!")
	for {
		select {
		case <-stop:
			return nil
		case tx := <-txCH:
			txs = append(txs, tx)
		case <-tick.C:
			length := len(txs)
			if length == 0 {
				break
			}

			transaction := db.Begin()
			for i := 0; i < length; i++ {
				if err := txs[i](transaction); err != nil {
					return err
				}
			}
			if err := transaction.Commit().Error; err != nil {
				transaction.Rollback()
				return err
			}
			txs = txs[length:]
		}
	}
}
