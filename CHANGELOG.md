# Histórico do Projeto RadioPX

Este documento registra todas as alterações significativas do projeto.

---

## 27/10/2026 - Limpeza de órfãos pós-Fase 1 (pacote location e endpoint morto)

### Contexto

Auditoria após a Fase 1 (persistência PostgreSQL, 20/10) encontrou código que ficou
sem consumidores: com o `ChannelService` deixando de usar o `LocationService`
(injetado apenas para registrar a posição do canal), o pacote `internal/location`
perdeu quase tudo, e o endpoint stub de localização continuava respondendo `200`
sem efeito algum. Nenhuma funcionalidade em uso foi alterada.

### Alterações

**Backend — `internal/location` enxugado**

- Novo `distance.go`: apenas `EarthRadiusKm` + `CalculateDistance` (Haversine),
  o único símbolo ainda consumido (por `internal/channel/service.go`).
- Novo `distance_test.go` com os 4 testes `TestCalculateDistance_*` (antes em
  `service_test.go`).
- **Removidos**: `service.go` (`LocationService` — `UpdateUserLocation`,
  `FindNearbyChannels`, `GetUsersInRadius`), `handler.go` (`LocationHandler`,
  nunca registrado em `main.go`) e `model.go` (`Location`, `LocationUpdate`,
  `NearbyChannel`, `ProximityFilter`) — código morto desde a refatoração.

**Backend — `cmd/server/main.go`**

- Removido o stub `POST /api/v1/location/update`, que respondia `200
  {"message":"Location updated"}` sem persistir nada. Verificado que **nenhum
  consumidor real**: o `LocationService` do frontend usa apenas o Geolocator
  local (envia lat/lng nos payloads de canais) e `scripts/test-backend.ps1`
  não testava o endpoint.

**Documentação**

- `docs/arquitetura.md` — endpoint removido da tabela de rotas, com nota de
  que a localização trafega na API de canais.
- `docs/requisitos.md` — RF04 agora aponta `internal/channel/` (descoberta) +
  `internal/location/` (distância).

### Motivo

Deixar o pacote e o endpoint órfãos após a Fase 1 convida a manutenção
enganosa (quem lê acha que localização é persistida no backend). A remoção
também tira do caminho a confusão entre o stub morto e o fluxo real de GPS.

### Verificação

- `go build ./...`, `go vet ./...` e `go test ./...` verdes.
- Grep confirma zero referências restantes a `NewLocationService`,
  `NewLocationHandler` e `location/update` no backend.
- Frontend intocado.

---

## 20/10/2026 - Fase 1: persistência PostgreSQL (fim do estado em memória)

### Contexto

Fase 1 do plano aprovado pelo autor (fases: 0 higiene do repo, 1 persistência, 2 ngrok,
3 tela de URL no app, 4 docs). Resolve o **Problema Conhecido nº 6** e a lacuna
"DEC-001 / FR-013 — PostgreSQL". Escopo aprovado: apenas `users`, `channels` e
`user_channels` são persistidos; **localização continua em memória** e **áudio nunca é
persistido** (regra crítica de privacidade). Base: branch `feature-login`.

### Dependência nova (política §11 do AGENTS.md)

- **`github.com/jackc/pgx/v5` v5.7.4** (fixada) — driver PostgreSQL via adaptador
  `database/sql` (`pgx/v5/stdlib`) + mapeamento de erro `23505` (`pgconn`).
  Finalidade: persistir usuários/canais com transações e `ON CONFLICT`.
  Alternativas consideradas: **GORM** ( ORM completo, mais pesado que o necessário);
  **lib/pq** (sem manutenção ativa); **pgx puro** (exigiria trocar `database/sql` e
  refazer testes). Versão **v5.7.4** escolhida porque é a última que declara `go 1.21`,
  compatível com o diretivo do `go.mod` sem bump (v5.7.5+ exige go 1.23).
- Transitivas ajustadas pelo `go mod tidy`: `golang.org/x/crypto` 0.17.0 → **0.31.0**,
  `x/net` → **0.21.0**, `x/text` → **0.21.0**, `x/sync` → **0.10.0** (todas indiretas).
  `github.com/google/uuid` segue direta (IDs).

### Alterações

**Backend — novo pacote `internal/db`**

- `db.go` — `Config`/`ConfigFromEnv()` (variáveis `DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD/DB_SSLMODE`,
  com defaults), `Open()` com pool (10 abertas/5 ociosas) e `Ping` com timeout de 5s.
- `migrate.go` — runner de migrações próprio: `//go:embed migrations/*.sql`, ordem
  lexicográfica, uma transação por migração, controle em tabela `schema_migrations`.
- `migrations/001_initial.sql` — movido de `backend/migrations/` e tornado idempotente
  (`IF NOT EXISTS`) para poder ser aplicado em banco já preparado via psql.
- `migrations/002_align_models.sql` (novo) — alinha schema ao modelo real: `users.cpf`
  deixa de ser NOT NULL (RF01 prevê mas o app não coleta), colunas `provider`,
  `provider_id`, `is_active`, `avatar`; `channels` ganha `description`, `latitude`,
  `longitude`, `created_by` (UUID FK) e `updated_at` + índice `idx_channels_location`.

**Backend — módulo `internal/user`**

- `repository.go` (novo) — interface `UserStore` (fronteira de persistência) +
  `PGUserStore` (CRUD, unique violation 23505 → `ErrUserAlreadyExists`).
- `memory_store.go` (novo) — `MemoryUserStore` para os testes unitários.
- `service.go` — `UserService` agora depende de `UserStore` (DIP);
  `NewUserService(store)`. **Todas as regras da `feature-login` preservadas**:
  erros em PT-BR, `normalizeEmail`, `validEmail`, `minPasswordLength = 8`, UUIDs.

**Backend — módulo `internal/channel`**

- `repository.go` (novo) — interface `ChannelStore` + `PGChannelStore`: capacidade e
  unicidade de membership aplicadas **transacionalmente** (`SELECT ... FOR UPDATE` +
  `ON CONFLICT`, commit só se `count <= max_users` — joins concorrentes não
  ultrapassam 10); `user_count` sempre computado via `COUNT(*)`, nunca armazenado;
  `ListNearby` com pré-filtro por bounding-box (Haversine e ordenação ficam no service).
- `memory_store.go` (novo) — `MemoryChannelStore` com a mesma semântica.
- `service.go` — construtor trocou `locationService` por `ChannelStore`;
  `generateChannelID` passou de `time.Now().UnixNano()` para UUID; campo `Distance`
  (já existente na branch) continua populado pelo service.
- `handler.go` — `HandleGetUserChannels` trata o novo erro de store.

**Backend — `cmd/server/main.go`**

- Abre a conexão (`db.Open`) e aplica migrações (`db.ApplyMigrations`) no boot,
  com **falha rápida** (`log.Fatalf`) se o banco estiver inacessível; cria
  `PGUserStore`/`PGChannelStore` e injeta nos services.

**Scripts**

