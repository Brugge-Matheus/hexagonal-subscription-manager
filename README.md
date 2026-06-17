# Gerenciador de Assinaturas Multi-Canal

Projeto acadêmico em Go para estudo de Arquitetura Hexagonal (Ports & Adapters).

| Campo | Informação |
|---|---|
| Disciplina | Arquitetura de Software |
| Alunos | Matheus Brugge, Stela David, Douglas Wilhan, Leonardo Santana e Ana Babiak |
| Arquitetura | Hexagonal (Ports & Adapters) |
| Linguagem | Go (sem framework web) |

## Arquitetura

O sistema é organizado em três zonas que nunca invertem a direção de dependência:

```
Adaptadores  →  Portas  →  Domínio
```

- **Domínio** (`internal/domain/entities`): entidades e regras de negócio puras. Não importa nenhum pacote externo.
- **Portas** (`internal/application/ports`): interfaces Go que definem os contratos de entrada e saída.
- **Casos de uso** (`internal/application/usecases`): orquestram as operações usando apenas as portas.
- **Adaptadores de entrada** (`internal/adapters/input`): CLI e Webhook HTTP.
- **Adaptadores de saída** (`internal/adapters/output`): repositório em arquivo JSON, PostgreSQL e notificação via log.
- **Composição** (`cmd/app/main.go`): único ponto onde as dependências são instanciadas e injetadas.

### Estrutura de diretórios

```
cmd/app/
  main.go                               # ponto de entrada e injeção de dependências
internal/
  domain/entities/
    subscription.go                     # entidade central com estados e transições
    customer.go
    plan.go
  application/
    ports/
      subscription_repository.go        # interface de persistência
      notification_service.go           # interface de notificação
    usecases/
      create_subscription.go
      list_subscriptions.go
      get_subscription.go
      update_subscription.go
      cancel_subscription.go
      delete_subscription.go
      process_payment_event.go          # reativa ou suspende via evento de pagamento
      errors.go
      subscription_usecases_test.go
  adapters/
    input/
      cli/cli.go                        # interpreta os.Args e exibe resultados
      webhook/handler.go                # recebe eventos HTTP e despacha em goroutine
    output/
      repositories/
        file_subscription.go            # persiste assinaturas como JSON em disco
        db_subscription.go              # persiste assinaturas em PostgreSQL
        in_memory_subscription.go       # usado nos testes (sem I/O)
        errors.go
      notification/
        log_notification.go             # escreve notificações no stdout
migrations/
  init.sql                              # cria a tabela subscriptions
data/subscriptions/                     # um arquivo .json por assinatura (modo file)
docker-compose.yml
```

## Repositórios disponíveis

O repositório ativo é selecionado pela variável de ambiente `REPOSITORY`. Os dois implementam a mesma interface `SubscriptionRepository` — nenhuma linha do domínio ou dos casos de uso muda ao trocar entre eles.

| `REPOSITORY` | Adaptador | Persistência |
|---|---|---|
| (não definido) | `FileSubscriptionRepository` | Arquivos JSON em `data/subscriptions/` |
| `db` | `DBSubscriptionRepository` | PostgreSQL |

## Como executar

### Modo arquivo (padrão)

Não requer nenhuma infraestrutura adicional.

```bash
go run ./cmd/app create-subscription CUSTOMER_ID PLAN_ID
go run ./cmd/app list-subscriptions
go run ./cmd/app show-subscription SUBSCRIPTION_ID
go run ./cmd/app update-subscription SUBSCRIPTION_ID CUSTOMER_ID PLAN_ID STATUS
go run ./cmd/app cancel-subscription SUBSCRIPTION_ID
go run ./cmd/app delete-subscription SUBSCRIPTION_ID
```

### Modo banco de dados (PostgreSQL)

**1. Subir o banco com Docker:**

```bash
docker compose up -d
```

O container inicializa com a migration `migrations/init.sql` aplicada automaticamente.

**2. Executar comandos com o repositório DB:**

```bash
REPOSITORY=db go run ./cmd/app create-subscription customer-1 basic-plan
REPOSITORY=db go run ./cmd/app list-subscriptions
```

Por padrão o `DATABASE_URL` aponta para o container Docker:

```
postgres://subscription_manager:subscription_manager@localhost:5432/subscription_manager?sslmode=disable
```

Para usar outro banco, defina a variável:

```bash
REPOSITORY=db DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable go run ./cmd/app list-subscriptions
```

## Canal webhook de pagamento

Inicia um servidor HTTP que recebe eventos do gateway de pagamento. Cada evento é processado em uma goroutine separada — o servidor responde `202 Accepted` imediatamente.

```bash
# Modo arquivo
go run ./cmd/app serve

# Modo banco de dados
REPOSITORY=db go run ./cmd/app serve

# Porta customizada
go run ./cmd/app serve :9090
```

**Endpoint:** `POST /webhook/payment`

```json
{
  "subscription_id": "1780356723552945000",
  "customer_id": "customer-1",
  "plan_id": "basic-plan",
  "status": "confirmed"
}
```

| `status` | Efeito |
|---|---|
| `confirmed` | Assinatura suspensa é reativada (`active`) |
| `refused` | Assinatura ativa é suspensa (`suspended`) |

```bash
curl -X POST http://localhost:8080/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"subscription_id":"SEU_ID","customer_id":"customer-1","plan_id":"basic","status":"confirmed"}'
```

## Testes

```bash
# Testes unitários (sem infraestrutura)
go test ./...

# Testes de integração com PostgreSQL (requer Docker rodando)
DATABASE_URL=postgres://subscription_manager:subscription_manager@localhost:5432/subscription_manager?sslmode=disable go test ./...
```

Os testes unitários usam `InMemorySubscription` — sem arquivo, sem rede, sem banco. Os testes de integração do `DBSubscriptionRepository` são pulados automaticamente quando `DATABASE_URL` não está definido.

## Regras de negócio

- Assinatura criada via CLI começa com status `active`.
- Cancelamento com menos de 7 dias gera elegibilidade a reembolso proporcional.
- Apenas assinaturas `suspended` podem ser reativadas.
- Apenas assinaturas `active` podem ser suspensas.
- `FileSubscriptionRepository` e `DBSubscriptionRepository` são protegidos por `sync.RWMutex` para acesso concorrente seguro via webhook.

## Critérios de coerência arquitetural (TP2)

- Nenhum arquivo em `internal/domain/` importa `internal/adapters/`.
- Casos de uso recebem `SubscriptionRepository` e `NotificationService` por injeção no construtor.
- Trocar `FileSubscriptionRepository` por `DBSubscriptionRepository` não altera nenhuma linha do domínio ou dos casos de uso — apenas a variável `REPOSITORY` em `main.go`.
- Testes de casos de uso não dependem de arquivo ou infraestrutura.
