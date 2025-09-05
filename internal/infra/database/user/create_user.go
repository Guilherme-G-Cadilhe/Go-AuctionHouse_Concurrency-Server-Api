package user

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/user_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

// CreateUser insere novo usuário no MongoDB
func (ur *UserRepository) CreateUser(ctx context.Context, user *user_entity.User) *internal_error.InternalError {
	// Converte entidade para modelo MongoDB
	userEntityMongo := &UserEntityMongo{
		Id:   user.Id,
		Name: user.Name,
	}

	// Insere no banco
	_, err := ur.Collection.InsertOne(ctx, userEntityMongo)
	if err != nil {
		logger.Error("Error trying to create user", err)
		return internal_error.NewInternalServerError("error trying to create user")
	}

	return nil // Sucesso
}
