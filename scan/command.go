package scan

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Command = &cli.Command{
	Name:  "scanner",
	Flags: []cli.Flag{flags.ConfigFlag},
	Action: func(ctx *cli.Context) error {
		// log 설정
		log := logrus.New()
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:      true,
			DisableColors:    false, // Ensure colors are not disabled
			DisableTimestamp: false,
			TimestampFormat:  "2006-01-02 15:04:05",
		})
		// TODO log 설정 flag 및 Config 로 변경 가능할수 있도록 수정
		log.SetLevel(logrus.TraceLevel)
		log.SetOutput(os.Stdout)

		log.Info("Read Config...")
		config, err := flags.ReadConfig[Config](ctx)
		if err != nil {
			return err
		}

		log.Info("Connect Database...")
		db, err := gorm.Open(postgres.Open(config.GetPostgreDns()), &gorm.Config{})
		if err != nil {
			return err
		}

		log.Info("Connect EventLogger...")
		conn, err := grpc.NewClient(config.EventLogger.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

		// API Open
		go func() {
			log.Info("Open Rest Api...", "listen", "0.0.0.0:8090")

			engine := gin.Default()
			v1 := engine.Group("/api/v1")
			for _, api := range []interface {
				RegisterApi(*gin.RouterGroup) error
			}{
				NewERC20Api(db), NewERC1155Api(db), NewFaucetApi(db), NewGovernorApi(db),
			} {
				if err := api.RegisterApi(v1); err != nil {
					log.WithField("message", err.Error()).Panic("fail to register api")
				}
			}
			if err := engine.Run("0.0.0.0:8090"); err != nil {
				stopCh <- os.Interrupt
			}
		}()

		return Scan(ctx.Context, stopCh, config.Contracts, db, logger.NewLoggerClient(conn), log)
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
