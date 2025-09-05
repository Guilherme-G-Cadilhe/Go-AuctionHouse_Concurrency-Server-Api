// Package validation centraliza a lógica de validação de dados da API
// Utiliza a biblioteca "validator" (go-playground/validator) para validações automáticas
// É similar ao Joi, Yup ou class-validator do Node.js
package validation

import (
	"encoding/json"
	"errors"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/rest_err"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	validator_en "github.com/go-playground/validator/v10/translations/en"
)

// Variáveis globais do package para validação e traduções
var (
	// Validate é a instância global do validador
	// Similar a ter uma instância configurada do Joi no Node.js
	Validate = validator.New()

	// transl é o tradutor para mensagens de erro em inglês
	// Converte erros técnicos em mensagens amigáveis
	transl ut.Translator
)

// init() configura o sistema de validação e tradução automaticamente
// Executa quando o package é importado pela primeira vez
func init() {
	// binding.Validator.Engine() obtém o validador usado pelo Gin framework
	// Type assertion (*validator.Validate) verifica se é do tipo correto
	// "ok" indica se a conversão foi bem-sucedida
	if value, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Configura idioma inglês para traduções
		en := en.New()                           // Cria localizador inglês
		enTransl := ut.New(en, en)               // Cria tradutor universal
		transl, _ = enTransl.GetTranslator("en") // Obtém tradutor específico

		// Registra traduções padrão em inglês para as regras de validação
		// Isso faz com que "required" vire "Field is required" automaticamente
		validator_en.RegisterDefaultTranslations(value, transl)
	}
}

// validateErr converte erros de validação para formato padronizado da API
// Esta função trata diferentes tipos de erro que podem ocorrer na validação
func ValidateErr(validation_err error) *rest_err.RestErr {
	// Variáveis para diferentes tipos de erro
	var jsonErr *json.UnmarshalTypeError          // Erro de tipo de JSON (string onde esperava int)
	var jsonValidation validator.ValidationErrors // Erros de validação de regras

	// errors.As() verifica se o erro é de um tipo específico e faz casting
	// É mais seguro que type assertion direta

	// CASO 1: Erro de tipo de dados JSON
	if errors.As(validation_err, &jsonErr) {
		// Ex: mandou "abc" onde esperava um número
		return rest_err.NewBadRequestError("Invalid field type")

		// CASO 2: Erro de validação de regras (required, min, max, etc.)
	} else if errors.As(validation_err, &jsonValidation) {
		// Slice para acumular todos os erros de validação
		errorCauses := []rest_err.Causes{}

		// Itera sobre cada erro de validação individual
		// validation_err.(validator.ValidationErrors) é type assertion forçada
		for _, err := range validation_err.(validator.ValidationErrors) {
			// Cria uma causa específica para cada campo com erro
			cause := rest_err.Causes{
				// err.Translate(transl) converte erro técnico para mensagem amigável
				// Ex: "required" vira "Field is required"
				Message: err.Translate(transl),
				// err.Field() retorna o nome do campo que falhou
				Field: err.Field(),
			}
			// Adiciona esta causa ao slice de causas
			errorCauses = append(errorCauses, cause)
		}

		// Retorna erro com todas as causas específicas
		// O "..." expande o slice como argumentos variádicos
		return rest_err.NewBadRequestError("Validation error", errorCauses...)

		// CASO 3: Qualquer outro tipo de erro
	} else {
		return rest_err.NewBadRequestError("error trying to convert fields")
	}
}

/*
BIBLIOTECA VALIDATOR - Como funciona:

1. TAGS DE VALIDAÇÃO:
type User struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=0,lte=130"`
}

2. VALIDAÇÃO:
err := validator.Validate.Struct(user)
if err != nil {
    // Trata erros de validação
}

3. TRADUÇÕES:
Sem tradução: "Key: 'User.Name' Error: Field validation for 'Name' failed on the 'required' tag"
Com tradução: "Name is required"

COMPARAÇÃO com Node.js:

Joi (Node.js):
const schema = Joi.object({
    name: Joi.string().min(2).max(100).required(),
    email: Joi.string().email().required(),
    age: Joi.number().integer().min(0).max(130)
});

Go validator:
type User struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=0,lte=130"`
}
*/
