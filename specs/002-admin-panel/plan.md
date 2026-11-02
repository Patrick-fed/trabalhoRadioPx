# Implementation Plan: Admin Panel

**Branch**: `feature-login` | **Date**: 27/10/2026 | **Spec**: `docs/requisitos.md` → RF09
**Input**: Decisões aprovadas em 27/10/2026 (SPA embed, métricas + gestão de usuários, `is_admin` + `ADMIN_EMAILS`)
**Status (02/11/2026)**: Fases A–E concluídas; Fase F (commit/push) pendente de aprovação

## Summary

Painel administrativo web servido pelo próprio backend Go (`go:embed`, rota
`/admin`), com **dashboard de métricas** e **gestão de usuários**
(listar/suspender/reativar/excluir). Identidade administrativa vem do novo
campo `is_admin` (migração 003) com bootstrap idempotente pela env
`ADMIN_EMAILS`. Requisito novo **RF09** em `docs/requisitos.md`; decisão
**DEC-017** em `docs/decisoes.md`.

## Technical Context

**Language/Version**: Go 1.21 (backend) · HTML/JS/CSS vanilla (SPA, sem build)
**Primary Dependencies**: `gorilla/websocket` (já usado), `rs/cors` (já usado)
**Storage**: PostgreSQL — nova coluna `users.is_admin`; `users.is_active` já existe
**Testing**: `go test ./...` (httptest para handlers) · `flutter test` intocado
**Target Platform**: Servidor local + túnel ngrok (demo)
**Constraints**: sem novo projeto/dependência; sem mudança de claims JWT;
RNF01 (áudio nunca persistido) intacto
**Scale/Scope**: ~6 endpoints novos + 1 tela; v1 com métricas e usuários apenas

## Decisões aprovadas

1. **Forma**: SPA leve embutida no Go (`//go:embed` em `/admin`) — sem
   Flutter Web, sem tela no app mobile, sem projeto separado
2. **Escopo v1**: Métricas do sistema + Gerenciar usuários. **Fora**:
   gestão de canais, moderação de voz (kick), auditoria, roles/Casbin
3. **Identidade**: coluna `is_admin` + bootstrap via env `ADMIN_EMAILS`
   (csv, idempotente no boot). JWT sem novas claims — `RequireAdmin` consulta
   o banco a cada request (fonte da verdade), promoção vale sem novo token

## Constitution Check (regras do AGENTS.md)

- RF09 criado em `docs/requisitos.md` antes da feature ✅
- Monolito modular: novo pacote `internal/admin` (não microsserviço) ✅
- PostgreSQL mantido; áudio nunca persistido ✅
- Senhas sempre bcrypt; JWT secret não commitado; env só em `.env.example` ✅
- Documentação (CHANGELOG, PROJECT_STATE, arquitetura, decisões) na entrega ✅

## Project Structure

### Documentation (this feature)

```text
specs/002-admin-panel/
├── plan.md                  # este arquivo
docs/
├── requisitos.md            # RF09 (novo)
├── decisoes.md              # DEC-017 (novo)
├── banco-de-dados.md        # migração 003
├── arquitetura.md           # pacote internal/admin + rotas
README.md                    # status + como acessar
CHANGELOG.md                 # entrada 27/10
PROJECT_STATE.md             # estado/pendências
AGENTS.md                    # seção Regras de API
backend/.env.example         # ADMIN_EMAILS
```

### Source (backend)

```text
backend/
├── cmd/server/main.go               # bootstrap ADMIN_EMAILS, rotas /api/v1/admin/*
├── internal/db/migrations/003_admin_users.sql   # novo
├── internal/user/{model,repository,service}.go   # IsAdmin + SetUserActive/SetUserAdmin + Authenticate(!IsActive)
├── internal/middleware/auth.go      # RequireAdmin
├── internal/voice/service.go        # ActiveChannelsCount()
├── pkg/websocket/hub.go             # GetLiveRoomCount()
└── internal/admin/                  # novo pacote
    ├── handler.go                   # rotas admin
    ├── stats.go                     # montagem das métricas
    ├── handler_test.go
    └── web/{index.html,app.js,style.css}  # SPA embutida
```

## Fases

### Fase A — `is_admin` no backend ✅ (02/11/2026)
1. `003_admin_users.sql`: `ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;` + índice parcial
2. `user/model.go`: campo `IsAdmin bool` (`json:"is_admin" db:"is_admin"`)
3. `user/repository.go`: `is_admin` em `userColumns`, `Create`, `Update` e todos os `Scan`
4. `user/service.go`:
   - `Authenticate` rejeita `!IsActive` (corrige pendência conhecida — suspensão passa a valer no login; erro vira 401 genérico)
   - `SetUserActive(userID, active)` e `SetUserAdmin(userID, admin)` com guarda:
     admin não pode suspender, demover nem excluir **a si mesmo** (`ErrAdminSelfAction`)
