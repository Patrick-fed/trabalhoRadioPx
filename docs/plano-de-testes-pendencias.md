# Plano de Testes — Pendências Executáveis (RadioPX)

> **Data**: 12/11/2026 · **Atualizado**: 17/11/2026 (Blocos A e B executados)
> **Escopo**: apenas o que **falta testar/validar** para fechar a v1. Este documento
> complementa o [`plano-de-testes.md`](plano-de-testes.md) (referência completa) sem
> duplicá-lo. Não altera código.
> **Fontes**: `docs/requisitos.md`, `PROJECT_STATE.md` (seção *Lacunas / Aderência
> aos Requisitos*), `docs/plano-de-testes.md` (§6 checklist de campo, §7 priorização).

---

## 1. Objetivo

Executar as pendências de teste e validação identificadas na auditoria de
requisitos, em ordem de prioridade, de modo a deixar o gate de release
(`plano-de-testes.md` → §7) sem itens P1 em aberto.

**Fora do escopo deste documento:**

- Testes de carga/desempenho (RNF02/RNF03 — fase futura, §8 do plano completo)
- Google OAuth (RF02) e Casbin (RNF05) — trabalho futuro decidido
- Teste de penetração formal — não realizado, fora da v1
- Confirmação de e-mail (RF01) e revogação ativa de token de
  suspenso — são **pendências de requisito/decisão**, não de teste (ver §6;
  o codec opus/PCM16 foi encerrado pela DEC-021 em 17/11/2026)

---

## 2. Pré-requisitos / ambiente

| Item | Detalhe |
|------|---------|
| Backend | Rodando local (`go run cmd/server/main.go`) ou via túnel `ngrok http 8080` |
| Env | `ADMIN_EMAILS` com pelo menos 1 e-mail de admin (bootstrap no boot) |
| Banco | PostgreSQL 14+ com `DB_*` configurado (migrações aplicadas no boot) |
| Script | `scripts/test-backend.ps1` (46 checks desde 17/11; header ngrok embutido; admin via `RADIOPX_ADMIN_EMAIL`) |
| APK | `flutter build apk --release --dart-define=API_BASE_URL=<url>` (ou app rodando via `flutter run`) |
| Aparelhos | **2 dispositivos Android** para o Bloco C (C1–C3, C5–C7); sem AVD nesta máquina |
| Contas | E-mails de teste com sufixo Unix timestamp; 1 admin via `ADMIN_EMAILS`; 1 usuário comum |
| Limpeza | DB descartável ou truncate de `users`/`channels`/`user_channels` entre rodadas |

**Comandos de regressão de referência** (rodar antes e depois de qualquer mudança):

```bash
# Backend
cd backend
go build ./... && go vet ./... && go test ./... -count=1
RADIOPX_TEST_DB=1 go test ./internal/db/... -count=1   # exige PostgreSQL

# Frontend
cd frontend
flutter analyze    # baseline: 14 issues, zero novos
flutter test       # baseline: 62

# E2E (servidor rodando)
powershell -File scripts/test-backend.ps1   # baseline: 46 checks
```

---

## 3. Bloco A — Novos testes automatizados (P1/P2)

> **✅ Executado em 17/11/2026** — A1–A5 cobertos (script 46/46 + testes Go);
> registro em §8. O texto abaixo fica como referência do que foi feito.
> Priorização conforme `plano-de-testes.md` → §7. Cada caso indica o arquivo a
> alterar. Após implementar, rodar a regressão completa (§2).

### A1 — Bloco admin no `test-backend.ps1` — **P1** (TC-RF09-10)

- **Objetivo**: cobrir RF09 na regressão E2E (hoje só `admin/handler_test.go`).
- **Arquivo**: `scripts/test-backend.ps1`
- **Passos** (após login de admin):
  1. `GET /api/v1/admin/stats` com token de admin → 200 com as 6 métricas
  2. `GET /api/v1/admin/users` com token de admin → 200; resposta **sem** `password_hash`; com `last_lat`/`last_lng`/`last_location_at`
  3. `GET /api/v1/admin/stats` **sem** token → 401; com token de usuário comum → 403
  4. `PATCH /api/v1/admin/users/{id}` `{"is_active": false}` → 200; em seguida login desse usuário → **401**
  5. `PATCH` do admin **sobre si mesmo** → 403 (self-action); `DELETE` de si mesmo → 403
  6. `DELETE /api/v1/admin/users/{id}` (usuário de teste) → 200; `GET` da lista → some
- **Esperado**: script passa de **28 para ≥33 checks**, todos verdes.
- **Referência**: TC-RF09-02/03/04/05/06.

### A2 — voiceGuard no script — **P1**

- **Objetivo**: E2E de "apenas um usuário fala por vez" (RF03) e de liberação no
  disconnect (fix `releaseVoice`).
