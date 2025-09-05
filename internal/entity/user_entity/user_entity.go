// Package user_entity define a entidade de domínio User e suas interfaces
// Esta é a CAMADA DE DOMÍNIO da Clean Architecture
// Equivale aos "models" ou "entities" no Node.js, mas mais focado em regras de negócio
package user_entity

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

// User representa a entidade de domínio principal para usuários
// Esta struct define APENAS os dados essenciais do usuário
// Diferente do Node.js/Mongoose onde misturamos dados + métodos, aqui separamos
type User struct {
	Id   string // ID único do usuário (sem tags BSON aqui - entidade pura)
	Name string // Nome do usuário
}

// UserRepositoryInterface define o CONTRATO para acesso a dados de usuário
// É o padrão Repository Pattern - abstração sobre como os dados são persistidos
// Em Node.js seria como definir uma interface/classe abstrata para o DAO
type UserRepositoryInterface interface {
	// FindUserById busca um usuário por ID
	// Parâmetros:
	//   - ctx context.Context: Context para timeout/cancelamento
	//   - id string: ID do usuário a ser buscado
	// Retorna:
	//   - *User: Ponteiro para a entidade User (nil se não encontrado)
	//   - *internal_error.InternalError: Erro customizado (nil se sucesso)
	FindUserById(ctx context.Context, id string) (*User, *internal_error.InternalError)
}

/*
IMPORTANTE: Interface vs Implementação em Go

No Node.js fazemos algo como:
class UserRepository {
    async findUserById(id) { ... }
}

Em Go, separamos:
1. INTERFACE (contrato) = UserRepositoryInterface
2. IMPLEMENTAÇÃO (concreta) = UserRepository (no arquivo de infra)

A interface fica na camada de domínio, a implementação na camada de infraestrutura.
Isso permite trocar a implementação sem afetar as regras de negócio.
*/
