# Build stage
FROM golang:1.23-alpine AS builder

# Instala certificados SSL e git
RUN apk add --no-cache ca-certificates git

# Define diretório de trabalho
WORKDIR /app

# Copia arquivos de dependência
COPY go.mod go.sum ./

# Baixa dependências (fica em cache se não mudaram)
RUN go mod download

# Copia código fonte
COPY . .

# Compila a aplicação
# CGO_ENABLED=0: desabilita CGO para binary estático
# GOOS=linux: compila para Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/auction

# Final stage
FROM alpine:latest

# Instala certificados SSL e timezone
RUN apk --no-cache add ca-certificates tzdata

# Cria usuário não-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Define diretório de trabalho
WORKDIR /app

# Copia binary da stage anterior
COPY --from=builder /app/main .

# Copia arquivo de config (se necessário)
COPY --from=builder /app/cmd/auction/.env .env

# Muda proprietário dos arquivos
RUN chown -R appuser:appgroup /app

# Muda para usuário não-root
USER appuser

# Expõe porta
EXPOSE 8080

# Comando para executar
CMD ["./main"]