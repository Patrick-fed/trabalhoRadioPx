<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->

# Regras do Projeto RadioPX

## Contexto

Este é um projeto acadêmico (TCC/trabalho de faculdade) para simulação de um sistema de rádio PX digital.

## Regras Gerais

### 1. Entender antes de modificar

- **NÃO** altere arquivos sem entender a arquitetura existente
- Leia a documentação em `docs/` antes de fazer mudanças
- Consulte `docs/decisoes.md` para entender decisões arquiteturais anteriores
- Revise `docs/arquitetura.md` para entender a estrutura do sistema

### 2. Seguir requisitos

- **NÃO** crie funcionalidades que não estejam nos requisitos
- Consulte `docs/requisitos.md` para lista completa de requisitos
- Toda feature deve ter RF (Requisito Funcional) ou RNF (Requisito Não Funcional) correspondente
- Requisitos são a fonte da verdade

### 3. Comunicação

- Antes de implementar algo grande, explique o que pretende fazer
- Aguarde aprovação antes de prosseguir com mudanças significativas
- Documente mudanças significativas

### 4. Documentação

- Após alterações importantes, atualize a documentação correspondente
- Atualize `CHANGELOG.md` com data e motivo da alteração
- Atualize `PROJECT_STATE.md` com status atual do projeto
- Mantenha `docs/requisitos.md` atualizado com novos requisitos

## Tecnologias Utilizadas

| Camada | Tecnologia | Uso |
|--------|-----------|-----|
| **Frontend** | Flutter/Dart | App mobile cross-platform |
| **Backend** | Go/Golang | Servidor API |
| **Banco** | PostgreSQL | Dados permanentes |
| **Comunicação** | REST + WebSocket | API e tempo real |
| **Autenticação** | JWT/OAuth/Casbin | Sessões e acesso |

## Estrutura de Pastas

```
RadioPx/
├── backend/           # Backend Go
│   ├── cmd/          # Pontos de entrada
│   ├── internal/     # Pacotes privados
│   └── pkg/          # Pacotes públicos
│
├── frontend/         # App Flutter
│   ├── lib/          # Código Dart
│   ├── android/      # Android específico
│   └── ios/          # iOS específico
│
├── docs/             # Documentação
│   ├── requisitos.md
│   ├── arquitetura.md
│   ├── banco-de-dados.md
│   └── decisoes.md
│
└── specs/            # Especificações de features
```

## Padrão de Código

### Go
- Seguir padrões `gofmt`
- Usar naming conventions do Go (camelCase para funções, PascalCase para exportados)
- Comentários em inglês
- Tests em arquivos `_test.go`

### Dart/Flutter
- Seguir padrões `dartfmt`
- Usar naming conventions do Dart (camelCase para variáveis, PascalCase para classes)
- Widget naming: substantivo + descrição (ex: `AudioPlayerWidget`)

### Commits
- Mensagens claras e descritivas
- Formato: `tipo(módulo): descrição`
- Exemplos: `feat(auth): implement JWT authentication`, `fix(voice): resolve audio latency issue`

## Como Rodar o Projeto

### Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend
flutter pub get
flutter run
```

### Testes

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
flutter test
```

## Regras do Banco de Dados

- Apenas **usuários** e **canais** são persistidos em PostgreSQL
- **Áudio** NUNCA é armazenado (requisito crítico de privacidade)
- Buffer de áudio existe apenas em memória (30 segundos máximo)
- UUIDs são usados como chaves primárias
- Índices criados para campos de busca frequente

## Regras de API

### REST (Operações Tradicionais)

- `POST /api/v1/auth/register` - Cadastro
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/channels` - Listar canais
- `POST /api/v1/channels/join` - Entrar no canal
- `POST /api/v1/channels/leave` - Sair do canal
- `POST /api/v1/location/update` - Atualizar localização

### WebSocket (Tempo Real)

- `ws://localhost:8080/ws/audio/{channel_id}` - Streaming de áudio
- Conexão persistente para baixa latência
- Reconexão automática em caso de queda

## Autenticação

- **JWT**: Sessões stateless para APIs REST
- **OAuth (Google)**: Login social para melhor experiência
- **Casbin**: Controle de acesso baseado em papéis
- Tokens expiram após 24 horas
- Refresh tokens para renovação silenciosa

## Coisas que NÃO Devem ser Alteradas

### Requisitos Críticos

1. **Privacidade de áudio**: NENHUM áudio de conversa deve ser persistido
2. **Raio de proximidade**: 10km (não alterar sem justificativa)
3. **Máximo de usuários por canal**: 10
4. **Máximo de canais visíveis**: 15
5. **Buffer temporário**: 30 segundos (não aumentar significativamente)

### Arquitetura

1. **Monolito modular**: Não converta para microsserviços
2. **PostgreSQL**: Não troque de banco de dados sem justificativa
3. **WebSocket para áudio**: Não use HTTP polling para streaming

### Segurança

1. **JWT Secret**: Nunca committar em repositório
2. **Credenciais Google**: Usar variáveis de ambiente
3. **Senhas**: Sempre hasheadas com bcrypt
4. **TLS**: Obrigatório em produção

## Fluxo de Trabalho Recomendado

1. **Planejamento**: Revisar requisitos e arquitetura
2. **Implementação**: Seguir estrutura de pastas definida
3. **Testes**: Escrever testes antes ou durante implementação
4. **Documentação**: Atualizar docs após mudanças significativas
5. **Commit**: Mensagens descritivas seguindo padrão
