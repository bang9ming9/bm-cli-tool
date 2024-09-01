package scan_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/scan"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	gov "github.com/bang9ming9/bm-governance/types"
	"github.com/bang9ming9/go-hardhat/bms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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

func TestERC20Transfer(t *testing.T) {
	ctx := context.Background()
	// 0: Transfer 를 발생시키기 위한 동장
	node := bms.NewBacked(t)
	// 0-1 : ERC20 배포
	go func() {
		for {
			node.Commit()
			time.Sleep(1e9)
		}
	}()
	contracts, err := gov.DeployBMGovernor(ctx, node.Owner, node, 0, struct {
		Name   string
		Symbol string
	}{"", ""}, struct {
		Name    string
		Version string
		Uri     string
	}{"", "", ""}, struct{ Name string }{""})
	require.NoError(t, err)
	node.Commit()
	ERC20 := contracts.Erc20
	// 0-2 : ERC20 발급 (From:0, To:Owner, Value:COST)
	callOpts := new(bind.CallOpts)
	cost, err := ERC20.Funcs().COST(callOpts)
	require.NoError(t, err)
	node.Owner.Value = cost
	_, err = ERC20.Funcs().Mint(node.Owner, node.Owner.From)
	require.NoError(t, err)
	node.Commit()
	// 0: end

	// 1: scan.Scan() 함수 실행
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.ErrorLevel)

	// 1-1: 테스트 디비 생성
	db := testDB(t)
	defer db.Migrator().DropTable(&dbtypes.ERC20Transfer{})
	// 1-2: 테이블 설정
	require.NoError(t, db.AutoMigrate(&dbtypes.ERC20Transfer{}))
	// 1-3: ERC20 스캐너 생성
	erc20s, err := scan.NewERC20Scanner(ERC20.Address(), logger)
	require.NoError(t, err)
	// 1-4: Scan 실행
	stopCh := make(chan os.Signal, 1)
	go func() {
		time.Sleep(0.01e9)
		stopCh <- os.Interrupt
	}()
	require.NoError(t, scan.Scan(ctx, logger, stopCh, node, db, erc20s))
	// 1:end

	//  데이터 확인
	records := []dbtypes.ERC20Transfer{}
	require.NoError(t, db.Find(&records).Error)
	require.Equal(t, 1, len(records))
	record := records[0]
	require.Equal(t, record.From, common.Address{})
	require.Equal(t, record.To, node.Owner.From)
	require.True(t, record.Value.Big().Cmp(cost) == 0)
}