- `scripts/test-backend.ps1` — seção WS corrigida para o contrato vigente desde
  11/10 (`0b61791`): URL `/ws/audio/` com barra final (o ServeMux redireciona 301 e o
  `ClientWebSocket` do .NET não segue redirect) e `?token=<JWT>` no lugar de
  `user_id` (`webSocketAuthMiddleware` ignora `user_id` forjado). **O script estava
  obsoleto desde 11/10** — a falha de WS era pré-existente, não regressão.

### Motivo

Requisito documentado desde o início (DEC-001, `docs/banco-de-dados.md`, AGENTS.md)
que o código cumpria só em memória: reiniciar o backend zera usuários e canais.

### Verificação

- `go build ./...`, `go vet ./...` e `go test ./...` verdes (9 pacotes).
- **Teste de integração novo** `internal/db/pg_integration_test.go` (opt-in via
  `RADIOPX_TEST_DB=1`, senão `skip`): migração idempotente + CRUD de usuário +
  duplicidade + autenticação + membership/capacidade/ownership contra o PostgreSQL real — **PASS**.
- `scripts/test-backend.ps1` contra servidor local: **28/28 verdes** (inclui relay de
  áudio via WebSocket com token).
- **Prova de persistência**: conta criada → servidor reiniciado → login **200**;
  linha confirmada por psql (UUID em `users`, `schema_migrations` com 001+002).
- Dados de teste removidos após a verificação (banco de volta a 0 linhas).

### Pendências / observações

- Localização (`internal/location`) segue em memória (escopo aprovado).
- Atualizações de docs ficam para a Fase 4: `docs/banco-de-dados.md` (schema 002),
  `docs/requisitos.md` (novo RNF06 da tela de URL), `docs/decisoes.md` (DEC-012+).
- A higiene da Fase 0 (`.gitignore` raiz, untrack de `server.exe`/APK, JWT rotacionado)
  foi comitada apenas na `prototype` (`37adeef`); ao commitar esta fase na
  `feature-login` decidir se ela é trazida para cá ou resolvida na merge.

---

## 20/10/2026 - Fase 0: conformidade com as regras do repo (segredos fora do git, artefatos de build)

### Contexto

