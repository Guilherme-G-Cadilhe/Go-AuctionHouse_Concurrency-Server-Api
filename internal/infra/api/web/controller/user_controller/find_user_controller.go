// Package user_controller implementa os controllers HTTP para operações de usuário
// CAMADA DE INTERFACE/APRESENTAÇÃO - recebe requests HTTP e retorna responses
package user_controller

import (
	"context"
	"net/http"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/rest_err"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/usecase/user_usecase"
	"github.com/gin-gonic/gin" // Framework web similar ao Express.js
	"github.com/google/uuid"   // Para validação de UUIDs
)

// userController é a struct que agrupa os handlers HTTP relacionados a usuário
// Em Go, não temos classes - usamos structs + métodos
// É similar a um controller class no NestJS ou Express
type UserController struct {
	// Injeção de dependência - recebe o useCase via construtor
	// userUseCaseInterface implementa as regras de negócio
	userUseCase user_usecase.UserUseCaseInterface
}

// NewUserController é a função FACTORY para criar instâncias do controller
// Padrão de injeção de dependência manual em Go
// Recebe as dependências como parâmetros
func NewUserController(userUseCase user_usecase.UserUseCaseInterface) *UserController {
	return &UserController{
		userUseCase: userUseCase, // Injeta o useCase
	}
}

// FindUserById é o HANDLER HTTP para buscar usuário por ID
// METHOD RECEIVER "(u *userController)" vincula à struct userController
// gin.Context é similar ao Request/Response do Express.js
func (u *UserController) FindUserById(c *gin.Context) {
	// c.Param() extrai parâmetro da URL
	// Rota: GET /users/:userId -> c.Param("userId") pega o valor
	// É como req.params.userId no Express.js
	userId := c.Param("userId")

	// VALIDAÇÃO DE UUID
	// uuid.Validate() verifica se a string é um UUID válido
	// Evita queries desnecessárias no banco com IDs inválidos
	if err := uuid.Validate(userId); err != nil {
		// rest_err.Causes{} cria uma causa específica para o erro
		errRest := rest_err.NewBadRequestError("invalid fields", rest_err.Causes{
			Field:   "userId",             // Campo que causou o erro
			Message: "Invalid UUID Value", // Mensagem específica
		})

		// c.JSON() retorna resposta JSON com status code
		// Similar a res.status(400).json(errRest) no Express.js
		c.JSON(errRest.Code, errRest)
		return // Para a execução aqui (similar ao return no Express)
	}

	// CHAMA O USE CASE para executar a lógica de negócio
	// context.Background() cria um contexto vazio (sem timeout/cancelamento)
	// Em produção, melhor usar contexto com timeout: c.Request.Context()
	user, err := u.userUseCase.FindUserById(context.Background(), userId)
	if err != nil {
		// ConvertErrors() converte erro interno para erro HTTP
		// Abstrai detalhes internos e expõe apenas o necessário para o cliente
		errRest := rest_err.ConvertErrors(err)
		c.JSON(errRest.Code, errRest)
		return
	}

	// SUCESSO - retorna o usuário encontrado
	// http.StatusOK é a constante para 200
	// user (DTO) já está formatado para JSON
	c.JSON(http.StatusOK, user)
}

/*
FLUXO COMPLETO DE UMA REQUEST:

1. Cliente faz GET /users/123e4567-e89b-12d3-a456-426614174000
2. Gin router chama FindUserById()
3. Controller extrai "userId" da URL
4. Controller valida se é UUID válido
5. Controller chama UseCase.FindUserById()
6. UseCase chama Repository.FindUserById()
7. Repository busca no MongoDB
8. Repository retorna User entity
9. UseCase converte para DTO
10. UseCase retorna DTO para Controller
11. Controller retorna JSON para cliente

COMPARAÇÃO Express.js vs Gin:

Express.js:
app.get('/users/:userId', async (req, res) => {
    try {
        const { userId } = req.params;
        const user = await userService.findById(userId);
        res.json(user);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

Gin (Go):
func (u *userController) FindUserById(c *gin.Context) {
    userId := c.Param("userId")
    user, err := u.userUseCase.FindUserById(ctx, userId)
    if err != nil {
        c.JSON(err.Code, err)
        return
    }
    c.JSON(http.StatusOK, user)
}
*/
