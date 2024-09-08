# Quick Start
```
docker compose up -d
```
동작중인 블록체인에 접근이 가능해야 합니다.
노드를 실행시키기 위해서 [eth-pos-devnet](https://github.com/bang9ming9/eth-pos-devnet) 을 확인해 주세요.
## 동작 순서
1. .docker/keystore 의 0번째 key (*14d95b8ecc31875f409c1fe5cd58b3b5cddfddfd*) 을 사용하여 연결된 노드에 bm contracts 를 배포합니다. (bct deploy)
   > 최초로 배포되는 컨트랙트의 주소는 동일 합니다.
2. postgres DB 를 띄웁니다.
3. postgres DB 에 테이블을 설정합니다. (bct scanner init)
4. 띄워진 노드에서 발생한 이벤트를 수집하여 postgres DB 에 저장합니다. (bct scanner)

# BUILD
``` bash
git clone https://github.com/bang9ming9/bm-cli-tool && cd bm-cli-tool/cmd
sudo go build -o /usr/local/bin/bct . # Bang9ming9 Cmmand line interface Tool
```

# Deploy
bm-governance, bm-erc721 컨트랙트를 배포합니다.

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

# Scanner
bm-governance 에서 발생하는 몇가지 이벤트를 수집합니다.
이벤트는 postgresDB 에 저장합니다.
DB 에서 지원하지 않는 타입인 *big.Int 및 리스트 타입은 Blob 타입으로 저장합니다.
common.Address, common.Hash 는 고정길이 Byte 로 저장합니다. char(20), char(32)

- BmErc20 (Transfer) # 홀더 확인
- BmErc1155 (TransferSinge, TransferBatch) # 활동중인 유저 확인
- BmGovernor (ProposalCreated, ProposalCanceled) # 진행중인 Proposal 확인
- Faucet (Claimed)

## Databas init
```bash
bct sacnner init --config ./scanner.toml
```
데이터 베이스에 준비된 테이블을 셋팅 합니다.

## Scan
```bash
bct sacnner --config ./scanner.toml
```
위의 5개의 이밴트를 수집하여 DB 에 저장합니다.

# TODO
1. CLI 기능 작업
   1. execute
        - ERC20 토큰발급
        - ERC1155 대표자 설정
        - 제안 신청
        - 제안 취소
        - 투표 (찬성,반대,기권)
   2. call
      - 진행중인 Proposal 목록 조회
      - 하나의 Proposal 상세 조회