# Arquitetura do Sistema RadioPX

## Visão Geral

Arquitetura monolítica modular com separação por responsabilidades. O sistema é dividido em módulos claros, cada um com uma responsabilidade específica.

## Status de Implementação

| Módulo | Backend | Frontend | Status |
|--------|---------|----------|--------|
| Auth | ✅ Completo | ✅ Completo | ✅ |
| User | ✅ Completo | ✅ Completo | ✅ |
| Channel | ✅ Completo | ✅ Completo | ✅ |
| Voice | ✅ Completo | ✅ Completo | ✅ |
| Location | ✅ Completo | ✅ Completo | ✅ |
| Profile | N/A | ✅ Completo | ✅ |

> **Nota (27/10/2026)**: a tabela acima é aspiracional. Estado real: Voice com um falante por vez
> (`voiceGuard`, 27/10), Location reduzida a cálculo de distância (handler removido), canais e
> usuários em **PostgreSQL desde 20/10**, buffer/replay de voz ainda **não conectados**.
> Detalhes em `PROJECT_STATE.md` → *Lacunas / Aderência aos Requisitos*.

## Diagrama de Alto Nível

```
┌─────────────────────────────────────────────────────┐
│                    Flutter App                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   Auth      │  │  Channels   │  │   Voice     │ │
│  │   Module    │  │   Module    │  │   Module    │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘ │
│         │                │                │         │
│  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐ │
│  │  Location   │  │   Profile   │  │   Audio     │ │
│  │   Module    │  │   Module    │  │   Queue     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘ │
└─────────┼────────────────┼────────────────┼─────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────┐
│                   Go Backend                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   Auth      │  │   User      │  │   Channel   │ │
│  │   Service   │  │   Service   │  │   Service   │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘ │
│         │                │                │         │
│  ┌──────┴──────┐  ┌──────┴──────┐  ┌──────┴──────┐ │
│  │   Voice     │  │  Location   │  │  WebSocket  │ │
│  │   Service   │  │   Service   │  │   Manager   │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘ │
└─────────┼────────────────┼────────────────┼─────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────┐
│                Infrastructure                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │  PostgreSQL │  │  WebSocket  │  │    ngrok    │ │
│  │  Database   │  │   Server    │  │   Tunnel    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Backend (Go)

### Estrutura

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    # Ponto de entrada
│
├── internal/
│   ├── auth/
│   │   └── service.go                  # Lógica de negócio de auth
│   │
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go               # Interface UserStore + PGUserStore
│   │   ├── memory_store.go             # MemoryUserStore (testes)
│   │   └── model.go
│   │
│   ├── channel/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go               # Interface ChannelStore + PGChannelStore
│   │   ├── memory_store.go             # MemoryChannelStore (testes)
│   │   └── model.go
│   │
│   ├── db/
│   │   ├── db.go                       # Config env + Open (pool/ping)
│   │   ├── migrate.go                  # Runner de migrações (go:embed)
│   │   └── migrations/
│   │       ├── 001_initial.sql
│   │       └── 002_align_models.sql
│   │
│   ├── voice/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── websocket.go                # WebSocket handling
│   │   ├── buffer_adapter.go           # Adapta buffer_service à interface do WS (DIP)
│   │   ├── buffer_manager.go           # Buffer por canal (não conectado, fora de escopo)
│   │   ├── buffer_model.go
│   │   └── buffer_service.go           # Buffer de 30s conectado ao relay (RF06)
│   │
│   ├── location/
│   │   └── distance.go                 # EarthRadiusKm + CalculateDistance
│   │
│   └── middleware/
│       ├── auth.go                     # Autenticação (REST + WS)
│       ├── cors.go
│       ├── logging.go
│       └── ratelimit.go                # Rate limiting (RemoteAddr)
│
├── pkg/
│   ├── websocket/
│   │   ├── hub.go                      # Hub de conexões
│   │   └── client.go                   # Cliente WebSocket + voiceGuard
│   │
│   ├── security/
│   │   └── config.go                   # Config (não utilizado no boot)
│   │
│   └── utils/
│       └── helpers.go
│
├── go.mod
└── go.sum
```

### Módulos

#### Auth Module
- **Responsabilidade**: Autenticação e autorização
- **Dependências**: User module
- **Endpoints**: `/api/v1/auth/*`

#### User Module
- **Responsabilidade**: Gestão de usuários
- **Dependências**: Nenhuma
- **Endpoints**: `/api/v1/user/profile`, `/api/v1/user/change-password`

#### Channel Module
- **Responsabilidade**: Alocação e gestão de canais
- **Dependências**: User module; localização via `internal/location` (distância)
- **Endpoints**: `/api/v1/channels/*`

#### Voice Module
- **Responsabilidade**: PTT e streaming de áudio
- **Dependências**: Channel module, WebSocket
- **Endpoint**: `/ws/audio/?channel=...&token=...` (autenticado por JWT)

