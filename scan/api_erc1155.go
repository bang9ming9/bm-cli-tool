package scan

import (
	"math/big"
	"net/http"

	"github.com/bang9ming9/bm-cli-tool/scan/dbtypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ERC1155Api struct {
	db *gorm.DB
}

func NewERC1155Api(db *gorm.DB) *ERC1155Api {
	return &ERC1155Api{db}
}

func (api *ERC1155Api) RegisterApi(engine *gin.RouterGroup) error {
	group := engine.Group("/erc1155")
	{
		group.GET("/holders/:tid", api.holders)
		group.GET("/history/:addr", checkParamAddress, api.history)
	}
	return nil
}

func (api *ERC1155Api) holders(ctx *gin.Context) {
	tid, ok := ctx.Params.Get("tid")
	if tid == "" || !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tid"})
	} else if tokenID, ok := new(big.Int).SetString(tid, 0); !ok || tokenID.Sign() == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "tid is not number format"})
	} else {
		var result []common.Address
		err := api.db.WithContext(ctx).Model(&dbtypes.ERC1155Transfer{}).
			Where("id = ?", (*dbtypes.BigInt)(tokenID)).
			Select("DISTINCT _to").Pluck("_to", &result).Error
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"data": result})
		}
	}
}

func (api *ERC1155Api) history(ctx *gin.Context) {
	address, _ := ctx.Get("address")
	var result []dbtypes.ERC1155Transfer
	err := api.db.WithContext(ctx).
		Where("_from = ?", address).Or("_to = ?", address).Or("operator = ?", address).
		Find(&result).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"data": result})
	}

}
