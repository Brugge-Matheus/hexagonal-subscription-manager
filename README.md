# Gerenciador de Assinaturas Multi-Canal

Projeto acadêmico em Go para estudo de Arquitetura Hexagonal (Ports & Adapters).

## Estado atual

O projeto foi migrado de Ruby para **Go** e, no estágio atual, implementa o **CRUD completo de subscriptions pela CLI**, com persistência em arquivos JSON dentro de `data/subscriptions`.

## Arquitetura escolhida

A arquitetura escolhida é a **Hexagonal (Ports & Adapters)**.

- `internal/domain/entities`: entidades do domínio.
- `internal/application/usecases`: casos de uso da aplicação.
- `internal/application/ports`: contratos que a aplicação usa para falar com fora.
- `internal/adapters/input`: adaptadores de entrada, como a CLI.
- `internal/adapters/output`: adaptadores de saída, como persistência em arquivos.
- `cmd/app`: ponto de entrada e composição das dependências.

## Descricao das implementacoes e fluxo arquitetural

### Funcionalidade 1: Criar assinatura

- **Onde o fluxo começa:** CLI, pelo comando `create-subscription`
- **Quais componentes participam:** `CLI -> CreateSubscription -> SubscriptionRepository -> FileSubscription`
- **Onde fica a regra de negócio:** no caso de uso `CreateSubscription`, que cria a entidade `Subscription` com status inicial `active`
- **Onde os dados são armazenados ou consultados:** arquivos JSON em `data/subscriptions`
- **Como executar ou testar:** `go run ./cmd/app create-subscription customer-1 basic-plan`
- **Resultado esperado:** uma nova assinatura é criada e um arquivo JSON é salvo em `data/subscriptions`

### Funcionalidade 2: Listar assinaturas

- **Onde o fluxo começa:** CLI, pelo comando `list-subscriptions`
- **Quais componentes participam:** `CLI -> ListSubscriptions -> SubscriptionRepository -> FileSubscription`
- **Onde fica a regra de negócio:** no caso de uso `ListSubscriptions`, que coordena a consulta
- **Onde os dados são armazenados ou consultados:** arquivos JSON em `data/subscriptions`
- **Como executar ou testar:** `go run ./cmd/app list-subscriptions`
- **Resultado esperado:** todas as assinaturas armazenadas são exibidas no terminal

### Funcionalidade 3: Consultar assinatura

- **Onde o fluxo começa:** CLI, pelo comando `show-subscription SUBSCRIPTION_ID`
- **Quais componentes participam:** `CLI -> GetSubscription -> SubscriptionRepository -> FileSubscription`
- **Onde fica a regra de negócio:** no caso de uso `GetSubscription`, que busca a assinatura pelo id
- **Onde os dados são armazenados ou consultados:** arquivos JSON em `data/subscriptions`
- **Como executar ou testar:** `go run ./cmd/app show-subscription SUBSCRIPTION_ID`
- **Resultado esperado:** a assinatura é exibida no terminal ou a CLI informa que ela não foi encontrada

### Funcionalidade 4: Atualizar assinatura

- **Onde o fluxo começa:** CLI, pelo comando `update-subscription SUBSCRIPTION_ID CUSTOMER_ID PLAN_ID STATUS`
- **Quais componentes participam:** `CLI -> UpdateSubscription -> SubscriptionRepository -> FileSubscription`
- **Onde fica a regra de negócio:** no caso de uso `UpdateSubscription`, que verifica a existência da assinatura antes de sobrescrever os dados
- **Onde os dados são armazenados ou consultados:** arquivos JSON em `data/subscriptions`
- **Como executar ou testar:** `go run ./cmd/app update-subscription SUBSCRIPTION_ID customer-2 premium-plan suspended`
- **Resultado esperado:** a assinatura é atualizada no arquivo JSON correspondente

### Funcionalidade 5: Apagar assinatura

- **Onde o fluxo começa:** CLI, pelo comando `delete-subscription SUBSCRIPTION_ID`
- **Quais componentes participam:** `CLI -> DeleteSubscription -> SubscriptionRepository -> FileSubscription`
- **Onde fica a regra de negócio:** no caso de uso `DeleteSubscription`, que verifica se a assinatura existe antes de apagar
- **Onde os dados são armazenados ou consultados:** arquivos JSON em `data/subscriptions`
- **Como executar ou testar:** `go run ./cmd/app delete-subscription SUBSCRIPTION_ID`
- **Resultado esperado:** o arquivo JSON da assinatura é removido da pasta `data/subscriptions`

## Estrutura principal

```text
cmd/app/main.go
internal/domain/entities/subscription.go
internal/application/ports/subscription_repository.go
internal/application/usecases/
internal/adapters/input/cli/cli.go
internal/adapters/output/repositories/file_subscription.go
data/subscriptions/
```

### Papel das pastas

- `cmd/app`: inicializa a aplicação e injeta as dependências.
- `internal/domain/entities`: modela as entidades centrais do domínio.
- `internal/application/ports`: define os contratos usados pela aplicação.
- `internal/application/usecases`: concentra os casos de uso do sistema.
- `internal/adapters/input`: recebe entradas externas, como a CLI.
- `internal/adapters/output`: implementa infraestrutura concreta, como persistência em arquivos.
- `data/subscriptions`: armazena um arquivo JSON por assinatura.

## Como executar

```bash
go run ./cmd/app create-subscription customer-1 basic-plan
go run ./cmd/app list-subscriptions
go run ./cmd/app show-subscription SUBSCRIPTION_ID
go run ./cmd/app update-subscription SUBSCRIPTION_ID customer-2 premium-plan suspended
go run ./cmd/app delete-subscription SUBSCRIPTION_ID
```

Cada assinatura criada gera um arquivo `.json` próprio dentro de `data/subscriptions`.

## Como rodar os testes

```bash
go test ./...
```

## O que já está funcionando

- Criação de assinatura via CLI.
- Listagem de assinaturas via CLI.
- Consulta de assinatura por id via CLI.
- Atualização de assinatura via CLI.
- Exclusão de assinatura via CLI.
- Persistência em arquivos JSON.
- Testes automatizados para repositório e casos de uso.

## Repositório Git

- https://github.com/Brugge-Matheus/hexagonal-subscription-manager
