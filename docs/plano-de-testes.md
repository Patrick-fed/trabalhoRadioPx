# Plano de Testes — RadioPX

> **Data**: 10/11/2026
> **Escopo**: plano de testes do protótipo funcional (v1). Documento de referência —
> não altera código. Fontes: `docs/requisitos.md`, `PROJECT_STATE.md` (seção
> *Lacunas*), `CHANGELOG.md`, suítes existentes e `scripts/test-backend.ps1`.

---

## 1. Objetivo

Garantir que o sistema RadioPX atenda aos requisitos funcionais (RF01–RF09) e não
funcionais (RNF01–RNF06) definidos em `docs/requisitos.md`, com foco nos fluxos
críticos: autenticação, PTT (voz), descoberta por proximidade, admin e privacidade
de áudio.

## 2. Estratégia de testes

| Nível | Ferramenta | Comando | Status atual |
|-------|-----------|---------|--------------|
| Unitário (backend) | `go test` | `cd backend && go test ./... -count=1` | ✅ 81 testes em 11 arquivos |
| Unitário (frontend) | `flutter test` | `cd frontend && flutter test` | ✅ 62 testes em 9 arquivos |
| Integração (banco) | `go test` opt-in | `RADIOPX_TEST_DB=1 go test ./internal/db/... -count=1` | ✅ stores PG + migrações |
| E2E REST+WS | script PowerShell | `powershell -File scripts/test-backend.ps1 [-BaseUrl <url>]` | ✅ 46 checks (servidor rodando; 17/11/2026) |
| Estático | `go vet` / `flutter analyze` / `golangci-lint` | ver §7 | ✅ baseline (14 issues analyze) |
| Manual em campo | checklist (§6) | — | ⏳ pendências listadas em §6 |
| Carga / desempenho | fase futura (§8) | — | ❌ fora do escopo v1 (RNF02/RNF03) |

**Ordem de execução da regressão completa (gate de release):**

```bash
# 1. Backend
cd backend
go build ./... && go vet ./... && go test ./... -count=1
RADIOPX_TEST_DB=1 go test ./internal/db/... -count=1   # exige PostgreSQL

# 2. Frontend
cd frontend
flutter analyze    # baseline: 14 issues, zero novos
flutter test

# 3. E2E (servidor local ou túnel ngrok rodando)
powershell -File scripts/test-backend.ps1
powershell -File scripts/test-backend.ps1 -BaseUrl https://<tunnel>.ngrok-free.dev
```

## 3. Matriz de rastreabilidade (RF/RNF → cobertura → lacunas)

| Req | Escopo resumido | Cobertura automatizada | Lacuna |
|-----|-----------------|------------------------|--------|
| RF01 | Cadastro (e-mail único, senha ≥8) | `main_test.go` (register), `user/service_test.go`, script §1 | Confirmação de e-mail **não existe** (fora do código) |
| RF02 | Login JWT + refresh (OAuth fora da v1) | `auth/service_test.go`, `main_test.go`, script §1/§6, `auth_service_test.dart` | Google OAuth = trabalho futuro |
| RF03 | PTT, 1 falante/vez, latência <500ms | WS auth em `main_test.go`; relay no script §3 | voiceGuard (2º transmissor rejeitado) **sem teste automatizado**; latência só em campo |
| RF04 | Canais por GPS (raio 10km, 30s) | `location/distance_test.go`, `location/handler_test.go` (POST /location), `channel/service_test.go` (nearby), script §2 | Precisão 100m e timer no aparelho: validação em campo |
| RF05 | Max 10/canal, max 15 visíveis, canal automático | `channel/service_test.go` (`ChannelFull`, `MaxVisibleChannels`) | Criação automática de canal novo ao lotar: **sem teste** |
| RF06 | Buffer 30s + replay na reconexão | `voice/buffer_service_test.go` (5 casos) | E2E de `request_replay` com 2 clientes: **pendente em campo** |
| RF07 | Perfil: ver/editar/trocar senha | `user/handler_test.go`, `user/service_test.go`, script §4/§5 | — |
| RF08 | Fila de áudio (FIFO, teto 30s, sentinel) | `audio_queue_test.dart` (5 casos) | E2E de ponta a ponta no CI: **não existe** |
| RF09 | Painel admin (métricas, gestão, localização) | `admin/handler_test.go` (12), `user/service_test.go` (+6), `pg_integration_test` | **Não coberto pelo `test-backend.ps1`**; endereço no popup (Nominatim) e painel no navegador: validação manual |
| RNF01 | Áudio nunca persistido | — (sem teste) | **Sem critério automatizado** — ver TC-RNF01 |
| RNF02 | Latência <500ms, alocação <5s | — | Carga: fase futura (§8) |
| RNF03 | 10.000 concorrentes | — | Carga: fase futura (§8) |
| RNF04 | Buffer/reconexão/replay | parcial (buffer unit + script WS) | Desconexão simulada automatizada: pendente |
| RNF05 | JWT, rate limit 429, WS autenticado, TLS (ngrok) | `ratelimit_test.go`, `auth/service_test.go`, WS auth tests, script | Teste de penetração formal: não realizado |
| RNF06 | URL do servidor editável | `config_test.dart` (8), `server_url_sheet_test.dart` (2) | — |