- **Arquivo**: `scripts/test-backend.ps1` (bloco WS §3)
- **Passos**:
  1. Usuário A conecta no WS do canal e envia `audio is_ptt=true` (inicia transmissão)
  2. Usuário B conecta no mesmo canal e tenta enviar `is_ptt=true`
  3. **Esperado**: B é rejeitado (não assume a transmissão); A segue falando
  4. A desconecta (fecha o WS)
  5. B envia `is_ptt=true`
  6. **Esperado**: B agora consegue transmitir; o canal **não** fica preso "ocupado"
- **Referência**: TC-RF03-02, TC-RF03-03.

### A3 — `request_replay` no script — **P1** (RF06)

- **Objetivo**: E2E do replay pós-reconexão (hoje só unitário em `buffer_service_test.go`).
- **Arquivo**: `scripts/test-backend.ps1`
- **Passos**:
  1. A transmite no canal; A envia ≥3 chunks com `sequence` incremental; B recebe até a sequência N
  2. B desconecta; A envia mais 2 chunks (sequências N+1, N+2)
  3. B reconecta e envia `request_replay {room, since_sequence: N}`
  4. **Esperado**: B recebe os segmentos perdidos por usuário e a replay é encerrada por um pacote `is_ptt=false` (sentinel); membership checada (B precisa ser membro)
- **Referência**: TC-RF06-02.

### A4 — Canal lotado → canal novo — **P2** (RF05)

- **Objetivo**: cobrir "criação automática de canal ao atingir 10 usuários" (hoje sem teste).
- **Arquivo**: `scripts/test-backend.ps1` **ou** `backend/internal/channel/service_test.go`
- **Passos**:
  1. Criar canal; fazer join de **10** usuários distintos → 10º aceito, canal cheio
  2. 11º usuário tenta join → **Esperado**: rejeição (`ChannelFull`) ou redirecionado/criação de novo canal, conforme o fluxo implementado; lista nearby limitada a 15 itens permanece
- **Referência**: TC-RF05-01, TC-RF05-02.

### A5 — Áudio nunca persistido — **P2** (RNF01 / TC-RNF01-01)

- **Objetivo**: critério automatizável de privacidade (RNF01).
- **Arquivo**: novo teste em `backend/internal/db/` (opt-in `RADIOPX_TEST_DB=1`) **ou** check no `test-backend.ps1`
- **Passos**:
  1. Trafegar áudio via WS por alguns segundos (ou reutilizar o relay do script)
  2. Varredura: `SELECT` em `users`, `channels`, `user_channels`, `schema_migrations` + greps de logs
  3. **Esperado**: nenhum campo/blob de áudio; só as 4 tabelas esperadas; sem payload de áudio em log
- **Referência**: TC-RNF01-01, TC-RF06-04.

---

## 4. Bloco B — Testes unitários faltantes (P2, opcional)

> **✅ Executado em 17/11/2026** — B1 (`voice/service_test.go`) e B2
> (`channel/service_test.go`, concorrência) implementados; registro em §8.

| ID | Teste | Arquivo |
|----|-------|---------|
| B1 | **voiceGuard automatizado**: segundo `StartTransmission` no mesmo canal retorna erro; `StopTransmission` libera | `backend/pkg/websocket/` ou `backend/internal/voice/` (`*_test.go`) |
| B2 | **Criação automática de canal**: `ChannelFull` na junção → novo canal criado/acionado, com transação (join concorrente não estoura 10) | `backend/internal/channel/service_test.go` |

> B2 depende do comportamento real confirmado em A4 — implementar A4 primeiro.

---

## 5. Bloco C — Checklist de campo (manual)

> Itens detalhados a partir de `plano-de-testes.md` §6 e `PROJECT_STATE.md`
> → *Problemas Conhecidos*. Marcar [x] ao concluir, com data.

- [ ] **C1 — RF03/RF06 (2 aparelhos, P1)**: PTT real no mesmo canal — latência
  percetível < 500ms (medir com cronômetro/gravador); derrubar a conexão de B no
  meio da fala → reconexão + `request_replay` **sem buraco perceptível**
- [ ] **C2 — RF03 (P1)**: enquanto A fala, B tenta falar → bloqueado; A fecha o
  app → canal liberado e B consegue falar (sem ficar preso "ocupado")
- [ ] **C3 — RF04 em aparelho Android (P1)**: GPS real + lifecycle — reporter de
  30s (`POST /api/v1/location`) pausa em background e retoma com tick imediato
  no foreground; `last_location_at` atualizado em `/admin/users`
- [ ] **C4 — Timeout/reconexão (P1)**: backend parado → entrar num canal exibe
  erro em ~10s (não spinner eterno); 5 tentativas de backoff (2/4/8/16/30s) →
  estado terminal "Tentar novamente"