5. `cmd/server/main.go`: env `ADMIN_EMAILS` (csv) → promove idempotente no boot (log por email)
6. `backend/.env.example`: `ADMIN_EMAILS=`
7. Testes em `user/service_test.go` (+6: inactive login, self-suspend, bootstrap…)

### Fase B — Endpoints + autorização ✅ (02/11/2026)
1. `middleware.RequireAdmin(isAdmin func(string) bool)`: 401 sem token, 403 se não-admin (após `AuthMiddleware`)
2. Novo pacote `internal/admin` (deps injetadas: user service, channel store, voice service, hub):
   - `GET    /api/v1/admin/stats` → `{users_total, users_active, channels_total, ws_connections, live_rooms, transmitting_channels}`
   - `GET    /api/v1/admin/users` → lista sem `password_hash` (campo `json:"-"`)
   - `PATCH  /api/v1/admin/users/{id}` → body `{"is_active"|"is_admin": bool}`
   - `DELETE /api/v1/admin/users/{id}`
3. Getters novos: `hub.GetLiveRoomCount()`, `VoiceService.ActiveChannelsCount()`
4. Rotas no `main.go` no padrão `injector` (Bearer → `X-User-ID` → `RequireAdmin`)
5. Testes: `internal/admin/handler_test.go` (11 casos)

### Fase C — SPA `/admin` ✅ (02/11/2026)
1. `internal/admin/web/` com `//go:embed` (`spa.go`); handler serve em `/admin`, `/admin/*` com fallback index
2. UI em vanilla JS: **Login** (`POST /api/v1/auth/login` → token em memória),
   **Dashboard** (cards + refresh 10s), **Usuários** (busca, badges, Suspender/
   Reativar/Promover/Demover/Excluir com `confirm`), **Logout**; trata 401/403/429
3. Mesmo host do backend → sem CORS; login já protegido pelo rate limit de auth

### Fase B2 — Localização passiva ✅ (02/11/2026, DEC-018)
1. `004_user_last_location.sql`: `last_lat`, `last_lng`, `last_location_at` (nullable)
2. Model/repo/service com ponteiros + `UpdateLastLocation`
3. `channel/handler.go`: interface `LocationUpdater` + `recordLocation` best-effort em nearby/create
4. SPA: coluna Localização com link OSM (ou "—")

### Fase D — Verificação ✅ (02/11/2026)
- `go build ./...` + `go vet ./...` + `go test ./... -count=1` → verdes
- `flutter test` (57/57) e `flutter analyze` (14, baseline) — intocado
- E2E manual: `/admin` 200, assets 200, API 401/403, self-patch 403,
  login→stats→users com `last_*` no payload; psql confirma gravação
- `scripts/test-backend.ps1` → **28/28** (não estendido, conforme plano)

### Fase E — Documentação ✅ (02/11/2026)
- `docs/requisitos.md`: **RF09 — Painel Administrativo** + rastreabilidade + nota
- `docs/decisoes.md`: **DEC-017** (SPA embutida/is_admin) e **DEC-018** (localização passiva)
- `docs/banco-de-dados.md`: migrações 003 e 004
- `docs/arquitetura.md`: módulo Admin, árvore, endpoints
- `README.md`: seção "Painel Administrativo" + `ADMIN_EMAILS`
- `AGENTS.md`: seção Admin em "Regras de API"
- `CHANGELOG.md` (02/11/2026) + `PROJECT_STATE.md`

### Fase F — Entrega ⏳ (pendente de aprovação)
- Commit `feat(admin): painel administrativo com metricas e gestao de usuarios (RF09)`
- Push `feature-login` → sync FF `dev` e `main` (padrão aprovado)

## Riscos / Pendências

- Token de usuário **suspenso** segue válido até expirar (24h); efeito
  garantido no próximo login — revogação ativa fica para futuro
- `userColumns`/`Create`/`Update`/`Scan`: todos os pontos do repositório
  precisam do `is_admin` (esquecimento = null scan / coluna não persistida)
- `Authenticate` passando a checar `is_active` pode afetar dados de teste
  existentes (usuários já inativos) — verificar testes
- Escopo fora da v1: gestão de canais, moderação de voz, auditoria,
  roles no JWT, Casbin (RNF05 segue futuro)
