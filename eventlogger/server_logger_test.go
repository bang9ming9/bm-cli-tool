package eventlogger_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/eventlogger"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logtypes"
	abis "github.com/bang9ming9/bm-governance/abis"
	gov "github.com/bang9ming9/bm-governance/test"
	"github.com/bang9ming9/go-hardhat/bms"
	"github.com/bang9ming9/go-hardhat/bms/bmsutils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewLoggerServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		args, _, cancel := makeLogServerArgs(t)
		defer cancel()
		go func() {
			time.Sleep(1e9)
			args.stopCh <- os.Interrupt
			for {
				time.Sleep(1e9)
				args.client.Commit()
			}
		}()
		require.NoError(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			args.log,
			args.client,
			args.collection,
			args.query,
		))
	})
	t.Run("Input error", func(t *testing.T) {
		args, _, cancel := makeLogServerArgs(t)
		defer cancel()

		require.Error(t, eventlogger.NewLoggerServer(
			nil,
			args.addr,
			args.log,
			args.client,
			args.collection,
			args.query,
		))
		require.Error(t, eventlogger.NewLoggerServer(
			args.stopCh,
			"",
			args.log,
			args.client,
			args.collection,
			args.query,
		))
		require.Error(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			nil,
			args.client,
			args.collection,
			args.query,
		))
		require.Error(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			args.log,
			nil,
			args.collection,
			args.query,
		))
		require.Error(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			args.log,
			args.client,
			nil,
			args.query,
		))
		go func() {
			time.Sleep(1e9)
			args.stopCh <- os.Interrupt
			for {
				time.Sleep(1e9)
				args.client.Commit()
				args.stopCh <- os.Interrupt
			}
		}()
		require.NoError(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			args.log,
			args.client,
			args.collection,
			nil,
		))
	})
}

func TestLogging(t *testing.T) {
	ctx := context.Background()
	// ERC1155_OwnershipTransferred (1)
	args, contracts, cancel := makeLogServerArgs(t)
	defer cancel()

	go func() {
		require.NoError(t, eventlogger.NewLoggerServer(args.stopCh, args.addr, args.log, args.client, args.collection, args.query))
	}()

	owner := args.client.Owner
	txPool, callOpts := bmsutils.NewTxPool(args.client), new(bind.CallOpts)

	// ERC20_Transfer (2)
	cost, err := contracts.Erc20.Funcs().COST(callOpts)
	require.NoError(t, err)
	owner.Value = cost
	require.NoError(t, txPool.Exec(contracts.Erc20.Funcs().Mint(owner, owner.From)))
	owner.Value = nil
	// ERC20_Approved (3)
	require.NoError(t, txPool.Exec(contracts.Erc20.Funcs().Approve(owner, contracts.Governor.Address(), new(big.Int).SetBytes(common.MaxHash[:]))))
	require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))

	// ERC1155_TransferSingle, ERC1155_DelegateVotesChanged, ERC1155_DelegateChanged (4,5,6)
	// Governance_Proposal (7)
	contracts.NextProposalTime(t, args.client)
	proposal := contracts.NewProposalToTarget(t, "TestLoggin", 1, 2, "Hello")
	tx, err := contracts.Governor.Funcs().Propose(owner, proposal.Targets, proposal.Values, proposal.CallDatas, proposal.Description)
	require.NoError(t, txPool.Exec(tx, err))
	require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))

	args.stopCh <- os.Interrupt

	time.Sleep(2e9)
	cur, err := args.collection.Find(ctx, bson.D{})
	require.NoError(t, err)

	result := []bson.M{}
	require.NoError(t, cur.All(ctx, &result))
	require.Equal(t, 7, len(result))
}

func makeLogServerArgs(t *testing.T) (struct {
	stopCh     chan os.Signal
	addr       string
	log        *logrus.Logger
	client     *bms.Backend
	collection *mongo.Collection
	query      *ethereum.FilterQuery
}, *gov.BMGovernor, func()) {

	ctx := context.Background()

	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://bang9ming9:password@localhost:27017"))
	require.NoError(t, err)
	collection := dbClient.Database("test").Collection("testcollection")
	collection.Drop(ctx)

	backend, contracts := gov.DeployBMGovernorWithBackend(t)

	addresses := []common.Address{contracts.Erc20.Address(), contracts.Erc1155.Address(), contracts.Governor.Address()}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	returnVal := struct {
		stopCh     chan os.Signal
		addr       string
		log        *logrus.Logger
		client     *bms.Backend
		collection *mongo.Collection
		query      *ethereum.FilterQuery
	}{
		stopCh:     make(chan os.Signal),
		addr:       "localhost:50501",
		log:        log,
		client:     backend,
		collection: collection,
		query: &ethereum.FilterQuery{
			Addresses: addresses,
			FromBlock: common.Big1,
		},
	}

	return returnVal, contracts, func() { dbClient.Disconnect(ctx) }
}