#### Location Module
- **Responsabilidade**: Cálculo de distância (Haversine) para proximidade
- **Dependências**: Nenhuma
- **Código**: apenas `distance.go` (handler/service removidos em 27/10/2026 —
  a localização trafega nos payloads da API de canais)

#### Middleware
- **Responsabilidade**: Interceptadores HTTP
- **Dependências**: Auth module
- **Uso**: Autenticação, CORS, Rate limiting

## Frontend (Flutter)

### Estrutura

```
lib/
├── core/
│   ├── api_client.dart                 # Cliente HTTP (package:http, timeout 10s)
│   ├── config.dart                     # URL da API em runtime (RNF06)
│   │
│   ├── websocket/
│   │   └── websocket_service.dart      # Serviço WebSocket (wss derivado)
│   │
│   └── permissions/
│       ├── location_permission.dart    # Permissão de localização
│       └── microphone_permission.dart  # Permissão de microfone
│
├── features/
│   ├── auth/
│   │   ├── screens/
│   │   │   ├── login_screen.dart
│   │   │   └── register_screen.dart
│   │   ├── widgets/
│   │   ├── services/
│   │   │   └── auth_service.dart
│   │   └── models/
│   │       └── user_model.dart
│   │
│   ├── channels/
│   │   ├── screens/
│   │   │   ├── channels_screen.dart
│   │   │   └── channel_detail_screen.dart
│   │   ├── widgets/
│   │   ├── services/
│   │   │   └── channel_service.dart
│   │   └── models/
│   │       └── channel_model.dart
│   │
│   ├── voice/
│   │   ├── screens/
│   │   │   └── voice_screen.dart
│   │   ├── widgets/
│   │   │   ├── ptt_button.dart
│   │   │   ├── audio_visualizer.dart
│   │   │   └── audio_player.dart
│   │   ├── services/
│   │   │   ├── audio_service.dart
│   │   │   ├── ptt_service.dart
│   │   │   └── audio_queue.dart
│   │   └── models/
│   │       └── audio_packet.dart
│   │
│   ├── location/
│   │   ├── screens/
│   │   ├── widgets/
│   │   ├── services/
│   │   │   └── location_service.dart
│   │   └── models/
│   │       └── location_model.dart
│   │
│   ├── profile/
│   │   ├── screens/
│   │   │   └── profile_screen.dart
│   │   ├── widgets/
│   │   ├── services/
│   │   │   └── profile_service.dart
│   │   └── models/
│   │
│   └── settings/
│       └── widgets/
│           └── server_url_sheet.dart    # Sheet de URL do servidor (RNF06)
│
├── main.dart
└── app.dart
```

### Configuração do servidor (RNF06)

- `core/config.dart` (`AppConfig`) mantém a URL da API em **runtime**: semente
  do build (`--dart-define=API_BASE_URL`), sobrescrita pela URL salva pelo
  usuário em `shared_preferences` (chave `server_url`); `main()` chama
  `AppConfig.load()` antes do `runApp`.
