# Gerenciador de Assinaturas Multi-Canal

Projeto acadêmico em Go para estudo de Arquitetura Hexagonal (Ports & Adapters).

| Campo | Informação |
|---|---|
| Disciplina | Arquitetura de Software |
| Alunos | Matheus Brugge, Stela David, Douglas Wilhan, Leonardo Santana e Ana Babiak |
| Arquitetura | Hexagonal (Ports & Adapters) |
| Linguagem | Go (sem framework web) |

---

## Build

```bash
go build -o subscription-manager ./cmd/app
```

Gera o binário `./subscription-manager` na raiz do projeto. Todos os comandos abaixo assumem que esse binário já existe.

---

## Modo CLI (arquivo JSON)

O repositório padrão persiste cada assinatura como um arquivo `.json` em `data/subscriptions/`. Não requer nenhuma infraestrutura adicional.

### Criar assinatura

```bash
./subscription-manager create-subscription CUSTOMER_ID PLAN_ID
```

Cria uma nova assinatura com status `active` e exibe o ID gerado.

```bash
./subscription-manager create-subscription customer-1 basic-plan
```

### Listar todas as assinaturas

```bash
./subscription-manager list-subscriptions
```

### Buscar uma assinatura

```bash
./subscription-manager show-subscription SUBSCRIPTION_ID
```

### Atualizar uma assinatura

```bash
./subscription-manager update-subscription SUBSCRIPTION_ID CUSTOMER_ID PLAN_ID STATUS
```

`STATUS` aceita: `active`, `suspended`, `canceled`.

```bash
./subscription-manager update-subscription 1234 customer-1 basic-plan suspended
```

### Cancelar uma assinatura

```bash
./subscription-manager cancel-subscription SUBSCRIPTION_ID
```

Cancela a assinatura e informa se há elegibilidade a reembolso proporcional (cancelamentos feitos em até 7 dias da criação).

### Reativar uma assinatura

```bash
./subscription-manager reactivate-subscription SUBSCRIPTION_ID
```

Reativa uma assinatura com status `suspended`. Retorna erro se a assinatura não estiver suspensa.

### Deletar uma assinatura

```bash
./subscription-manager delete-subscription SUBSCRIPTION_ID
```

---

## Modo servidor com interface web

Inicia um servidor HTTP na porta `8080` com:
- **Interface web** em `GET /` para testar todas as operações sem curl
- **API REST** completa para CRUD de assinaturas
- **Webhook de pagamento** em `POST /webhook/payment`

```bash
./subscription-manager serve
```

Acesse **http://localhost:8080** no navegador.

Para usar uma porta diferente:

```bash
./subscription-manager serve :9090
```

### Interface web

A UI possui um formulário para cada operação. Ao criar uma assinatura, o ID é preenchido automaticamente nos outros formulários. O painel direito exibe a resposta JSON com o código HTTP.

### API REST

| Método | Endpoint | Operação |
|---|---|---|
| `GET` | `/subscriptions` | Listar todas |
| `POST` | `/subscriptions` | Criar |
| `GET` | `/subscriptions/{id}` | Buscar por ID |
| `PUT` | `/subscriptions/{id}` | Atualizar |
| `POST` | `/subscriptions/{id}/cancel` | Cancelar |
| `POST` | `/subscriptions/{id}/reactivate` | Reativar |
| `DELETE` | `/subscriptions/{id}` | Deletar |
| `POST` | `/webhook/payment` | Evento de pagamento |

### Webhook de pagamento

Cada evento é processado em uma goroutine — o servidor responde `202 Accepted` imediatamente e processa em background.

```bash
curl -X POST http://localhost:8080/webhook/payment \
  -H "Content-Type: application/json" \
  -d '{"subscription_id":"SEU_ID","customer_id":"customer-1","plan_id":"basic-plan","status":"confirmed"}'
```

| `status` | Efeito |
|---|---|
| `confirmed` | Assinatura suspensa é reativada (`active`) |
| `refused` | Assinatura ativa é suspensa (`suspended`) |

---

## Modo banco de dados (PostgreSQL)

Troque o repositório de arquivo pelo PostgreSQL definindo `REPOSITORY=db`. Nenhuma linha de domínio ou caso de uso é alterada — apenas o adaptador de saída muda.

### Subir o banco com Docker

```bash
docker compose up -d postgres
```

O container inicializa com `migrations/init.sql` aplicado automaticamente.

### Executar com PostgreSQL

```bash
REPOSITORY=db ./subscription-manager create-subscription customer-1 basic-plan
REPOSITORY=db ./subscription-manager list-subscriptions
REPOSITORY=db ./subscription-manager serve
```

O `DATABASE_URL` padrão aponta para o container Docker:

```
postgres://subscription_manager:subscription_manager@localhost:5432/subscription_manager?sslmode=disable
```

Para usar outro banco:

```bash
REPOSITORY=db DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable ./subscription-manager list-subscriptions
```

---

## Docker Compose completo (app + banco)

Sobe o PostgreSQL e o app Go juntos. O app aguarda o banco estar saudável antes de iniciar.

```bash
docker compose up --build
```

Acesse **http://localhost:8080** no navegador.