## 4. Casos de teste funcionais

> Formato: **ID | Prioridade | Pré-condição | Passos | Esperado**.
> P1 = crítico (release), P2 = importante, P3 = complementar.

### RF01 — Criação de conta

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF01-01 | P1 | E-mail não cadastrado | `POST /api/v1/auth/register` com dados válidos | 201 + token + user (sem senha) |
| TC-RF01-02 | P1 | E-mail já existe | Repetir register com mesmo e-mail | 409 com mensagem PT-BR |
| TC-RF01-03 | P1 | — | Senha com 7 caracteres | 400, senha mínima 8 |
| TC-RF01-04 | P2 | — | E-mail com maiúsculas/espaços | Normalizado (`trim`+lowercase); unicidade preservada |
| TC-RF01-05 | P3 | Cadastro OK | Verificar UI do app (register_screen) | Cadastro conclui sem spinner infinito; e-mail duplicado mostra mensagem amigável |

> **Nota**: confirmação de e-mail (critério do RF01) não está implementada na v1 — registrar como pendência de requisito, não como falha de teste.

### RF02 — Login e sessão

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF02-01 | P1 | Conta válida | Login e-mail+senha | 200 com access + refresh token (HS256) |
| TC-RF02-02 | P1 | Conta válida | Senha errada | 401 mensagem PT-BR |
| TC-RF02-03 | P1 | Access token | `POST /auth/refresh` com access token | 401 (tipo trocado rejeitado) |
| TC-RF02-04 | P1 | Refresh token | `POST /auth/refresh` com refresh token | 200, novos tokens |
| TC-RF02-05 | P1 | — | 11 requisições de login/refresh em <1 min (mesmo IP) | 11ª retorna 429 |
| TC-RF02-06 | P2 | Sessão expirada no app | Abrir app sem access válida, com refresh | Renovação silenciosa ou logout limpo (`isLoggedIn`) |
| TC-RF02-07 | P3 | Token HS384/none forjado | Enviar token com algoritmo não-HS256 | 401 |

### RF03 — Push-to-Talk (voz)

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF03-01 | P1 | 2 usuários no mesmo canal | A conecta WS (`?token=`), envia `audio is_ptt=true` | B recebe o chunk; A **não** recebe eco próprio |
| TC-RF03-02 | P1 | A transmitindo | B tenta iniciar transmissão | B é rejeitado (voiceGuard) — **atualmente só unitário; automatizar** |
| TC-RF03-03 | P1 | A transmitindo | A desconecta (fecha WS) | Canal é liberado; B pode falar (fix `releaseVoice`) — **validar em campo** |
| TC-RF03-04 | P1 | — | Conectar WS sem token / token inválido | Handshake rejeitado (401) |
| TC-RF03-05 | P1 | 2 aparelhos, mesmo canal | PTT real: falar no A, ouvir no B | Latência percetível < 500ms (medir com cronômetro/gravador) |
| TC-RF03-06 | P2 | Usuário no canal | Indicação visual de canal ocupado na UI | Elemento visível ao ocupar |

### RF04 — Descoberta por proximidade

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF04-01 | P1 | Canal em SP (-23.55/-46.63) | `GET /channels/nearby?lat=&lng=&radius=10` próximo | Canal aparece com `distance` |
| TC-RF04-02 | P1 | Canal em SP | Query a >10km (ex.: Rio) | Canal **não** aparece |
| TC-RF04-03 | P2 | App com GPS | Tela "Perto de mim" | Canais ordenados por distância; card mostra `distanceFormatted` |
| TC-RF04-04 | P2 | App aberto | Esperar 30s / voltar do background | Localização atualizada (polling/TTL 2min local) |
| TC-RF04-05 | P1 | Logado em qualquer tela | `POST /api/v1/location` com lat/lng válidos | 200 `{"ok":true}`; `last_location_at` atualizado (teste unitário: `location/handler_test.go`) |
| TC-RF04-06 | P1 | Logado, app aberto | Observar ~30s (ou log do reporter) | Reporte periódico automático; falha de rede não gera erro na UI (best-effort) |
| TC-RF04-07 | P2 | App em background | Pausar/retomar app | Reporter pausa em `paused` e retoma com tick imediato no `resumed` |
| TC-RF04-08 | P2 | — | Body com lat 91 / lng 181 ou sem `X-User-ID` | 400 (validação do handler) |