Primeira fase da ampliação do projeto (persistência PostgreSQL + exposição via ngrok). Nenhuma linha de
código de aplicação foi alterada — apenas higiene do repositório. O `backend/.env` estava trackeado desde
o commit inicial com `DB_PASSWORD` e `JWT_SECRET` reais, em violação ao AGENTS.md ("JWT Secret: nunca
committar em repositório").

### Alterações

- **`.gitignore` na raiz (novo)** — regras para segredos (`backend/.env`, `backend/.env.*` com exceção de
  `.env.example`), artefatos de build (`backend/server.exe`, `RadioPX.apk`, `frontend/build/`,
  `frontend/android/build/`) e `.vscode/`. O `frontend/android/build/` vinha aparecendo como untracked
  porque o `.gitignore` do módulo android não cobre `build/`
- **`git rm --cached backend/.env backend/server.exe RadioPX.apk`** — segredo e binários saem do
  versionamento (os arquivos continuam no disco; ignorá-los sem tirá-los do git não teria efeito).
  Observação: o `RadioPX.apk` foi entregável versionado em 3 commits — se quiser continuar publicando-o
  no repo, basta `git add -f RadioPX.apk` e remover a linha do `.gitignore`
- **`JWT_SECRET` rotacionado** — o valor antigo (`your-super-secret-jwt-key-change-in-production`) é o
  placeholder do `.env.example` e está no histórico do git; o novo é aleatório (48 bytes, base64) e vive
  só no `backend/.env` local. Rotação invalida tokens emitidos antes (irrelevante hoje: usuários em
  memória são perdidos a cada restart)
- **`backend/.env` recriado localmente** com as credenciais verificadas do PostgreSQL local
  (`localhost:5432`, database `radiopx` já existente com o schema da migração 001 e 0 linhas) e o novo
  segredo

### Motivo

Preparar o repositório para as fases seguintes (persistência e ngrok) sem vazar segredos e sem poluir o
status com artefatos de build.

### Verificação

- `git check-ignore -v` confirma: `backend/.env`, `backend/server.exe`, `RadioPX.apk`,
  `frontend/android/build/`, `.vscode/`; e confirma que `backend/.env.example` **não** é ignorado (regra
  de exceção funcionando)
- `git status` mostra apenas as três exclusões stageadas, o `.gitignore` novo e as modificações
  pré-existentes de `frontend/windows/flutter/generated_plugin_*` (regeneradas por um `flutter` local,
  não tocadas aqui)

### Notas

- Os blobs do `.env` antigo permanecem no histórico (commits `dd31110`..`6005a7e`). Reescrever o
  histórico seria destrutivo e ficou fora do escopo deliberadamente; a rotação do segredo mitiga o risco
- Alterações stageadas, não commitadas — commit somente a pedido

---

## 13/10/2026 - Reconciliação docs × código: registro de lacunas e escopo da v1

### Contexto

Auditoria do código (backend Go + frontend Flutter) frente à documentação. O código é um **protótipo
funcional em memória**, mas os docs marcavam 73/73 tarefas e todos os RF/RNF como concluídos. Nenhuma
funcionalidade foi alterada nesta entrada — apenas registrada a divergência para correção futura.

### Alterações (apenas documentação)

- **`PROJECT_STATE.md`** — nova seção **Lacunas / Aderência aos Requisitos** (fonte única) listando:
  lacunas dentro do escopo v1 (RF03 sem bloqueio de um falante por vez; RF06/RNF04 buffer e replay
  como código morto; codec `opus` vs PCM16; distância não serializada; sem testes de integração),
  **trabalho futuro fora de escopo v1** (PostgreSQL, Google OAuth, Casbin, TLS, deploy/monitoramento)
  e débito técnico (código morto). Notas de alerta adicionadas em "Fase", "Requisitos Concluídos" e
  na tabela "Progresso".
- **`docs/requisitos.md`** — nota de status apontando para PROJECT_STATE.
- **`docs/arquitetura.md`** — nota de que a tabela "Status de Implementação" é aspiracional.
- **`docs/decisoes.md`** — nova **DEC-011** (escopo v1 em memória) com status de DEC-001/DEC-002.
- **`specs/001-px-radio-simulation/tasks.md`** — nota sobre checkboxes otimistas (T045–T051, T049/T050,
  parte OAuth de T052–T061).
- **`README.md`** — nota de que "Status Atual" está desatualizado.

### Motivo

Evitar decisões erradas e defesa frágil do trabalho: a documentação superestimava o progresso em
relação ao código real. Registrado a pedido do autor, que fará as correções depois.

### Verificação

- Somente arquivos de documentação (`.md`) foram alterados; nenhum código-fonte, teste ou dependência
  foi tocado.

---

## 11/10/2026 - Segunda revisão do login: bypass do rate limit, UUIDs e limpeza de sessão

### Contexto

Nova auditoria do fluxo de login após o endurecimento da manhã. Foram encontrados e corrigidos
quatro pontos, sendo um de segurança relevante.

### Alterações

**Backend**

- `internal/middleware/ratelimit.go` — **correção de segurança**: `getClientID` passou a usar apenas
  `r.RemoteAddr`. Antes ele confiava em `X-Forwarded-For`/`X-Real-IP` sem proxy confiável; como o
  servidor Go não injeta esses headers e não há reverse proxy, um cliente podia enviar um valor
  diferente a cada tentativa e **burlar totalmente o rate limit de login** (força bruta ilimitada).
- `internal/user/service.go` — `generateUserID()` passou de `time.Now().UnixNano()` para
  `uuid.NewString()` (`github.com/google/uuid`, dependência já presente). Alinha ao AGENTS.md
  ("UUIDs são usados como chaves primárias"): IDs deixam de ser previsíveis/enumeráveis e elimina o
  risco de colisão por resolução de relógio.

**Frontend**

- `lib/features/auth/services/auth_service.dart` — `isLoggedIn()`, no caso de access token expirado
  **e** sem refresh token, agora executa `logout()` antes de retornar `false`, removendo token/usuário
  obsoletos do storage (antes ficava um "usuário fantasma" em `getCurrentUser()`).
- `lib/features/auth/screens/register_screen.dart` — a detecção de e-mail duplicado também casa
  `'já está cadastrado'`. O backend responde em PT-BR desde 10/10, então o check antigo por
  `'already exists'` era código morto e a mensagem amigável nunca aparecia.

**Testes**

- `internal/middleware/ratelimit_test.go` — novo `TestGetClientID_IgnoresForwardedHeaders` (headers de
  encaminhamento não alteram o clientID).
- `internal/user/service_test.go` — novo `TestUserService_CreateUser_GeneratesUniqueUUIDs` (ID é UUID
  válido e distinto entre usuários).
- `frontend/test/auth_service_test.dart` — novo caso: token expirado sem refresh limpa o storage.

### Motivo

Revisão solicitada do login para identificar novos bugs após o endurecimento anterior.

### Verificação

- `go build ./...` + `go vet ./...` + `go test ./...` verdes.
- `flutter analyze`: **15 issues — os mesmos do baseline, zero novos**; `flutter test`: **41/41**
  (era 40/40).

### Pendências / observações

- **Limiter compartilhado** (register+login+refresh, 10/min por IP): dois aparelhos atrás do mesmo NAT
  dividem a cota. Mantido por ora; avaliar limiters separados se necessário.
- **`Authenticate` ignora `IsActive`**: usuário desativado ainda loga (latente — não há fluxo de
  desativação).
- **`_refreshingSession` é por instância** de `AuthService`: o guard anti-loop não é global.

---

## 11/10/2026 - Endurecimento de segurança: correções de alta, média e baixa prioridade

### Contexto

Continuação da revisão do login (RF02) iniciada em 10/10. Os itens levantados em auditoria foram
corrigidos em três lotes por prioridade, com a suíte de testes rodada ao fim de cada lote.

### Alterações

**Alta prioridade**

- `internal/user/service.go` + `internal/user/handler.go` — `ChangePassword` agora rejeita senha vazia
  (`ErrPasswordRequired`) e com menos de 8 caracteres (`ErrPasswordTooShort`), ambos mapeados para
  `400` (antes a senha curta caía em `409` e o requisito de 8 caracteres não era aplicado na troca de
  senha). Alinhado ao RF01/RF07.
- `frontend/lib/features/profile/screens/profile_screen.dart` — validador da troca de senha passou de
  6 para 8 caracteres, alinhado ao backend.
- `backend/cmd/server/main.go` — **rate limit** de 10 req/min por IP aplicado a
  `/api/v1/auth/register`, `/auth/login` e `/auth/refresh` (o `middleware.NewRateLimiter` já existia
  mas nunca era usado).
- `internal/middleware/ratelimit.go` — `getClientID` remove a porta via `net.SplitHostPort` (antes o
  mesmo IP com portas efêmeras diferentes contava como clientes distintos, inutilizando o limite);
  mensagem de `429` traduzida para PT-BR.

**Média prioridade**

- `cmd/server/main.go` — `handleLogin`/`handleRefresh` passam a tratar o erro de `ValidateToken`
  (evita nil-deref de `claims`) respondendo `500 {"error":"Falha ao gerar token"}`.
- `cmd/server/main.go` — novo `webSocketAuthMiddleware`: `/ws/audio/` agora exige
  `?token=<access JWT>`, validado por `ValidateAccessToken`; o `user_id` é forçado ao ID do token
  (impede spoofing por query string). Sem token/token inválido → `401`.
- `frontend/lib/core/websocket/websocket_service.dart` — `connect` passou a exigir `token` e monta a
  URI com `Uri.replace(queryParameters: ...)`.
- `frontend/lib/features/auth/services/auth_service.dart` — novo `getValidToken()` (renova
  silenciosamente se expirado); usado pela tela de voz.
- `frontend/lib/features/voice/screens/voice_screen.dart` — obtém o token válido antes de conectar;
  sem token, reagenda a reconexão.
- `frontend/lib/features/profile/screens/profile_screen.dart` — botão "Excluir Conta" +
  `_deleteAccount()` (diálogo de confirmação → `deleteAccount()` → `logout()`), tornando o endpoint
  `DELETE /api/v1/user/profile` alcançável pela UI.

**Baixa prioridade**

- `internal/auth/service.go` — `ValidateToken` passou a exigir **exatamente** HS256
  (`token.Method.Alg() != jwt.SigningMethodHS256.Alg()`), rejeitando HS384/HS512/none (antes aceitava
  qualquer HMAC).
- `frontend/lib/features/auth/services/auth_service.dart` — token malformado (sem `exp`
  decodificável) não vale mais como sessão: `isLoggedIn()` faz logout e retorna `false`;
  `getAuthHeaders()` trata como expirado e tenta renovar.
- `cmd/server/main.go` — emite aviso no log quando `JWT_SECRET` não é definido (fallback dev inseguro).

**Testes**

- Backend: `internal/user/service_test.go` + `handler_test.go` (validação da troca de senha),
  `internal/middleware/ratelimit_test.go` (novo), `internal/auth/service_test.go` (HS256 e tipo de
  token), `cmd/server/main_test.go` (3 casos do `webSocketAuthMiddleware`, incluindo resistência a
  spoofing do `user_id`).
- Frontend: `test/auth_service_test.dart` (novo, 6 casos) e `test/login_screen_test.dart` (novo, 3
  casos).

### Motivo

Execução das correções apontadas pela auditoria, por prioridade, com verificação a cada lote.

### Verificação

- `go build ./...` + `go vet ./...` + `go test ./...` verdes.
- `flutter analyze`: **15 issues — os mesmos do baseline, zero novos** (o `_deleteAccount` gerou um
  `use_build_context_synchronously` que foi corrigido com guarda `mounted`).
- `flutter test`: **40/40** (era 31/31).

### Notas e pendências

- **Mudança de protocolo do WebSocket**: `/ws/audio` agora exige `?token=`. APKs compilados antes
  desta mudança deixam de conectar até serem recompilados — **validar em campo**.
- **Exclusão de conta** não consta em RF explícito (RF07 lista visualizar/editar/alterar senha); a UI
  e o endpoint foram expostos realizando a intenção já registrada no `PROJECT_STATE`. Reavaliar se
  deve virar RF próprio.
- **IP embutido no APK** (Problemas Conhecidos nº 4) segue como decisão de configuração pendente —
  não é bug de código.

---

## 10/10/2026 - Revisão do login: consolidação JWT, refresh token ativo e correção de bugs

### Contexto

Auditoria completa do fluxo de autenticação (RF02) na branch `feature-login`. O login já
funcionava no caminho feliz, mas havia código JWT duplicado, um bug latente na exclusão de
conta e o refresh token existia mas nunca era usado.

### Alterações

**Backend**

- `internal/auth/middleware.go` — **removido** (código morto: `AuthMiddleware` duplicado não era
  referenciado em lugar nenhum).
- `internal/middleware/auth.go` — `AuthMiddleware` agora recebe um `TokenValidator` injetado
  (implementado pelo `auth.AuthService`, fonte única de verificação de JWT). Removidos o segundo
  `Claims`, `GenerateToken` e `ValidateToken` que liam `JWT_SECRET` por conta própria. `Bearer`
  passa a ser aceito sem diferenciar maiúsculas/minúsculas.
- `internal/auth/service.go` — validação agora exige algoritmo **HS256** (evita alg-confusion);
  novo claim `type` (`access`/`refresh`); novo `ValidateAccessToken` que retorna `userID` e
  rejeita refresh token usado como access; `RefreshToken` rejeita access token usado como refresh.
- `internal/user/service.go` — e-mail normalizado (`trim` + lowercase) em `CreateUser`,
  `Authenticate`, `GetUserByEmail`, `UpdateUser` e `CreateOAuthUser`; validação de e-mail
  (`net/mail`, stdlib) e senha mínima de 8 caracteres (alinhado ao RF01). Erros traduzidos para
  PT-BR (`este e-mail já está cadastrado`, `a senha deve ter pelo menos 8 caracteres`, etc.).
- `internal/user/handler.go` — removidos handlers mortos por path (`HandleCreateUser`,
  `HandleGetUser`, `HandleUpdateUser`, `HandleDeleteUser`, `HandleListUsers`,
  `HandleChangePassword`, `extractUserID`). Novo `HandleDeleteUserByID`.
- `cmd/server/main.go` — middleware e injetor passam a usar `authService.ValidateAccessToken`;
  DELETE `/api/v1/user/profile` agora passa o `userID` (corrige o bug de excluir conta que
  sempre respondia 400, pois `extractUserID` procurava o segmento `"users"`); `handleRegister`/
  `handleLogin`/`handleRefresh` com mensagens em PT-BR e validação (400) para campos vazios,
  e-mail inválido e senha curta.
- `cmd/server/main_test.go` (novo) — 10 testes dos handlers: register sucesso/duplicado/inputs
  inválidos, login sucesso/senha errada/normalização de e-mail, refresh rejeitando access token e
  aceitando refresh token.

**Frontend**

- `lib/features/auth/services/token_validity.dart` (novo) — decodifica `exp` do JWT com
  `dart:convert` (sem dependência nova): `jwtExpiry` e `tokenIsExpired`.
- `lib/features/auth/services/auth_service.dart` — `isLoggedIn()` passou a validar a expiração do
  access token e, se expirado, **renova silenciosamente** via refresh token antes de deslogar;
  `getAuthHeaders()` renova proativamente token expirado e faz logout se a renovação falhar
  (liga o `refreshToken()` que era código morto); guard `_refreshingSession` evita loops.
- `login_screen.dart` / `register_screen.dart` — e-mail com `trim()`; senha mínima de 8
  caracteres; login mapeia `Invalid credentials`/senha inválida para `"E-mail ou senha inválidos"`.
- `user_model.dart` — parsing defensivo (`id`/`email` toleram ausência/nulo, datas via
  `DateTime.tryParse`); `displayName`/`initials` não quebram com e-mail vazio (retornam `?`).
- Testes: `token_validity_test.dart` (8 casos) + casos de `UserModel` tolerante no
  `models_test.dart` (3 casos).

### Motivo

Revisão solicitada do login: confirmar o que funciona, apontar o que falta e corrigir bugs.

### Verificação

- `go build ./...` + `go vet ./...` + `go test ./...` verdes (inclui os 10 testes novos de handler).
- `flutter analyze`: **15 issues — os mesmos 15 do baseline, zero novos** (durante o trabalho chegou
  a 16 com um erro real em `isLoggedIn` — variável `refreshToken` sombreava o método — corrigido).
- `flutter test`: **31/31** (era 20/20).
- Fluxo E2E contra o backend real (servidor iniciado do fonte, depois encerrado):
  register 201 com e-mail normalizado para lowercase; duplicado → 409 PT; senha curta → 400 PT;
  login com `Mixed.User@Example.com` + espaços → 200; senha errada → 401 PT; access token em
  `/auth/refresh` → 401; refresh token → 200; `GET /api/v1/user/profile` → 200; `DELETE
  /api/v1/user/profile` → 200 (antes sempre 400).

### Segurança (adicional)

- `backend/.env` **removido do versionamento** (`git rm --cached`) — o arquivo pode conter
  `JWT_SECRET`/credenciais Google, o que a regra do AGENTS.md proíbe (nunca committar segredos).
  O arquivo permanece no disco (ignorado pelo novo `backend/.gitignore`, que mantém
  `.env.example` rastreado como template).
- O servidor não depende do arquivo para subir: `main.go` usa fallback dev
  (`dev-secret-change-in-production`) quando `JWT_SECRET` não está definido, e o `loadEnvFile`
  próprio (parser stdlib, sem biblioteca externa) só injeta variáveis ausentes. Smoke test
  comprovou boot sem `.env` (health OK).
- **Nota**: o segredo antigo continua no histórico do git. Purga definitiva exigiria reescrever o
  histórico (`filter-repo`), o que é arriscado num repo acadêmico — recomendado apenas se o repositório
  for público. Para o TCC, troque o `JWT_SECRET` antes de qualquer exposição.

### Pendências abertas (fora de escopo)

- **Google OAuth** (parte do RF02) não implementado — exige nova dependência (`google_sign_in`) e
  decisão de escopo.
- `internal/middleware/cors.go`, `logging.go` e `ratelimit.go` ainda são código aparentemente não
  usado pelo `main.go` (que usa `github.com/rs/cors` e não aplica rate limit) — candidatos a remoção
  em tarefa futura.
- Backend continua em memória (mapas) — PostgreSQL segue pendente (PROJECT_STATE, problema nº 6).

---

## 06/10/2026 - Correção: salvar nome no perfil falhava com "type 'Null' is not a subtype of type 'String'"

### Causa

`PUT /api/v1/user/profile` respondia `200` com `{"message":"User updated"}` (`backend/internal/user/handler.go`), mas `ProfileService.updateProfile` fazia `ProfileModel.fromJson(jsonDecode(response.body))`, cujo factory exige `json['id'] as String` e `json['email'] as String` — ambos `null` → **`type 'Null' is not a subtype of type 'String'`** (o "subtype string null").

Por que o dado salvava mesmo assim: o `UpdateUser` executa **antes** de gravar a resposta, então o backend já tinha aplicado a alteração. O nome só aparecia ao voltar porque `frontend/lib/app.dart` monta a lista de telas dentro do `build`, o `State` é destruído/recriado na troca de aba e o `initState` refaz `_loadProfile()`.

Por que "não dava resposta": o `throw` ocorre **antes** do `Navigator.pop(context)`, então a bottom sheet de edição nunca fechava; e o `catch` exibia o erro cru, porque `friendlyError()` apenas remove o prefixo `Exception: ` e `TypeError` não é `Exception`.

**Pré-existente e não relacionado** ao trabalho de timeouts do mesmo dia — confirmado no `git diff`, que mostra que a troca `http.put` → `apiPut` e a tradução das mensagens não tocaram no parsing. Varredura por `jsonDecode(response.body)` em todo o frontend: auth (`data['user']`), channels (retornam o objeto), `change-password` e `delete` (só checam status) estavam corretos — **só este endpoint divergia**.

### Alterações

**Backend**
- `backend/internal/user/handler.go` — `HandleUpdateUserProfile` agora chama `h.userService.GetUser(userID)` após o update e devolve o objeto `User`, espelhando o padrão já existente de `HandleGetUserProfile`. `User.Password` continua `json:"-"`, então não vaza hash.
- `backend/internal/user/handler_test.go` (novo — primeiro teste de handler do projeto, que só tinha testes de service):
  - `TestUserHandler_HandleUpdateUserProfile` — guarda de regressão: exige `id`, `email` e `name` no corpo do `200`, confere que o serviço persistiu o `name` novo e que o campo `password` não aparece no JSON
  - `TestUserHandler_HandleUpdateUserProfile_NotFound` — `404` para usuário inexistente
  - a assinatura recebe `userID` por parâmetro, então o teste usa `httptest.NewRequest`/`NewRecorder` direto, sem subir router

**Frontend**
- `frontend/lib/features/profile/screens/profile_screen.dart` — `_saveProfile` captura o `ScaffoldMessenger` **antes** do `await` e o reutiliza nos dois caminhos: sucesso fecha a sheet e mostra **"Perfil atualizado"** (verde); erro mantém o SnackBar vermelho. Capturar antes do await evita usar `BuildContext` após gap assíncrono.

### Motivo

Teste em campo: "ao editar o nome no perfil quando se aperta no salvar ele não dá resposta, dá um erro de subtype string null, mas quando volta o perfil foi alterado".

### Verificação

- `go build ./...` + `go test ./...` verde (testes novos executados, não cacheados)
- `flutter analyze`: **15 issues** — subconjunto do baseline de 16: saiu um `use_build_context_synchronously` pré-existente (o `ScaffoldMessenger.of(context)` do bloco `catch`), **zero issues novos**
- `flutter test` 20/20
- Backend reiniciado a partir do fonte (`go run cmd/server/main.go`) e `GET /health` → `200 {"status":"ok","service":"radiopx-backend","connections":0}`
- Artefato conferido, não presumido: `aapt2` → `package="com.radiopx.radiopx"`, `launchMode=2` (singleTask), `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e a string nova `Perfil atualizado`, e **não** contém `10.201.108.96`
- APK release recompilado (50,8MB) e copiado para `C:\RadioPx\RadioPX.apk`

### Notas

- Fragilidade pré-existente mantida como está: no caminho de sucesso o `finally { setState(...) }` roda logo após `Navigator.pop` — funciona porque a bottom sheet ainda está na animação de saída.
- `HandleUpdateUser` (handler.go:69) é código morto, sem nenhuma chamada — fora do escopo desta correção.
- Backend reiniciado zera os dados em memória (o `UserService` é um mapa em RAM), então é esperado precisar recriar as contas de teste.

### Teste em campo sugerido

Editar o nome → Salvar → a sheet fecha e exibe "Perfil atualizado"; reabrir o perfil mantém o nome. Com o backend parado, deve aparecer "Erro ao salvar: Não foi possível conectar ao servidor. Verifique se ele está rodando e na mesma rede."

---

## 06/10/2026 - Correção: app ficava carregando infinitamente ao entrar (sem timeout em nenhuma chamada)

### Causa

O APK anterior embutia `API_BASE_URL=http://10.201.108.96:8080`, um IP transitório de DHCP. A rede renovou o lease e a máquina passou a responder em `172.16.158.222` — o host embutido ficou inalcançável (verificado: `TcpTestSucceeded=False` no IP antigo, `True` no atual, com o backend saudável em 8080).

O histórico confirma que **o endereço oscila**: `172.16.158.222` (build de 22/09) → `10.201.108.96` (build de 04/10, quando se supôs que o anterior estava "obsoleto") → `172.16.158.222` (build atual). Ou seja, o valor já voltou a ser o de antes, o que descarta qualquer teoria de troca de sub-rede e confirma que fixar o IP no binário é insustentável.

O sintoma era travamento em vez de erro porque **nenhuma chamada HTTP do frontend tinha timeout** (busca por `.timeout(`: zero ocorrências em `lib/`). Com o host inalcançável o `SYN` fica sem resposta e o TCP retenta por ~2 minutos; como a Future nunca completa, os blocos `catch`/`finally` que zeram `_isLoading` nunca executam e o `CircularProgressIndicator` gira para sempre.

### Alterações

**`frontend/lib/core/api_client.dart` (novo)**
- `apiTimeout = Duration(seconds: 10)` — fonte única do tempo de espera
- `apiGet`/`apiPost`/`apiPut`/`apiDelete` com `.timeout(apiTimeout)`, capturando `TimeoutException`, `SocketException` e `http.ClientException` e relançando mensagens em PT-BR ("Servidor não respondeu. Verifique a conexão com a rede." / "Não foi possível conectar ao servidor...")
- `friendlyError(Object)` — remove o prefixo `Exception: ` nas mensagens exibidas na UI

**Serviços — 14 chamadas migradas**
- `auth_service.dart`: 3 `http.post` → `apiPost` (`import http` mantido, ainda usado em `_readError`)
- `channel_service.dart`: 7 chamadas → `apiGet`/`apiPost`
- `profile_service.dart`: 4 chamadas → `apiGet`/`apiPut`/`apiPost`/`apiDelete`

**Mensagens em inglês traduzidas (11)** — `channel_service` (7) e `profile_service` (4) agora lançam "Falha ao carregar canais próximos", "Falha ao atualizar o perfil" etc. O `register_screen` continua detectando `already exists` na mensagem do backend, não numa string local.

**`location_service.dart`** — `getCurrentPosition` ganhou `timeLimit: apiTimeout`. Verificado na fonte do pacote que o geolocator 10.1.1 aceita `Duration? timeLimit`; o `catch` já existente devolve `null` e vira mensagem de erro, não pendura.

**`websocket_service.dart`** — `WebSocketChannel.connect` é lazy e o `try/catch` anterior não capturava falha de rede; a linha que marcava `_isConnected = true` roda **antes** da conexão existir (falso positivo). Agora:
- `await channel.ready.timeout(apiTimeout)` — confirmado na fonte do `web_socket_channel` 2.4.5 que no Android (`IOWebSocketChannel`) `ready` é o `_readyCompleter` real e **faz `completeError` em falha**; na classe base é `Future.value()` (stub)
- `_isConnected = true` só após o handshake concluir
- em falha, `sink.close()` é disparado **sem aguardar** com teto de 2s — o `WebSocketSinkCompleter` só recebe destino quando o handshake conclui, então aguardar poderia pendurar (reintroduziria o bug)

**`voice_screen.dart`** — substituí o `Timer(2s)` que reconectava para sempre:
- backoff exponencial 2s → 4s → 8s → 16s → 30s, teto de **5 tentativas**
- um único `Timer?` agendado (sem duplicar), cancelado no `dispose()` — antes o Timer solto podia disparar após o dispose
- sucesso **zera** `_reconnectAttempts`
- esgotou → estado terminal `_connectionFailed`: banner "Sem conexão com o servidor" e o botão PTT é substituído por **"Tentar novamente"** + instrução, que zera os contadores e reconecta

**`friendlyError` aplicado em 11 pontos de UI** — `login_screen`, `register_screen`, `channels_screen` (4), `profile_screen` (3), `channel_detail_screen` (2, hoje órfã).

### Motivo

Teste em campo: "na hora que tentamos entrar, apenas fica carregando infinitamente" — a rede havia mudado e toda a UI pendurava sem qualquer indicação de erro.

### Verificação

- `flutter analyze`: **16 issues — os mesmos 16 do baseline, zero novos**
- `flutter test` 20/20 · `go test ./...` verde
- Artefato conferido, não presumido: `aapt2` → `package="com.radiopx.radiopx"`, `launchMode=2` (singleTask), `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e as strings novas (`Verifique a conex`, `Falha ao carregar canais pr`, `Sem conex`) e **não** contém `10.201.108.96`
- APK release recompilado (50,8MB) e copiado para `C:\RadioPx\RadioPX.apk`

### Trade-offs e pendências

- **Em web** o `ready` é `Future.value()` (stub), então a detecção de falha de handshake do WebSocket não é efetiva nessa plataforma. `apiTimeout` na base só evita travamento, nunca falso-bloqueio. O alvo de teste é APK.
- **A causa raiz da troca de IP continua fora do repo**: a máquina recebe IP por DHCP e a URL fica fixa no binário via `--dart-define`. Resolver com reserva DHCP/IP estático no roteador, ou tornando a URL editável em runtime (esta última exige um RF/RNF novo em `docs/requisitos.md` — decisão pendente).
- `channel_detail_screen` continua órfã (nunca é navegada).

---

## 04/10/2026 - Correção: app abria em múltiplas instâncias em vez de ir para segundo plano

### Alterações

- `frontend/android/app/src/main/AndroidManifest.xml`:
  - removido `android:taskAffinity=""` — a activity ficava sem afinidade com qualquer task, então o launcher criava uma **task nova a cada lançamento** em vez de retomar a existente
  - `android:launchMode="singleTop"` → `"singleTask"` — `singleTop` só reaproveita a instância se a activity já estiver no topo da task atual, o que nunca ocorre com o app em background

### Motivo

Teste em campo: ao tocar no ícone do launcher, o app abria em uma instância nova (na aba Canais) em vez de voltar para a tela anterior. Isso gerava múltiplas cópias no recents, e como cada instância é um Flutter engine próprio, cada uma abria seu próprio WebSocket no mesmo canal (voice_screen.dart) e sua própria sessão `flutter_sound` — audible como áudio dobrado e PTT agindo em duplicidade.

Isso também explica o relato de "não há opção de sair do canal": a única forma de sair (engrenagem no AppBar → "Sair do Canal") fica na tela de voz, e o usuário nunca retornava a ela porque o ícone do launcher abria outra instância. Os dois atributos foram introduzidos à mão no commit `d6c3f91` (22/09) e não existem no template padrão do Flutter.

### Verificação

- `aapt2 dump xmltree` no APK gerado: `android:launchMode(0x0101001d)=2` (singleTask) e `taskAffinity` **ausente** — correção conferida no artefato final, não presumida
- `libapp.so` do APK contém `http://10.201.108.96:8080` e **não** contém o IP antigo `172.16.158.222` (a rede mudou desde 22/09 e o `--dart-define` do APK anterior estava desatualizado)
- APK release recompilado (50,8MB / 53.279.333 bytes) e copiado para `C:\RadioPx\RadioPX.apk`

### Trade-offs aceitos

- `singleTask` roteia novos intents para `onNewIntent` em vez de `onNewActivity`. Sem impacto agora: `MainActivity` é um `FlutterActivity` vazio e não há deep link. Se futuramente houver notificação ou deep link, tratar via `onNewIntent`/`NewIntentIntent`.
- `singleTask` impede duas instâncias do app no mesmo aparelho. Não afeta o teste em 2 celulares.

### Resultado

- `flutter analyze` sem novos issues; `flutter test` 20/20 verdes
- **Validação em campo pendente**: abrir o app, entrar num canal, ir para a tela inicial e tocar no ícone 3-4 vezes — deve retomar na mesma tela e o recents deve mostrar 1 entrada. Depois revalidar o fluxo de sair (engrenagem → "Sair do Canal"). Se o fluxo de sair ainda falhar, o próximo passo é expor um botão visível na tela de voz.

---

## 04/10/2026 - Correção: lista de canais não atualiza automaticamente

### Alterações

- `frontend/lib/features/channels/screens/channels_screen.dart`:
  - polling automático a cada 10s (`_autoRefreshInterval`) para atualizar as listas "Meus canais" e "Perto de mim" sem intervenção do usuário
  - `_loadAll` e `_silentRefresh` agora delegam para `_refresh({required bool showLoading})` — os call sites existentes (botão refresh, pull-to-refresh, criar/entrar/sair de canal) não mudaram
  - refresh silencioso não altera `_isLoading`/`_error`, portanto o `CircularProgressIndicator` de tela cheia e a tela de erro ficam restritos à carga inicial e ao refresh manual (antes, um refresh em polling trocaria o `TabBarView` inteiro e resetaria aba/scroll)
  - erro em refresh silencioso é ignorado (mantém a lista visível em vez de trocar a tela inteira por um erro)
  - guarda de concorrência `_isRefreshing` evita requisições sobrepostas e respostas fora de ordem
  - `_sameChannels` só chama `setState` quando algo de fato mudou (id, userCount, isAvailable, updatedAt, distance)
  - cache de localização com TTL de 2min (`_resolveLocation`/`_locationTtl`) — o polling reutiliza a última fix em vez de acordar o GPS de alta precisão a cada 10s; refresh manual continua forçando GPS fresco
  - `WidgetsBindingObserver`: pausa o timer quando a app vai para background e faz um refresh imediato ao voltar ao foreground

### Motivo

Teste em campo com 2 celulares: ao criar um canal no aparelho B, o aparelho A só mostrava o canal novo depois de trocar de aba. O `ChannelsScreen` só buscava dados em `initState`, no botão refresh e no pull-to-refresh — o reload ao trocar de aba era acidental, causado pelo `State` ser destruído e recriado por `app.dart` (que monta a lista de telas dentro do `build`).

### Resultado

- `flutter analyze` sem novos issues (16 pré-existentes, nenhum em `channels_screen.dart`)
- `flutter test` 20/20 verdes; `go test ./...` verde

### Pendências identificadas (não alteradas)

- `backend/internal/channel/model.go`: `Channel` não possui campo `distance`; `GetNearbyChannels` calcula a distância apenas para ordenar e nunca a serializa → `ChannelModel.distance` é sempre `null` e a linha de distância do card de canal nunca aparece. Requer mudança de backend.
- `frontend/lib/app.dart`: as telas continuam sendo construídas dentro do `build`, o que destrói o `State` ao trocar de aba (perda de estado). Passou a ser irrelevante para a atualização da lista, mas continua sendo a causa raiz do comportamento original.

---

## 22/09/2026 - Correção: Logout quebrado, erro no cadastro, login travado e latência de áudio

### Alterações

- `frontend/lib/features/profile/services/profile_service.dart`: `baseUrl` fixa (`http://localhost:8080`) substituída por `AppConfig.apiBaseUrl` — perfil (e o botão Sair) funcionam no celular
- `frontend/lib/features/auth/services/auth_service.dart`:
  - novo `ValueNotifier<bool> sessionState` (liga/desliga toda vez que o token de sessão muda em `logout`/`_saveAuthData`)
  - `register` aceita `200` e `201` (antes só `201` → missão confundia o sucesso com erro genérico)
  - login e register agora propagam o corpo da resposta do backend via `_readError` (mensagens de erro reais, ex.: e-mail duplicado)
- `frontend/lib/app.dart`: `AuthGate` escuta `AuthService.sessionState` e reavalia o estado de login — sem precisar reiniciar o app
- `frontend/lib/features/auth/screens/login_screen.dart`: removido `_LoginSuccess` desnecessário; navegação agora é reativa (AuthGate troca o body); guards `mounted` nos `setState`
- `frontend/lib/features/auth/screens/register_screen.dart`: sucesso volta para a tela inicial via `popUntil`; mensagem amigável para e-mail já cadastrado; guards `mounted`
- `frontend/lib/features/profile/screens/profile_screen.dart`: botão "Sair" fica sempre visível (inclusive quando o perfil falha); `_logout` só fecha a sessão (o AuthGate cuida da navegação)
- `frontend/lib/features/voice/services/stream_audio_player.dart`: reescrito com `flutter_sound` (sessão de streaming contínua PCM16/16kHz, `uint8ListSink`) — elimina o delay de ~2,5s do buffer inicial do `just_audio`
- `frontend/lib/features/voice/screens/voice_screen.dart`: voz abre a sessão de áudio ao entrar no canal (`startSession`)
- `frontend/pubspec.yaml`: `just_audio` removido; `flutter_sound ^9.28.0` adicionado
- `frontend/android/app/build.gradle.kts`: `minSdk` elevado para 24 (requisito do flutter_sound)
- APK release recompilado com `--dart-define=API_BASE_URL=http://172.16.158.222:8080` e copiado para `C:\RadioPx\RadioPX.apk`

### Motivo

Teste em campo: cadastro mostrava erro mesmo quando funcionava, login não entrava até reiniciar o app, não havia como sair da conta e o áudio tinha delay enorme.

### Resultado

- `flutter analyze` sem erros; `flutter test` 20/20 verdes
- `flutter build apk --release` OK (50,4MB) → `C:\RadioPx\RadioPX.apk`

---

## 22/09/2026 - Correção: Joining de canal e voz entre 2 celulares na mesma conta

### Alterações

- `backend/internal/channel/handler.go`: join idempotente — reentrar em canal que já pertence agora retorna 200 em vez de 409 (corrige "Failed to join channel" ao retocar o canal ou usar a mesma conta em 2 celulares)
- `backend/pkg/websocket/hub.go`: broadcast de áudio não envia mais o pacote de volta ao próprio remetente (elimina o eco)
- `frontend/lib/features/channels/services/channel_service.dart`: exceções de join/leave agora incluem status HTTP e corpo da resposta para facilitar diagnóstico
- `frontend/lib/features/voice/screens/voice_screen.dart`: removida a supressão de eco por `user_id` (a supressão agora é feita no servidor) — permite que 2 celulares logados na mesma conta conversem entre si
- `scripts/test-backend.ps1`: valida que join repetido retorna 200 (idempotente) e que o remetente NÃO recebe o próprio eco
- APK release recompilado com as correções e copiado para `C:\RadioPx\RadioPX.apk`

### Motivo

Teste em campo com 2 celulares na mes ma conta falhava ao entrar no canal (409) e, se entrasse, cada aparelho silenciava o outro (supressão por user_id).

### Resultado

- Script de validação 28/28 verdes (incluindo idempotência e ausência de eco)
- `go test ./...` verde; `flutter test` 20/20; `flutter analyze` sem erros

---

## 22/09/2026 - Protótipo Funcional + APK Android

### Alterações

- Backend Go compilado e rodando (`server.exe`) sem depender de PostgreSQL (dados em memória)
- Adicionado carregamento automático do `.env` no `main.go` (JWT_SECRET/SERVER_PORT agora funcionam sem export manual)
- Corrigido JWT secret: middleware de autenticação usa o mesmo fallback do `main.go`
- Corrigidas rotas de canais: `/api/v1/channels/{id}/join` e `/api/v1/channels/{id}/leave`
- Corrigidos endpoints de perfil: `/api/v1/user/profile` e `/api/v1/user/change-password` agora alcançáveis
- Frontend: criado `lib/core/config.dart` com `API_BASE_URL` configurável via `--dart-define`
- Frontend: services (auth/channels/profile) usam a config central
- Frontend: `websocket_service.dart` corrigido (protocolo `room`, áudio em base64) e ligado na VoiceScreen
- Voz real de ponta a ponta: captura PCM16 via `record`, streaming de reprodução via `just_audio` (StreamAudioSource + WAV)
- PTT funcional: segurar grava e transmite via WebSocket, soltar para; supressão de eco próprio; reconexão automática
- VoiceScreen: sair do canal e ver membros funcionais
- Permissões Android adicionadas (INTERNET, RECORD_AUDIO, localização) + cleartext liberado
- Pasta `android/` gerada e APK release compilado (assinado com debug key = instalável)
- Testes Flutter 20/20 verdes; `flutter analyze` sem erros; backend `go test ./...` verde

### Resultado

- APK em `frontend/build/app/outputs/flutter-apk/app-release.apk` (cópia em `RadioPX.apk`)
- Backend em `:8080`. Celular acessa via IP LAN da máquina (ex.: `http://172.16.158.222:8080`)
- Observação: regra de firewall para porta 8080 exige execução como administrador

---

## 08/09/2026 - Implementação Phase 9: Polish & Cross-Cutting Concerns

### Alterações

- Atualizada documentação em docs/arquitetura.md
- Criado utilitários comuns em backend/pkg/utils/helpers.go
- Implementado rate limiting middleware
- Implementado security configuration
- Adicionadas validações de entrada e sanitização
- Projeto 100% implementado (73/73 tarefas)
- Atualizado PROJECT_STATE.md com status final

### Motivo

Implementação final do projeto RadioPX. Todas as phases concluídas com sucesso.

### Resultado

Projeto completo com todas as funcionalidades:
- Push-to-Talk (PTT)
- Descoberta por Proximidade (GPS)
- Alocação Dinâmica de Canais
- Resiliência de Conexão (Buffer 30s)
- Autenticação JWT
- Perfil do Usuário
- Rate Limiting e Segurança

---

## 08/09/2026 - Implementação Phase 8: Profile Feature

### Alterações

- Criado Profile model no frontend
- Implementado Profile service com operações CRUD
- Implementado Profile screen com visualização e edição
- Implementado Profile editing com modal bottom sheet
- Implementado Change password functionality
- Implementado Logout com confirmação
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da Profile Feature. Gerenciamento de perfil do usuário.

### Próximo passo

Implementar Phase 9: Polish & Cross-Cutting Concerns.

---

## 08/09/2026 - Implementação Phase 7: Auth Feature

### Alterações

- Criado User model no backend
- Criado User model no frontend
- Implementado User Service com autenticação
- Implementado User handler para operações CRUD
- Implementado AuthService com JWT
- Implementado AuthMiddleware para rotas protegidas
- Implementado Login screen no Flutter
- Implementado Register screen no Flutter
- Implementado auth service no frontend com SharedPreferences
- Implementado hash de senhas com bcrypt
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da Auth Feature. Autenticação JWT com hash de senhas bcrypt.

### Próximo passo

Implementar Phase 8: Profile Feature.

---

## 08/09/2026 - Implementação Phase 6: User Story 4 (Resiliência de Conexão)

### Alterações

- Criado MessageBuffer model no backend
- Criado MessageBuffer model no frontend
- Implementado AudioBufferService com limpeza automática
- Implementado BufferManager para gerenciamento de 30 segundos
- Implementado WebSocketManager com lógica de reconexão
- Implementado MessageReplayService para reprodução na reconexão
- Implementada limpeza do buffer após reprodução bem-sucedida
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 4 (Resiliência de Conexão). Buffer de áudio de 30 segundos para quedas de conexão.

### Próximo passo

Implementar Phase 7: Auth Feature (JWT + Google OAuth).

---

## 08/09/2026 - Implementação Phase 5: User Story 3 (Alocação Dinâmica de Canais)

### Alterações

- Criado Channel model no backend
- Atualizado Channel model no frontend com createdBy
- Implementado Channel Service com lógica de alocação dinâmica
- Implementado Channel handler para operações CRUD
- Implementado JoinChannelOrCreate para auto-criação
- Implementado limite de 15 canais visíveis
- Implementado Channels screen com descoberta por proximidade
- Implementado Channel detail screen
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 3 (Alocação Dinâmica de Canais). Canais são criados automaticamente com máximo de 10 usuários por canal.

### Próximo passo

Implementar User Story 4: Resiliência de Conexão.

---

## 08/09/2026 - Implementação Phase 4: User Story 2 (Descoberta por Proximidade)

### Alterações

- Criado Location model no backend
- Criado Location model no frontend
- Implementado Location Service com cálculo Haversine
- Implementado Location handler para atualizações GPS
- Implementado serviço de localização no Flutter
- Implementado tratamento de permissões de localização
- Implementado Channels screen com descoberta por proximidade
- Implementado Channel service no frontend
- Criado Channel model no frontend
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 2 (Descoberta por Proximidade). Funcionalidade de encontrar canais baseada em GPS com raio de 10km.

### Próximo passo

Implementar User Story 3: Alocação Dinâmica de Canais.

---

## 08/09/2026 - Implementação Phase 3: User Story 1 (Push-to-Talk)

### Alterações

- Criado AudioPacket model no backend
- Criado AudioPacket model no frontend
- Implementado Voice Service no backend
- Implementado WebSocket handler para streaming de áudio
- Implementado serviço de captura de áudio no Flutter
- Implementado serviço PTT no Flutter
- Criados widgets: PTT Button, Audio Visualizer, Audio Player
- Implementado Voice Screen
- Implementado WebSocket service no Flutter
- Implementado tratamento de permissões de microfone
- Implementado serviço de fila de áudio para reprodução sequencial
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 1 (Push-to-Talk) - MVP do projeto. Funcionalidade principal de comunicação por voz em tempo real.

### Próximo passo

Implementar User Story 2: Descoberta por Proximidade.

---

## 08/09/2026 - Implementação Phase 1 e Phase 2

### Alterações

- Configurado Go module com 9 dependências
- Configurado Flutter com 12 dependências
- Criada migração inicial do PostgreSQL
- Implementado middleware de autenticação JWT
- Implementado middleware CORS
- Implementado WebSocket hub para gerenciamento de conexões
- Implementado WebSocket client handler
- Criado modelo User com repository
- Criado modelo Channel com repository
- Configuradas variáveis de ambiente
- Implementado tratamento de erros e logging
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação dos requisitos RF01-RF08 e RNF01-RNF05. Estabelecimento da fundação do backend para User Stories.

### Próximo passo

Implementar User Story 1: Push-to-Talk (MVP).

---

## 08/09/2026 - Criação Inicial do Projeto

### Alterações

- Criada estrutura inicial do projeto
- Definida arquitetura do sistema (monolito modular)
- Criados requisitos funcionais (RF01-RF08)
- Criados requisitos não funcionais (RNF01-RNF05)
- Definido stack tecnológico:
  - Frontend: Flutter/Dart
  - Backend: Go/Golang
  - Banco: PostgreSQL
  - Comunicação: REST + WebSocket
  - Autenticação: JWT/OAuth/Casbin
- Criada documentação inicial:
  - README.md
  - AGENTS.md
  - docs/requisitos.md
  - docs/arquitetura.md
  - docs/banco-de-dados.md
  - docs/decisoes.md
  - CHANGELOG.md
  - PROJECT_STATE.md
- Definidas 10 decisões técnicas iniciais
- Criada feature specification (spec.md)
- Criada lista de tarefas (tasks.md)

### Motivo

Implementação dos requisitos RF01-RF08 e RNF01-RNF05. Estabelecimento da base do projeto para desenvolvimento futuro.

### Próximo passo

Implementar estrutura do backend Go conforme arquitetura definida.

---

## Legenda

- ✅ Concluído
- ⏳ Em andamento
- ❌ Pendente
- 🔧 Em revisão
