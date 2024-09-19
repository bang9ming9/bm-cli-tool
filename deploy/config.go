package deploy

import (
	"context"
	"fmt"

	"github.com/bang9ming9/bm-cli-tool/cmd/flags"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Chain struct {
		URI string `toml:"uri"`
	} `toml:"chain"`
	Account struct {
		Keystore string         `toml:"keystore"`
		Address  common.Address `toml:"address"`
		Password string         `toml:"password"`
	} `toml:"account"`
}

func GetConfig(ctx *cli.Context) (*Config, error) {
	config, err := flags.ReadConfig[Config](ctx)
	if err != nil {
		return nil, err
	}

	if ctx.IsSet(flags.ChainFlag.Name) {
		config.Chain.URI = ctx.Path(flags.ChainFlag.Name)
	}

	if ctx.IsSet(flags.KeyStoreDirFlag.Name) {
		config.Account.Keystore = ctx.Path(flags.KeyStoreDirFlag.Name)
	}

	if ctx.IsSet(flags.AccountFlag.Name) {
		config.Account.Address = common.HexToAddress(ctx.String(flags.AccountFlag.Name))
	}

	if config.Account.Password == "" {
		config.Account.Password = flags.GetPassword(ctx)
	}

	return config, nil
}

func (cfg *Config) GetAccount(c context.Context, client *ethclient.Client) (*bind.TransactOpts, func() error, error) {
	ctx, cancel := context.WithTimeout(c, 5e9)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, nil, err
	}

	ks := keystore.NewKeyStore(cfg.Account.Keystore, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.Find(accounts.Account{Address: cfg.Account.Address})
	if err != nil {
		return nil, nil, err
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, account, chainID)
	if err != nil {
		return nil, nil, err
	}

	password := cfg.Account.Password
	if password == "" {
		password, err = (&promptui.Prompt{
			Label: fmt.Sprintf("%s's password", account.Address),
			Mask:  '*',
		}).Run()
		if err != nil {
			return nil, nil, err
		}
	}

	return opts, func() error { return ks.Lock(account.Address) }, ks.Unlock(account, password)
}
