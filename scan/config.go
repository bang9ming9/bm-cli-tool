package scan

import (
	"fmt"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

type ContractConfig struct {
	FromBlock  uint64         `toml:"from"` // 블록을 스캔할 시작 블럭
	Faucet     common.Address `toml:"faucet"`
	ERC20      common.Address `toml:"erc20"`
	ERC1155    common.Address `toml:"erc1155"`
	Governance common.Address `toml:"governance"`
}

type Config struct {
	EndPoint struct {
		URL string `toml:"url"`
	} `toml:"end-point"`
	Contracts ContractConfig `toml:"contracts"`
	Database  struct {
		DBName   string `toml:"name"`
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"passowrd"`
	} `toml:"db"`
}

func GetConfig(ctx *cli.Context) (*Config, error) {
	config, err := flags.ReadConfig[Config](ctx)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (cfg *Config) GetPostgreDns() string {
	dbConfig := cfg.Database
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName,
	)
}
