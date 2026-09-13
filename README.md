# RadioPX - Simulação Digital de Rádio PX

## Descrição

Aplicativo mobile que simula um sistema de rádio PX (Private Exchange) para comunicação por voz em tempo real baseada na proximidade dos usuários. O sistema permite que pessoas dentro de um raio de 10km se comuniquem via push-to-talk, com canais dinâmicos e sem armazenamento permanente de áudio.

## Objetivo

Permitir que usuários dentro de um raio de 10km se comuniquem via push-to-talk, com canais dinâmicos a cada 10 usuários e sem armazenamento permanente de áudio, garantindo privacidade e resiliência de conexão.

## Tecnologias

| Componente | Tecnologia | Finalidade |
|-----------|-----------|------------|
| **Frontend** | Flutter/Dart | App mobile cross-platform (iOS/Android) |
| **Backend** | Go/Golang | Servidor API de alta performance |
| **Banco de Dados** | PostgreSQL | Dados permanentes (usuários, canais) |
| **Comunicação** | REST/HTTP | Operações tradicionais |
| **Tempo Real** | WebSocket | Streaming de áudio bidirecional |
| **Autenticação** | JWT/OAuth (Google)/Casbin | Sessões e controle de acesso |
| **Localização** | GPS do dispositivo | Serviços de localização |

## Como Executar

### Pré-requisitos

- **Backend**: Go 1.21+, PostgreSQL 14+
- **Frontend**: Flutter 3.16+, Dart SDK
- **Desenvolvimento**: Git, IDE (VS Code/GoLand)

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

### Variáveis de Ambiente

```bash
# Backend
DB_HOST=localhost
DB_PORT=5432
DB_NAME=radiopx
DB_USER=postgres
DB_PASSWORD=your-password
SERVER_PORT=8080
JWT_SECRET=your-secret-key
GOOGLE_CLIENT_ID=your-google-client-id
PROXIMITY_RADIUS_KM=10
MAX_USERS_PER_CHANNEL=10
MAX_CHANNELS_VISIBLE=15

# Frontend
API_BASE_URL=http://localhost:8080
WS_BASE_URL=ws://localhost:8080
GOOGLE_CLIENT_ID=your-google-client-id
```

## Estrutura do Projeto

```
RadioPx/
├── backend/                        # Backend Go
│   ├── cmd/server/                 # Ponto de entrada
│   │   └── main.go
│   ├── internal/                   # Pacotes privados
│   │   ├── auth/                  # Autenticação
│   │   ├── user/                  # Gestão de usuários
│   │   ├── channel/               # Alocação de canais
│   │   ├── voice/                 # PTT e streaming de áudio
│   │   ├── location/              # GPS e proximidade
│   │   └── middleware/            # Auth, rate limiting
│   ├── pkg/                       # Pacotes públicos
│   │   ├── websocket/            # Handling WebSocket
│   │   └── audio/                # Processamento de áudio
│   └── go.mod                     # Módulo Go
│
├── frontend/                       # App Flutter
│   ├── lib/
│   │   ├── core/                  # Utilitários compartilhados
│   │   │   ├── network/
│   │   │   ├── websocket/
│   │   │   └── permissions/
│   │   ├── features/              # Módulos de funcionalidades
│   │   │   ├── auth/
│   │   │   ├── channels/
│   │   │   ├── voice/
│   │   │   ├── location/
│   │   │   └── profile/
│   │   └── main.dart
│   ├── android/                   # Código específico Android
│   ├── ios/                       # Código específico iOS
│   └── pubspec.yaml               # Dependências Flutter
│
├── docs/                          # Documentação
│   ├── requisitos.md              # Requisitos do sistema
│   ├── arquitetura.md             # Arquitetura do sistema
│   ├── banco-de-dados.md          # Design do banco de dados
│   └── decisoes.md                # Decisões técnicas
│
├── specs/                          # Especificações de features
├── AGENTS.md                       # Regras para OpenCode
├── CHANGELOG.md                    # Histórico de alterações
└── PROJECT_STATE.md                # Estado atual do projeto
```

## Status Atual

- **Fase**: Inicialização do Projeto
- **Requisitos definidos**: ✅
- **Arquitetura definida**: ✅
- **Documentação inicial**: ✅
- **Backend**: ⏳ Pendente
- **Frontend**: ⏳ Pendente
- **Última atualização**: 08/09/2026
