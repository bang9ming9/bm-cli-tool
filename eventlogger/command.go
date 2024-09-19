package eventlogger

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Chain struct {
		URI string `toml:"uri"`
	} `toml:"chain"`
	Database struct {
		URI        string `toml:"uri"`
		Database   string `toml:"database"`
		Collection string `toml:"collection"`
	} `toml:"db"`
	Server struct {
		Host string `toml:"host"`
	} `toml:"server"`
	Logger struct {
		Level string `toml:"level"` // panic,fatal,error,warn,info,debug,trace
		File  string `toml:"file"`
	} `toml:"log"`
	FilterQuery struct {
		ScanBlock uint64           `toml:"scan-block"`
		Addresses []common.Address `toml:"addresses"`
	} `toml:"filter-query"`
}

var Command = &cli.Command{
	Name:  "event-logger",
	Flags: []cli.Flag{flags.ConfigFlag},
	Action: func(ctx *cli.Context) error {
		config, err := flags.ReadConfig[Config](ctx)
		if err != nil {
			return err
		}
		logger, err := config.NewLogger()
		if err != nil {
			return err
		}

		logger.Info("Set Filter Query from Config...")
		query, err := config.NewFilterQuery()
		if err != nil {
			return err
		}
		logger.Info("Dial ETH Client...")
		client, err := ethclient.DialContext(ctx.Context, config.Chain.URI)
		if err != nil {
			return err
		}
		defer client.Close()

		logger.Info("Connect Mongodb...")
		collection, err := config.ConnectDatabase()
		if err != nil {
			return err
		}
		defer collection.Database().Client().Disconnect(ctx.Context)

		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
		defer close(stopCh)

		logger.Info("Open Query Server...")
		return NewLoggerServer(stopCh, config.Server.Host, logger, client, collection, query)
	},
}

// Set Logger
func (config *Config) NewLogger() (*logrus.Logger, error) {
	logger := logrus.New()
	{
		cfg := config.Logger
		level, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			return nil, err
		}
		logger.SetLevel(level)
		if cfg.File == "" {
			logger.SetOutput(os.Stdout)
			logger.SetFormatter(&logrus.TextFormatter{
				ForceColors:      true,
				DisableColors:    false,
				DisableTimestamp: false,
				TimestampFormat:  "2006-01-02 15:04:05",
			})
		} else {
			logFile, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			logger.SetOutput(logFile)
			logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat:  "2006-01-02 15:04:05",
				DisableTimestamp: false,
			})
		}
	}
	return logger, nil
}

func (config *Config) ConnectDatabase() (*mongo.Collection, error) {
	cfg := config.Database
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	return client.Database(cfg.Database).Collection(cfg.Collection), nil
}

func (config *Config) NewFilterQuery() (*ethereum.FilterQuery, error) {
	cfg := config.FilterQuery
	return &ethereum.FilterQuery{
		Addresses: cfg.Addresses,
		FromBlock: new(big.Int).SetUint64(cfg.ScanBlock),
	}, nil
}
