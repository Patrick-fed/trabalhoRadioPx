# Estado Atual do Projeto RadioPX

Este documento registra o estado atual do projeto para facilitar a continuidade do desenvolvimento.

---

## Fase

**Desenvolvimento - Protótipo Funcional + APK Android (27/10/2026 - Fase 4: documentação final)**

Em 27/10/2026 a Fase 4 fechou a documentação do plano aprovado: `banco-de-dados.md`
reescrito para o schema pós-`002`, `decisoes.md` com DEC-012..015 (persistência
PG+DIP, ngrok/TLS/CORS, rate limit global, URL editável), `arquitetura.md` corrigido
(árvores reais, endpoints, Redis/OAuth/Casbin marcados), `requisitos.md` com nota
atualizada, `README.md` reescrito e `.env.example` limpo. Ver
[CHANGELOG](CHANGELOG.md) de 27/10/2026.

Em 27/10/2026 a Fase 3 entrou em vigor: a URL da API é editável em runtime
pelo app (RNF06), com semente do build (`--dart-define`), persistência em
`shared_preferences` e sheet com Testar/Salvar acessível no Login e no Perfil.
Trocar de URL (ex.: ngrok reiniciado) não exige mais novo build. Ver
[CHANGELOG](CHANGELOG.md) de 27/10/2026.

Em 27/10/2026 a Fase 2 entrou em vigor: o backend fica exposto por túnel ngrok
(`ngrok http 8080`) com URL HTTPS por sessão — o APK passa a ser buildado com
`--dart-define=API_BASE_URL=<url do túnel>`, eliminando a dependência do IP de
DHCP local. CORS do backend agora aceita `https://*.ngrok-free.{app,dev}` e o
script de testes envia o header de bypass do "browser warning". Ver
[CHANGELOG](CHANGELOG.md) de 27/10/2026.

Em 20/10/2026 a Fase 1 do plano aprovado entrou em vigor: usuários, canais e memberships
agora são persistidos em PostgreSQL (pacote `internal/db` + stores PG em `user`/`channel`),
com runner de migrações embutido (`001` idempotente + `002_align_models`). Localização
segue em memória e áudio nunca é persistido (escopo aprovado). Ver
[CHANGELOG](CHANGELOG.md) de 20/10/2026.

Em 11/10/2026 o login passou por nova auditoria: corrigido um bypass do rate limit via headers de
encaminhamento, IDs de usuário passaram a ser UUIDs, sessão expirada sem refresh é limpa e a detecção
de e-mail duplicado no cadastro voltou a funcionar.