- [ ] **C5 — Instância única (P1)**: app em background, tocar no ícone do
  launcher 3–4× → retoma na mesma tela, 1 entrada no recents; fluxo
  "Sair do Canal" (engrenagem → Sair) funciona
- [ ] **C6 — RF05 (P2)**: criar canal no celular B → aparece em A em ≤10s
  (polling); ida ao background→foreground atualiza a lista na hora
- [ ] **C7 — RNF06 (P2)**: URL nova do ngrok colada no sheet (Testar → Salvar)
  conecta **sem recompilar** o APK; URL sem `http(s)://` é recusada
- [ ] **C8 — RF09 ações manuais (P2)**: no `/admin` do navegador — suspender,
  reativar, promover, demover e excluir um usuário de teste (confirmação em
  cada ação); suspenso não consegue logar no app

---

## 6. Pendências de requisito/decisão (não são falha de teste)

Registrar a decisão nos documentos correspondentes quando forem tratadas —
**não** contam contra o gate de release:

1. **RF01 — Confirmação de e-mail**: critério de aceitação sem implementação na
   v1 (fora do código). Tratar como pendência de requisito em `docs/requisitos.md`.
2. ~~**Codec**: pacote declara `Codec: "opus"` mas o tráfego é PCM16.~~
   **Encerrado em 17/11/2026 (DEC-021)**: v1 declara e trafega PCM16 (backend
   já dizia `pcm16`; default do model Dart corrigido de `opus` para `pcm16`);
   Opus fica como trabalho futuro (RNF02).
3. **Token de suspenso**: segue válido até expirar (24h); efeito garantido no
   próximo login — revogação ativa fica para futuro (pendência aceita, RF09).
4. **RF05 — redirecionamento não ligado na API**: `JoinChannelOrCreate`
   (`backend/internal/channel/service.go:70`) não é chamado por nenhum
   handler — o fluxo real retorna `409 ChannelFull` e o app exibe "Canal está
   cheio" (nenhum redireciona para canal novo). O comportamento pretendido
   está coberto por teste unitário (A4/B2); **ligar o fluxo no handler/app é
   decisão de implementação pendente**.

---

## 7. Ordem de execução e gate

```
1. Bloco A (automatizado, P1 → P2):  A1 → A2 → A3 → A4 → A5
2. Bloco B (unitários, opcional):    B1, B2
3. Bloco C (campo, 2 aparelhos):     C1–C8
4. Regressão completa (§2)
5. Gate de release
```

**Gate de release (saída):**

1. `go test ./... -count=1` → 100% pass
2. `RADIOPX_TEST_DB=1 go test ./internal/db/...` → pass (quando PG disponível)
3. `flutter test` → 100% pass (baseline 62)
4. `scripts/test-backend.ps1` → **46/46** (29 atuais + A1/A2/A3 desde 17/11)
5. Bloco C sem itens **P1** pendentes (C1–C5 concluídos)
6. Nenhuma regressão conhecida aberta em `PROJECT_STATE.md` → *Problemas Conhecidos*

---

## 8. Registro de resultados

Preencher durante a execução.

| ID | Data | Executor | Resultado (✅/❌) | Observação |
|----|------|----------|-------------------|------------|
| A1 | 17/11/2026 | opencode | ✅ | Bloco 7 do script: stats/users/authz/self/delete (10 checks) |
| A2 | 17/11/2026 | opencode | ✅ | Bloco 3b do script: bloqueio do 2º falante + release no disconnect |
| A3 | 17/11/2026 | opencode | ✅ | Bloco 3c do script: 3 chunks + replay N+1/N+2 + sentinel |
| A4 | 17/11/2026 | opencode | ✅ | `channel/service_test.go`: redireciona p/ canal novo; ver §6 nº 4 (handler não ligado) |
| A5 | 17/11/2026 | opencode | ✅ | `db/privacy_integration_test.go` (`RADIOPX_TEST_DB=1`): 4 tabelas, sem colunas de áudio |
| B1 | 17/11/2026 | opencode | ✅ | `voice/service_test.go`: ErrChannelBusy, stop only-speaker, release |
| B2 | 17/11/2026 | opencode | ✅ | 25 joins concorrentes nunca estouram 10/canal |
| C1 | | | | |
| C2 | | | | |
| C3 | | | | |
| C4 | | | | |
| C5 | | | | |
| C6 | | | | |
| C7 | | | | |
| C8 | | | | |

---

## 9. Referências

- Plano completo (matriz de rastreabilidade, casos TC-*, segurança, carga): [`plano-de-testes.md`](plano-de-testes.md)
- Requisitos: [`requisitos.md`](requisitos.md)
- Lacunas e problemas conhecidos: [`../PROJECT_STATE.md`](../PROJECT_STATE.md)
- Decisões: [`decisoes.md`](decisoes.md)
- Script E2E: `scripts/test-backend.ps1`
- Histórico: [`../CHANGELOG.md`](../CHANGELOG.md)
