// Package user implementa a camada de INFRAESTRUTURA para acesso a dados de usuário
// Esta é a CAMADA DE INFRAESTRUTURA da Clean Architecture
// Aqui temos os detalhes de como persistir dados (MongoDB neste caso)
package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/user_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserEntityMongo representa como o User é armazenado no MongoDB
// Esta struct é específica para MongoDB (note as tags `bson`)
// Separamos a entidade de domínio (User) da representação no banco (UserEntityMongo)
// No Node.js com Mongoose, isso seria um Schema
type UserEntityMongo struct {
	Id   string `bson:"_id"`  // Mapeia para o campo "_id" do MongoDB
	Name string `bson:"name"` // Mapeia para o campo "name" do MongoDB
}

// UserRepository é a implementação CONCRETA da UserRepositoryInterface
// Esta struct "implementa" a interface definida na camada de domínio
// Collection é um ponteiro para a coleção do MongoDB
type UserRepository struct {
	Collection *mongo.Collection // Referência para a coleção "users" no MongoDB
}

// NewUserRepository é uma função FACTORY para criar instâncias do UserRepository
// Em Go, é padrão usar funções New* como construtores
// Parâmetros:
//   - database *mongo.Database: Ponteiro para o database MongoDB
//
// Retorna:
//   - *UserRepository: Nova instância configurada com a coleção "users"
func NewUserRepository(database *mongo.Database) *UserRepository {
	return &UserRepository{
		// database.Collection("users") obtém referência para a coleção "users"
		Collection: database.Collection("users"),
	}
}

// FindUserById implementa o método definido na UserRepositoryInterface
// O "(ur *UserRepository)" é um METHOD RECEIVER - vincula este método à struct UserRepository
// É similar aos métodos de classe no Node.js/JavaScript
// Parâmetros e retorno são os mesmos definidos na interface
func (ur *UserRepository) FindUserById(ctx context.Context, id string) (*user_entity.User, *internal_error.InternalError) {
	// bson.M{} cria um filtro MongoDB (equivale ao {_id: id} no MongoDB/Node.js)
	// bson.M é um tipo Map[string]interface{} otimizado para MongoDB
	filter := bson.M{"_id": id}

	// Declara uma variável do tipo UserEntityMongo para receber os dados
	var user UserEntityMongo

	// ur.Collection.FindOne() executa query MongoDB para buscar UM documento
	// .Decode(&user) decodifica o resultado BSON para a struct Go
	// O "&user" passa o ENDEREÇO da variável (ponteiro) para que seja preenchida
	err := ur.Collection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		// errors.Is() verifica se o erro é de um tipo específico
		// mongo.ErrNoDocuments indica que nenhum documento foi encontrado
		// É como verificar se result.length === 0 no Node.js
		if errors.Is(err, mongo.ErrNoDocuments) {
			// fmt.Sprintf() é como template literals ou string interpolation
			logger.Error(fmt.Sprintf("user with id %s not found", id), err)
			// Retorna erro customizado de "not found" (404)
			return nil, internal_error.NewNotFoundError(fmt.Sprintf("user with id %s not found", id))
		}

		// Qualquer outro erro é considerado erro interno do servidor
		logger.Error(fmt.Sprintf("error trying to find user with id %s", id), err)
		// Retorna erro customizado de "internal server error" (500)
		return nil, internal_error.NewInternalServerError(fmt.Sprintf("error trying to find user with id %s", id))
	}

	// Se chegou aqui, encontrou o usuário com sucesso
	// Converte de UserEntityMongo (representação do banco) para User (entidade de domínio)
	// &user_entity.User{} cria uma nova instância e retorna seu ponteiro
	return &user_entity.User{
		Id:   user.Id,
		Name: user.Name,
	}, nil // nil indica que não houve erro
}

/*
PADRÃO REPOSITORY em Go vs Node.js:

Node.js (direto com Mongoose):
const user = await UserModel.findById(id);
if (!user) throw new NotFoundError('User not found');

Go (com Repository Pattern):
1. Interface define O QUE fazer (FindUserById)
2. Implementação define COMO fazer (MongoDB, Postgres, etc.)
3. UseCase usa a interface, não se importa com a implementação
4. Facilita testes (mock da interface) e mudança de banco
*/
