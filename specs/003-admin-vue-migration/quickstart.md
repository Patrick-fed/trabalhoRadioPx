# Quickstart: Migração do Painel Administrativo para Vue 3

**Feature**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)  
**Date**: 2026-11-21

## Pré-requisitos

- Node 24.14.1 e npm 11.11.0 (ou versões compatíveis fixadas em `admin-ui/package-lock.json`).
- Go 1.21+ para compilar o backend (o ambiente local usa 1.26.0).
- BackendRadioPX configurado com `ADMIN_EMAILS` para login administrativo.
- PostgreSQL acessível com as credenciais em `backend/.env` para o script E2E.
- Nenhuma dependência de CDN para o painel; apenas os tiles do OpenStreetMap exigem internet.

> DEC-022 foi registrada em `docs/decisoes.md`: autoriza Vite/Vue, dependências e
> versionamento dos artefatos embutidos.

## Instalação

```powershell
cd admin-ui
npm install
```

O `package-lock.json` está versionado para reproduzir as versões exatas. Foram
observados warnings `EBADENGINE` em dependências transitivas e 11 avisos de
vulnerabilidade no `npm install`; não foram alterados porque ficam fora do escopo
desta migração.

## Desenvolvimento

Em um terminal, suba o backend:

```powershell
cd backend
go run cmd/server/main.go
```

Em outro terminal, suba a SPA:

```powershell
cd admin-ui
npm run dev
```

Acesse `http://localhost:5173/admin/`. O Vite faz proxy de `/api` para
`http://localhost:8080`, mantendo a aplicação com `API = ""`.

## Build para incorporação no Go

```powershell
cd admin-ui
npm run build
```

O build gera:

1. `backend/internal/admin/web/index.html` com `<base href="/admin/">`;
2. `backend/internal/admin/web/assets/` (JS/CSS com hash);
3. `backend/internal/admin/web/vendor/leaflet/images/`;
4. remoção dos arquivos vanilla antigos (`app.js`, `style.css`, JS/CSS antigos do
   vendor) do diretório de saída;
5. `backend/internal/admin/spa.go` sem alteração funcional (somente comentário atualizado).

Os arquivos gerados precisam ser versionados porque `//go:embed web/*` os exige
no momento da compilação.

## Executar com o binário Go

```powershell
cd backend
go build ./cmd/server
go run cmd/server/main.go
```

Verificações manuais rápidas:

- `http://localhost:8080/admin` redireciona para `/admin/`;
- `http://localhost:8080/admin/` exibe o login;
- login com usuário `is_admin=true` abre o dashboard;
- login com usuário comum mostra "Acesso negado: apenas administradores";
- alternar Dashboard/Usuários não altera a URL;
- os sete cards e o mapa correspondem aos dados da API;
- nenhum recurso do mapa é carregado de CDN (somente tiles OSM).

## Testes

```powershell
cd admin-ui
npm test
```

Resultado observado: **39 testes em 11 arquivos, todos passando**. Cobertura:

- `tests/unit/api.spec.js`: headers, 401, 429, falha de rede e mensagens;
- `tests/unit/adminApi.spec.js`: métodos, endpoints, headers e bodies;
- `tests/unit/adminStore.spec.js`: login, logout, limpeza em 401 e deltas;
- `tests/unit/adminActions.spec.js`: confirmações, mensagens e recarga;
- `tests/unit/router.spec.js`: apenas duas rotas, guard e base `/admin`;
- `tests/unit/format.spec.js`: datas `pt-BR`, coordenadas e deltas;
- `tests/components/LoginView.spec.js`: validação de admin e mensagens;
- `tests/components/AdminShell.spec.js`: troca de aba sem URL/histórico;
- `tests/components/UsersTable.spec.js`: filtro local, self-action, ações e FR-031 (primeira abertura após pré-carregamento do dashboard);
- `tests/components/DashboardStats.spec.js`: sete cards, subtítulos e deltas;
- `tests/components/AdminMap.spec.js`: marcadores, `fitBounds` e `invalidateSize`.

