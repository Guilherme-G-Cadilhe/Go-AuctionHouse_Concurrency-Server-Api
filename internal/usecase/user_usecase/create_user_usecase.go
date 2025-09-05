// internal/usecase/user_usecase/create_user_usecase.go
package user_usecase

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/user_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

// DTO para input de criação
type UserInputDTO struct {
	Name string `json:"name" binding:"required"` // binding:"required" = validação obrigatória
}

// CreateUser implementa criação de usuário
func (uc *UserUseCase) CreateUser(ctx context.Context, userInput UserInputDTO) (*UserOutputDTO, *internal_error.InternalError) {
	// Cria entidade usando factory function
	user := user_entity.CreateUser(userInput.Name)

	// Chama repository para persistir
	err := uc.UserRepository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Retorna DTO do usuário criado
	return &UserOutputDTO{
		Id:   user.Id,
		Name: user.Name,
	}, nil
}
