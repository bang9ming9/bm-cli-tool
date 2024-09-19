package scan_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bang9ming9/bm-cli-tool/scan"
	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/bang9ming9/bm-cli-tool/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	db := testutils.NewSQLMock(t)
	db.AutoMigrate(dbtypes.AllTables...)

	engine := gin.Default()
	v1 := engine.Group("/test")
	{
		require.NoError(t, scan.NewERC20Api(db).RegisterApi(v1))
		require.NoError(t, scan.NewERC1155Api(db).RegisterApi(v1))
		require.NoError(t, scan.NewFaucetApi(db).RegisterApi(v1))
		require.NoError(t, scan.NewGovernorApi(db).RegisterApi(v1))
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: engine,
	}
	go func() {
		require.Equal(t, http.ErrServerClosed, srv.ListenAndServe())
	}()
	time.Sleep(1e9)

	holders := []common.Address{common.BytesToAddress([]byte("1")), common.BytesToAddress([]byte("2"))}
	t.Run("ERC20Api", func(t *testing.T) {
		// DB 데이터 저장
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("1")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[0],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("11")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[0],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("21")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[0],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("2")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[1],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("12")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[1],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC20Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("22")),
				Block:  1,
			},
			From:  common.Address{},
			To:    holders[1],
			Value: (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)

		t.Run("holders", func(t *testing.T) {
			status, body, err := GetRequest[[]common.Address]("/erc20/holders")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.True(t, reflect.DeepEqual(holders, body))
		})
		t.Run("history", func(t *testing.T) {
			status, _, err := GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history")
			require.Error(t, err)
			require.Equal(t, http.StatusNotFound, status)

			status, _, err = GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history/hello")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			status, _, err = GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history/" + "bang9ming9")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			zero := common.Address{}
			status, _, err = GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history/" + zero.Hex())
			require.Error(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, status)

			status, body, err := GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history/" + (common.Address{1}).Hex())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Equal(t, 0, len(body))

			for _, holder := range holders {
				status, body, err = GetRequest[[]dbtypes.ERC20Transfer]("/erc20/history/" + holder.Hex())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 3, len(body))
			}
		})
	})
	t.Run("ERC1155Api", func(t *testing.T) {
		// DB 데이터 저장
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("1")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[0],
			Id:       (*dbtypes.BigInt)(big.NewInt(1)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("11")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[0],
			Id:       (*dbtypes.BigInt)(big.NewInt(1)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("21")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[0],
			Id:       (*dbtypes.BigInt)(big.NewInt(1)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("2")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[1],
			Id:       (*dbtypes.BigInt)(big.NewInt(1)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("12")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[1],
			Id:       (*dbtypes.BigInt)(big.NewInt(2)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)
		require.NoError(t, db.Create(&dbtypes.ERC1155Transfer{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("22")),
				Block:  1,
			},
			Index:    0,
			Operator: common.Address{},
			From:     common.Address{},
			To:       holders[1],
			Id:       (*dbtypes.BigInt)(big.NewInt(3)),
			Value:    (*dbtypes.BigInt)(big.NewInt(1)),
		}).Error)

		t.Run("holders", func(t *testing.T) {
			status, _, err := GetRequest[[]common.Address]("/erc1155/holders")
			require.Error(t, err)
			require.Equal(t, http.StatusNotFound, status)
			status, _, err = GetRequest[[]common.Address]("/erc1155/holders/0")
			require.Error(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, status)
			status, _, err = GetRequest[[]common.Address]("/erc1155/holders/hello")
			require.Error(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, status)

			status, body, err := GetRequest[[]common.Address]("/erc1155/holders/1")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.True(t, reflect.DeepEqual(holders, body))

			status, body0x01, err := GetRequest[[]common.Address]("/erc1155/holders/0x01")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.True(t, reflect.DeepEqual(body, body0x01))

			status, body, err = GetRequest[[]common.Address]("/erc1155/holders/2")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Truef(t, reflect.DeepEqual(holders[1:], body), "\nexpected: %v\nactual  : %v", holders[1:], body)

			status, body0x02, err := GetRequest[[]common.Address]("/erc1155/holders/0x02")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Truef(t, reflect.DeepEqual(body, body0x02), "\nexpected: %v\nactual  :%v", body, body0x02)
		})
		t.Run("history", func(t *testing.T) {
			status, _, err := GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history")
			require.Error(t, err)
			require.Equal(t, http.StatusNotFound, status)

			status, _, err = GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history/hello")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			status, _, err = GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history/" + "bang9ming9")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			zero := common.Address{}
			status, _, err = GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history/" + zero.Hex())
			require.Error(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, status)

			status, body, err := GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history/" + (common.Address{1}).Hex())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Equal(t, 0, len(body))

			for _, holder := range holders {
				status, body, err = GetRequest[[]dbtypes.ERC1155Transfer]("/erc1155/history/" + holder.Hex())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 3, len(body))
			}
		})
	})
	t.Run("FaucetApi", func(t *testing.T) {
		// DB 데이터 저장
		require.NoError(t, db.Create(&dbtypes.FaucetClaimed{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("1")),
				Block:  1,
			},
			Account: holders[0],
		}).Error)
		require.NoError(t, db.Create(&dbtypes.FaucetClaimed{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("11")),
				Block:  2,
			},
			Account: holders[0],
		}).Error)
		require.NoError(t, db.Create(&dbtypes.FaucetClaimed{
			Raw: dbtypes.Raw{
				TxHash: common.BytesToHash([]byte("2")),
				Block:  1,
			},
			Account: holders[1],
		}).Error)
		t.Run("history", func(t *testing.T) {
			status, _, err := GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history")
			require.Error(t, err)
			require.NotEqual(t, http.StatusOK, status)

			status, _, err = GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/hello")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			status, _, err = GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/" + "bang9ming9")
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, status)

			zero := common.Address{}
			status, _, err = GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/" + zero.Hex())
			require.Error(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, status)

			status, body, err := GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/" + (common.Address{1}).Hex())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Equal(t, 0, len(body))

			status, body, err = GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/" + holders[0].Hex())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Equal(t, 2, len(body))

			status, body, err = GetRequest[[]dbtypes.FaucetClaimed]("/faucet/history/" + holders[1].Hex())
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.Equal(t, 1, len(body))
		})
	})

	var Targets dbtypes.AddressList = holders
	var Values dbtypes.BigIntList = []*big.Int{}
	var Signatures dbtypes.StringList = []string{}
	var Calldatas dbtypes.BytesList = [][]byte{}
	now := uint64(time.Now().Unix())
	curStart, curEnd := now-1, now+100
	pastStart, pastEnd := now-200, now-100
	// DB 데이터 저장 (proposal)
	require.NoError(t, db.Create(&dbtypes.GovernorProposal{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("1")),
			Block:  1,
		},
		Active:      true,
		ProposalId:  (*dbtypes.BigInt)(big.NewInt(1)),
		Proposer:    holders[0],
		Targets:     &Targets,
		Values:      &Values,
		Signatures:  &Signatures,
		Calldatas:   &Calldatas,
		VoteStart:   pastStart,
		VoteEnd:     pastEnd,
		Description: strings.Repeat("a", 4096),
	}).Error)
	require.NoError(t, db.Create(&dbtypes.GovernorProposal{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("2")),
			Block:  2,
		},
		Active:      false,
		ProposalId:  (*dbtypes.BigInt)(big.NewInt(2)),
		Proposer:    holders[0],
		Targets:     &Targets,
		Values:      &Values,
		Signatures:  &Signatures,
		Calldatas:   &Calldatas,
		VoteStart:   pastStart,
		VoteEnd:     pastEnd,
		Description: strings.Repeat("a", 4096),
	}).Error)
	require.NoError(t, db.Create(&dbtypes.GovernorProposal{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("3")),
			Block:  3,
		},
		Active:      true,
		ProposalId:  (*dbtypes.BigInt)(big.NewInt(3)),
		Proposer:    holders[0],
		Targets:     &Targets,
		Values:      &Values,
		Signatures:  &Signatures,
		Calldatas:   &Calldatas,
		VoteStart:   curStart,
		VoteEnd:     curEnd,
		Description: strings.Repeat("a", 4096),
	}).Error)
	require.NoError(t, db.Create(&dbtypes.GovernorProposal{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("4")),
			Block:  4,
		},
		Active:      false,
		ProposalId:  (*dbtypes.BigInt)(big.NewInt(4)),
		Proposer:    holders[0],
		Targets:     &Targets,
		Values:      &Values,
		Signatures:  &Signatures,
		Calldatas:   &Calldatas,
		VoteStart:   curStart,
		VoteEnd:     curEnd,
		Description: strings.Repeat("a", 4096),
	}).Error)
	require.NoError(t, db.Create(&dbtypes.GovernorProposal{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("5")),
			Block:  5,
		},
		Active:      true,
		ProposalId:  (*dbtypes.BigInt)(big.NewInt(5)),
		Proposer:    holders[1],
		Targets:     &Targets,
		Values:      &Values,
		Signatures:  &Signatures,
		Calldatas:   &Calldatas,
		VoteStart:   curStart,
		VoteEnd:     curEnd,
		Description: strings.Repeat("a", 4096),
	}).Error)
	require.NoError(t, db.Create(&dbtypes.GovernorVoteCast{
		Raw: dbtypes.Raw{
			TxHash: common.BytesToHash([]byte("6")),
			Block:  6,
		},
		Voter:      holders[0],
		ProposalId: (*dbtypes.BigInt)(big.NewInt(5)),
		Support:    0,
		Weight:     (*dbtypes.BigInt)(big.NewInt(5)),
		Reason:     strings.Repeat("a", 1024),
	}).Error)
	t.Run("GovernorApi", func(t *testing.T) {
		t.Run("proposals", func(t *testing.T) {
			t.Run("/", func(t *testing.T) {
				api := "/proposals"
				status, body, err := GetRequest[[]dbtypes.GovernorProposal](api)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 5, len(body))
				require.True(t, reflect.DeepEqual(holders, body[0].Targets.Get()))
			})
			t.Run("/voteable-items", func(t *testing.T) {
				api := "/proposals/voteable-items"
				status, _, err := GetRequest[[]dbtypes.GovernorProposal](api)
				require.Error(t, err)
				require.Equal(t, http.StatusNotFound, status)

				status, body, err := GetRequest[[]dbtypes.GovernorProposal](api + "/" + holders[0].Hex())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 1, len(body))
				require.Equal(t, big.NewInt(3), body[0].ProposalId.Get())
			})
			t.Run("/executable-items", func(t *testing.T) {
				api := "/proposals/executable-items"
				status, body, err := GetRequest[[]dbtypes.GovernorProposal](api)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 1, len(body))
				require.Equal(t, big.NewInt(1), body[0].ProposalId.Get())
			})
		})
		t.Run("/votes", func(t *testing.T) {
			t.Run("/", func(t *testing.T) {
				api := "/votes"
				status, _, err := GetRequest[[]dbtypes.GovernorVoteCast](api)
				require.Error(t, err)
				require.Equal(t, http.StatusNotFound, status)
			})
			t.Run("/history", func(t *testing.T) {
				api := "/votes/history/"
				status, _, err := GetRequest[[]dbtypes.GovernorVoteCast](api)
				require.Error(t, err)
				require.Equal(t, http.StatusNotFound, status)

				status, body, err := GetRequest[[]dbtypes.GovernorVoteCast](api + holders[0].Hex())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 1, len(body))

				status, body, err = GetRequest[[]dbtypes.GovernorVoteCast](api + holders[1].Hex())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, status)
				require.Equal(t, 0, len(body))
			})
		})
	})
	require.NoError(t, srv.Shutdown(context.TODO()))
}

func GetRequest[T any](api string) (int, T, error) {
	data := new(T)

	response, err := http.Get("http://localhost:8080/test" + api)
	if err != nil {
		return 0, *data, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return response.StatusCode, *data, errors.New("404 page not found")
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, *data, err
	}

	body := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &body); err != nil {
		return response.StatusCode, *data, err
	}
	if message, ok := body["error"]; ok {
		return response.StatusCode, *data, errors.New(message.(string))
	} else {
		if dataBytes, err := json.Marshal(body["data"]); err != nil {
			return response.StatusCode, *data, err
		} else {
			return response.StatusCode, *data, json.Unmarshal(dataBytes, data)
		}
	}
}