### RF05 — Alocação dinâmica de canais

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF05-01 | P1 | Canal com 9 membros | 10º usuário faz join | Join aceito; canal cheio |
| TC-RF05-02 | P1 | Canal com 10 membros | 11º usuário faz join | Rejeitado ou redirecionado a novo canal (fluxo de criação automática — **sem teste, cobrir**) |
| TC-RF05-03 | P1 | Usuário em ≥15 canais | Listar nearby | Lista limitada a 15 itens |
| TC-RF05-04 | P2 | Mesmo usuário | Join repetido | 200 idempotente (já no script) |

### RF06 — Buffer de conexão / replay

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF06-01 | P1 | A transmitindo no canal | A envia ≥3 chunks com sequence | B recebe com `sequence` incremental; buffer registra (status) |
| TC-RF06-02 | P1 | B recebeu até seq N | B cai e reconecta, envia `request_replay {room, since_sequence}` | Recebe segmentos perdidos por usuário, encerrado por `is_ptt=false` — **E2E pendente (2 clientes)** |
| TC-RF06-03 | P2 | Buffer com >30s de áudio | Chamar cleanup | Mensagens antigas descartadas; só memória (sem disco/PG) |
| TC-RF06-04 | P2 | — | Auditoria: procurar blobs de áudio em `users`/`channels`/logs | Nenhum áudio persistido (liga TC-RNF01-01) |

### RF07 — Perfil

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF07-01 | P1 | Logado | `GET /user/profile` | 200 com dados; sem `password` |
| TC-RF07-02 | P1 | Logado | `PUT /user/profile` renomeando | 200 devolve objeto atualizado (id/email/name); UI mostra "Perfil atualizado" |
| TC-RF07-03 | P1 | Logado | Change password com senha atual errada | 401/400 PT-BR |
| TC-RF07-04 | P2 | Logado | Nova senha <8 caracteres | 400 (validação backend e UI) |
| TC-RF07-05 | P2 | Logado | Excluir conta (`DELETE /user/profile`) | 200; login subsequente 401 |

### RF08 — Fila de áudio

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF08-01 | P1 | Fila vazia | Enfileirar 3 chunks + sentinel | Reproduz em ordem FIFO; sentinel encerra segmento sem emitir áudio |
| TC-RF08-02 | P1 | Fila perto do teto (~960KB) | Enfileirar além do teto | Chunks mais antigos descartados; não estoura memória |
| TC-RF08-03 | P2 | Em reprodução | `clear()` | Fila esvazia e para de tocar |
| TC-RF08-04 | P2 | Segmento ativo | Sem chunks por 2,5s (ocioso) | Segmento encerra automaticamente |

### RF09 — Painel administrativo

| ID | P | Pré-condição | Passos | Esperado |
|----|---|--------------|--------|----------|
| TC-RF09-01 | P1 | Admin logado (via `ADMIN_EMAILS`) | `GET /admin` | 200, SPA carrega (HTML + assets + Leaflet vendorado) |
| TC-RF09-02 | P1 | Não-admin / sem token | `GET /api/v1/admin/stats` e `/users` | 401/403 |
| TC-RF09-03 | P1 | Admin | `GET /admin/users` | Lista **sem** `password_hash`; com `last_lat/last_lng/last_location_at` |
| TC-RF09-04 | P1 | Admin | `PATCH users/{id}` `is_active=false` | 200; usuário suspenso **não consegue logar** (Authenticate rejeita) |
| TC-RF09-05 | P1 | Admin logado como si mesmo | Self-suspend / self-demote / self-delete | 403 em todos |
| TC-RF09-06 | P1 | Admin | `DELETE users/{id}` | 200; usuário some da lista |
| TC-RF09-07 | P2 | Admin no navegador | Dashboard com refresh 10s | Cards com subtexto/badges de variação; pan do mapa estável |
| TC-RF09-08 | P2 | Usuário com posição conhecida | Ver mapa | Marcador por usuário; suspenso âmbar; popup com link OSM; sem posição = sem marcador |
| TC-RF09-09 | P2 | Boot do servidor | Env `ADMIN_EMAILS=a@x,b@y` | Bootstrap idempotente promove ambos |
| TC-RF09-10 | P1 | — | **Adicionar ao `test-backend.ps1`** (bloco admin: 02/03/04/05/06) | Script cobre RF09 na regressão |
| TC-RF09-11 | P2 | Admin no navegador, usuário com posição | Abrir popup do marcador | Linha de endereço aparece ("endereço…" → endereço pt-BR via Nominatim); falha mostra "endereço indisponível"; 2º popup do mesmo ponto vem do cache |