Para rebuildar apenas o app após mudanças no código:

```bash
docker compose up --build app
```

---

## Testes

```bash
# Unitários — sem infraestrutura, roda sempre
go test ./...

# Com integração PostgreSQL (requer Docker rodando)
DATABASE_URL=postgres://subscription_manager:subscription_manager@localhost:5432/subscription_manager?sslmode=disable go test ./...
```

Os testes de integração do `DBSubscriptionRepository` são pulados automaticamente quando `DATABASE_URL` não está definido.

---

## Arquitetura

O sistema é organizado em três zonas que nunca invertem a direção de dependência:

```
Adaptadores  →  Portas  →  Domínio
```

- **Domínio** (`internal/domain/entities`): entidades e regras de negócio puras. Não importa nenhum pacote externo.
- **Portas** (`internal/application/ports`): interfaces Go que definem os contratos de entrada (driving) e saída (driven).
- **Casos de uso** (`internal/application/usecases`): orquestram as operações usando apenas as portas.
- **Adaptadores de entrada** (`internal/adapters/input`): CLI, REST HTTP com UI web.
- **Adaptadores de saída** (`internal/adapters/output`): repositório em arquivo JSON, PostgreSQL, notificação via log e gateway de pagamento fake.
- **Composição** (`cmd/app/main.go`): único ponto onde as dependências são instanciadas e injetadas.

### Portas

**Driving ports** (chamadas pelos adaptadores de entrada):

| Interface | Operação |
|---|---|
| `SubscribeUseCase` | Criar assinatura |
| `ListSubscriptionsUseCase` | Listar todas as assinaturas |
| `GetSubscriptionUseCase` | Buscar assinatura por ID |
| `UpdateSubscriptionUseCase` | Atualizar campos de uma assinatura |
| `CancelSubscriptionUseCase` | Cancelar com cálculo de elegibilidade de reembolso |
| `ReactivateSubscriptionUseCase` | Reativar assinatura suspensa |
| `DeleteSubscriptionUseCase` | Remover assinatura |
| `ProcessPaymentEventUseCase` | Processar evento do gateway de pagamento |

**Driven ports** (implementadas pelos adaptadores de saída):

| Interface | Responsabilidade |
|---|---|
| `SubscriptionRepository` | Persistência (`Save`, `FindByID`, `FindAll`, `Delete`) |
| `NotificationService` | Envio de notificações ao cliente |
| `PaymentGateway` | Cobrança e estorno no gateway externo |

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
      subscription_usecases.go          # 8 driving port interfaces + CancellationResult
      subscription_repository.go        # interface de persistência
      notification_service.go           # interface de notificação
      payment_gateway.go                # interface de cobrança/estorno
      errors.go                         # sentinelas compartilhados (ErrNotFound, etc.)
    usecases/
      create_subscription.go            # cria assinatura; cobra gateway antes de salvar
      list_subscriptions.go
      get_subscription.go
      update_subscription.go
      cancel_subscription.go            # cancela com cálculo de reembolso proporcional
      reactivate_subscription.go        # reativa assinatura suspensa
      delete_subscription.go
      process_payment_event.go          # reativa ou suspende via evento de pagamento
      errors.go
      subscription_usecases_test.go
  adapters/
    input/
      cli/cli.go                        # interpreta os.Args e exibe resultados
      rest/
        server.go                       # handlers REST + roteamento
        ui.html                         # interface web embutida no binário
    output/
      repositories/
        file_subscription.go            # persiste assinaturas como JSON em disco
        db_subscription.go              # persiste assinaturas em PostgreSQL
        in_memory_subscription.go       # usado nos testes (sem I/O)
        errors.go
      notification/
        log_notification.go             # escreve notificações no stdout
      gateway/
        fake_payment_gateway.go         # gateway fake para testes e desenvolvimento
migrations/
  init.sql                              # cria a tabela subscriptions
data/subscriptions/                     # um arquivo .json por assinatura (modo file)
Dockerfile
docker-compose.yml
```

### Critérios de coerência arquitetural (TP2)

- Nenhum arquivo em `internal/domain/` importa `internal/adapters/`.
- Casos de uso recebem dependências exclusivamente por injeção — `SubscriptionRepository`, `NotificationService` e `PaymentGateway` são interfaces definidas em `ports`.
- Adaptadores de entrada (`cli`, `rest`) dependem apenas de `ports` — não importam nenhum pacote de `usecases`.
- O CLI e o servidor REST são dois adaptadores independentes que chamam os mesmos casos de uso pelas mesmas interfaces de porta — sem duplicação de lógica.
- Trocar `FileSubscriptionRepository` por `DBSubscriptionRepository` não altera nenhuma linha do domínio ou dos casos de uso — apenas a variável `REPOSITORY` em `main.go`.
- Testes de casos de uso não dependem de arquivo, rede ou banco de dados.
- Eventos de pagamento idempotentes são ignorados: se a assinatura já está no estado alvo, nenhuma escrita ocorre e nenhuma notificação é enviada.
- A cobrança no gateway ocorre antes da persistência: se o `Save` falhar após um `Charge` bem-sucedido, um estorno (`Refund`) é disparado automaticamente como transação compensatória.
