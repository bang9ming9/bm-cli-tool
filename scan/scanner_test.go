package scan_test

import (
	"context"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/eventlogger"
	"github.com/bang9ming9/bm-cli-tool/eventlogger/logger"
	"github.com/bang9ming9/bm-cli-tool/scan"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/bang9ming9/bm-cli-tool/testutils"
	gabis "github.com/bang9ming9/bm-governance/abis"
	gov "github.com/bang9ming9/bm-governance/test"
	"github.com/bang9ming9/go-hardhat/bms"
	"github.com/bang9ming9/go-hardhat/bms/bmsutils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestScannerBasic(t *testing.T) {
	// 0: log 데이터 생성
	// 0-1 : 컨트랙트 배포
	// 0-1-1 : bm-governance 및 실행 테스트 컨트랙트 배포
	args, backend, contracts, close := makeLogServerArgs(t)
	defer close()
	loggerClient, adminClient, close2 := startEventLogger(t, args)
	defer close2()

	ERC20, ERC1155 := contracts.Erc20, contracts.Erc1155
	backend.Commit()
	// 0-1-2 : Faucet 컨트랙트 배포 및 코인 충전
	faucet, _, FAUCET, err := gabis.DeployFaucet(backend.Owner, backend)
	require.NoError(t, err)
	backend.Commit()
	backend.Owner.Value = bmsutils.ToWei(100)
	_, err = bmsutils.SendDynamicTx(backend, backend.Owner, &faucet, []byte{})
	require.NoError(t, err)
	backend.Commit()
	backend.Owner.Value = nil
	_, err = adminClient.Add(context.TODO(), &logger.AddressReqMessage{Address: faucet[:]})
	require.NoError(t, err)
	// 0-2 : ERC20 발급 (From:0, To:Owner, Value:COST)
	// 0-2-1 : 테스터 계정 생성 및 Faucet
	tester := bms.GetTEoa(t)
	_, err = FAUCET.Claim(tester, tester.From)
	require.NoError(t, err)
	backend.Commit()

	callOpts := new(bind.CallOpts)
	cost, err := ERC20.Funcs().COST(callOpts)
	require.NoError(t, err)
	tester.Value = cost
	_, err = ERC20.Funcs().Mint(tester, tester.From)
	require.NoError(t, err)
	backend.Commit()
	tester.Value = nil
	// 0-3 : ERC1155 발급 TransferSingle(From : 0, To:Owner, Ids:[block.timestamp/86400], Values:[COST])
	_, err = ERC1155.Funcs().Mint(tester, tester.From)
	require.NoError(t, err)
	backend.Commit()
	currentID, err := ERC1155.Funcs().CurrentID(callOpts)
	require.NoError(t, err)
	// 0-4 : Deleate Self
	_, err = ERC1155.Funcs().Delegate(tester, tester.From)
	require.NoError(t, err)
	backend.Commit()
	// 0-5 : Proposal 등록 (ProposalCreated)
	proposal, proposalID := func() (*gov.Proposal, *big.Int) {
		retry := 0
		for {
			proposal := contracts.NewProposalToTarget(t, "TC1", 1, common.Address{1}, common.Hash{1}, "1")
			tx, err := contracts.Governor.Funcs().Propose(tester, proposal.Targets, proposal.Values, proposal.CallDatas, proposal.Description)
			// 주말 시간이라면 에러가 발생할 수 있기 때문에 하루씩 더하면서 에러가 발생하지 않을때까지 시도한다.
			if retry > 7 {
				require.NoError(t, err)
			}
			if err == nil {
				receipt, err := bind.WaitMined(context.Background(), backend, tx)
				require.NoError(t, err)
				require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
				return proposal, contracts.UnpackProposalCreated(t, receipt).ProposalId
			}
			backend.AdjustTime(86400)
			backend.Commit()
			retry++
		}
	}()
	require.NotNil(t, proposal)
	require.True(t, proposalID.Sign() > 0)
	// 0-6 : Proposal 취소 (ProposalCanceled)
	_, err = contracts.Governor.Funcs().Cancel(tester, proposal.Targets, proposal.Values, proposal.CallDatas, crypto.Keccak256Hash([]byte(proposal.Description)))
	require.NoError(t, err)
	backend.Commit()
	// 0-7 ERC1155 발급 및 전송
	currentID2 := func() *big.Int {
		for {
			ID, err := ERC1155.Funcs().CurrentID(callOpts)
			require.NoError(t, err)
			if currentID.Cmp(ID) != 0 {
				return ID
			}
			require.NoError(t, backend.AdjustTime(86400*7))
			backend.Commit()
		}
	}()
	_, err = ERC1155.Funcs().Mint(tester, tester.From)
	require.NoError(t, err)
	backend.Commit()
	receiver := bms.GetTEoa(t)
	_, err = ERC1155.Funcs().SafeBatchTransferFrom(tester, tester.From, receiver.From, []*big.Int{currentID, currentID2}, []*big.Int{cost, cost}, []byte{})
	require.NoError(t, err)
	backend.Commit()
	// 0: end

	// 1: scan.Scan() 함수 실행
	logentry := logrus.New()
	logentry.SetFormatter(&logrus.TextFormatter{})
	logentry.SetOutput(os.Stderr)
	logentry.SetLevel(logrus.ErrorLevel)
	// 1-1: 테스트 디비 생성
	db := testutils.NewSQLMock(t)
	require.NoError(t, db.AutoMigrate(dbtypes.AllTables...))
	// 1-2: Scan 실행
	go func() {
		require.NoError(t, scan.Scan(context.Background(), make(chan os.Signal), scan.ContractConfig{
			FromBlock:  1,
			Faucet:     faucet,
			ERC20:      ERC20.Address(),
			ERC1155:    ERC1155.Address(),
			Governance: contracts.Governor.Address(),
		}, db, loggerClient, logentry))
	}()
	time.Sleep(3e9)
	// 1:end

	// 데이터 확인 (ERC20_Transfer)
	{
		records := []dbtypes.ERC20Transfer{}
		require.NoError(t, db.Find(&records).Error)
		require.Equal(t, 1, len(records))
		record := records[0]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, tester.From)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
	}

	// 데이터 확인 (ERC1155_Transfer)
	{
		records := []dbtypes.ERC1155Transfer{}
		require.NoError(t, db.Find(&records).Error)
		require.Equal(t, 4, len(records))
		record := records[0]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, tester.From)
		require.True(t, record.Id.Get().Cmp(currentID) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[1]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, tester.From)
		require.True(t, record.Id.Get().Cmp(currentID2) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[2]
		require.Equal(t, record.From, tester.From)
		require.Equal(t, record.To, receiver.From)
		require.True(t, record.Id.Get().Cmp(currentID) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[3]
		require.Equal(t, record.From, tester.From)
		require.Equal(t, record.To, receiver.From)
		require.True(t, record.Id.Get().Cmp(currentID2) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
	}
	// 데이터 확인 (Faucet_Claimed)
	{
		records := []dbtypes.FaucetClaimed{}
		require.NoError(t, db.Find(&records).Error)
		require.Equal(t, 1, len(records))
		record := records[0]
		require.Equal(t, tester.From, record.Account)
	}

	// 데이터 확인 (Governor ProposalCreated, ProposalCanceled)
	{
		// ProposalCreated
		{
			records := []dbtypes.GovernorProposal{}
			require.NoError(t, db.Find(&records).Error)
			require.Equal(t, 1, len(records))
			record := records[0]
			require.Equal(t, record.ProposalId.Get(), proposalID)
			require.Equal(t, record.Proposer, tester.From)
			reflect.DeepEqual(record.Targets.Get(), proposal.Targets)
			reflect.DeepEqual(record.Values.Get(), proposal.Values)
			reflect.DeepEqual(record.Description, proposal.Description)
			require.False(t, record.Active) // canceled
		}
	}
}

type EventLoggerArgs struct {
	stopCh     chan os.Signal
	addr       string
	log        *logrus.Logger
	client     *bms.Backend
	collection *mongo.Collection
	query      *ethereum.FilterQuery
}

func makeLogServerArgs(t *testing.T) (EventLoggerArgs, *bms.Backend, *gov.BMGovernor, func()) {
	ctx := context.Background()
	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://bang9ming9:password@localhost:27017"))
	require.NoError(t, err)
	collection := dbClient.Database("test").Collection("testcollection")
	collection.Drop(ctx)

	backend, contracts := gov.DeployBMGovernorWithBackend(t)

	addresses := []common.Address{contracts.Erc20.Address(), contracts.Erc1155.Address(), contracts.Governor.Address()}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	args := EventLoggerArgs{
		stopCh:     make(chan os.Signal),
		addr:       "localhost:50503",
		log:        log,
		client:     backend,
		collection: collection,
		query: &ethereum.FilterQuery{
			Addresses: addresses,
			FromBlock: common.Big1,
		},
	}

	return args, backend, contracts, func() { dbClient.Disconnect(ctx) }
}

func startEventLogger(t *testing.T, args EventLoggerArgs) (logger.LoggerClient, logger.AdminClient, func()) {
	go func() {
		require.NoError(t, eventlogger.NewLoggerServer(
			args.stopCh,
			args.addr,
			args.log,
			args.client,
			args.collection,
			args.query,
		))
	}()

	for i := 0; i < 3; i++ {
		time.Sleep(1e9)
		conn, err := grpc.NewClient("localhost:50503", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			return logger.NewLoggerClient(conn), logger.NewAdminClient(conn), func() { args.stopCh <- os.Interrupt }
		}
	}
	panic("fail to connect grpc")
}
