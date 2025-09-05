# 🏛️ Go AuctionHouse - Sistema de Leilões Concorrente

Sistema de leilões em tempo real desenvolvido em Go, implementando Clean Architecture e processamento concorrente para alta performance. Utiliza MongoDB para persistência e técnicas avançadas de batch processing com goroutines e channels.

## 🏗️ Arquitetura

```
Cliente → Controller → UseCase → Repository → MongoDB
            ↓
     Batch Processing (Concorrência)
          ↓
Goroutines + Channels + Mutex
```

**Camadas da Clean Architecture**

- **Controllers:** Camada de apresentação HTTP (Gin framework)
- **UseCases:** Regras de negócio e orquestração
- **Entities:** Entidades de domínio puras
- **Repositories:** Abstração de acesso a dados
- **Infrastructure:** Implementação concreta (MongoDB)
  **Sistema de Concorrência**
- **Batch Processing:** Agrupa lances para inserção eficiente
- **Channels:** Comunicação assíncrona entre goroutines
- **Mutex:** Proteção de recursos compartilhados (cache)
- **WaitGroups:** Coordenação de múltiplas goroutines

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados
- Opcionalmente: Go 1.23+ para desenvolvimento local
- Executar o sistema completo

```bash
# Clone o repositório
git clone <repo>
cd Go-AuctionHouse_Concurrency-Server-Api

# Inicia todos os serviços
docker-compose up --build -d

# Para parar
docker-compose down
```

### Executar localmente (desenvolvimento)

```bash
# Instalar dependências
go mod download

# Subir MongoDB via Docker
docker run -d --name mongodb -p 27017:27017 mongo:latest

# Executar aplicação
go run ./cmd/auction/main.go
```

## 🧪 Testando o Sistema

**1. Health Check**

```bash
curl http://localhost:8080/health
```

**2. Criar Usuário**

```bash
curl -X POST http://localhost:8080/user \
  -H "Content-Type: application/json" \
  -d '{"name": "João Silva"}'
```

```json
Resposta:
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "João Silva"
}
```

**3. Criar Leilão**

```bash
curl -X POST http://localhost:8080/auctions \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 14 Pro",
    "category": "Electronics",
    "description": "iPhone 14 Pro usado em ótimo estado, sem arranhões",
    "condition": 1
  }'
Condições disponíveis:

0 = Novo
1 = Usado
2 = Recondicionado
```

**4. Fazer Lance (Concorrente)**

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "USER_ID_AQUI",
    "auction_id": "AUCTION_ID_AQUI",
    "amount": 1500.50
  }'
```

5. Buscar Lances de um Leilão
   bash
   Copiar

curl http://localhost:8080/bid/AUCTION_ID_AQUI 6. Ver Lance Vencedor
bash
Copiar

curl http://localhost:8080/auctions/winner/AUCTION_ID_AQUI 7. Listar Todos os Leilões
bash
Copiar

# Todos os leilões

curl http://localhost:8080/auctions

# Filtrar por categoria

curl "http://localhost:8080/auctions?category=Electronics"

# Filtrar por nome do produto

curl "http://localhost:8080/auctions?productName=iPhone"

## ⚡ Sistema de Concorrência

### Batch Processing de Lances

O sistema implementa processamento em lote para máxima performance:

### Cache Inteligente

- Status dos leilões é cacheado em memória
- Tempo de fim calculado dinamicamente
- Mutex protege acesso concorrente ao cache
- Evita consultas repetidas ao banco

## 📁 Estrutura do Projeto

Go-AuctionHouse/
├── cmd/auction/ # Aplicação principal
│ └── main.go
├── internal/
│ ├── entity/ # Entidades de domínio
│ │ ├── user_entity/
│ │ ├── auction_entity/
│ │ └── bid_entity/
│ ├── usecase/ # Casos de uso
│ │ ├── user_usecase/
│ │ ├── auction_usecase/
│ │ └── bid_usecase/ # ← Batch processing aqui
│ ├── infra/
│ │ ├── database/ # Repositories
│ │ │ ├── user/
│ │ │ ├── auction/
│ │ │ └── bid/ # ← Concorrência aqui
│ │ └── api/web/
│ │ ├── controller/ # HTTP handlers
│ │ └── validation/ # Validações
│ └── internal_error/ # Tratamento de erros
├── configuration/
│ ├── database/mongodb/ # Conexão MongoDB
│ ├── logger/ # Sistema de logs
│ └── rest_err/ # Erros HTTP
├── docker-compose.yml
├── Dockerfile
└── go.mod

## 🧩 Conceitos Implementados

### Clean Architecture

- **Separação clara** de responsabilidades
- **Inversão de dependência** com interfaces
- **Entities** independentes de frameworks
- **DTOs** para controle de dados expostos

### Concorrência em Go

- **Goroutines** para processamento paralelo
- **Channels** para comunicação assíncrona
- **Select statements** para multiplexing
- **Mutex** para proteção de recursos

### Patterns de Design

- **Repository Pattern** para abstração de dados
- **Factory Functions** para criação de entidades
- **Dependency Injection** manual
- **Error Handling** consistente

### Performance

- **Batch Processing** para alta throughput
- **Cache em memória** para reduzir latência
- **Connection pooling** do MongoDB driver
- **Processamento assíncrono** de lances

## 📚 Aprendizados

- **Clean Architecture** em Go com separation of concerns
- **Concorrência avançada** com goroutines, channels e mutex
- **Batch processing** para sistemas de alta performance
- **Error handling** robusto e consistente
- **Docker** para containerização de aplicações Go
- **MongoDB** com Go driver oficial
- **Gin framework** para APIs REST performáticas
- **Dependency Injection** manual vs frameworks
  **Desenvolvido com ❤️ em Go para aprendizado de arquitetura e concorrência**