### RNF — Não funcionais selecionados

| ID | P | Passos | Esperado |
|----|---|--------|----------|
| TC-RNF01-01 | P1 | Após 30min de conversa: `pg_dump`/consulta em todas as tabelas + varredura de logs | Zero registros de áudio; só `users`, `channels`, `user_channels`, `schema_migrations` |
| TC-RNF01-02 | P2 | Conferir retenção | Política documentada (README/docs); sem payload de áudio em log |
| TC-RNF05-01 | P1 | 11× register/login/refresh mesmo IP | 429 na 11ª (já unit + manual) |
| TC-RNF05-02 | P1 | Headers `X-Forwarded-For` forjados | Rate limit continua por `RemoteAddr` (não burlável) |
| TC-RNF05-03 | P2 | WS sem `?token` | Rejeitado; `user_id` forçado ao do token (anti-spoofing) |
| TC-RNF05-04 | P3 | Produção | TLS via ngrok (DEC-013); Casbin/ TLS nativo = fora da v1 |
| TC-RNF06-01 | P1 | Trocar URL no sheet do app (Testar → Salvar) | `GET /health` validado; persiste após restart; sem recompilar |
| TC-RNF06-02 | P1 | URL sem `http(s)://` | Recusa com mensagem clara |
| TC-RNF06-03 | P2 | `https:` salvo | WS deriva `wss:` automaticamente |
| TC-RNF04-01 | P1 | Matar backend ~5s no meio do PTT de B, subir de novo | B reconecta com backoff (2/4/8/16/30s, 5 tentativas); falha vira "Tentar novamente" |

## 5. Testes de segurança (RNF05 — checklist)

- [ ] Registro/login/refresh com força bruta → 429 após 10/min por IP
- [ ] JWT forjado (secret errado, alg trocado, exp vencido) → 401
- [ ] Access token usado como refresh → 401; refresh usado como access → rejeitado
- [ ] Endpoints `/api/v1/admin/*` sem token → 401; usuário comum → 403
- [ ] `GET /admin/users` nunca retorna `password_hash` (grep na resposta)
- [ ] WS com token de outro usuário / sem canal de membership → rejeição
- [ ] Exclusão de conta: tokens antigos deixam de valer
- [ ] CORS: origem ngrok aceita; origem arbitrária não
- [ ] `.env` / `JWT_SECRET` fora do versionamento (verificar `git status` + `.gitignore`)

## 6. Testes manuais em campo (checklist pendente)

Itens extraídos de `PROJECT_STATE.md` → *Problemas Conhecidos* e *Lacunas*:

- [ ] **RF03/RF06** — 2 aparelhos no mesmo canal: PTT com latência <500ms; reconexão no meio da fala com replay (`request_replay`) sem buraco perceptível
- [ ] **RF03** — 2º transmissor é bloqueado e o canal não fica preso "ocupado" após disconnect
- [x] **RF09** — painel `/admin` no navegador: dashboard + mapa com marcadores (Chrome headless/CDP, 10/11/2026); ações de gestão seguem pendentes de clique manual
- [x] **RF04/DEC-020** — ticks reais de `POST /location` → banco + `GET /admin/users` com `last_location_at` fresco (10/11/2026); reporter coberto por `location_reporter_test.dart` (intervalo/guard/start/stop)
- [x] **RF09/DEC-020** — popup do mapa resolve endereço pt-BR via Nominatim em browser real; CORS `ACAO=*` verificado (10/11/2026)
- [ ] **RF04 em aparelho** — GPS real + lifecycle (pause/resume do reporter) no Android: pendente de dispositivo/emulador (sem toolchain VS/AVD nesta máquina)
- [ ] **Timeout/reconexão** — backend parado: erro em ~10s (não spinner eterno); 5 tentativas de backoff → "Tentar novamente"
- [ ] **Instância única** — tocar no ícone do launcher 3-4× com app em background: retoma na mesma tela, 1 entrada no recents; fluxo "Sair do Canal" funciona
- [ ] **RF05** — criar canal no celular B e ver aparecer em A em ≤10s (polling); ida ao background→foreground atualiza na hora
- [ ] **RNF06** — URL nova do ngrok trocada no app sem rebuild

