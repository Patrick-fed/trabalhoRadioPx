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
│  │  PostgreSQL │  │   Redis     │  │  WebSocket  │ │
│  │  Database   │  │   Cache     │  │   Server    │ │
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
│   │   ├── handler.go                 # HTTP handlers
│   │   ├── service.go                 # Lógica de negócio
│   │   ├── repository.go              # Acesso a dados
│   │   └── middleware.go              # JWT middleware
│   │
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── channel/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── voice/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── websocket.go               # WebSocket handling
│   │   └── buffer.go                  # Buffer temporário
│   │
│   ├── location/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   └── middleware/
│       ├── auth.go                    # Autenticação
│       ├── cors.go                    # CORS
│       └── ratelimit.go               # Rate limiting
│
├── pkg/
│   ├── websocket/
│   │   ├── hub.go                     # Hub de conexões
│   │   └── client.go                  # Cliente WebSocket
│   │
│   └── audio/
│       ├── codec.go                   # Codificação de áudio
│       └── buffer.go                  # Buffer de áudio
│
├── migrations/
│   └── 001_initial.sql                # Migrações do banco
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
- **Endpoints**: `/api/v1/users/*`

#### Channel Module
- **Responsabilidade**: Alocação e gestão de canais
- **Dependências**: User module, Location service
- **Endpoints**: `/api/v1/channels/*`

#### Voice Module
- **Responsabilidade**: PTT e streaming de áudio
- **Dependências**: Channel module, WebSocket
- **Endpoints**: `/ws/audio/{channel_id}`

#### Location Module
- **Responsabilidade**: GPS e cálculo de proximidade
- **Dependências**: Nenhuma
- **Endpoints**: `/api/v1/location/*`

#### Middleware
- **Responsabilidade**: Interceptadores HTTP
- **Dependências**: Auth module
- **Uso**: Autenticação, CORS, Rate limiting

## Frontend (Flutter)

### Estrutura

```
lib/
├── core/
│   ├── network/
│   │   ├── api_client.dart            # Cliente HTTP
│   │   ├── dio_client.dart            # Configuração Dio
│   │   └── interceptors.dart          # Interceptadores
│   │
│   ├── websocket/
│   │   ├── websocket_service.dart     # Serviço WebSocket
│   │   └── websocket_manager.dart     # Gerenciador
│   │
│   ├── permissions/
│   │   ├── location_permission.dart   # Permissão de localização
│   │   └── microphone_permission.dart # Permissão de microfone
│   │
│   └── utils/
│       ├── constants.dart             # Constantes
│       └── helpers.dart               # Funções auxiliares
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
│   └── profile/
│       ├── screens/
│       │   └── profile_screen.dart
│       ├── widgets/
│       ├── services/
│       │   └── profile_service.dart
│       └── models/
│
├── main.dart
└── app.dart
```

### Módulos

#### Auth Feature
- **Responsabilidade**: Login e registro
- **Screens**: Login, Register
- **Services**: Auth service (JWT, OAuth)

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
- `POST /api/v1/location/update` - Atualizar localização

### WebSocket (Tempo Real)

```
┌─────────────┐      WebSocket       ┌─────────────┐
│   Flutter   │ ◄──────────────────► │   Go WS     │
│   App       │      Bidirecional    │   Server    │
└─────────────┘                      └─────────────┘
```

**Endpoint:**
- `ws://localhost:8080/ws/audio/{channel_id}` - Streaming de áudio

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
│  3. Áudio codificado/comprimido (Opus/AAC)                  │
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
│  Áudio recebido armazenado em buffer (30s)                  │
│       │                                                     │
│       ▼                                                     │
│  Reconexão bem-sucedida                                     │
│       │                                                     │
│       ▼                                                     │
│  Replay automático de mensagens perdidas                    │
│       │                                                     │
│       ▼                                                     │
│  Buffer limpo após replay                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Segurança

### Autenticação

1. **JWT**: Tokens stateless para APIs REST
2. **OAuth**: Login social via Google
3. **Casbin**: Controle de acesso baseado em papéis

### Criptografia

1. **TLS**: Comunicação em trânsito
2. **bcrypt**: Senhas hasheadas
3. **JWT Secret**: Nunca em repositório

### Autorização

1. **Middlewares**: Verificação em cada requisição
2. **Roles**: Admin, User
3. **Permissions**: Baseadas em módulos

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
3. **Índices**: Otimização de queries
4. **Cache**: Redis para dados frequentes
