package scan

import (
	"net/http"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FaucetApi struct {
	db *gorm.DB
}

func NewFaucetApi(db *gorm.DB) *FaucetApi {
	return &FaucetApi{db}
}

func (api *FaucetApi) RegisterApi(engine *gin.RouterGroup) error {
	group := engine.Group("/faucet")
	{
		group.GET("/history/:addr", checkParamAddress, api.history)
	}
	return nil
}

func (api *FaucetApi) history(ctx *gin.Context) {
	address, _ := ctx.Get("address")
	var result []dbtypes.FaucetClaimed
	err := api.db.WithContext(ctx).
		Where("account = ?", address).
		Find(&result).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}
