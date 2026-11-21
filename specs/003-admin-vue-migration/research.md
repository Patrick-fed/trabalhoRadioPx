# Research: Migração do Painel Administrativo para Vue 3

**Feature**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)  
**Date**: 2026-11-21

## Resumo

Não restaram dúvidas em aberto na spec ou no Technical Context. As pesquisas abaixo resolvem as decisões de implementação necessárias para o planejamento, sem introduzir funcionalidades novas.

## 1. Estrutura de build e integração com `go:embed`

**Decision**: Manter as fontes Vue em `admin-ui/` na raiz e configurar o Vite para gerar o bundle diretamente em `backend/internal/admin/web/`, com `base: '/admin/'` e `emptyOutDir: true`. Os artefatos gerados (`index.html`, `assets/`, `vendor/leaflet/images/`) devem ser versionados porque o `go:embed web/*` exige arquivos físicos no momento da compilação do Go.

**Rationale**: O handler atual usa `//go:embed web/*` e `fs.Sub(webFS, "web")`, sem qualquer código para servir arquivos de outro diretório. Colocar `node_modules` ou fontes dentro de `web/` faria o binário embutir dependências desnecessárias. Manter `admin-ui/` fora de `backend/internal/admin/web/` preserva o deploy de binário único (DEC-003) e evita alterar a lógica de `spa.go`.

**Alternatives considered**: (1) servir `admin-ui/dist` com um servidor estático separado — rejeitado por quebrar binário único e mesma origem; (2) colocar o projeto Vue dentro de `backend/internal/admin/web/` — rejeitado porque `go:embed` incluiria `node_modules`; (3) marcar a saída como ignorada pelo Git — rejeitado porque `go build` falharia em clone limpo sem o build prévio.

## 2. Modelo de navegação com Vue Router

**Decision**: Usar `createWebHistory('/admin/')` com duas rotas lógicas: `/login` (LoginView) e `/` (AdminShell). Transições de autenticação usam `router.replace` para não acumular histórico. Dashboard e Usuários são um `ref` local em `AdminShell`, sem rotas filhas e sem chamadas ao Router.

**Rationale**: Atende à opção C escolhida pelo solicitante: o Router controla apenas o limite login/casca autenticada, enquanto as abas internas preservam a URL `/admin/`. O `SPAHandler` já devolve `index.html` para caminhos desconhecidos sob `/admin/`, então um acesso direto a `/admin/login` também funciona. Um guard global verifica o token em memória e redireciona para `/login` quando ausente.

**Alternatives considered**: (1) rotas `/dashboard` e `/users` — rejeitado por alterar URLs e histórico; (2) rota única com renderização condicional — rejeitado por não usar o Router para o controle de autenticação como solicitado; (3) `createWebHashHistory` — rejeitado por introduzir `#` visível na URL.

## 3. Estado, token e ciclo de vida

**Decision**: Pinia com um único store `admin`, sem plugin de persistência. Estado inclui `token`, `me`, `users`, `stats`, `prevStats`, `activeTab`, `toast` e estado do mapa. O token é preenchido no login e limpo em logout, 401 e recarga da página.

**Rationale**: Reproduz as variáveis do IIFE atual (`token`, `me`, `usersCache`, `prevStats`, `mapMarkerCount`) com reatividade organizada, sem alterar a semântica de memória volátil. Nenhuma API de armazenamento do navegador é usada.

**Alternatives considered**: (1) `localStorage` para sobreviver a recarga — rejeitado por violar FR-003/FR-012 e a decisão de segurança atual; (2) provides/inject sem Pinia — rejeitado porque a stack exigida inclui Pinia; (3) múltiplos stores por domínio — rejeitado por complexidade desnecessária para uma SPA pequena.

## 4. Serviço de API e mensagens de erro

**Decision**: `services/api.js` centraliza `fetch` com `API = ""`, `Content-Type: application/json` em toda requisição e `Authorization: Bearer <token>` quando há sessão. Mapeamento de erros permanece idêntico: 401 → logout + "Sessão expirada. Entre novamente."; 429 → "Muitas tentativas. Aguarde um momento."; falha de rede → "Não foi possível conectar ao servidor"; demais casos → `error`/`message` do corpo ou `Erro <status>`.

**Rationale**: Preserva o comportamento observável atual e evita duplicar lógica em componentes. O login continua sem header de autorização e, após receber o token, consulta `GET /api/v1/admin/users` para localizar o admin por e-mail normalizado e validar `is_admin`.

**Alternatives considered**: (1) Axios — rejeitado por adicionar dependência sem necessidade; (2) tratamento de erro por componente — rejeitado por divergência de mensagens; (3) usar `GET /api/v1/user/profile` para obter `is_admin` — rejeitado nesta feature por ser uma melhoria fora de escopo, embora tecnicamente válido.

## 5. Leaflet em Vue e assets locais

**Decision**: Importar `leaflet` 1.9.4 e seu CSS pelo bundle; copiar os cinco arquivos atuais de `backend/internal/admin/web/vendor/leaflet/images/` para `admin-ui/public/vendor/leaflet/images/`. O componente `AdminMap.vue` usa `markRaw` para a instância do mapa, inicializa no `onMounted`, observa a lista de usuários, chama `invalidateSize` quando a aba volta a ficar visível e desmonta o mapa no `onUnmounted`. Os ícones padrão continuam apontando para `vendor/leaflet/images/...`, resolvidos pelo `<base href="/admin/">`. Tiles do OSM permanecem remotos.

**Rationale**: Mantém DEC-019 e FR-025: nenhum JS/CSS/imagem do mapa vem de CDN, inclusive atrás de túnel. O comportamento de `fitBounds` apenas quando a quantidade de marcadores muda evita que o refresh de 10s roube pan/zoom do administrador.

