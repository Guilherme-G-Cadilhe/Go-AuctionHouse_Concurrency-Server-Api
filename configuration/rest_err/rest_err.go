// Package rest_err define estruturas customizadas para tratamento de erros HTTP
// É uma abordagem similar ao criar classes de erro customizadas no Node.js
package rest_err

import (
	"net/http"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
)

// RestErr é uma struct que representa um erro estruturado para APIs REST
// Em Go, structs são similares a classes/objetos, mas sem herança
// As tags `json:"..."` definem como os campos serão serializados para JSON
type RestErr struct {
	Message string   `json:"message"` // Mensagem principal do erro
	Err     string   `json:"err"`     // Tipo/categoria do erro
	Code    int      `json:"code"`    // Código HTTP do erro
	Causes  []Causes `json:"causes"`  // Array de causas específicas (para validação)
}

// Causes representa erros específicos de campos (útil para validação de formulários)
// Similar a ter um array de erros de validação no Node.js
type Causes struct {
	Field   string `json:"field"`   // Nome do campo que causou erro
	Message string `json:"message"` // Mensagem específica do erro do campo
}

// Error() faz RestErr implementar a interface error nativa do Go
// É similar a implementar toString() ou valueOf() no JavaScript
// Qualquer tipo que tenha o método Error() string é considerado um error
func (r *RestErr) Error() string {
	return r.Message
}

// ConvertErrors converte erros internos da aplicação para erros HTTP
// Esta função faz a PONTE entre a camada de domínio e a camada de apresentação
// Parâmetro:
//   - internalError: Erro da camada de domínio/aplicação
//
// Retorna:
//   - *RestErr: Erro formatado para HTTP response
func ConvertErrors(internalError *internal_error.InternalError) *RestErr {
	// Switch baseado no tipo de erro interno
	// Mapeia erros de domínio para códigos HTTP apropriados
	switch internalError.Err {
	case "bad_request":
		// Erro de validação/dados inválidos -> 400 Bad Request
		return NewBadRequestError(internalError.Error())
	case "not_found":
		// Recurso não encontrado -> 404 Not Found
		return NewNotFoundError(internalError.Error())
	default:
		// Qualquer outro erro -> 500 Internal Server Error
		// Fallback seguro para erros inesperados
		return NewInternalServerError(internalError.Error())
	}
}

/*
PADRÃO DE CONVERSÃO DE ERROS:

Este padrão é crucial na Clean Architecture porque:

1. SEPARAÇÃO DE RESPONSABILIDADES:
  - Domínio: define tipos de erro de negócio
  - Infraestrutura: define códigos HTTP específicos

2. ABSTRAÇÃO:
  - UseCase retorna erro de domínio
  - Controller converte para erro HTTP

3. REUTILIZAÇÃO:
  - Mesmo erro de domínio pode virar diferentes códigos HTTP
  - Ex: "not_found" pode ser 404 na API REST, mas 204 em GraphQL

FLUXO:
Repository -> InternalError -> UseCase -> InternalError -> Controller -> RestErr -> JSON
*/

// NewBadRequestError é uma função factory para criar erros de Bad Request (400)
// Em Go, é comum usar funções New* para criar instâncias (como construtores)
// Parâmetro:
//   - message string: Mensagem customizada do erro
//
// Retorna:
//   - *RestErr: Ponteiro para a struct de erro criada
func NewBadRequestError(message string, causes ...Causes) *RestErr {
	// &RestErr{} cria uma nova instância e retorna seu endereço (ponteiro)
	// Similar ao new RestErr() no JavaScript, mas retornando referência
	return &RestErr{
		Message: message,               // Mensagem customizada passada
		Err:     "bad_request",         // Identificador do tipo de erro
		Code:    http.StatusBadRequest, // 400 - constante do pacote http
		Causes:  causes,
	}
}

// NewInternalServerError cria erros de servidor interno (500)
// Usado quando algo deu errado no servidor, não por culpa do cliente
func NewInternalServerError(message string) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "internal_server",
		Code:    http.StatusInternalServerError, // 500
		Causes:  nil,
	}
}

// NewNotFoundError cria erros de recurso não encontrado (404)
// Usado quando um recurso solicitado não existe
func NewNotFoundError(message string) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "not_found",
		Code:    http.StatusNotFound, // 404
		Causes:  nil,
	}
}

/*
EXEMPLO de uso comparado ao Node.js:

Node.js com classes:
class RestError extends Error {
    constructor(message, code, type) {
        super(message);
        this.code = code;
        this.type = type;
    }
}

const error = new RestError("User not found", 404, "not_found");

Go:
err := rest_err.NewNotFoundError("User not found")
// err.Code = 404
// err.Err = "not_found"
// err.Message = "User not found"

No handler HTTP, você pode fazer:
if err != nil {
    c.JSON(err.Code, err) // Gin framework example
    return
}
*/
