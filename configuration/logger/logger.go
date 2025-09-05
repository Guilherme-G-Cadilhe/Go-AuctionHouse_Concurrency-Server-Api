// Package logger encapsula a funcionalidade de logging usando Zap
// Zap é uma biblioteca de logging de alta performance criada pela Uber
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Variável global que armazena a instância do logger
// Em Go, variáveis em nível de package são globais
// O * indica que é um ponteiro para zap.Logger
var (
	log *zap.Logger
)

// init() é uma função especial do Go que executa automaticamente quando o package é importado
// É equivalente a um código que roda na inicialização do módulo no Node.js
func init() {
	// Configuração personalizada do Zap logger
	// zap.Config é uma struct que define como o logger deve se comportar
	logConfiguration := zap.Config{
		// Level define o nível mínimo de log que será registrado
		// InfoLevel significa que vai logar: Info, Warn, Error, Fatal (mas não Debug)
		Level: zap.NewAtomicLevelAt(zap.InfoLevel),

		// Encoding define o formato de saída dos logs
		// "json" significa que os logs serão estruturados em JSON (ótimo para produção)
		// Alternativa seria "console" para logs mais legíveis durante desenvolvimento
		Encoding: "json",

		// EncoderConfig configura como cada campo do log será formatado
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message", // Campo que conterá a mensagem principal do log
			LevelKey:   "level",   // Campo que indica o nível do log (info, error, etc.)
			TimeKey:    "time",    // Campo que conterá o timestamp

			// EncodeLevel define como o nível será formatado
			// LowercaseLevelEncoder = "info", "error" (minúsculo)
			EncodeLevel: zapcore.LowercaseLevelEncoder,

			// EncodeTime define o formato do timestamp
			// ISO8601TimeEncoder = formato padrão internacional (2023-12-01T15:30:45Z)
			EncodeTime: zapcore.ISO8601TimeEncoder,

			// EncodeCaller mostra de onde o log foi chamado (arquivo:linha)
			// ShortCallerEncoder = apenas o nome do arquivo e linha (não o path completo)
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	// Tenta construir o logger com a configuração definida
	var err error
	log, err = logConfiguration.Build()
	if err != nil {
		// panic() para erros críticos que impedem a aplicação de funcionar
		// É similar ao throw de uma exceção não capturada no Node.js
		panic(err)
	}
}

// info é uma função helper para logs de informação
// Parâmetros:
//   - message string: Mensagem principal do log
//   - tags ...zap.Field: Campos adicionais (variadic - aceita N argumentos)
func info(message string, tags ...zap.Field) {
	// log.Info() registra um log de nível informativo
	log.Info(message, tags...)
	// log.Sync() força a escrita imediata do buffer (importante para garantir que o log seja escrito)
	log.Sync()
}

// Error é uma função helper para logs de erro (note que é exportada - começa com maiúscula)
// Parâmetros:
//   - message string: Mensagem de contexto do erro
//   - err error: O erro específico que ocorreu
//   - tags ...zap.Field: Campos adicionais opcionais
func Error(message string, err error, tags ...zap.Field) {
	// append() adiciona o erro como um campo nomeado "error" ao slice de tags
	// zap.NamedError() cria um campo estruturado com o erro
	tags = append(tags, zap.NamedError("error", err))

	// Registra o log de erro com todos os campos
	log.Error(message, tags...)
	log.Sync()
}

/*
EXEMPLO de uso do Zap vs console.log do Node.js:

Node.js:
console.log('User created:', { userId: 123, email: 'user@example.com' });
console.error('Database error:', error);

Go com Zap:
logger.info("User created",
    zap.Int("userId", 123),
    zap.String("email", "user@example.com"))

logger.Error("Database connection failed", err,
    zap.String("database", "mongodb"),
    zap.Int("retryAttempt", 3))

Saída JSON do Zap:
{
    "level": "info",
    "time": "2023-12-01T15:30:45Z",
    "message": "User created",
    "userId": 123,
    "email": "user@example.com"
}
*/
