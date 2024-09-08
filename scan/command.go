package scan

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/abis"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
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

		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

		// subscribe new block
		blockTick, newTick := make(chan uint64), make(chan struct{})
		go func() {
			newHead := make(chan *types.Header)
			sub, err := client.SubscribeNewHead(ctx.Context, newHead)
			if err != nil {
				logger.WithField("message", err.Error()).Error("fail to subscribe new block head")
				stopCh <- os.Interrupt
				return
			}
			defer sub.Unsubscribe()
			for {
				select {
				case head := <-newHead:
					blockTick <- head.Number.Uint64()
					newTick <- struct{}{}
				case err := <-sub.Err():
					logger.WithField("message", err.Error()).Error("fail to subscribe new block head")
					stopCh <- os.Interrupt
				}
			}
		}()

		// API Open
		go func() {
			logger.Info("Open Rest Api...", "listen", ":8080")
			governorCaller, err := gov.NewBmGovernorCaller(config.Contracts.Governance, client)
			if err != nil {
				logger.WithField("message", err.Error()).Panic("fail to new governance caller")
			}

			engine := gin.Default()
			v1 := engine.Group("/api/v1")
			for _, api := range []interface {
				RegisterApi(*gin.RouterGroup) error
			}{
				NewERC20Api(db), NewERC1155Api(db), NewFaucetApi(db),
				NewGovernorApi(db).Loop(governorCaller, logger, newTick),
			} {
				if err := api.RegisterApi(v1); err != nil {
					logger.WithField("message", err.Error()).Panic("fail to register api")
				}
			}
			if err := engine.Run(":8080"); err != nil {
				stopCh <- os.Interrupt
			}
		}()

		return Scan(ctx.Context, config.Contracts, logger, stopCh, db, client, blockTick)
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

				err = db.AutoMigrate(dbtypes.AllTables...)
				if err == nil {
					fmt.Println("Scanner Init Successed!")
				}
				return err
			},
		},
	},
}
