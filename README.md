# BUILD
``` bash
git clone https://github.com/bang9ming9/bm-cli-tool && cd bm-cli-tool/cmd
sudo go build -o /usr/local/bin/bct . # bang9ming9 command line interface tool
```

# Deploy

## deploy with options
``` bash
# example
bct deploy \
--chain "https://localhost:8545" \
--keystore /Users/wm-bl000094/go/src/github.com/bang9ming9/eth-pos-devnet/execution/keystore \
--account 0x14d95b8ecc31875f409c1fe5cd58b3b5cddfddfd \
--password /Users/wm-bl000094/go/src/github.com/bang9ming9/eth-pos-devnet/execution/geth_passowrd.txt
```

## deploy with config
``` toml
# ./deploy.toml
[end-point]
url = "http://localhost:8545"

[account]
keystore = "/keystore"
address = "0x14d95b8ecc31875f409c1fe5cd58b3b5cddfddfd"
password = "password"
```
``` bash
bct deploy --config ./deploy.toml
```