## 7. Critérios

### Entrada
- Build verde (`go build`, `flutter analyze` sem issues novos vs baseline de 14)
- Servidor local ou túnel ngrok disponível; PostgreSQL para teste de integração
- Dados de teste limpos (e-mails com sufixo timestamp, como no script)

### Saída (gate de release)
1. `go test ./... -count=1` → 100% pass
2. `RADIOPX_TEST_DB=1 go test ./internal/db/...` → pass (quando PG disponível)
3. `flutter test` → 100% pass (baseline 62)
4. `scripts/test-backend.ps1` → 46/46 (29 base + admin/voiceGuard/replay desde 17/11)
5. Checklist de campo (§6) sem P1 pendente
6. Nenhuma regressão conhecida aberta em `PROJECT_STATE.md` → *Problemas Conhecidos*

### Priorização de novos testes (recomendações, fora desta etapa)

> **✅ Itens 1–5 executados em 17/11/2026** (ver `plano-de-testes-pendencias.md` §8).

1. **P1** — bloco admin no `test-backend.ps1` (RF09 sem cobertura E2E)
2. **P1** — caso de voiceGuard no script (2º transmissor)
3. **P1** — caso de `request_replay` no script (RF06)
4. **P2** — caso de canal lotado → canal novo (RF05)
5. **P2** — teste automatizado de ausência de áudio persistido (TC-RNF01-01)

## 8. Fase futura — testes de carga e desempenho (RNF02/RNF03)

> **Fora do escopo v1** (decisão registrada em `docs/requisitos.md` e
> `PROJECT_STATE.md`). Deixado como planejamento para quando houver ambiente.

| Cenário | Meta | Ferramenta sugerida |
|---------|------|---------------------|
| Latência de ponta a ponta do áudio | < 500ms (mediana), p95 < 800ms | 2 clientes WS + carimbo de tempo no payload / Wireshark |
| Alocação de canal | < 5s sob carga | script medindo tempo do `POST /channels` + `join` |
| 1.000 usuários simultâneos (REST) | Sem erro 5xx; rate limit estável | `k6` ou `vegeta` contra `/health`, register/login, nearby |
| 10.000 conexões WS por região (RNF03) | Sem OOM; relay estável | harness Go com N goroutines de WS por canal |
| Estresse multi-canal | Criação automática de canais sob demanda | mistura WS + joins em lote |

**Requisitos mínimos do ambiente de carga**: máquina dedicada (não o dev laptop),
PostgreSQL isolado, métricas de CPU/memória, e execução **fora** do rate limit de
10/min (ou com IPs múltiplos / cota ajustada — ver Problema Conhecido nº 9).

## 9. Ambiente e dados de teste

| Item | Valor |
|------|-------|
| Backend local | `backend/server.exe` (porta 8080, `.env` com `DB_*`, `JWT_SECRET` de dev) |
| Túnel | `ngrok http 8080` + header `ngrok-skip-browser-warning: 1` (já no script) |
| Banco | PostgreSQL 14+; integração exige `RADIOPX_TEST_DB=1` |
| APK | `flutter build apk --release --dart-define=API_BASE_URL=<url>` |
| Contas de teste | e-mails com sufixo Unix timestamp (auto-limpas por expiração); admin via `ADMIN_EMAILS` |
| Limpeza | script não limpa dados — usar DB descartável ou truncate de `users`/`channels`/`user_channels` entre rodadas |
| Dois aparelhos | necessários apenas para TC-RF03-05, TC-RF06-02 e checklist de campo |

## 10. Rastreabilidade e referências

- Requisitos: [`docs/requisitos.md`](requisitos.md)
- Lacunas e problemas conhecidos: [`PROJECT_STATE.md`](../PROJECT_STATE.md)
- Decisões (rate limit, ngrok, replay): [`docs/decisoes.md`](decisoes.md)
- Script E2E: `scripts/test-backend.ps1`
- Histórico de correções e verificações: [`CHANGELOG.md`](../CHANGELOG.md)
- Comandos oficiais: [`README.md`](../README.md) → *Testes*
