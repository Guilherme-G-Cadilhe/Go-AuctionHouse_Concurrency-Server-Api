// Package mongodb contém as configurações e funções para conexão com MongoDB
// Diferente do Node.js onde usamos mongoose, aqui usamos o driver oficial do MongoDB para Go
package mongodb

import (
	"context"
	"os"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constantes para as variáveis de ambiente
// Em Go, é uma boa prática usar constantes para strings que não mudam
// Isso evita erros de digitação e facilita manutenção
const (
	MONGODB_URI      = "MONGODB_URI"
	MONGODB_DATABASE = "MONGODB_DATABASE"
)

// NewMongoDBConnection estabelece conexão com MongoDB e retorna uma instância do database
// Parâmetros:
//   - ctx context.Context: Context do Go para controle de timeout/cancelamento (diferente do Node.js)
//
// Retorna:
//   - *mongo.Database: Ponteiro para o database (em Go usamos ponteiros para evitar cópias desnecessárias)
//   - error: Interface de erro do Go (ao invés de try/catch como no Node.js)
func NewMongoDBConnection(ctx context.Context) (*mongo.Database, error) {
	// os.Getenv() busca variável de ambiente (equivale ao process.env do Node.js)
	mongoURI := os.Getenv(MONGODB_URI)
	mongoDatabase := os.Getenv(MONGODB_DATABASE)

	// mongo.Connect() conecta ao MongoDB usando o context
	// options.Client().ApplyURI() configura as opções de conexão
	// Em Go, muitas funções retornam (valor, erro) - padrão da linguagem
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		// Se houver erro, loga usando nosso sistema customizado e retorna
		// Em Go, tratamos erros explicitamente (não há exceções como no Node.js)
		logger.Error("Error connecting to MongoDB", err)
		return nil, err
	}

	// client.Ping() testa se a conexão está funcionando
	// É como fazer um "health check" da conexão
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("Error pinging MongoDB", err)
		return nil, err
	}

	// client.Database() seleciona o database específico
	// Retorna um ponteiro para o database (sucesso) e nil para erro
	return client.Database(mongoDatabase), nil

}
