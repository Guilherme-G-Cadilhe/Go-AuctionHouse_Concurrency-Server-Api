package bid

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/auction_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/bid_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/infra/database/auction"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"go.mongodb.org/mongo-driver/mongo"
)

type BidEntityMongo struct {
	Id        string  `bson:"_id"`
	UserId    string  `bson:"user_id"`
	AuctionId string  `bson:"auction_id"`
	Amount    float64 `bson:"amount"`
	Timestamp int64   `bson:"timestamp"`
}

// BidRepository agora possui campos para CONCORRÊNCIA e CACHE
type BidRepository struct {
	Collection        *mongo.Collection
	AuctionRepository *auction.AuctionRepository

	// CACHE MAPS - evitam consultas repetidas ao banco
	auctionStatusMap  map[string]auction_entity.AuctionStatus // Cache do status dos leilões
	auctionEndTimeMap map[string]time.Time                    // Cache do tempo de fim dos leilões

	// MUTEXES - protegem acesso concorrente aos maps
	// sync.Mutex garante que apenas uma goroutine acesse o resource por vez
	auctionStatusMapMutex *sync.Mutex // Protege auctionStatusMap
	auctionEndTimeMutex   *sync.Mutex // Protege auctionEndTimeMap

	auctionInterval time.Duration // Duração padrão dos leilões
}

func NewBidRepository(database *mongo.Database, auctionRepository *auction.AuctionRepository) *BidRepository {
	return &BidRepository{
		auctionInterval: getAuctionInterval(),
		// make() cria maps vazios (similar a {} no JavaScript)
		auctionStatusMap:  make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap: make(map[string]time.Time),
		// &sync.Mutex{} cria novos mutexes
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
		Collection:            database.Collection("bids"),
		AuctionRepository:     auctionRepository,
	}
}

// CreateBidBatch processa múltiplos lances CONCORRENTEMENTE
// Esta é a função mais complexa - usa goroutines + WaitGroup + Mutex
func (bd *BidRepository) CreateBidBatch(ctx context.Context, bidEntities []bid_entity.Bid) *internal_error.InternalError {
	// sync.WaitGroup coordena múltiplas goroutines
	// É como Promise.all() no JavaScript, mas mais flexível
	var wg sync.WaitGroup

	// Itera sobre cada lance no batch
	for _, bid := range bidEntities {
		// wg.Add(1) incrementa o contador de goroutines ativas
		wg.Add(1)

		// GOROUTINE - executa função em paralelo
		// go func() é como criar uma nova thread/processo
		go func(bidValue bid_entity.Bid) {
			// defer wg.Done() decrementa contador quando função termina
			// É executado independente de como a função sai (return, panic, etc.)
			defer wg.Done()

			// === SEÇÃO CRÍTICA 1: Leitura do cache de status ===
			// Lock() garante acesso exclusivo ao map
			bd.auctionStatusMapMutex.Lock()
			auctionStatus, okStatus := bd.auctionStatusMap[bidValue.AuctionId]
			// Unlock() libera o lock imediatamente após uso
			bd.auctionStatusMapMutex.Unlock()

			// === SEÇÃO CRÍTICA 2: Leitura do cache de tempo ===
			bd.auctionEndTimeMutex.Lock()
			auctionEndTime, okEndTime := bd.auctionEndTimeMap[bidValue.AuctionId]
			bd.auctionEndTimeMutex.Unlock()

			// Converte entidade para modelo MongoDB
			bidEntityMongo := &BidEntityMongo{
				Id:        bidValue.Id,
				UserId:    bidValue.UserId,
				AuctionId: bidValue.AuctionId,
				Amount:    bidValue.Amount,
				Timestamp: bidValue.Timestamp.Unix(),
			}

			// CACHE HIT - se temos dados do leilão em cache
			if okEndTime && okStatus {
				now := time.Now()
				// Verifica se leilão já fechou
				if auctionStatus == auction_entity.Completed || now.After(auctionEndTime) {
					return // Lance rejeitado - leilão fechado
				}

				// Lance válido - insere no banco
				if _, err := bd.Collection.InsertOne(ctx, bidEntityMongo); err != nil {
					logger.Error("Error trying to insert bid", err)
					return
				}
				return
			}

			// CACHE MISS - precisa buscar dados do leilão no banco
			auctionEntity, err := bd.AuctionRepository.FindAuctionById(ctx, bidValue.AuctionId)
			if err != nil {
				logger.Error(fmt.Sprintf("error trying to find auction by id %s", bidValue.AuctionId), err)
				return
			}

			// Verifica se leilão está ativo
			if auctionEntity.Status != auction_entity.Active {
				logger.Error(fmt.Sprintf("auction with id %s is not open", bidValue.AuctionId), err)
				return
			}

			// === SEÇÃO CRÍTICA 3: Atualização do cache de status ===
			bd.auctionStatusMapMutex.Lock()
			bd.auctionStatusMap[bidValue.AuctionId] = auctionEntity.Status
			bd.auctionStatusMapMutex.Unlock()

			// === SEÇÃO CRÍTICA 4: Atualização do cache de tempo ===
			bd.auctionEndTimeMutex.Lock()
			// Calcula tempo de fim = timestamp inicial + intervalo
			bd.auctionEndTimeMap[bidValue.AuctionId] = auctionEntity.Timestamp.Add(bd.auctionInterval)
			bd.auctionEndTimeMutex.Unlock()

			// Insere lance válido no banco
			if _, err := bd.Collection.InsertOne(ctx, bidEntityMongo); err != nil {
				logger.Error("error trying to insert bid", err)
				return
			}

		}(bid) // Passa bid como parâmetro para evitar closure issues
	}

	// wg.Wait() bloqueia até todas as goroutines terminarem
	// É como await Promise.all() no JavaScript
	wg.Wait()
	return nil
}

// getAuctionInterval lê configuração de duração dos leilões
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	// time.ParseDuration() converte string para Duration
	// Ex: "5m", "30s", "2h45m"
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // Fallback: 5 minutos
	}
	return duration
}

/*
CONCEITOS DE CONCORRÊNCIA:

1. GOROUTINES:
  - Threads leves do Go (milhares podem rodar simultaneamente)
  - go func() cria nova goroutine
  - Mais eficientes que threads do OS

2. SYNC.WAITGROUP:
  - Coordena múltiplas goroutines
  - Add(n) incrementa contador
  - Done() decrementa contador
  - Wait() bloqueia até contador = 0

3. SYNC.MUTEX:
  - Mutual Exclusion - apenas 1 goroutine por vez
  - Protege recursos compartilhados (maps, variáveis)
  - Lock() + Unlock() define seção crítica

4. RACE CONDITIONS:
  - Problema: múltiplas goroutines acessam mesmo recurso
  - Solução: mutex protege acesso concorrente
  - Go tem detector de race: go run -race main.go

PADRÃO DE CACHE + CONCORRÊNCIA:
- Cache evita consultas repetidas ao banco
- Mutex protege cache de corruption
- Goroutines processam lances em paralelo
- Performance muito superior ao processamento sequencial
*/
