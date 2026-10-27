# RadioPX - Simulação Digital de Rádio PX

## Descrição

Aplicativo mobile que simula um sistema de rádio PX (Private Exchange) para comunicação por voz em tempo real baseada na proximidade dos usuários. O sistema permite que pessoas dentro de um raio de 10km se comuniquem via push-to-talk, com canais dinâmicos e sem armazenamento permanente de áudio.

Arquitetura **monolítica modular**: backend Go único com pacotes internos por responsabilidade, app Flutter cross-platform e PostgreSQL para persistência de usuários/canais. Áudio nunca é persistido (requisito crítico de privacidade).

## Objetivo

Permitir que usuários dentro de um raio de 10km se comuniquem via push-to-talk, com canais dinâmicos a cada 10 usuários e sem armazenamento permanente de áudio, garantindo privacidade e resiliência de conexão.

## Tecnologias

| Componente | Tecnologia | Finalidade |
|-----------|-----------|------------|
| **Frontend** | Flutter/Dart | App mobile cross-platform (Android/iOS) |
| **Backend** | Go/Golang | Servidor API monolítico modular |
| **Banco de Dados** | PostgreSQL | Dados permanentes (usuários, canais, memberships) |
| **Comunicação** | REST/HTTP | Operações tradicionais (JSON) |
| **Tempo Real** | WebSocket | Streaming de áudio bidirecional (JWT no handshake) |
| **Autenticação** | JWT (access + refresh) | Sessões stateless (OAuth/Casbin fora do escopo v1) |
| **Exposição** | ngrok | Túnel HTTPS/WSS por sessão em desenvolvimento |
| **Localização** | GPS do dispositivo | Coordenadas enviadas nos payloads de canais |

## Como Executar

### Pré-requisitos

- **Backend**: Go 1.21+, PostgreSQL 14+ (obrigatório — o servidor falha rápido sem banco)
- **Frontend**: Flutter 3.16+, Dart SDK; Android SDK para APK
- **Desenvolvimento**: Git, IDE (VS Code/GoLand), ngrok (opcional, para expor o backend)

### Backend

```bash
cd backend
go mod download
# requer PostgreSQL rodando + backend/.env (ver abaixo); aplica migrações no boot
go run cmd/server/main.go
```

### Frontend (app)

```bash
cd frontend
flutter pub get
# API_BASE_URL é a semente do build; a URL pode ser editada no app (RNF06)
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

### Variáveis de Ambiente (`backend/.env`)

Cópie o exemplo: `cp .env.example .env` e ajuste.

```bash
# As únicas variáveis lidas pelo backend hoje:
SERVER_PORT=8080             # porta HTTP (padrão 8080)
JWT_SECRET=your-secret-key   # segredo JWT (fallback de dev com aviso no log)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=radiopx
DB_USER=postgres
DB_PASSWORD=your-password
DB_SSLMODE=disable           # enable em produção com TLS
```

> `API_BASE_URL` no app é passada via `--dart-define` no **build** e pode ser
> alterada dentro do app (Login → "Configurar servidor" ou Perfil → "Servidor").
> A URL WebSocket é derivada automaticamente (`https:` → `wss:`).

### Testes

```bash
# Backend (unidade)
cd backend && go test ./...

# Backend (integração com PostgreSQL real; requer DB_*)
RADIOPX_TEST_DB=1 go test ./internal/db/...

# Frontend (unit + widget)
cd frontend && flutter test

# E2E REST+WS contra o servidor rodando (opcional -BaseUrl para túnel)
powershell -File scripts/test-backend.ps1
powershell -File scripts/test-backend.ps1 -BaseUrl https://<tunnel>.ngrok-free.dev
```

### APK e ngrok (demonstração)

```bash
# 1. Expor o backend (após configurar o authtoken do ngrok)
ngrok http 8080
# URL pública: http://127.0.0.1:4040/api/tunnels → public_url (https)

# 2. Buildar o APK com a URL do túnel como semente
cd frontend
flutter build apk --release --dart-define=API_BASE_URL=https://<tunnel>.ngrok-free.dev

# 3. Copiar para a raiz (gitignored — não versionar o binário)
cp build/app/outputs/flutter-apk/app-release.apk ../RadioPX.apk
```

> No plano free do ngrok a URL muda por sessão — no app, atualize via
> "Configurar servidor" (não precisa rebuildar).

## Estrutura do Projeto

```
RadioPx/
├── backend/                        # Backend Go (monolito modular)
│   ├── cmd/server/                 # Ponto de entrada (main.go)
│   ├── internal/                   # Pacotes privados
│   │   ├── auth/                   # Autenticação (service)
│   │   ├── user/                   # Usuários (handler/service/stores)
│   │   ├── channel/                # Canais (handler/service/stores)
│   │   ├── db/                     # Open + runner de migrações
│   │   │   └── migrations/         # 001_initial.sql, 002_align_models.sql
│   │   ├── voice/                  # PTT, WebSocket e buffer (parcial)
│   │   ├── location/               # Cálculo de distância (Haversine)
│   │   └── middleware/             # Auth, CORS, rate limit, logging
│   ├── pkg/                        # Pacotes reutilizáveis
│   │   ├── websocket/              # Hub + client (voiceGuard)
│   │   ├── security/               # Config (reservado)
│   │   └── utils/                   # Helpers
│   └── go.mod
│
├── frontend/                       # App Flutter
│   ├── lib/
│   │   ├── core/                   # api_client, config (URL runtime), websocket, permissions
│   │   ├── features/               # auth, channels, voice, location, profile, settings
│   │   ├── app.dart                # MaterialApp + AuthGate + NavigationBar
│   │   └── main.dart               # AppConfig.load() + runApp
│   ├── android/ | ios/
│   └── pubspec.yaml
│
├── scripts/
│   └── test-backend.ps1            # E2E REST+WS (28 checks)
│
├── docs/
│   ├── requisitos.md               # RF/RNF (inclui RNF06)
│   ├── arquitetura.md              # Arquitetura do sistema
│   ├── banco-de-dados.md           # Schema PostgreSQL (pós-002)
│   └── decisoes.md                 # DEC-001..DEC-015
│
├── specs/                          # Especificações speckit
├── AGENTS.md                       # Regras para OpenCode
├── CHANGELOG.md                    # Histórico de alterações
└── PROJECT_STATE.md                # Estado atual do projeto
```

## Status Atual

- **Arquitetura**: monolito modular Go + Flutter + PostgreSQL
- **Fases do plano aprovado**: 0 (higiene) ✅, 1 (PostgreSQL) ✅, 2 (ngrok) ✅, 3 (URL editável/RNF06) ✅ — 4 (docs) em andamento
- **Protótipo funcional**: backend Go com auth JWT + canais + voz via WebSocket; app Flutter com APK Android
- **Conhecido e mapeado**: OAuth/Casbin fora do escopo v1; buffer/replay de voz não conectado; validação em campo pendente de alguns itens
- **Estado e pendências**: `PROJECT_STATE.md` → *Lacunas / Aderência aos Requisitos*
- **Última atualização**: 27/10/2026
