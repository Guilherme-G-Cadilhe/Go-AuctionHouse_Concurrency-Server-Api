// internal/infra/api/web/controller/user_controller/create_user_controller.go
package user_controller

import (
	"context"
	"net/http"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/rest_err"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/usecase/user_usecase"
	"github.com/gin-gonic/gin"
)

// CreateUser é o handler HTTP para criar usuário
// POST /users com JSON {"name": "João"}
func (u *UserController) CreateUser(c *gin.Context) {
	var userInput user_usecase.UserInputDTO

	// c.ShouldBindJSON() faz parse do JSON e valida automaticamente
	// Se JSON for inválido ou "name" estiver vazio, retorna erro
	if err := c.ShouldBindJSON(&userInput); err != nil {
		errRest := rest_err.NewBadRequestError("Invalid JSON body")
		c.JSON(errRest.Code, errRest)
		return
	}

	// Chama UseCase para criar usuário
	user, err := u.userUseCase.CreateUser(context.Background(), userInput)
	if err != nil {
		errRest := rest_err.ConvertErrors(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	// Retorna usuário criado com status 201 (Created)
	c.JSON(http.StatusCreated, user)
}
