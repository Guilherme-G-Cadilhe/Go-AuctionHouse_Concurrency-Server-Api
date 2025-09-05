// Package auction_entity define a entidade de domínio Auction e suas regras de negócio
// Esta é a CAMADA DE DOMÍNIO - contém as regras fundamentais do leilão
package auction_entity

import (
	"context"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"github.com/google/uuid" // Biblioteca para gerar UUIDs únicos
)

// CreateAuctionBody é uma FUNÇÃO FACTORY para criar uma nova instância de Auction
// Este é um padrão comum em Go - função que cria e valida entidades
// No Node.js seria como um método estático ou um constructor com validação
func CreateAuctionBody(
	productName string,
	category string,
	description string,
	condition ProductCondition) (*Auction, *internal_error.InternalError) {

	// Cria uma nova instância de Auction com valores iniciais
	auction := &Auction{
		Id:          uuid.New().String(),
		ProductName: productName,
		Category:    category,
		Description: description,
		Condition:   condition,
		Status:      Active,     // Todo leilão inicia como "Active"
		Timestamp:   time.Now(), // Timestamp de criação
	}

	// Valida a entidade antes de retornar
	// Se inválida, retorna erro sem criar o objeto
	err := auction.Validate()
	if err != nil {
		return nil, err
	}

	return auction, nil
}

// Validate é um METHOD da struct Auction que valida suas regras de negócio
// "(au *Auction)" é o METHOD RECEIVER - vincula o método à struct
// Este método implementa as REGRAS DE DOMÍNIO da entidade
func (au *Auction) Validate() *internal_error.InternalError {
	if len(au.ProductName) <= 1 || len(au.Category) <= 2 || len(au.Description) <= 10 && (au.Condition != New && au.Condition != Used && au.Condition != Refurbished) {
		return internal_error.NewBadRequestError("invalid data")
	}
	return nil
}

// Auction é a ENTIDADE PRINCIPAL de domínio para leilões
// Define a estrutura de dados e comportamentos de um leilão
type Auction struct {
	Id          string           `json:"id"` // UUID único do leilão
	ProductName string           `json:"product_name"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Condition   ProductCondition `json:"condition"` // Estado do produto (enum)
	Status      AuctionStatus    `json:"status"`    // Status do leilão (enum)
	Timestamp   time.Time        // Data/hora de criação (sem tag JSON - não exposto na API)
}

// ProductCondition é um TIPO CUSTOMIZADO baseado em int
// Em Go, podemos criar tipos baseados em tipos primitivos
// É similar aos enums do TypeScript/Java
type ProductCondition int
type AuctionStatus int

// Constantes que definem os valores válidos para AuctionStatus
// "iota" é um identificador especial do Go que gera valores sequenciais
// Active = 0, Completed = 1
const (
	Active    AuctionStatus = iota // 0 - Leilão ativo
	Completed                      // 1 - Leilão finalizado
)

// Constantes para ProductCondition
// New = 0, Used = 1, Refurbished = 2
const (
	New         ProductCondition = iota // 0 - Produto novo
	Used                                // 1 - Produto usado
	Refurbished                         // 2 - Produto recondicionado
)

// AuctionRepositoryInterface define o CONTRATO para persistência de leilões
// Interface na camada de domínio = independente de implementação (MongoDB, PostgreSQL, etc.)
type AuctionRepositoryInterface interface {
	// CreateAuction persiste um novo leilão no banco
	CreateAuction(ctx context.Context, auction *Auction) *internal_error.InternalError
	// FindAuctionById busca leilão por ID específico
	FindAuctionById(ctx context.Context, id string) (*Auction, *internal_error.InternalError)
	// FindAllAuctions busca leilões com filtros opcionais
	// Se os filtros forem vazios/zero, busca todos
	FindAllAuctions(
		ctx context.Context,
		status AuctionStatus,
		category, productName string) ([]Auction, *internal_error.InternalError) // Retorna slice de leilões
}

/*
CONCEITOS IMPORTANTES:

1. ENUMS em Go (iota):
No Node.js:
enum AuctionStatus {
    ACTIVE = 0,
    COMPLETED = 1
}

No Go:
type AuctionStatus int
const (
    Active AuctionStatus = iota
    Completed
)

2. FACTORY FUNCTIONS:
No Node.js:
class Auction {
    static create(productName, category, ...) {
        const auction = new Auction(...);
        auction.validate();
        return auction;
    }
}

No Go:
func CreateAuctionBody(...) (*Auction, *internal_error.InternalError)

3. METHOD RECEIVERS:
No Node.js:
class Auction {
    validate() { ... }
}

No Go:
func (au *Auction) Validate() *internal_error.InternalError
*/
