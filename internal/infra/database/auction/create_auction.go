// Package auction implementa a camada de infraestrutura para persistência de leilões
// CAMADA DE INFRAESTRUTURA - detalhes de implementação do MongoDB
package auction

import (
	"context"
	"os"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/auction_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuctionEntityMongo representa como o Auction é armazenado no MongoDB
// Separação entre entidade de domínio (Auction) e modelo de persistência (AuctionEntityMongo)
// Note as diferenças: Timestamp vira int64, tipos mantidos como referência à entidade
type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"` // MongoDB usa "_id" por padrão
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition // Mantém referência ao tipo da entidade
	Status      auction_entity.AuctionStatus    // Mantém referência ao tipo da entidade
	Timestamp   int64                           // MongoDB: timestamp como Unix epoch (int64)
}

// AuctionRepository é a implementação concreta da AuctionRepositoryInterface
// Esta struct "implementa" implicitamente a interface definida na camada de domínio
type AuctionRepository struct {
	Collection *mongo.Collection // Referência para coleção "auctions" do MongoDB
}

// NewAuctionRepository é a função FACTORY para criar instâncias do repository
// Padrão de injeção de dependência manual em Go
func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"), // Define coleção "auctions"
	}
}

// CreateAuction implementa o método da interface AuctionRepositoryInterface
// METHOD RECEIVER "(ar *AuctionRepository)" vincula à struct AuctionRepository
func (ar *AuctionRepository) CreateAuction(ctx context.Context, auction *auction_entity.Auction) *internal_error.InternalError {
	// CONVERSÃO: Entidade de domínio -> Modelo de persistência
	// Este mapeamento é necessário porque:
	// 1. Entidade não deve saber sobre MongoDB
	// 2. MongoDB pode precisar de formato específico (timestamps, etc.)
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auction.Id,
		ProductName: auction.ProductName,
		Category:    auction.Category,
		Description: auction.Description,
		Condition:   auction.Condition,
		Status:      auction.Status,
		// .Unix() converte time.Time para int64 (Unix timestamp)
		// MongoDB armazena melhor como número que como objeto complexo
		Timestamp: auction.Timestamp.Unix(),
	}

	// ar.Collection.InsertOne() insere documento no MongoDB
	// ctx para timeout/cancelamento, auctionEntityMongo é o documento
	// "_" ignora o resultado da inserção (só nos importa com erros)
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		// Retorna erro genérico - não expõe detalhes internos do MongoDB
		return internal_error.NewInternalServerError("error trying to create auction")
	}

	go func() {
		select {
		case <-time.After(getAuctionInterval()):
			update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
			filter := bson.M{"_id": auctionEntityMongo.Id}
			_, err := ar.Collection.UpdateOne(ctx, filter, update)
			if err != nil {
				logger.Error("error trying to update auction to close", err)
				return
			}

		}
	}()

	return nil // Sucesso - sem erro
}

func getAuctionInterval() time.Duration {
	interval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}

/*
PADRÃO ENTITY vs MODEL:

Node.js (Mongoose):
const auctionSchema = new Schema({...});
const auction = new AuctionModel(data);
await auction.save();

Go (Separação clara):
1. auction_entity.Auction (DOMÍNIO - regras de negócio)
2. AuctionEntityMongo (INFRAESTRUTURA - formato MongoDB)
3. Conversão explícita entre eles

BENEFÍCIOS:
- Domínio independente do banco
- Mudança de banco não afeta regras de negócio
- Controle total sobre mapeamento
- Testabilidade (mock da interface)
*/