var (
	faucet     common.Address = common.HexToAddress("0x0000000000000000000000000000000000004000")
	governance                = common.HexToAddress("0x6CEE2F2836abb07535a16AEf26e2C6326f7e2640")
	erc20                     = common.HexToAddress("0xc65Ef3Dc8D75769b02928778774eaA288A429403")
	erc1155                   = common.HexToAddress("0xc56dbaBCEd1a57f77209076bB6d711871a934f1f")
	erc721                    = common.HexToAddress("0x8d4B69F0308293ed37a154369E5A2c91A13CCD65")
)

func TestMakeLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	txPool, txOpts, contracts := getContracts(t, ctx)

	receiver, err := bms.GetEoaAt(contracts.ChainID, 1)
	require.NoError(t, err)
	t.Run("ERC20", func(t *testing.T) {
		t.Run("mint", func(t *testing.T) {
			minted, err := contracts.ERC20.Minted(txOpts.From)
			require.NoError(t, err)
			if !minted {
				cost, err := contracts.ERC20.COST()
				require.NoError(t, err)
				contracts.ERC20.TransactOpts.Value = cost
				require.NoError(t, txPool.Exec(contracts.ERC20.Mint(txOpts.From)))
				require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))
				contracts.ERC20.TransactOpts.Value = nil
			}
		})
		t.Run("transfer", func(t *testing.T) {
			err := txPool.Exec(contracts.ERC20.Transfer(receiver.From, common.Big1))
			require.NoError(t, bmsutils.ToRevert(err))
		})
		require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))
	})

	t.Run("ERC1155", func(t *testing.T) {
		currentID, err := contracts.ERC1155.CurrentID()
		t.Run("mint", func(t *testing.T) {
			require.NoError(t, err)
			balance, err := contracts.ERC1155.BalanceOf(txOpts.From, currentID)
			require.NoError(t, err)
			if balance.Sign() == 0 {
				require.NoError(t, txPool.Exec(contracts.ERC1155.Mint(txOpts.From)))
				require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))
			}
		})
		t.Run("delegate", func(t *testing.T) {
			delegator, err := contracts.ERC1155.Delegates(txOpts.From)
			require.NoError(t, err)
			if delegator != txOpts.From {
				require.NoError(t, txPool.Exec(contracts.ERC1155.Delegate(txOpts.From)))
				require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))
			}
		})
		t.Run("transfer", func(t *testing.T) {
			require.NoError(t, txPool.Exec(contracts.ERC1155.SafeTransferFrom(txOpts.From, receiver.From, currentID, common.Big1, []byte{})))
		})
		require.NoError(t, txPool.AllReceiptStatusSuccessful(ctx))
	})
}