**Alternatives considered**: (1) CDN do Leaflet — rejeitado por DEC-019; (2) Google Maps — rejeitado por chave de API e termos de uso; (3) marcar ícones como dados do bundle sem `public/` — rejeitado porque as URLs relativas esperadas pelo CSS/vendor precisam permanecer estáveis em `/admin/vendor/leaflet/images/`.

## 6. Dashboard, deltas e timer

**Decision**: `DashboardStats.vue` dispara `Promise.all([adminApi.stats(), adminApi.users()])` ao montar e a cada 10.000 ms. Os deltas são calculados no store em relação ao snapshot anterior e exibidos somente para Usuários, Ativos, Canais e Localização. O timer é registrado no store e limpo no logout/desmonte.

**Rationale**: Preserva `startStats`/`stopStats`/`setDelta` do `app.js`, incluindo sinal "+" ou "−" e ocultação quando não há variação. Calcular no store mantém o componente apresentacional.

**Alternatives considered**: (1) `setInterval` dentro do componente sem limpeza — rejeitado por timers duplicados em trocas de aba; (2) polling separado para stats e usuários — rejeitado por aumentar requisições e quebrar a atualização atômica do mapa; (3) WebSocket para métricas — rejeitado por novo contrato e fora de escopo.

## 7. Tabela de usuários, busca e ações

**Decision**: `UsersTable.vue` renderiza as colunas atuais e recebe a lista do store. A busca é local, case-insensitive e com `trim`, sem nova chamada à API. As ações chamam `adminApi` com `PATCH {is_active}`, `PATCH {is_admin}` ou `DELETE`, usando `window.confirm` com os mesmos textos e exibindo os toasts atuais. A linha do próprio admin mostra "você" e nenhum botão.

**Rationale**: Mantém FR-031..FR-041 e a proteção de self-action já existente no backend. A renderização com interpolação do Vue já escapa conteúdo; o popup do Leaflet usa helper `escapeHtml` explícito porque recebe HTML string.

**Alternatives considered**: (1) busca no backend — rejeitado por exigir novo parâmetro de API; (2) modal customizado de confirmação — rejeitado por ser melhoria de UX fora de escopo; (3) `v-html` com dados da API — rejeitado por risco de XSS.

## 8. Estratégia de testes

**Decision**: Vitest + `@vue/test-utils` cobrindo: API (headers, 401/429/rede), store (login/logout/deltas/timer), router (redirect sem token e ausência de rotas de aba), `UsersTable` (filtro, self-action, ações) e `AdminMap` (quantidade de marcadores, `fitBounds` e `invalidateSize`). Regressão adicional: `go test ./...` e `scripts/test-backend.ps1` inalterados.

**Rationale**: A SPA não tem testes hoje; Vitest integra-se ao Vite e permite testar a lógica migrada sem navegador. O backend não muda, então seus testes existentes continuam como contrato de não regressão.

**Alternatives considered**: (1) Jest — exigiria configuração adicional com Vite; (2) Playwright/Cypress — custo de setup e navegador para uma migração sem mudança de fluxo; (3) apenas testes manuais — insuficiente para SC-002/SC-009.

## 9. Ambiente de desenvolvimento e proxy

**Decision**: `vite.config.js` define `base: '/admin/'`, `server.port = 5173` e proxy de `/api` para `http://localhost:8080`. A aplicação continua chamando caminhos absolutos `/api/...`; em produção, o mesmo host do Go atende a API e a SPA.

**Rationale**: Permite desenvolver a SPA com hot reload sem alterar `API = ""` nem configurar URLs por ambiente. O proxy existe apenas no dev server e não gera artefato nem variável de runtime.

**Alternatives considered**: (1) configurar `API` por env — rejeitado por mudar comportamento e adicionar configuração; (2) rodar sempre `go run` após cada build — rejeitado para iteração de UI, mantido apenas como smoke test; (3) servir Vite na raiz — rejeitado por divergir de `/admin/`.

## 10. Contrato do backend e estabilidade

**Decision**: Nenhum arquivo Go, SQL, Flutter ou WebSocket será alterado. Os contratos consumidos estão documentados em [contracts/admin-rest.md](contracts/admin-rest.md) e devem ser tratados como imutáveis nesta feature.

**Rationale**: FR-044/FR-045 e SC-010 exigem que a migração seja exclusivamente de tecnologia de frontend. O middleware `RequireAdmin` continua consultando `users.is_admin` a cada requisição, então promoção/demovação valem sem novo token.

**Alternatives considered**: (1) mover autorização para claims JWT — rejeitado por DEC-017; (2) adicionar endpoint de perfil admin — rejeitado por novo contrato; (3) alterar CORS para incluir `PATCH` — desnecessário porque a SPA é same-origin em produção e o proxy cobre o dev.

## 11. Lacuna da constituição e decisão arquitetural pendente

**Decision**: Como `.specify/memory/constitution.md` contém apenas placeholders, os gates do plano foram derivados de `AGENTS.md`, `docs/` e da spec. Antes da implementação, deve ser registrada uma nova decisão (DEC-020) autorizando build Vite, dependências Vue e versionamento dos artefatos embutidos, superando parcialmente a DEC-017.

**Rationale**: A spec já documenta o conflito com DEC-017. Registrar a decisão antes do código mantém rastreabilidade e evita que a mudança pareça silenciosa.

**Alternatives considered**: (1) editar DEC-017 retroativamente — rejeitado por apagar histórico; (2) seguir DEC-017 e cancelar a migração — rejeitado por contrariar o objetivo aprovado; (3) implementar sem registrar decisão — rejeitado pelas regras de documentação do `AGENTS.md`.
