package bid_usecase

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/bid_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

type BidInputDTO struct {
	UserId    string  `json:"user_id"`
	AuctionId string  `json:"auction_id"`
	Amount    float64 `json:"amount"`
}
type BidOutputDTO struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	AuctionId string    `json:"auction_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp" time_format:"2006-01-02 15:04:05"`
}

// BidUseCase implementa BATCH PROCESSING com CHANNELS
type BidUseCase struct {
	BidRepository       bid_entity.BidEntityRepository
	timer               *time.Timer         // Timer para flush periódico
	maxBatchSize        int                 // Tamanho máximo do batch
	batchInsertInterval time.Duration       // Intervalo entre flushes
	bidChannel          chan bid_entity.Bid // CHANNEL para comunicação entre goroutines
}

func NewBidUseCase(bidRepository bid_entity.BidEntityRepository) BidUseCaseInterface {
	maxSizeInterval := getMaxBatchSizeInterval()
	maxBatchSize := getMaxBatchSize()

	bidUseCase := &BidUseCase{
		BidRepository:       bidRepository,
		maxBatchSize:        maxBatchSize,
		batchInsertInterval: maxSizeInterval,
		timer:               time.NewTimer(maxSizeInterval),
		// BUFFERED CHANNEL - pode armazenar N elementos sem bloquear
		// Similar a uma queue com capacidade limitada
		bidChannel: make(chan bid_entity.Bid, maxBatchSize),
	}

	// Inicia goroutine de processamento em background
	bidUseCase.triggerCreateRoutine(context.Background())

	return bidUseCase
}

type BidUseCaseInterface interface {
	CreateBid(ctx context.Context, bidInputDto BidInputDTO) *internal_error.InternalError
	FindBidByAuctionId(ctx context.Context, auctionId string) ([]BidOutputDTO, *internal_error.InternalError)
	FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*BidOutputDTO, *internal_error.InternalError)
}

// Variável GLOBAL para batch atual (shared entre goroutines)
var bidBatch []bid_entity.Bid

// triggerCreateRoutine roda em background processando lances em batches
// Esta é uma GOROUTINE DE LONGA DURAÇÃO (long-running goroutine)
func (bu *BidUseCase) triggerCreateRoutine(ctx context.Context) {
	// defer close() garante que channel seja fechado ao sair
	go func() {
		defer close(bu.bidChannel)

		// LOOP INFINITO processando eventos
		for {
			// SELECT - similar ao switch, mas para channels
			// Espera até um dos cases estar pronto
			select {
			// CASE 1: Recebeu novo lance do channel
			case bidEntity, ok := <-bu.bidChannel:
				// ok = false significa que channel foi fechado
				if !ok {
					// Flush final dos lances restantes
					if len(bidBatch) > 0 {
						if err := bu.BidRepository.CreateBidBatch(ctx, bidBatch); err != nil {
							logger.Error("[A] error trying to create bid batch on goroutine", err)
						}
					}
					return // Termina goroutine
				}

				// Adiciona lance ao batch atual
				bidBatch = append(bidBatch, bidEntity)

				// Se batch atingiu tamanho máximo, processa imediatamente
				if len(bidBatch) >= bu.maxBatchSize {
					if err := bu.BidRepository.CreateBidBatch(ctx, bidBatch); err != nil {
						logger.Error("[B] error trying to create bid batch on goroutine", err)
					}
					// bidBatch = []bid_entity.Bid{}
					// Limpa batch (bidBatch = nil é mais eficiente que slice vazio)
					bidBatch = nil
					// Reset timer para próximo intervalo
					bu.timer.Reset(bu.batchInsertInterval)
				}

				// CASE 2: Timer expirou (intervalo de tempo passou)
			case <-bu.timer.C:
				// Processa batch atual mesmo que não esteja cheio
				if err := bu.BidRepository.CreateBidBatch(ctx, bidBatch); err != nil {
					logger.Error("[C] error trying to create bid batch on goroutine", err)
				}
				// bidBatch = []bid_entity.Bid{}
				bidBatch = nil
				bu.timer.Reset(bu.batchInsertInterval)
			}
		}

	}()
}

// CreateBid é ASSÍNCRONO - não espera processamento completar
func (bu *BidUseCase) CreateBid(ctx context.Context, bidInputDto BidInputDTO) *internal_error.InternalError {
	// Cria entidade de lance
	bidEntity, err := bid_entity.CreateBid(bidInputDto.UserId, bidInputDto.AuctionId, bidInputDto.Amount)
	if err != nil {
		return err
	}

	// ENVIA para channel (operação não-bloqueante se channel tem buffer)
	// Equivale a uma queue.push() assíncrono
	bu.bidChannel <- *bidEntity
	// Retorna IMEDIATAMENTE - não espera processamento
	return nil
}

/*
PADRÕES DE CONCORRÊNCIA AVANÇADOS:

1. CHANNELS:
  - Pipes para comunicação entre goroutines
  - make(chan Type, buffer) cria channel buffered
  - channel <- value envia valor
  - value := <-channel recebe valor

2. SELECT STATEMENT:
  - Multiplexing de channels
  - Similar ao switch, mas para operações de channel
  - Não-bloqueante se tem default case

3. BATCH PROCESSING:
  - Agrupa operações para melhor performance
  - Flush por tamanho OU por tempo
  - Muito usado em sistemas de alta performance

4. LONG-RUNNING GOROUTINES:
  - Goroutines que rodam durante toda vida da aplicação
  - Processam eventos em background
  - Similar a workers/background jobs

FLUXO DO SISTEMA:
1. Cliente envia POST /bid
2. Controller chama UseCase.CreateBid()
3. UseCase envia bid para channel (retorna imediatamente)
4. Background goroutine processa batch quando:
  - Batch atinge tamanho máximo OU
  - Timer expira
5. Repository processa batch concorrentemente
6. Múltiplos lances são inseridos em paralelo

BENEFÍCIOS:
- Alta throughput (milhares de lances/segundo)
- Baixa latência (resposta imediata)
- Eficiência (batch inserts)
- Tolerância a picos de tráfego
*/

func getMaxBatchSizeInterval() time.Duration {
	batchInsertInterval := os.Getenv("BATCH_INSERT_INTERVAL")
	duration, err := time.ParseDuration(batchInsertInterval)

	if err != nil {
		return 3 * time.Minute
	}

	return duration
}

func getMaxBatchSize() int {
	batchSize := os.Getenv("MAX_BATCH_SIZE")
	batchSizeInt, err := strconv.Atoi(batchSize)
	if err != nil {
		return 5
	}
	return batchSizeInt
}
