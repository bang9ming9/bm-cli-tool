package flags

import (
	"os"
	"strings"

	"github.com/naoina/toml"
	"github.com/urfave/cli/v2"
)

const (
	AccountCategory = "ACCOUNT"
)

var (
	ConfigFlag = &cli.PathFlag{
		Name:      "config",
		TakesFile: true,
		Usage:     "TOML configuration file",
	}

	ChainFlag = &cli.StringFlag{
		Name:  "chain",
		Usage: "chain directory",
	}

	KeyStoreDirFlag = &cli.PathFlag{
		Name:     "keystore",
		Usage:    "Directory for the keystore (default = inside the datadir)",
		Category: AccountCategory,
	}

	AccountFlag = &cli.StringFlag{
		Name:     "account",
		Usage:    "account to unlock",
		Category: AccountCategory,
	}

	PasswordFileFlag = &cli.PathFlag{
		Name:      "password",
		Usage:     "Password file to use for non-interactive password input",
		TakesFile: true,
		Category:  AccountCategory,
	}
)

func GetPassword(ctx *cli.Context) string {
	if ctx.IsSet(PasswordFileFlag.Name) {
		bytes, _ := os.ReadFile(ctx.Path(PasswordFileFlag.Name))
		return strings.TrimSpace(string(bytes))
	}
	return ""
}

func ReadConfig[T any](ctx *cli.Context) (*T, error) {
	config := new(T)
	if ctx.IsSet(ConfigFlag.Name) {
		bytes, err := os.ReadFile(ctx.Path(ConfigFlag.Name))
		if err != nil {
			return nil, err
		}
		return config, toml.Unmarshal(bytes, &config)
	}
	return config, nil
}
