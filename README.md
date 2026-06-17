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
- **Adaptadores de entrada** (`internal/adapters/input`): CLI e Webhook HTTP — traduzem o exterior para chamadas aos casos de uso.
- **Adaptadores de saída** (`internal/adapters/output`): repositório em arquivo JSON, em memória (testes) e serviço de notificação via log.
- **Composição** (`cmd/app/main.go`): único ponto onde as dependências são instanciadas e injetadas.

### Estrutura de diretórios

```
cmd/app/
  main.go                             # ponto de entrada e injeção de dependências
internal/
  domain/entities/
    subscription.go                   # entidade central com estados e transições
    customer.go
    plan.go
  application/
    ports/
      subscription_repository.go      # interface de persistência
      notification_service.go         # interface de notificação
    usecases/
      create_subscription.go
      list_subscriptions.go
      get_subscription.go
      update_subscription.go
      cancel_subscription.go
      delete_subscription.go
      process_payment_event.go        # reativa ou suspende via evento de pagamento
      errors.go
      subscription_usecases_test.go
  adapters/
    input/
      cli/cli.go                      # interpreta os.Args e exibe resultados
      webhook/handler.go              # recebe eventos HTTP e despacha em goroutine
    output/
      repositories/
        file_subscription.go          # persiste assinaturas como JSON em disco
        in_memory_subscription.go     # usado nos testes (sem I/O)
        errors.go
      notification/
        log_notification.go           # escreve notificações no stdout
data/subscriptions/                   # um arquivo .json por assinatura
```

## Canais de entrada

### CLI

Opera assinaturas via terminal. Todos os comandos compartilham o mesmo repositório em `data/subscriptions`.

```bash
# Criar assinatura
go run ./cmd/app create-subscription CUSTOMER_ID PLAN_ID

# Listar todas as assinaturas
go run ./cmd/app list-subscriptions

# Consultar assinatura por ID
go run ./cmd/app show-subscription SUBSCRIPTION_ID

# Atualizar assinatura
go run ./cmd/app update-subscription SUBSCRIPTION_ID CUSTOMER_ID PLAN_ID STATUS

# Cancelar assinatura (verifica elegibilidade a reembolso em 7 dias)
go run ./cmd/app cancel-subscription SUBSCRIPTION_ID

# Apagar assinatura
go run ./cmd/app delete-subscription SUBSCRIPTION_ID
```

### Webhook de pagamento

Inicia um servidor HTTP que recebe eventos do gateway de pagamento. Cada evento é processado em uma goroutine separada — o servidor responde `202 Accepted` imediatamente sem bloquear.

```bash
# Iniciar o servidor (padrão: :8080)
go run ./cmd/app serve

# Porta customizada
go run ./cmd/app serve :9090
```

**Endpoint:** `POST /webhook/payment`

**Corpo da requisição (JSON):**

```json
{
  "subscription_id": "1780356723552945000",
  "customer_id": "customer-1",
  "plan_id": "basic-plan",
  "status": "confirmed"
}
```

| Campo `status` | Efeito na assinatura |
|---|---|
| `confirmed` | Assinatura suspensa é reativada (`active`) |
| `refused` | Assinatura ativa é suspensa (`suspended`) |

**Exemplo com curl:**

```bash
# Pagamento confirmado — reativa a assinatura
curl -X POST http://localhost:8080/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"subscription_id":"SEU_ID","customer_id":"customer-1","plan_id":"basic","status":"confirmed"}'

# Pagamento recusado — suspende a assinatura
curl -X POST http://localhost:8080/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"subscription_id":"SEU_ID","customer_id":"customer-1","plan_id":"basic","status":"refused"}'
```

## Regras de negócio

- Assinatura criada via CLI começa com status `active`.
- Cancelamento com menos de 7 dias gera elegibilidade a reembolso proporcional.
- Apenas assinaturas `suspended` podem ser reativadas (`Reactivate`).
- Apenas assinaturas `active` podem ser suspensas (`Suspend`).
- O repositório de arquivo e o repositório em memória são protegidos por `sync.RWMutex` para acesso concorrente seguro via webhook.

## Testes

```bash
go test ./...
```

Os testes de casos de uso usam `InMemorySubscription` — sem arquivo, sem rede, sem banco. Cobrem:

- CRUD completo de assinaturas
- Cancelamento com e sem reembolso (janela de 7 dias)
- Processamento de pagamento confirmado → reativação
- Processamento de pagamento recusado → suspensão
- Status inválido → `ErrInvalidPaymentStatus`
- Assinatura inexistente → `ErrSubscriptionNotFound`

## Critérios de coerência arquitetural (TP2)

- Nenhum arquivo em `internal/domain/` importa `internal/adapters/`.
- Casos de uso recebem `SubscriptionRepository` e `NotificationService` por injeção no construtor.
- Trocar `FileSubscriptionRepository` por `InMemorySubscription` não altera nenhuma linha do domínio ou dos casos de uso.
- Testes de casos de uso não dependem de arquivo ou infraestrutura.
