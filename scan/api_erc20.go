package scan

import (
	"net/http"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ERC20Api struct {
	db *gorm.DB
}

func NewERC20Api(db *gorm.DB) *ERC20Api {
	return &ERC20Api{db}
}

func (api *ERC20Api) RegisterApi(engine *gin.RouterGroup) error {
	group := engine.Group("/erc20")
	{
		group.GET("/holders", api.holders)
		group.GET("/history/:addr", checkParamAddress, api.history)
	}
	return nil
}

func (api *ERC20Api) holders(ctx *gin.Context) {
	var result []common.Address
	err := api.db.WithContext(ctx).
		Model(&dbtypes.ERC20Transfer{}).Select("DISTINCT _to").Pluck("_to", &result).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}

func (api *ERC20Api) history(ctx *gin.Context) {
	address, _ := ctx.Get("address")
	var result []dbtypes.ERC20Transfer
	err := api.db.WithContext(ctx).
		Where("_from = ?", address).Or("_to = ?", address).
		Find(&result).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}
}
