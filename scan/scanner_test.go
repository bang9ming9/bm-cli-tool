package scan_test

import (
	"context"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/scan"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/test"
	"github.com/bang9ming9/go-hardhat/bms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	// SQLite 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestScannerBasic(t *testing.T) {
	// 0: log 데이터 생성
	// 0-1 : bm-governance 테스트 컨트랙트 배포
	backend, contracts := gov.DeployBMGovernorWithBackend(t)
	ERC20, ERC1155 := contracts.Erc20, contracts.Erc1155
	// 0-2 : ERC20 발급 (From:0, To:Owner, Value:COST)
	callOpts := new(bind.CallOpts)
	cost, err := ERC20.Funcs().COST(callOpts)
	require.NoError(t, err)
	backend.Owner.Value = cost
	_, err = ERC20.Funcs().Mint(backend.Owner, backend.Owner.From)
	require.NoError(t, err)
	backend.Commit()
	backend.Owner.Value = nil
	// 0-3 : ERC1155 발급 TransferSingle(From : 0, To:Owner, Ids:[block.timestamp/86400], Values:[COST])
	_, err = ERC1155.Funcs().Mint(backend.Owner, backend.Owner.From)
	require.NoError(t, err)
	backend.Commit()
	currentID, err := ERC1155.Funcs().CurrentID(callOpts)
	require.NoError(t, err)
	// 0-4 : Deleate Self
	_, err = ERC1155.Funcs().Delegate(backend.Owner, backend.Owner.From)
	require.NoError(t, err)
	backend.Commit()
	// 0-5 : Proposal 등록 (ProposalCreated)
	proposal, proposalID := func() (*gov.Proposal, *big.Int) {
		retry := 0
		for {
			proposal := contracts.NewProposalToTarget(t, "TC1", 1, common.Address{1}, common.Hash{1}, "1")
			tx, err := contracts.Governor.Funcs().Propose(backend.Owner, proposal.Targets, proposal.Values, proposal.CallDatas, proposal.Description)
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
	_, err = contracts.Governor.Funcs().Cancel(backend.Owner, proposal.Targets, proposal.Values, proposal.CallDatas, crypto.Keccak256Hash([]byte(proposal.Description)))
	require.NoError(t, err)
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
	_, err = ERC1155.Funcs().Mint(backend.Owner, backend.Owner.From)
	require.NoError(t, err)
	backend.Commit()
	receiver := bms.GetEOA(t)
	_, err = ERC1155.Funcs().SafeBatchTransferFrom(backend.Owner, backend.Owner.From, receiver.From, []*big.Int{currentID, currentID2}, []*big.Int{cost, cost}, []byte{})
	require.NoError(t, err)
	backend.Commit()
	// 0: end

	// 1: scan.Scan() 함수 실행
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.DebugLevel)
	// 1-1: 테스트 디비 생성
	db := testDB(t)
	// 1-2: 테이블 설정
	require.NoError(t, db.AutoMigrate(&dbtypes.ERC20Transfer{}))
	defer db.Migrator().DropTable(&dbtypes.ERC20Transfer{})
	require.NoError(t, db.AutoMigrate(&dbtypes.ERC1155Transfer{}))
	defer db.Migrator().DropTable(&dbtypes.ERC1155Transfer{})
	require.NoError(t, db.AutoMigrate(&dbtypes.GovernorProposalCreated{}))
	defer db.Migrator().DropTable(&dbtypes.GovernorProposalCreated{})
	require.NoError(t, db.AutoMigrate(&dbtypes.GovernorProposalCanceled{}))
	defer db.Migrator().DropTable(&dbtypes.GovernorProposalCanceled{})
	// 1-3: 스캐너 생성
	erc20s, err := scan.NewERC20Scanner(ERC20.Address(), logger)
	require.NoError(t, err)
	require.NotNil(t, erc20s)
	erc1155s, err := scan.NewERC1155Scanner(ERC1155.Address(), logger)
	require.NoError(t, err)
	require.NotNil(t, erc1155s)
	governors, err := scan.NewGovernorScanner(contracts.Governor.Address(), logger)
	require.NoError(t, err)
	require.NotNil(t, governors)
	// 1-4: Scan 실행
	ctx := context.Background()
	stopCh := make(chan os.Signal, 1)
	go func() {
		time.Sleep(1e9)
		stopCh <- os.Interrupt
	}()
	require.NoError(t, scan.Scan(ctx, logger, stopCh, backend, db,
		erc20s,
		erc1155s,
		governors,
	))
	// 1:end

	// 데이터 확인 (ERC20_Transfer)
	{
		records := []dbtypes.ERC20Transfer{}
		require.NoError(t, db.Find(&records).Error)
		require.Equal(t, 1, len(records))
		record := records[0]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, backend.Owner.From)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
	}

	// 데이터 확인 (ERC1155_Transfer)
	{
		records := []dbtypes.ERC1155Transfer{}
		require.NoError(t, db.Find(&records).Error)
		require.Equal(t, 4, len(records))
		record := records[0]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, backend.Owner.From)
		require.True(t, record.Id.Get().Cmp(currentID) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[1]
		require.Equal(t, record.From, common.Address{})
		require.Equal(t, record.To, backend.Owner.From)
		require.True(t, record.Id.Get().Cmp(currentID2) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[2]
		require.Equal(t, record.From, backend.Owner.From)
		require.Equal(t, record.To, receiver.From)
		require.True(t, record.Id.Get().Cmp(currentID) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
		record = records[3]
		require.Equal(t, record.From, backend.Owner.From)
		require.Equal(t, record.To, receiver.From)
		require.True(t, record.Id.Get().Cmp(currentID2) == 0)
		require.True(t, record.Value.Get().Cmp(cost) == 0)
	}

	// 데이터 확인 (Governor ProposalCreated, ProposalCanceled)
	{
		// ProposalCreated
		{
			records := []dbtypes.GovernorProposalCreated{}
			require.NoError(t, db.Find(&records).Error)
			require.Equal(t, 1, len(records))
			record := records[0]
			require.Equal(t, record.ProposalId.Get(), proposalID)
			require.Equal(t, record.Proposer, backend.Owner.From)
			reflect.DeepEqual(record.Targets.Get(), proposal.Targets)
			reflect.DeepEqual(record.Values.Get(), proposal.Values)
			reflect.DeepEqual(record.Description, proposal.Description)
		}
		// ProposalCanceled
		{
			records := []dbtypes.GovernorProposalCanceled{}
			require.NoError(t, db.Find(&records).Error)
			require.Equal(t, 1, len(records))
			record := records[0]
			require.Equal(t, record.ProposalId.Get(), proposalID)
		}
	}
}
