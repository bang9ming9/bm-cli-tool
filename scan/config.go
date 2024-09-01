package scan

import (
	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

type Config struct {
	EndPoint struct {
		URL string `toml:"url"`
	} `toml:"end-point"`
	Contracts struct {
		ERC20      common.Address `toml:"erc20"`
		ERC1155    common.Address `toml:"erc1155"`
		Governance common.Address `toml:"governance"`
	} `toml:"contracts"`
	Database struct {
		URL      string `toml:"url"`
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