Regressão do backend:

```powershell
cd backend
go test ./...
```

Resultado observado: **todos os pacotes OK**.

```powershell
cd ..
powershell -File scripts\test-backend.ps1
```

Resultado observado em 21/11/2026: **não concluído** — o PostgreSQL local
recusou a autenticação (`SQLSTATE 28P01`) e o servidor não subiu; o script
reportou falha no health check. Reexecutar com `DB_*` válidos em `backend/.env`.

## Auditorias finais

- `backend/internal/admin/web/`: nenhuma referência a CDN para JS/CSS/imagens;
  URLs externas restantes são namespaces W3C, documentação Leaflet/Vue,
  tiles OSM, atribuição OSM e links "ver no OpenStreetMap".
- `admin-ui/src/stores/admin.js`: nenhum uso de `localStorage`,
  `sessionStorage`, cookies ou plugin de persistência.
- FR-031: `usersTabLoading` em `admin.js` e condição em `UsersTable.vue`;
  teste de regressão em `tests/components/UsersTable.spec.js`.
- Entrega validada com teste temporário de `SPAHandler`: `/admin` → 302 para
  `/admin/`, `/admin/` → HTML, `/admin/rota/inexistente` → fallback HTML,
  assets JS e ícone Leaflet → 200; `spa.go` sem alteração funcional (somente
  comentário).

## SC-001..SC-010 — resultado observado

| Critério | Resultado | Evidência |
|----------|-----------|-----------|
| SC-001 | PASS | abas/ações/login/cards/mapa presentes; suíte Vue 39/39 |
| SC-002 | PASS | `api.spec.js`, `LoginView.spec.js` e `adminActions.spec.js` cobrem mensagens e 401/429/rede |
| SC-003 | PASS | auditoria de `admin.js` e do bundle; token só em memória |
| SC-004 | PASS | `AdminShell.spec.js` + `router.spec.js`: base `/admin`, sem push/replace nas abas |
| SC-005 | PASS | `DashboardStats.spec.js`: sete cards, subtítulos, deltas e timer em `admin.js` |
| SC-006 | PASS | `AdminMap.spec.js`: contagem de marcadores, `fitBounds` só ao mudar quantidade e `invalidateSize` |
| SC-007 | PASS | `adminApi.spec.js` + `adminActions.spec.js` + `UsersTable.spec.js` |
| SC-008 | PASS | build incorporado, `SPAHandler` inalterado, validação de assets locais |
| SC-009 | PARCIAL | cenários unitários/e2e do painel passaram; `scripts/test-backend.ps1` bloqueado por PostgreSQL local |
| SC-010 | PASS | backend, Flutter, WebSocket, banco e contratos de API inalterados |

## Solução de problemas

| Sintoma | Causa provável | Ação |
|--------|----------------|------|
| `/admin/` exibe página em branco | build não foi executado ou `index.html` gerado está ausente | rodar `npm run build` e recompilar o Go |
| Ícones do mapa quebrados | `public/vendor/leaflet/images/` não foi copiado | conferir saída em `backend/internal/admin/web/vendor/leaflet/images/` |
| API chamada em `localhost:5173/api` falha | backend não está na 8080 ou proxy do Vite não configurado | subir o backend e conferir `vite.config.js` |
| Login cai em "Acesso negado" | e-mail não está em `ADMIN_EMAILS` ou `is_admin=false` | configurar bootstrap admin no backend |
| URL muda ao trocar de aba | rotas filhas foram adicionadas ao Router | manter abas como estado local em `AdminShell.vue` |
| `test-backend.ps1` falha no health check | PostgreSQL inacessível ou credenciais `DB_*` inválidas | corrigir `backend/.env` e reexecutar o script |
| Token aparece no DevTools Storage | persistência foi adicionada ao Pinia | remover plugin de persistência; token deve viver só em memória |
