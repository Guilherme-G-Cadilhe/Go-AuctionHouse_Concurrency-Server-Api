package auction_controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/rest_err"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/usecase/auction_usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (au *AuctionController) FindAuctionById(c *gin.Context) {
	auctionId := c.Param("auctionId")

	if err := uuid.Validate(auctionId); err != nil {
		errRest := rest_err.NewBadRequestError("invalid fields", rest_err.Causes{
			Field:   "auctionId",          // Campo que causou o erro
			Message: "Invalid UUID Value", // Mensagem específica
		})

		c.JSON(errRest.Code, errRest)
		return
	}

	auction, err := au.auctionUseCase.FindAuctionById(context.Background(), auctionId)
	if err != nil {
		errRest := rest_err.ConvertErrors(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auction)
}

func (au *AuctionController) FindAllAuctions(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")
	productName := c.Query("productName")

	statusNumber, errConv := strconv.Atoi(status)
	if errConv != nil {
		errRest := rest_err.NewBadRequestError("Erro trying to validate auction status param")
		c.JSON(errRest.Code, errRest)
		return
	}

	auctions, err := au.auctionUseCase.FindAllAuctions(context.Background(), auction_usecase.AuctionStatus(statusNumber), category, productName)
	if err != nil {
		fmt.Println(err)
		errRest := rest_err.ConvertErrors(err)
		c.JSON(errRest.Code, errRest)
		return
	}
	//return empty array json if not found actions instead of null
	if len(auctions) == 0 {
		c.JSON(http.StatusOK, []any{})
		return
	}

	c.JSON(http.StatusOK, auctions)
}

func (au *AuctionController) FindWinningBidByAuctionId(c *gin.Context) {
	auctionId := c.Param("auctionId")

	if err := uuid.Validate(auctionId); err != nil {
		errRest := rest_err.NewBadRequestError("invalid fields", rest_err.Causes{
			Field:   "auctionId",          // Campo que causou o erro
			Message: "Invalid UUID Value", // Mensagem específica
		})

		c.JSON(errRest.Code, errRest)
		return
	}

	auction, err := au.auctionUseCase.FindWinningBidByAuctionId(context.Background(), auctionId)
	if err != nil {
		errRest := rest_err.ConvertErrors(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	c.JSON(http.StatusOK, auction)
}
