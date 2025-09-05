// Package user_usecase implementa a CAMADA DE APLICAÇÃO da Clean Architecture
// Esta camada contém as REGRAS DE NEGÓCIO e orquestra as operações
// É equivalente aos "services" no Node.js, mas mais estruturado
package user_usecase

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/user_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

// UserUseCase é a struct que implementa as regras de negócio para usuários
// Ela DEPENDE da abstração (interface), não da implementação concreta
// Este é o princípio da INVERSÃO DE DEPENDÊNCIA
type UserUseCase struct {
	// UserRepository é a interface, não a implementação concreta
	// Isso permite injetar diferentes implementações (MongoDB, PostgreSQL, Mock para testes)
	UserRepository user_entity.UserRepositoryInterface
}

// UserOutputDTO (Data Transfer Object) define como os dados do usuário serão expostos
// É diferente da entidade User - este é formatado para a API REST
// DTO separa representação interna (entidade) da externa (API)
// No Node.js seria como ter um "serializer" ou "transformer"
type UserOutputDTO struct {
	Id   string `json:"id"`   // Campo "id" no JSON de resposta
	Name string `json:"name"` // Campo "name" no JSON de resposta
}

func NewUserUseCase(userRepository user_entity.UserRepositoryInterface) UserUseCaseInterface {
	return &UserUseCase{
		userRepository,
	}
}

// UserUseCaseInterface define o CONTRATO dos casos de uso de usuário
// Interfaces em Go são implícitas - qualquer tipo que implementar estes métodos satisfaz a interface
// Facilita testes e permite múltiplas implementações
type UserUseCaseInterface interface {
	// FindUserById é o caso de uso para buscar usuário por ID
	// Retorna DTO (não a entidade) para controlar o que é exposto
	FindUserById(ctx context.Context, id string) (*UserOutputDTO, *internal_error.InternalError)
	CreateUser(ctx context.Context, userInput UserInputDTO) (*UserOutputDTO, *internal_error.InternalError)
}

// FindUserById implementa o caso de uso de busca de usuário
// METHOD RECEIVER "(uc *UserUseCase)" vincula este método à struct UserUseCase
// Esta função orquestra a operação: chama repository, trata erros, converte para DTO
func (uc *UserUseCase) FindUserById(ctx context.Context, id string) (*UserOutputDTO, *internal_error.InternalError) {
	// Chama o repository através da interface (DEPENDÊNCIA INVERTIDA)
	// Não sabemos se é MongoDB, PostgreSQL, etc. - só sabemos que implementa a interface
	user, err := uc.UserRepository.FindUserById(ctx, id)

	// Se houver erro, propaga para cima (controller/handler)
	if err != nil {
		return nil, err
	}

	// Converte a entidade User para UserOutputDTO
	// Esta conversão garante que apenas os dados necessários sejam expostos na API
	// É como fazer um "user.toJSON()" customizado no Node.js
	return &UserOutputDTO{
		Id:   user.Id,
		Name: user.Name,
	}, nil
}

/*
CLEAN ARCHITECTURE - Fluxo das camadas:

1. CONTROLLER (HTTP) -> 2. USE CASE (regras negócio) -> 3. REPOSITORY (dados) -> 4. DATABASE

Node.js tradicional:
router.get('/users/:id', async (req, res) => {
    const user = await UserModel.findById(req.params.id); // Direto no controller
    res.json(user);
});

Go com Clean Architecture:
1. Controller recebe HTTP request
2. Controller chama UseCase.FindUserById()
3. UseCase chama Repository.FindUserById()
4. Repository busca no MongoDB
5. Repository retorna User entity
6. UseCase converte para DTO
7. Controller retorna JSON

BENEFÍCIOS:
- Testabilidade (mock das interfaces)
- Separação de responsabilidades
- Independência de framework/banco
- Fácil manutenção e evolução
*/
