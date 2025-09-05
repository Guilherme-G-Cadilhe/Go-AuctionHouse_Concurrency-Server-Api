package auction

import (
	"context"
	"fmt"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/auction_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // Para regex e outras operações BSON
)

// FindAuctionById busca um leilão específico por ID
func (ar *AuctionRepository) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	// Cria instância vazia para receber os dados do MongoDB
	auctionEntityMongo := &AuctionEntityMongo{}

	// Busca documento por "_id" e decodifica para a struct
	err := ar.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(auctionEntityMongo)
	if err != nil {
		logger.Error(fmt.Sprintf("error trying to find auction by id %s", id), err)
		return nil, internal_error.NewNotFoundError(fmt.Sprintf("error trying to find auction by id %s", id))
	}

	// CONVERSÃO: Modelo de persistência -> Entidade de domínio
	auction := &auction_entity.Auction{
		Id:          auctionEntityMongo.Id,
		ProductName: auctionEntityMongo.ProductName,
		Category:    auctionEntityMongo.Category,
		Description: auctionEntityMongo.Description,
		Condition:   auctionEntityMongo.Condition,
		Status:      auctionEntityMongo.Status,
		// time.Unix() converte int64 Unix timestamp de volta para time.Time
		Timestamp: time.Unix(auctionEntityMongo.Timestamp, 0),
	}

	return auction, nil
}

// FindAllAuctions busca múltiplos leilões com filtros opcionais
func (ar *AuctionRepository) FindAllAuctions(
	ctx context.Context,
	status auction_entity.AuctionStatus,
	category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {

	// bson.M{} é um Map vazio que será populado com filtros
	// É equivalente a um objeto JavaScript: {}
	filter := bson.M{}

	// FILTROS CONDICIONAIS - só adiciona se valor não for vazio/zero

	// Se status não for zero (Active = 0), adiciona filtro por status
	// Em Go, zero values: int = 0, string = "", bool = false, etc.
	if status != 0 {
		filter["status"] = status
	}

	// Se categoria não estiver vazia, adiciona filtro exato
	if category != "" {
		filter["category"] = category
	}

	// Se productName não estiver vazio, adiciona filtro com REGEX (case-insensitive)
	if productName != "" {
		filter["product_name"] = primitive.Regex{
			Pattern: productName, // Padrão de busca
			Options: "i",         // "i" = case insensitive (MongoDB)
		}
	}

	// Slice vazio para receber os documentos do MongoDB
	// var slice []Type cria slice vazio (similar ao [] no JavaScript)
	var auctions []AuctionEntityMongo

	// Find() retorna um CURSOR (não os dados diretamente)
	// Cursor é como um iterator - permite processar grandes volumes de dados
	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("error trying to find auctions", err)
		return nil, internal_error.NewInternalServerError("error trying to find auctions")
	}

	// defer garante que cursor.Close() seja executado ao final da função
	// É como um try/finally - essencial para liberar recursos
	defer cursor.Close(ctx)

	// cursor.All() lê TODOS os documentos do cursor de uma vez
	// &auctions passa o endereço do slice para ser preenchido
	if err = cursor.All(ctx, &auctions); err != nil {
		logger.Error("error trying to decode auctions", err)
		return nil, internal_error.NewInternalServerError("error trying to decode auctions")
	}

	// CONVERSÃO: Slice de modelos MongoDB -> Slice de entidades de domínio
	auctionsEntities := []auction_entity.Auction{} // Slice vazio de entidades

	// range itera sobre o slice (como for...of no JavaScript)
	// "_" ignora o índice, auction é o valor atual
	for _, auction := range auctions {
		// append() adiciona elemento ao slice (como push() no JavaScript)
		auctionsEntities = append(auctionsEntities, auction_entity.Auction{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Description: auction.Description,
			Condition:   auction.Condition,
			Status:      auction.Status,
			Timestamp:   time.Unix(auction.Timestamp, 0), // Unix -> time.Time
		})
	}

	return auctionsEntities, nil
}

/*
CONCEITOS IMPORTANTES:

1. CURSOR vs ARRAY:
Node.js (Mongoose):
const auctions = await AuctionModel.find(filter); // Retorna array diretamente

Go (MongoDB Driver):
cursor, err := collection.Find(ctx, filter) // Retorna cursor
err = cursor.All(ctx, &results)             // Lê cursor para slice

2. FILTROS DINÂMICOS:
Node.js:
const filter = {};
if (status) filter.status = status;
if (category) filter.category = category;

Go:
filter := bson.M{}
if status != 0 { filter["status"] = status }
if category != "" { filter["category"] = category }

3. REGEX no MongoDB:
Node.js:
{ product_name: { $regex: productName, $options: 'i' } }

Go:
primitive.Regex{ Pattern: productName, Options: "i" }
*/
