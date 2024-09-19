package scan

import (
	"context"
	"net/http"
	"time"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GovernorApi struct {
	db *gorm.DB
}

func NewGovernorApi(db *gorm.DB) *GovernorApi {
	return &GovernorApi{db}
}

func (api *GovernorApi) RegisterApi(engine *gin.RouterGroup) error {
	proposals := engine.Group("/proposals")
	{
		proposals.GET("/", api.allProposals)
		proposals.GET("/voteable-items/:addr", checkParamAddress, api.voteableProposals)
		proposals.GET("/executable-items", api.executableProposals)
	}
	votes := engine.Group("/votes")
	{
		votes.GET("/history/:addr", checkParamAddress, api.voteHistory)
	}
	return nil
}

// ////////////
// Proposals //
// ////////////

func (api *GovernorApi) allProposals(ctx *gin.Context) {
	var result []*dbtypes.GovernorProposal
	if err := api.db.WithContext(ctx).Find(&result).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}

func (api *GovernorApi) voteableProposals(ctx *gin.Context) {
	address, _ := ctx.Get("address")
	now := uint64(time.Now().Unix())
	var result []*dbtypes.GovernorProposal
	err := api.db.WithContext(ctx).
		Where("active = ?", true).
		Where("vote_start <= ?", now).
		Where("vote_end > ?", now).
		Where("proposal_id NOT IN (?)",
			api.db.Model(&dbtypes.GovernorVoteCast{}).
				Where("voter = ?", address).
				Select("proposal_id")).
		Find(&result).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}

func (api *GovernorApi) executableProposals(ctx *gin.Context) {
	result, err := pastProposals(ctx, api.db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// ////////
// Votes //
// ////////

func (api *GovernorApi) voteHistory(ctx *gin.Context) {
	address, _ := ctx.Get("address")
	var result []*dbtypes.GovernorVoteCast
	err := api.db.WithContext(ctx).Where("voter = ?", address).Find(&result).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// ///////////////////////
// extracting functions //
// ///////////////////////

func checkParamAddress(ctx *gin.Context) {
	addr, ok := ctx.Params.Get("addr")
	if length := len(addr); !ok || !(length == 40 || length == 42) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid address format"})
	} else if address := common.HexToAddress(addr); address == (common.Address{}) {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "address is zero or not hex string"})
	} else {
		ctx.Set("address", address)
		ctx.Next()
	}
}

func pastProposals(ctx context.Context, db *gorm.DB) ([]*dbtypes.GovernorProposal, error) {
	now := uint64(time.Now().Unix())

	var result []*dbtypes.GovernorProposal
	return result, db.WithContext(ctx).
		Where("active = ?", true).
		Where("vote_end <= ?", now).
		Find(&result).Error
}