> **Atenção (13/10/2026)**: as listas de requisitos e o quadro de progresso abaixo **não refletem**
> o estado real do código. O mapeamento honesto de lacunas e de escopo está na seção
> [Lacunas / Aderência aos Requisitos](#lacunas--aderência-aos-requisitos).

---

## Lacunas / Aderência aos Requisitos

> **Fonte única de verdade** para o que ainda falta. Atualizado em 27/10/2026 (Fase 4).
> As seções de requisitos/progresso deste documento são anteriores a esta auditoria.

### Lacunas dentro do escopo v1 (a corrigir)

- ~~**RF03 — "apenas um usuário fala por vez" não é aplicado no fluxo real.**~~
  **Resolvido na `feature-login` (commit `f8251e8`)**: o WebSocket passou a aplicar
  `VoiceService.StartTransmission`/`StopTransmission` via `voiceGuard`
  (`backend/pkg/websocket/client.go:165-172`) — um segundo transmissor no mesmo
  canal é rejeitado. Resta validação em campo.

- **RF06 / RNF04 — buffer de 30s e replay são código morto.**
  `backend/internal/voice/buffer_service.go` e `buffer_manager.go` só são instanciados em testes
  (`buffer_service_test.go`); nada disso está ligado ao WebSocket. No frontend,
  `frontend/lib/features/voice/services/audio_queue.dart` e
  `frontend/lib/features/voice/models/message_buffer_model.dart` **não são importados** em lugar
  nenhum.

- **Codec — divergência de documentação.** O pacote declara `Codec: "opus"`
  (`backend/internal/voice/websocket.go:44`), mas o áudio trafegado é **PCM16 cru** em base64, sem
  compressão. Decidir se vira lacuna ou se é aceitável para v1.

- ~~**Distância do canal nunca é serializada.**~~ **Resolvido na `feature-login`**: o modelo
  `Channel` tem campo `distance,omitempty` e `GetNearbyChannels` o popula; o frontend
  (`channel_model.dart`) já faz parse e renderiza (`distanceFormatted`).

- **Testes de integração/E2E.** Backend ganhou `internal/db/pg_integration_test.go` (opt-in via
  `RADIOPX_TEST_DB=1`; cobre stores PG + migrações contra banco real) e há o script
  `scripts/test-backend.ps1` (28 checks, exige servidor rodando). Ainda não há teste
  automatizado do fluxo de áudio ponta-a-ponta no CI.

### Trabalho futuro — fora de escopo v1 (decisão explícita, 13/10/2026)

Estes itens **não** serão implementados na v1 (persistência em PostgreSQL desde 20/10):

- **RF02 — Google OAuth.** `CreateOAuthUser` existe, mas não há integração Google nem endpoint
  OAuth. (Ver tambem DEC-002 e DEC-011 em `docs/decisoes.md`.)
- **RNF05 — Casbin (controle de acesso por papéis).** Não implementado.
- **RNF05 — TLS.** `main.go` usa `ListenAndServe` HTTP puro (`backend/cmd/server/main.go:124`);
  no protótipo o TLS é entregue pelo túnel ngrok (DEC-013).
- ~~**DEC-001 / FR-013 — PostgreSQL.**~~ **Entregue em 20/10/2026 (Fase 1)**: driver
  `jackc/pgx/v5` v5.7.4 no `go.mod`, runner de migrações em `internal/db` e stores PG
  para `users`/`channels`/`user_channels`.
- **Deploy e monitoramento.** Sem `Dockerfile`/`docker-compose`, sem métricas.

### Débito técnico (código morto, sem chamadas)

- `backend/internal/user/handler.go` — `HandleUpdateUser`.
- `backend/internal/middleware/cors.go` e `logging.go` (o `main.go` usa `github.com/rs/cors`).
- `frontend/lib/features/channels/screens/channel_detail_screen.dart` (nunca é navegada).

### Documentação — estado em 27/10/2026 (Fase 4)

- ~~`docs/arquitetura.md`~~ — corrigido em 27/10 (árvores reais, endpoints, Redis/OAuth/Casbin marcados).
- ~~`docs/requisitos.md`~~ — nota de status atualizada em 27/10 (RF03/PG/RNF06).
- `specs/001-px-radio-simulation/tasks.md` → T001–T073 marcados `[x]`, mas T045–T051
  (buffer/replay) e a parte de OAuth de T052–T061 não foram entregues (**fora de escopo**).
- ~~`README.md`~~ — reescrito em 27/10 (status real, testes, APK/ngrok, env vars).
- `docs/banco-de-dados.md` e `docs/decisoes.md` atualizados na Fase 4 (schema pós-002; DEC-012..015).

---

## Requisitos Concluídos

> Lista histórica. "Definido" significa que o requisito está **especificado**, não necessariamente
> implementado. Ver [Lacunas / Aderência aos Requisitos](#lacunas--aderência-aos-requisitos).

### Requisitos Funcionais

- ✅ RF01 - Criação de Conta (Definido)
- ✅ RF02 - Login (Definido)
- ✅ RF03 - Push-to-Talk (Definido)
- ✅ RF04 - Descoberta por Proximidade (Definido)
- ✅ RF05 - Alocação Dinâmica de Canais (Definido)
- ✅ RF06 - Buffer de Conexão (Definido)
- ✅ RF07 - Gestão de Perfil (Definido)
- ✅ RF08 - Fila de Áudio (Definido)

### Requisitos Não Funcionais

- ✅ RNF01 - Privacidade (Definido)
- ✅ RNF02 - Desempenho (Definido)
- ✅ RNF03 - Escalabilidade (Definido)
- ✅ RNF04 - Resiliência (Definido)
- ✅ RNF05 - Segurança (Definido)
- ✅ RNF06 - Configurabilidade do servidor (adicionado em 27/10/2026, Fase 3)

---

## Implementação Concluída

### Phase 1: Setup ✅

- ✅ T001 - Go module com dependências
- ✅ T002 - Configuração de linting Go
- ✅ T003 - Configuração Flutter com dependências
- ✅ T004 - Migração inicial do PostgreSQL

### Phase 2: Foundational ✅

- ✅ T005 - Schema do banco de dados
- ✅ T006 - Middleware de autenticação JWT
- ✅ T007 - Middleware CORS
- ✅ T008 - WebSocket hub para gerenciamento de conexões
- ✅ T009 - WebSocket client handler
- ✅ T010 - Estrutura de rotas da API
- ✅ T011 - Modelo User
- ✅ T012 - Modelo Channel
- ✅ T013 - Variáveis de ambiente
- ✅ T014 - Tratamento de erros e logging

---

## Em Desenvolvimento

- ✅ User Story 1: Push-to-Talk (P1 - MVP) - Concluída
- ✅ User Story 2: Descoberta por Proximidade (P2) - Concluída
- ✅ User Story 3: Alocação Dinâmica de Canais (P3) - Concluída
- ✅ User Story 4: Resiliência de Conexão (P4) - Concluída
- ✅ Phase 7: Auth Feature - Concluída
- ✅ Phase 8: Profile Feature - Concluída
- ✅ Phase 9: Polish & Cross-Cutting Concerns - Concluída

---

## Pendentes

### Backend

- ✅ Módulo Voice (handler, service, websocket; buffer parcial — ver Lacunas RF06)
- ✅ Módulo Location (apenas `distance.go`; handler/service removidos em 27/10)
- ✅ Módulo Channel (handler, service, stores)
- ✅ Módulo Auth (service; JWT)
- ✅ Módulo User (handler, service, stores)
- ✅ Persistência PostgreSQL (stores PG em user/channel, migrações em `internal/db`)
- ✅ Testes unitários (9 arquivos `_test.go`)
- ✅ Testes de integração (opt-in `RADIOPX_TEST_DB=1`)

### Frontend

- ✅ Feature Voice (screens, widgets, services, models)
- ✅ Feature Location (services, models)
- ✅ Feature Channels (screens, services, models)
- ✅ Feature Auth (screens, services, models)
- ✅ Feature Profile (screens, services, models)
- ✅ Feature Settings (server_url_sheet — RNF06)
- ✅ Core (api_client, config runtime, websocket, permissions)
- ⏳ Itens mortos do pubspec: `dio`, `provider`, `google_fonts`, `path_provider` (sem imports; remover quando seguro)

### Infraestrutura

- ⏳ Configuração de deploy
- ⏳ Monitoramento
- ⏳ Testes de carga

---

## Última Alteração

**27/10/2026 - Fase 4: documentação final do plano aprovado**

**Data**: 27/10/2026

**Descrição**:
- **`banco-de-dados.md`**: DDLs reescritos para o estado pós-`002_align_models.sql`
  (`users` com `provider`/`is_active`/`avatar` e `cpf` opcional; `channels` com
  localização/`created_by`); índices corrigidos (`idx_users_phone` inexistente
  removido; `idx_channels_created`/`idx_channels_location` incluídos); seção do
  runner de migrações e `schema_migrations`; consulta de bounding-box
- **`decisoes.md`**: DEC-012 (persistência PG + stores injetados/DIP), DEC-013
  (ngrok + TLS + CORS wildcard), DEC-014 (rate limit global por RemoteAddr),
  DEC-015 (URL editável em runtime / RNF06); nota de atualização na DEC-011
- **`arquitetura.md`**: árvores backend/frontend com arquivos reais (remove
  `dio_client`, `pkg/audio`, handler/location mortos; inclui `internal/db` e
  migrações 001+002); endpoints corrigidos (`/ws/audio/?channel=...`,
  `/api/v1/user/profile`); Redis/OAuth/Casbin/roles marcados como não
  implementados; fluxo de buffer/replay marcado como não conectado
- **`requisitos.md`**: nota de status atualizada (RF03 resolvido, PG desde 20/10,
  RNF06 entregue)
- **`README.md`**: reescrito — monolito modular declarado, env vars reais
  (`SERVER_PORT`/`JWT_SECRET`/`DB_*`), seções de Testes e APK/ngrok, árvore
  corrigida, status real
- **`backend/.env.example`**: limpo — só variáveis lidas pelo código
- **`PROJECT_STATE.md`**: esta entrada; Fase 4 marcada como concluída; Lacunas,
  Pendentes e Dependências atualizadas

**Verificação**: greps de consistência (sem `dio_client`, `idx_users_phone`,
`WS_BASE_URL`, Redis não-marcado); DEC-012..015 sem buraco de numeração. Sem
mudança de código Go/Dart (docs + `.env.example` apenas). Detalhes em
`CHANGELOG.md`.

---

**27/10/2026 - Fase 3: URL do servidor editável em runtime no app**

**Data**: 27/10/2026

**Descrição**:
- **RNF06** adicionado a `docs/requisitos.md` (configurabilidade do servidor)
- **`core/config.dart`**: `AppConfig.apiBaseUrl` vira runtime — semente do
  build + `load()`/`save()` com `shared_preferences` (`server_url`), `main()`
  carrega antes do `runApp`; `normalizeUrl`/`displayHost`/`resetForTest`
- **3 services** (`auth`/`channels`/`profile`): `baseUrl` de `final` para
  getter (mudança de URL vale sem nova instanciação)
- **Novo** `features/settings/widgets/server_url_sheet.dart` (sheet Testar
  via `GET /health` / Salvar); acessível no **Login** ("Configurar servidor")
  e no **Perfil** ("Servidor: <host>" em Ações); sessão mantida ao trocar
- **Testes**: `config_test.dart` (8) + `server_url_sheet_test.dart` (2);
  suíte **52/52**; APK 50,8MB rebuildado com `--dart-define` da URL do túnel

**Verificação**: `flutter test` 52/52; `flutter analyze` sem novos issues;
build de release OK. Detalhes em `CHANGELOG.md`.

---

**27/10/2026 - Fase 2: exposição do backend via ngrok**

**Data**: 27/10/2026

**Descrição**:
- **Túnel ngrok** (`ngrok http 8080`, agente 3.39.13 após `ngrok update` — o
  winget instalou 3.3.1, abaixo do mínimo da conta) entrega URL HTTPS por
  sessão; authtoken fica em `%LOCALAPPDATA%\ngrok\ngrok.yml` (fora do repo)
- **CORS**: `AllowedOrigins` em `cmd/server/main.go` ganhou
  `https://*.ngrok-free.app` e `https://*.ngrok-free.dev` (conta usa `.dev`);
  verificado GET com Origin + preflight OPTIONS (204 + ACAO correto)
- **`scripts/test-backend.ps1`**: header `ngrok-skip-browser-warning: 1` em
  `Invoke-Json` e no handshake WS (sem ele, o aviso ERR_NGROK_6024 substitui a
  resposta da API)
- **APK**: `flutter build apk --release --dart-define=API_BASE_URL=<url do
  túnel>` → 50,8MB copiado para `RadioPX.apk` (gitignored), URL conferida no
  binário
- **Problema nº 4 mitigado** (IP de DHCP não quebra mais o APK); **novo nº 9**
  (rate limit global atrás do túnel, decisão: manter por ora)

**Verificação**: `GET /health` via túnel → 200; `test-backend.ps1 -BaseUrl
<url>` → 28/28 (inclui relay wss); CORS GET+preflight corretos. Detalhes em
`CHANGELOG.md`.

---

**27/10/2026 - Limpeza de órfãos pós-Fase 1 (location e endpoint morto)**

**Data**: 27/10/2026

**Descrição**:
- **`internal/location` enxugado** para só `distance.go` (`EarthRadiusKm` +
  `CalculateDistance`, o único uso real via `internal/channel`); removidos
  `LocationService`, `LocationHandler` (nunca registrado) e os modelos órfãos
- **`POST /api/v1/location/update` removido** — stub que respondia 200 sem efeito;
  o app usa Geolocator local e envia coordenadas na API de canais (verificado:
  nenhum consumidor nem no frontend nem no `scripts/test-backend.ps1`)
- **RF03 marcado como resolvido** na seção de lacunas (guard de um falante por
  vez aplicado no WebSocket desde `f8251e8`)
- Docs: `arquitetura.md` (endpoint fora da tabela) e `requisitos.md` (RF04)

**Verificação**: `go build` + `go vet` + `go test ./...` verdes; grep sem
referências a `NewLocationService`/`NewLocationHandler`/`location/update`;
frontend intocado. Detalhes em `CHANGELOG.md`.

---

**20/10/2026 - Fase 1: persistência PostgreSQL (fim do estado em memória)**

**Data**: 20/10/2026

**Descrição**:
- **Nova dependência**: `github.com/jackc/pgx/v5` **v5.7.4** (fixada; última com
  `go 1.21`), via `database/sql` — registro completo com alternativas no CHANGELOG
- **Novo pacote `internal/db`**: `ConfigFromEnv`/`Open` (pool + ping 5s) e
  `ApplyMigrations` (runner `go:embed`, tx por migração, tabela `schema_migrations`);
  migrações `001_initial.sql` (movido, idempotente) + `002_align_models.sql` (novo)
- **`internal/user`**: interfaces `UserStore` + `PGUserStore`/`MemoryUserStore`;
  `UserService` recebe o store por injeção (DIP) mantendo validações/erros PT-BR/UUID
- **`internal/channel`**: `ChannelStore` + `PGChannelStore`/`MemoryChannelStore`;
  capacidade transacional (join concorrente não estoura 10), `user_count` computado
  por `COUNT(*)`, `ListNearby` com bounding-box; service sem `locationService`,
  `generateChannelID` agora UUID
- **`cmd/server/main.go`**: abre conexão e aplica migrações no boot (falha rápida)
- **`scripts/test-backend.ps1`**: WS atualizado para `/ws/audio/?token=` (script
  obsoleto desde 11/10 — não era regressão)

**Verificação**:
- `go build` + `go vet` + `go test ./...` verdes; teste de integração PG
  (`RADIOPX_TEST_DB=1`) PASS; script e2e **28/28**
- Prova de persistência: conta criada → reinício do servidor → login **200**
  (linha confirmada via psql); banco limpo ao final
- Detalhes completos em `CHANGELOG.md`

**Observações abertas**:
- Localização segue em memória e áudio nunca é persistido (escopo aprovado)
- Fase 0 (higiene do repo) está só na `prototype` (`37adeef`) — decidir na merge
- Docs (`banco-de-dados.md`, `requisitos.md` RNF06, `decisoes.md` DEC-012+) ficam
  para a Fase 4

---

**11/10/2026 - Segunda revisão do login: bypass do rate limit, UUIDs e limpeza de sessão**

**Data**: 11/10/2026

**Descrição**:
- **Segurança**: `getClientID` (`internal/middleware/ratelimit.go`) passou a usar só `r.RemoteAddr` —
  antes confiava em `X-Forwarded-For`/`X-Real-IP` sem proxy, permitindo rotacionar o header e burlar o
  rate limit de login
- **Backend**: `generateUserID()` (`internal/user/service.go`) trocou `time.Now().UnixNano()` por
  `uuid.NewString()` (dependência `github.com/google/uuid` já presente), alinhando ao AGENTS.md
- **Frontend**: `isLoggedIn()` faz `logout()` quando o access token expirou e não há refresh token
  (remove token/usuário obsoletos); `register_screen` volta a reconhecer e-mail duplicado em PT-BR
  (`'já está cadastrado'`), antes código morto
- **Testes**: `ratelimit_test.go` (headers ignorados), `user/service_test.go` (UUID válido/distinto),
  `auth_service_test.dart` (limpeza de storage)

**Verificação**:
- `go build ./...` + `go vet ./...` + `go test ./...` verdes
- `flutter analyze`: 15 issues (baseline, zero novos) · `flutter test`: **41/41** (era 40/40)
- Detalhes completos em `CHANGELOG.md`

**Observações abertas**:
- Limiter é compartilhado por register/login/refresh (10/min por IP) — aparelhos no mesmo NAT dividem
  a cota
- `Authenticate` ignora `IsActive` (usuário desativado ainda loga; não há fluxo de desativação)
- `_refreshingSession` é por instância de `AuthService` (guard anti-loop não é global)

---

**11/10/2026 - Endurecimento de segurança: correções de alta, média e baixa prioridade**

**Data**: 11/10/2026

**Descrição**:
- **Alta**: `ChangePassword` valida senha vazia/mínima de 8 (400, alinhado a RF01/RF07) e o validador
  da UI passou de 6 para 8; `middleware.NewRateLimiter(10/min)` (antes código morto) agora envolve
  `/auth/register`, `/auth/login` e `/auth/refresh`, com `getClientID` removendo a porta via
  `net.SplitHostPort` e mensagem de 429 em PT-BR
- **Média**: `handleLogin`/`handleRefresh` tratam erro de `ValidateToken` (fim do nil-deref de
  `claims`); `/ws/audio/` ganhou `webSocketAuthMiddleware` (exige `?token=<access JWT>`, força o
  `user_id` ao ID do token — anti-spoofing); o frontend envia o token (`getValidToken()` +
  `WebSocketService.connect(token)`); botão "Excluir Conta" (`_deleteAccount`) torna o
  `DELETE /api/v1/user/profile` alcançável
- **Baixa**: `ValidateToken` exige **exatamente** HS256 (rejeita HS384/HS512/none); token malformado
  não vale mais como sessão (`isLoggedIn` desloga, `getAuthHeaders` trata como expirado); `main.go`
  avisa quando `JWT_SECRET` cai no fallback de dev
- **Testes**: backend — `user/service_test.go`, `user/handler_test.go`,
  `middleware/ratelimit_test.go` (novo), `auth/service_test.go`, `cmd/server/main_test.go` (3 casos
  do WS auth); frontend — `test/auth_service_test.dart` (novo, 6) e `test/login_screen_test.dart`
  (novo, 3)

**Verificação**:
- `go build ./...` + `go vet ./...` + `go test ./...` verdes
- `flutter analyze`: 15 issues (baseline, zero novos) · `flutter test`: **40/40** (era 31/31)
- Detalhes completos em `CHANGELOG.md`

**Pendências abertas**:
- **Protocolo do WebSocket mudou** (`/ws/audio?token=...`): APKs anteriores precisam ser
  recompilados e o fluxo validado em campo
- **Exclusão de conta** fora de RF explícito — decidir se vira requisito próprio
- **IP no APK** (Problema Conhecido nº 4) permanece como decisão de configuração pendente

---

**10/10/2026 - Revisão do login: consolidação JWT, refresh token ativo e correção de bugs**

**Data**: 10/10/2026

**Descrição**:
- **Backend**: removido `internal/auth/middleware.go` (código morto); `middleware.AuthMiddleware`
  passou a receber um `TokenValidator` injetado (fonte única = `auth.AuthService`), com exigência de
  algoritmo **HS256**; claim `type` (`access`/`refresh`) e rejeição de token do tipo trocado em
  `/auth/refresh`; e-mail normalizado (`trim`+lowercase) em todo o fluxo de usuário; validação de
  e-mail (`net/mail`) e senha mínima de 8 caracteres; mensagens de erro em PT-BR; `DELETE
  /api/v1/user/profile` corrigido (usava `extractUserID`, que procurava o segmento `"users"` e
  sempre retornava 400 → conta nunca era excluída)
- **Frontend**: `isLoggedIn()` valida a expiração do JWT (`token_validity.dart`, `dart:convert`) e
  renova silenciosamente via refresh token; `getAuthHeaders()` renova token expirado antes das
  chamadas (liga o `refreshToken()` que era código morto) e desloga se a renovação falhar; e-mail
  com `trim()`; senha mínima 8; `"E-mail ou senha inválidos"` amigável no login; `UserModel`
  tolerante a `id`/`email` ausentes
- **Segurança**: `backend/.env` saiu do versionamento (`git rm --cached` + `backend/.gitignore`;
  `.env.example` segue rastreado como template). Backend sobe sem `.env` graças ao fallback dev do
  `main.go` e so ignora variáveis já definidas (parser stdlib, sem dependência nova).
- **Testes**: `cmd/server/main_test.go` (10 casos de handler) no backend; `token_validity_test.dart`
  (8 casos) + `UserModel` defensivo (3 casos) no frontend

**Verificação**:
- `go build ./...` + `go vet ./...` + `go test ./...` verdes
- `flutter analyze`: 15 issues (baseline, zero novos) · `flutter test`: 31/31 (era 20/20)
- E2E com servidor real do fonte: register 201 (e-mail normalizado), duplicado 409 PT, senha curta
  400 PT, login com maiúsculas/espaços 200, senha errada 401 PT, access token em `/auth/refresh`
  401, refresh token 200, `GET`/`DELETE /api/v1/user/profile` 200
- Detalhes completos em `CHANGELOG.md`

---

**06/10/2026 - Correção: salvar nome no perfil falhava com "type 'Null' is not a subtype of type 'String'"**

**Data**: 06/10/2026

**Descrição**:
- **Causa**: `PUT /api/v1/user/profile` respondia `200` com `{"message":"User updated"}` (`backend/internal/user/handler.go`), mas `ProfileService.updateProfile` fazia `ProfileModel.fromJson(jsonDecode(response.body))` — `json['id'] as String` recebia `null` → `type 'Null' is not a subtype of type 'String'`
- O dado salvava porque o `UpdateUser` roda **antes** de responder; o nome só aparecia ao voltar porque `app.dart` recria as telas na troca de aba e o `initState` refaz `_loadProfile()`
- "Não dava resposta": o `throw` ocorre **antes** do `Navigator.pop`, então a bottom sheet não fechava; e o `catch` mostrava o erro cru, já que `friendlyError()` só remove o prefixo `Exception: ` (`TypeError` não é `Exception`)
- **Pré-existente, não relacionado** ao trabalho de timeouts do mesmo dia (confirmado no `git diff`); varredura em todo o frontend mostrou que **só** este endpoint tinha corpo de resposta divergente (auth, channels, change-password e delete estavam corretos)
- **Backend**: `HandleUpdateUserProfile` agora chama `GetUser` após o update e devolve o objeto `User`, espelhando `HandleGetUserProfile`; `Password` segue `json:"-"`
- **Backend**: `backend/internal/user/handler_test.go` (novo, primeiro teste de handler do projeto) — `TestUserHandler_HandleUpdateUserProfile` exige `id`/`email`/`name` no `200`, confere persistência e ausência do `password`; `..._NotFound` cobre `404`. Usa `httptest` direto, pois a assinatura recebe `userID` por parâmetro
- **Frontend**: `_saveProfile` captura o `ScaffoldMessenger` **antes** do `await` e o reutiliza nos dois caminhos; sucesso agora fecha a sheet e mostra **"Perfil atualizado"** (verde)
- Backend reiniciado a partir do fonte; APK recompilado (50,8MB) e copiado para `RadioPX.apk`

**Verificação** (conferida, não presumida): `go build ./...` + `go test ./...` verde com os testes novos; `flutter analyze` = **15 issues**, subconjunto do baseline de 16 (saiu um `use_build_context_synchronously` pré-existente, **zero novos**); `flutter test` 20/20; `GET /health` → `200 {"status":"ok",...}`; `aapt2` → `launchMode=2` e `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e `Perfil atualizado`, não contém `10.201.108.96`.

**Motivo**: teste em campo — "ao editar o nome no perfil quando se aperta no salvar ele não dá resposta, dá um erro de subtype string null, mas quando volta o perfil foi alterado".

**Fragilidade mantida**: no caminho de sucesso o `finally { setState(...) }` roda logo após `Navigator.pop` — funciona porque a bottom sheet ainda está na animação de saída.

**Teste em campo sugerido**: editar nome → Salvar → sheet fecha e exibe "Perfil atualizado"; reabrir o perfil mantém o nome. Com o backend parado deve aparecer "Erro ao salvar: Não foi possível conectar ao servidor…". Observação: reiniciar o backend zera os usuários em memória (o `UserService` é um mapa em RAM), então é esperado recriar as contas de teste.

---

## Alteração Anterior

**06/10/2026 - Correção: app ficava carregando infinitamente ao entrar (sem timeout)**

**Data**: 06/10/2026

**Descrição**:
- **Causa**: o APK embutia `API_BASE_URL=http://10.201.108.96:8080` e a máquina voltou a responder em `172.16.158.222` → host inalcançável
- **Por pendurar em vez de dar erro**: nenhuma das 14 chamadas HTTP do frontend tinha timeout (busca por `.timeout(` em `lib/`: zero) → com o `SYN` sem resposta o TCP retenta ~2min → a Future nunca completa → os `catch`/`finally` que zeram `_isLoading` nunca rodam → `CircularProgressIndicator` eterno
- **O IP oscila**: `172.16.158.222` (build 22/09) → `10.201.108.96` (build 04/10, quando se supôs que o anterior estava "obsoleto") → `172.16.158.222` (build atual). Já voltou a ser o valor antigo, então não é troca de sub-rede — embarcar IP transitório no binário é insustentável
- Novo `frontend/lib/core/api_client.dart`: `apiTimeout = 10s`, `apiGet`/`apiPost`/`apiPut`/`apiDelete` com timeout que traduz `TimeoutException`/`SocketException`/`http.ClientException` para mensagens em PT-BR, e `friendlyError()` que remove o prefixo `Exception: ` nas mensagens da UI
- 14 chamadas migradas: `auth_service` (3), `channel_service` (7), `profile_service` (4)
- 11 mensagens de serviço traduzidas de EN para PT-BR (`channel_service` 7, `profile_service` 4); `register_screen` segue detectando `already exists` na mensagem do backend
- `location_service.dart`: `timeLimit: apiTimeout` no `getCurrentPosition` (verificado que o geolocator 10.1.1 aceita `Duration? timeLimit`)
- `websocket_service.dart`: `await channel.ready.timeout()` passa a detectar falha real de handshake — antes a linha `_isConnected = true` rodava antes de existir conexão (falso positivo) e o `try/catch` não capturava nada, porque `WebSocketChannel.connect` é lazy. `sink.close()` em falha é disparado sem aguardar (teto de 2s), já que o `WebSocketSinkCompleter` só recebe destino após handshake
- `voice_screen.dart`: backoff exponencial 2/4/8/16/30s com teto de **5 tentativas**; um único `Timer?` cancelado no `dispose()` (antes o Timer solto podia disparar depois); sucesso zera os contadores. Esgotou → estado terminal `_connectionFailed`: banner "Sem conexão com o servidor" e o botão PTT é substituído por **"Tentar novamente"** + instrução
- `friendlyError` aplicado em 11 pontos de UI: `login_screen`, `register_screen`, `channels_screen` (4), `profile_screen` (3), `channel_detail_screen` (2)
- APK recompilado (50,8MB) com `API_BASE_URL=http://172.16.158.222:8080` e copiado para `RadioPX.apk`

**Verificação** (conferida no artefato, não presumida): `flutter analyze` = **16 issues, os mesmos 16 do baseline** (zero novos); `flutter test` 20/20; `go test ./...` verde; `aapt2` → `launchMode=2` (singleTask) e `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e as strings novas (`Verifique a conex`, `Falha ao carregar canais pr`, `Sem conex`) e **não** contém `10.201.108.96`.

**Motivo**: teste em campo — "na hora que tentamos entrar, apenas fica carregando infinitamente".

**Pendências abertas**:
- URL editável em runtime = **decisão pendente** (exige RF/RNF novo em `docs/requisitos.md`)
- Reserva DHCP/IP estático no roteador = causa raiz da troca, vive fora do repo
- Em web o `ready` é `Future.value()` (stub) → a detecção de falha de handshake do WebSocket não é efetiva nessa plataforma; o alvo de teste é APK
- Validar em campo: com o backend parado, entrar num canal deve exibir a mensagem de erro em ~10s (não spinner) e o login deve cair no retry da reconexão

**Relacionado**: corrige/substitui a entrada de 04/10/2026, que registrou o build com `10.201.108.96` acreditando que `172.16.158.222` estava obsoleto — hoje o valor correto é `172.16.158.222` de novo.

---

## Alterações Anteriores

### 04/10/2026 - Correção: app abria em múltiplas instâncias em vez de ir para segundo plano

**Data**: 04/10/2026

**Descrição**:
- `frontend/android/app/src/main/AndroidManifest.xml`: removido `android:taskAffinity=""` e `launchMode` alterado de `singleTop` para `singleTask`
- Consequência corrigida: launcher criava uma task nova por lançamento (múltiplas cópias no recents), cada uma com WebSocket e sessão `flutter_sound` próprios → áudio dobrado e PTT duplicado
- APK release recompilado com `API_BASE_URL=http://10.201.108.96:8080` (o IP do APK anterior, `172.16.158.222`, está obsoleto — a rede mudou) e copiado para `RadioPX.apk` (50,8MB)
- Correção conferida no artefato: `aapt2` reporta `launchMode=2` (singleTask) e ausência de `taskAffinity`; `libapp.so` contém o IP novo e não contém o antigo

**Relacionado**: este era provavelmente o mesmo bug reportado como "não há opção de sair do canal" — a única forma de sair fica na tela de voz (engrenagem → "Sair do Canal") e o ícone do launcher abria outra instância, então o usuário nunca retornava àquela tela.

**Validação em campo pendente**: abrir o app, entrar num canal, ir para a tela inicial e tocar no ícone 3-4 vezes (deve retomar na mesma tela, 1 entrada no recents); depois revalidar o fluxo de sair. Se ainda falhar, expor botão visível de sair na tela de voz.

### 04/10/2026 - Correção: lista de canais não atualiza automaticamente

**Data**: 04/10/2026

**Descrição**:
- `ChannelsScreen` agora faz polling automático a cada 10s nas listas "Meus canais" e "Perto de mim"
- Refresh de polling é **silencioso**: não mexe em `_isLoading`/`_error`, logo não troca o `TabBarView` por spinner de tela cheia nem reseta aba/scroll; erro de rede é ignorado em vez de virar tela de erro
- Refresh inicial, botão refresh e pull-to-refresh seguem mostrando loading via `_refresh({required bool showLoading})` — nenhum call site mudou
- Guarda de concorrência `_isRefreshing` + `_sameChannels` (só chama `setState` se algo mudou)
- Localização em cache com TTL de 2min; refresh manual continua forçando GPS fresco
- `WidgetsBindingObserver`: pausa o timer em background e atualiza imediatamente ao voltar ao foreground

**Motivo**: ao criar um canal no celular B, o celular A só exibia o canal novo depois de trocar de aba. O `ChannelsScreen` só buscava dados em `initState`/refresh manual, e o reload ao trocar de aba era acidental (o `State` é destruído e recriado por `app.dart`).

**Resultado**: `flutter analyze` sem novos issues, `flutter test` 20/20 verdes, `go test ./...` verde.

**Pendências abertas** (não alteradas aqui):
- `backend/internal/channel/model.go` — `Channel` não tem campo `distance`; `GetNearbyChannels` só usa a distância para ordenar e nunca a serializa → a linha de distância do card de canal nunca aparece (exige mudança de backend)
- `frontend/lib/app.dart` — telas construídas dentro do `build` destroem o `State` ao trocar de aba (perda de estado); irrelevante para a atualização da lista, mas é a causa raiz do comportamento original

**Teste em campo sugerido**: 2 celulares na mesma conta, criar canal no B e observar "Perto de mim" no A em até ~10s sem tocar em nada; depois testar ida ao background → foreground (deve atualizar na hora) e modo avião (não pode virar tela de erro).

### 22/09/2026 - Correção: logout, cadastro, login reativo e latência de áudio

**Data**: 22/09/2026

**Descrição**:
- Perfil usa `AppConfig.apiBaseUrl` (antes `localhost` fixo) → funciona no celular
- `sessionState` (ValueNotifier) no AuthService; `AuthGate` agora é reativo — login/logout sem reiniciar o app
- Cadastro aceita 200 e 201 e mostra o erro real do backend; e-mail duplicado vira mensagem amigável
- Botão "Sair" sempre visível no perfil (inclusive no estado de erro)
- Áudio replay reescrito com `flutter_sound` (sessão contínua PCM16 16kHz) — elimina o delay de ~2,5s do just_audio; `minSdk` 24
- APK recompilado (50,4MB) com `API_BASE_URL=http://172.16.158.222:8080` e copiado para `RadioPX.apk`
- `flutter analyze` sem erros; `flutter test` 20/20 verdes
- Teste em campo sugerido: cadastrar (deve ir para home sem erro), sair da conta, relogar sem reiniciar o app e conferir latência do PTT entre 2 celulares

---

## Próxima Tarefa

**Plano aprovado — Fases 0, 1, 2, 3 e 4 concluídas**

### Próximas Fases

1. ~~**Fase 2 — Expor o backend via ngrok**~~ — **Concluída em 27/10/2026**
   (túnel + CORS + script; ver CHANGELOG da data. URL por sessão no plano free)
2. ~~**Fase 3 — Tela para colar a URL do servidor no app**~~ — **Concluída em
   27/10/2026** (RNF06 adicionado; `AppConfig` em runtime + sheet Login/Perfil)
3. ~~**Fase 4 — Documentação**~~ — **Concluída em 27/10/2026** (`banco-de-dados.md`
   pós-002, DEC-012..015, `arquitetura.md` corrigido, `README.md` reescrito,
   `.env.example` limpo)

---

## Problemas Conhecidos

1. ~~**Distância dos canais nunca aparece**~~ — **Resolvido (backend, `feature-login`)**: `Channel` agora tem `distance,omitempty` serializado e `GetNearbyChannels` popula o campo; o frontend já parseia e renderiza (`channel_model.dart`, `channels_screen.dart`). Resta validação em campo.
2. **Estado das telas perdido ao trocar de aba** — `frontend/lib/app.dart` monta a lista de telas dentro do `build` (`body: _screens[_currentIndex]`), então o `State` é destruído e recriado a cada troca de aba. Não afeta mais a atualização da lista de canais (que agora é por polling), mas faz scroll e aba interna voltarem ao padrão. Correção sugerida: `IndexedStack` + atualização explícita ao selecionar a aba.
3. **Validação em campo pendente da correção de instância única** — a mudança no manifesto (`taskAffinity` removido, `singleTask`) foi conferida no APK gerado via `aapt2`, mas o comportamento no aparelho ainda precisa ser validado: tocar no ícone do launcher 3-4 vezes com o app em background deve retomar na mesma tela e manter 1 entrada no recents. Se o fluxo "sair do canal" (engrenagem → "Sair do Canal") continuar falhando depois disso, o próximo passo é expor um botão visível de sair na tela de voz — hoje ele está 2 níveis abaixo de um ícone de engrenagem.
4. ~~**IP do servidor embutido no APK fica obsoleto a cada troca de DHCP**~~ — **Resolvido em 27/10/2026 (Fases 2+3)**: a Fase 2 expôs o backend via túnel ngrok (o APK deixou de depender do IP local) e a Fase 3 tornou a URL editável em runtime no app (RNF06 — semente do build, `shared_preferences`, sheet no Login e no Perfil). Reinício do ngrok com URL nova é resolvido dentro do app, sem novo build.
5. **Validação em campo pendente do timeout/reconexão** — a detecção de falha de handshake do WebSocket foi conferida na fonte do `web_socket_channel 2.4.5` e no código, mas não em aparelho: com o backend parado, entrar num canal deve exibir a mensagem de erro em ~10s e a tela de voz deve passar por 5 tentativas com backoff até o botão "Tentar novamente". Em web esse comportamento não é garantido (o `ready` lá é stub).
6. ~~**Backend não usa PostgreSQL — tudo é em memória**~~ — **Resolvido em 20/10/2026 (Fase 1)**: pacote `internal/db` (config `DB_*`, `Open`, runner de migrações) + `PGUserStore`/`PGChannelStore` com `jackc/pgx/v5` v5.7.4; `001_initial.sql` movido para `internal/db/migrations/` e alinhado por `002_align_models.sql`. Persistem `users`, `channels` e `user_channels`; localização segue em memória e áudio nunca é gravado. Contas sobrevivem a reinícios (login 200 após restart verificado).
7. **`setState` logo após `Navigator.pop` no salvamento de perfil** — em `profile_screen.dart`, o `finally { setState(...) }` de `_saveProfile` executa logo após o `pop` do caminho de sucesso. Hoje funciona porque a bottom sheet ainda está na animação de saída, mas é frágil: se a animação mudar, pode virar `setState() called after dispose()`.
8. **WebSocket agora exige JWT e quebra clientes antigos** — desde 11/10/2026 o endpoint `/ws/audio/` exige `?token=<access JWT>` (validado por `webSocketAuthMiddleware`, com o `user_id` forçado ao ID do token). APKs compilados antes desta mudança deixam de conectar até serem recompilados. Mitigação: `AuthService.getValidToken()` renova o token antes de conectar. Validação em campo pendente.
9. **Rate limit global atrás do túnel** — desde 27/10/2026 (Fase 2) todo tráfego REST chega como o agente ngrok em `127.0.0.1`; `getClientID` (`internal/middleware/ratelimit.go`) usa `RemoteAddr` e ignora X-Forwarded-For por segurança, então os 10/min de register/login/refresh valem para o **somatório de todos os clientes** do túnel. Decisão registrada em 27/10/2026: manter como está e reavaliar se aparecer 429 em campo (contornos: subir cota, ou confiar XFF só quando o RemoteAddr for o agente local — decisão de segurança).

---

## Dependências

### Backend

- Go 1.21+
- PostgreSQL 14+
- `github.com/golang-jwt/jwt/v5` v5.2.0 (JWT)
- `github.com/google/uuid` v1.6.0 (UUIDs)
- `github.com/gorilla/websocket` v1.5.1 (WebSocket)
- `github.com/jackc/pgx/v5` v5.7.4 (driver PostgreSQL, Fase 1)
- `github.com/rs/cors` v1.10.1 (CORS, inclui wildcard ngrok)
- `golang.org/x/crypto` v0.31.0 (bcrypt)

### Frontend

- Flutter 3.16+ / Dart SDK / Android SDK
- Pacotes usados: `http`, `shared_preferences`, `web_socket_channel`, `record`,
  `flutter_sound`, `geolocator`
- Declarados sem uso: `dio`, `provider`, `google_fonts`, `path_provider`
  (candidatos a remoção)

### Desenvolvimento

- Git
- IDE (VS Code/GoLand)
- ngrok (exposição do backend em demonstração)
- Docker (opcional)

---

## Métricas do Projeto

### Arquivos Criados

| Arquivo | Tipo | Tamanho |
|---------|------|---------|
| README.md | Documentação | ~4KB |
| AGENTS.md | Regras | ~6KB |
| docs/requisitos.md | Requisitos | ~8KB |
| docs/arquitetura.md | Arquitetura | ~17KB |
| docs/banco-de-dados.md | Database | ~6KB |
| docs/decisoes.md | Decisões | ~8KB |
| CHANGELOG.md | Histórico | ~1KB |
| PROJECT_STATE.md | Estado | ~4KB |
| backend/go.mod | Configuração | ~200B |
| backend/.golangci.yml | Configuração | ~500B |
| backend/.env.example | Configuração | ~600B |
| backend/cmd/server/main.go | Backend | ~3KB |
| backend/internal/middleware/auth.go | Backend | ~2KB |
| backend/internal/middleware/cors.go | Backend | ~500B |
| backend/internal/middleware/logging.go | Backend | ~2KB |
| backend/internal/user/repository.go | Backend | ~2KB |
| backend/internal/channel/repository.go | Backend | ~3KB |
| backend/pkg/websocket/hub.go | Backend | ~3KB |
| backend/pkg/websocket/client.go | Backend | ~4KB |
| backend/migrations/001_initial.sql | Database | ~1KB |
| frontend/pubspec.yaml | Configuração | ~600B |
| frontend/lib/main.dart | Frontend | ~100B |
| frontend/lib/app.dart | Frontend | ~2KB |
| specs/.../spec.md | Especificação | ~7KB |
| specs/.../tasks.md | Tarefas | ~10KB |
| specs/.../requirements.md | Checklist | ~2KB |

### Progresso

| Fase | Tarefas | Concluídas | Status |
|------|---------|------------|--------|
| Phase 1: Setup | 4 | 4 | ✅ 100% |
| Phase 2: Foundational | 10 | 10 | ✅ 100% |
| Phase 3: US1 (PTT) | 13 | 13 | ✅ 100% |
| Phase 4: US2 (Location) | 8 | 8 | ✅ 100% |
| Phase 5: US3 (Channels) | 9 | 9 | ✅ 100% |
| Phase 6: US4 (Buffer) | 7 | 7 | ✅ 100% |
| Phase 7: Auth | 10 | 10 | ✅ 100% |
| Phase 8: Profile | 4 | 4 | ✅ 100% |
| Phase 9: Polish | 8 | 8 | ✅ 100% |
| **Total** | **73** | **73** | **100%** |

> ⚠️ Os percentuais acima são **históricos e otimistas**. Vários itens marcados como concluídos não
> estão ligados ao fluxo real (ex.: fases 3 e 6). A avaliação correta está em
> [Lacunas / Aderência aos Requisitos](#lacunas--aderência-aos-requisitos).

---

## Última Atualização

**Data**: 27/10/2026

**Hora**: 2026-10-27

**Por**: OpenCode (assistente AI)

---

## Notas para Continuação

1. **Testes**: Executar testes end-to-end completos
2. **Deploy**: Configurar ambiente de produção
3. **Monitoramento**: Implementar logs e métricas
4. **Documentação**: Atualizar README com instruções de deploy
5. **Google OAuth**: Implementar login social (futuro)
6. **IP do servidor**: ~~a URL do backend entra no binário via `--dart-define` e oscila junto com o DHCP~~ — **resolvido nas Fases 2+3** (túnel ngrok + URL editável em runtime, RNF06). Resta apenas a rotina operacional: ao reiniciar o ngrok, atualizar a URL no app (ou rebuildar com a nova semente)