- `wsBaseUrl` deriva automaticamente (`https:` → `wss:`, `http:` → `ws:`).
- Validação em `normalizeUrl`: `trim`, remoção de `/` final, só `http://`/`https://`.
- UI: `features/settings/widgets/server_url_sheet.dart` (bottom sheet com
  **Testar** via `GET /health` e **Salvar**), acessível no Login ("Configurar
  servidor") e no Perfil ("Servidor").

### Módulos

#### Auth Feature
- **Responsabilidade**: Login e registro
- **Screens**: Login, Register
- **Services**: Auth service (JWT; OAuth fora do escopo v1)

#### Channels Feature
- **Responsabilidade**: Gestão de canais
- **Screens**: Lista de canais, Detalhe do canal
- **Services**: Channel service

#### Voice Feature
- **Responsabilidade**: PTT e áudio
- **Screens**: Voice screen
- **Widgets**: PTT button, Audio visualizer, Audio player
- **Services**: Audio service, PTT service, Audio queue

#### Location Feature
- **Responsabilidade**: GPS e proximidade
- **Services**: Location service

#### Profile Feature
- **Responsabilidade**: Perfil do usuário
- **Screens**: Profile screen
- **Services**: Profile service

#### Settings Feature
- **Responsabilidade**: Configuração do servidor (RNF06)
- **Widgets**: Server URL sheet (Testar/Salvar)

## Comunicação

### REST (Operações Tradicionais)

```
┌─────────────┐      HTTP/HTTPS      ┌─────────────┐
│   Flutter   │ ───────────────────► │   Go API    │
│   App       │ ◄─────────────────── │   Server    │
└─────────────┘      JSON Response   └─────────────┘
```

**Endpoints:**
- `POST /api/v1/auth/register` - Cadastro
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/channels` - Listar canais
- `POST /api/v1/channels/join` - Entrar no canal
- `POST /api/v1/channels/leave` - Sair do canal

> Localização não tem endpoint próprio: o app envia as coordenadas do GPS nos
> payloads da API de canais (criar/listar próximos). O backend calcula a
> distância com `internal/location.CalculateDistance`.

### WebSocket (Tempo Real)

```
┌─────────────┐      WebSocket       ┌─────────────┐
│   Flutter   │ ◄──────────────────► │   Go WS     │
│   App       │      Bidirecional    │   Server    │
└─────────────┘                      └─────────────┘
```

**Endpoint:**
- `ws://<host>:<port>/ws/audio/?channel=<id>&user_id=<id>&token=<JWT>` —
  streaming de áudio (autenticado por query param; `wss://` quando a base é
  `https://`, derivado automaticamente no app)

## Fluxo de Voz (Crítico)

```
┌─────────────────────────────────────────────────────────────┐
│                    Fluxo de Transmissão                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Usuário pressiona PTT                                   │
│       │                                                     │
│       ▼                                                     │
│  2. Flutter captura áudio do microfone                       │
│       │                                                     │
│       ▼                                                     │
│  3. Áudio em PCM16 (sem compressão — divergência
│     da spec "Opus/AAC", registrada em PROJECT_STATE)
│       │                                                     │
│       ▼                                                     │
│  4. Enviado via WebSocket                                   │
│       │                                                     │
│       ▼                                                     │
│  5. Go Backend recebe pacote                                │
│       │                                                     │
│       ▼                                                     │
│  6. Verifica canal e disponibilidade                        │
│       │                                                     │
│       ▼                                                     │
│  7. Transmite para outros usuários do canal                 │
│       │                                                     │
│       ▼                                                     │
│  8. Flutter recebe via WebSocket                            │
│       │                                                     │
│       ▼                                                     │
│  9. Áudio decodificado                                      │
│       │                                                     │
│       ▼                                                     │
│  10. Adicionado à fila de áudio                             │
│       │                                                     │
│       ▼                                                     │
│  11. Reproduzido no alto-falante                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Fluxo de Conexão (Resiliência)

> **Status (27/10/2026)**: a reconexão com backoff existe no app (5 tentativas,
> validação em campo pendente) e o **buffer de replay está conectado**: o relay
> grava no `AudioBufferService` (memória, 30s) com `sequence` no payload, e após
> reconexão o app envia `request_replay` para receber o áudio perdido (RF06/RNF04).
> `BufferManager` e `message_buffer_model.dart` seguem sem uso (fora de escopo).

```
┌─────────────────────────────────────────────────────────────┐
│                    Fluxo de Reconexão                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Estado: Conectado                                          │
│       │                                                     │
│       ▼                                                     │
│  WebSocket ativo                                            │
│       │                                                     │
│       ▼                                                     │
│  Queda de conexão detectada                                 │
│       │                                                     │
│       ▼                                                     │
│  Iniciar tentativas de reconexão                            │
│       │                                                     │
│       ▼                                                     │
│  Servidor mantém áudio dos últimos 30s em buffer            │
│       │                                                     │
│       ▼                                                     │
│  Reconexão bem-sucedida                                     │
│       │                                                     │
│       ▼                                                     │
│  request_replay {room, since_sequence} → catch-up          │
│       │                                                     │
│       ▼                                                     │
│  Segmentos perdidos reenfileirados (AudioQueue)             │
│       │                                                     │
│       ▼                                                     │
│  Buffer do servidor expira por idade (30s)                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Segurança

### Autenticação

1. **JWT**: Tokens stateless para APIs REST (access + refresh, HS256)
2. ~~**OAuth**~~: Login social via Google **fora do escopo v1** (DEC-011)
3. ~~**Casbin**~~: Controle de acesso baseado em papéis **fora do escopo v1** (DEC-011)

### Criptografia

1. **TLS**: Em produção, obrigatório; no protótipo entregue pelo túnel ngrok (DEC-013)
2. **bcrypt**: Senhas hasheadas
3. **JWT Secret**: Nunca em repositório (fallback de dev com aviso no log)

### Autorização

1. **Middlewares**: Verificação em cada requisição REST e no handshake do WS
2. ~~**Roles**~~: Admin/User não implementados (conta única, escopo v1)

## Performance

### Metas

- **Latência de áudio**: < 500ms
- **Alocação de canal**: < 5 segundos
- **Detecção de proximidade**: 100m de precisão
- **Buffer**: 30 segundos
- **Usuários concorrentes**: 10.000+ por região

### Otimizações

1. **WebSocket**: Conexões persistentes
2. **Buffer**: Armazenamento em memória
3. **Índices**: Otimização de queries (ver `docs/banco-de-dados.md`)
4. ~~**Cache**~~: Redis **não utilizado** (sem dependência no `go.mod`; ver DEC-011)