const (
	BmErc721ABI string = `[{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"erc1155","type":"address"},{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"symbol","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"string","name":"value","type":"string"},{"internalType":"bytes32","name":"dataHash","type":"bytes32"}],"name":"BmErc721DuplicatedDataHash","type":"error"},{"inputs":[{"internalType":"uint256","name":"sum","type":"uint256"},{"internalType":"uint256","name":"need","type":"uint256"}],"name":"BmErc721InvalidERC1155Value","type":"error"},{"inputs":[{"internalType":"uint256","name":"tokenID","type":"uint256"}],"name":"BmErc721IsNotTransferable","type":"error"},{"inputs":[{"internalType":"string","name":"arg","type":"string"}],"name":"BmErc721NilInput","type":"error"},{"inputs":[],"name":"BmErc721ZeroERC1155ID","type":"error"},{"inputs":[],"name":"ERC721EnumerableForbiddenBatchMint","type":"error"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"owner","type":"address"}],"name":"ERC721IncorrectOwner","type":"error"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"ERC721InsufficientApproval","type":"error"},{"inputs":[{"internalType":"address","name":"approver","type":"address"}],"name":"ERC721InvalidApprover","type":"error"},{"inputs":[{"internalType":"address","name":"operator","type":"address"}],"name":"ERC721InvalidOperator","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"ERC721InvalidOwner","type":"error"},{"inputs":[{"internalType":"address","name":"receiver","type":"address"}],"name":"ERC721InvalidReceiver","type":"error"},{"inputs":[{"internalType":"address","name":"sender","type":"address"}],"name":"ERC721InvalidSender","type":"error"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"ERC721NonexistentToken","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"ERC721OutOfBoundsIndex","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"OwnableInvalidOwner","type":"error"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"OwnableUnauthorizedAccount","type":"error"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"approved","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":false,"internalType":"bool","name":"approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"tokenID","type":"uint256"},{"indexed":true,"internalType":"bool","name":"able","type":"bool"}],"name":"TransferAbleSet","type":"event"},{"inputs":[],"name":"BM_ERC1155","outputs":[{"internalType":"contract IERC1155","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"MINT_COST","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"approve","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenID","type":"uint256"}],"name":"burn","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"dataHash","type":"bytes32"}],"name":"existed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"burnID","type":"uint256"},{"internalType":"string","name":"value","type":"string"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256[]","name":"burnIDs","type":"uint256[]"},{"internalType":"uint256[]","name":"amounts","type":"uint256[]"},{"internalType":"string","name":"value","type":"string"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"bool","name":"approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"baseURI","type":"string"}],"name":"setBaseURI","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenID","type":"uint256"},{"internalType":"bool","name":"able","type":"bool"}],"name":"setTransferable","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes4","name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"tokenByIndex","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenID","type":"uint256"}],"name":"tokenData","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"tokenOfOwnerByIndex","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"tokenURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenID","type":"uint256"}],"name":"transferable","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
)

func getContracts(t *testing.T, ctx context.Context) (*bmsutils.TxPool, *bind.TransactOpts, struct {
	ChainID  *big.Int
	Faucet   abis.FaucetSession
	Governor abis.BmGovernorSession
	ERC20    abis.BmErc20Session
	ERC1155  abis.BmErc1155Session
	ERC721   *bind.BoundContract
}) {
	dialContext, cancel := context.WithTimeout(ctx, 5e9)
	defer cancel()
	client, err := ethclient.DialContext(dialContext, "http://localhost:8545")
	require.NoError(t, err)
	chainID, err := client.ChainID(ctx)
	require.NoError(t, err)
	callOpts := &bind.CallOpts{Context: ctx}
	require.NoError(t, err)

	keyIn, err := os.Open(filepath.Join(os.Getenv("HOME"), "/go/src/github.com/bang9ming9/eth-pos-devnet/execution/keystore/UTC--2024-07-29T22-33-43.797044000Z--14d95b8ecc31875f409c1fe5cd58b3b5cddfddfd"))
	require.NoError(t, err)
	txOpts, err := bind.NewTransactorWithChainID(keyIn, "password", chainID)
	require.NoError(t, err)

	Faucet, err := abis.NewFaucet(faucet, client)
	require.NoError(t, err)
	Governor, err := abis.NewBmGovernor(governance, client)
	require.NoError(t, err)
	Erc20, err := abis.NewBmErc20(erc20, client)
	require.NoError(t, err)
	Erc1155, err := abis.NewBmErc1155(erc1155, client)
	require.NoError(t, err)
	abi721, err := abi.JSON(strings.NewReader(BmErc721ABI))
	require.NoError(t, err)
	ERC721 := bind.NewBoundContract(erc721, abi721, client, client, client)

	abiFaucet, err := abis.FaucetMetaData.GetAbi()
	require.NoError(t, err)
	abiGovernor, err := abis.BmGovernorMetaData.GetAbi()
	require.NoError(t, err)
	abiERC20, err := abis.BmErc20MetaData.GetAbi()
	require.NoError(t, err)
	abiERC1155, err := abis.BmErc1155MetaData.GetAbi()
	require.NoError(t, err)

	bmsutils.EnrollErrors(abiFaucet, abiGovernor, abiERC20, abiERC1155, &abi721)
	return bmsutils.NewTxPool(client), txOpts, struct {
		ChainID  *big.Int
		Faucet   abis.FaucetSession
		Governor abis.BmGovernorSession
		ERC20    abis.BmErc20Session
		ERC1155  abis.BmErc1155Session
		ERC721   *bind.BoundContract
	}{
		ChainID:  chainID,
		Faucet:   abis.FaucetSession{Contract: Faucet, CallOpts: *callOpts, TransactOpts: *txOpts},
		Governor: abis.BmGovernorSession{Contract: Governor, CallOpts: *callOpts, TransactOpts: *txOpts},
		ERC20:    abis.BmErc20Session{Contract: Erc20, CallOpts: *callOpts, TransactOpts: *txOpts},
		ERC1155:  abis.BmErc1155Session{Contract: Erc1155, CallOpts: *callOpts, TransactOpts: *txOpts},
		ERC721:   ERC721,
	}
}

func TestGetLogs(t *testing.T) {
	conn, err := grpc.NewClient("localhost:50501", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	loggerClient := logger.NewLoggerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3e9)
	defer cancel()

	infoResMessage, err := loggerClient.Info(ctx, nil)
	require.NoError(t, err)
	addresses := make([]common.Address, len(infoResMessage.GetAddress()))
	for i, address := range infoResMessage.GetAddress() {
		addresses[i] = common.BytesToAddress(address)
	}
	t.Log(addresses)

	connStream, err := loggerClient.Connect(ctx, &logger.ConnectReqMessage{
		FromBlock: 1,
		Address:   erc20.Bytes(),
	})
	require.NoError(t, err)

	for {
		res, err := connStream.Recv()
		if err != nil {
			t.Log(err)
			break
		}
		log := logtypes.LogFromProtobuf(res)
		t.Log(log)
	}
}
