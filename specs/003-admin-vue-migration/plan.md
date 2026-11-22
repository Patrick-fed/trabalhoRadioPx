# Implementation Plan: Migração do Painel Administrativo para Vue 3

**Branch**: `feature/migracao-painel-adm-para-vue` (alias SpecKit para validação: `003-admin-vue-migration`) | **Date**: 2026-11-21 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/003-admin-vue-migration/spec.md`

## Summary

Substituir a SPA vanilla atual em `backend/internal/admin/web/` por uma aplicação Vue 3 construída com Vite, mantendo integralmente o comportamento observável do painel administrativo: login com token apenas em memória, dashboard com sete cards e mapa Leaflet, gestão de usuários com as mesmas ações/mensagens e entrega pelo próprio binário Go via `go:embed`/`SPAHandler`.

A abordagem escolhida mantém o backend intocado, usa Vue Router apenas para o limite login/casca autenticada (abas internas sem mudança de URL), Pinia para o estado em memória e direciona o build para o diretório já incorporado ao binário. Nenhum endpoint, regra de negócio ou dado persistido é alterado.

## Technical Context

**Language/Version**: JavaScript ES2022; Node 24.14.1 e npm 11.11.0 no ambiente de desenvolvimento; versões exatas de Vue/Vite fixadas em `admin-ui/package-lock.json` na implementação  
**Primary Dependencies**: Vue 3.x, Vue Router 4.x, Pinia 2.x, Leaflet 1.9.4, Vite, Vitest, `@vue/test-utils`  
**Storage**: PostgreSQL inalterado (somente usuários/canais); token e estado do painel apenas em memória, sem `localStorage`/`sessionStorage`/cookies  
**Testing**: Vitest + `@vue/test-utils` para a SPA; `go test ./...` e `scripts/test-backend.ps1` para regressão do backend (contractualmente inalterado)  
**Target Platform**: Navegadores modernos; SPA servida pelo backend Go em `/admin` no mesmo host da API  
**Project Type**: SPA web incorporada a um monolito modular Go; app Flutter permanece separado e fora de escopo  
**Performance Goals**: refresh do dashboard a cada 10s; `fitBounds` apenas na primeira renderização ou quando a quantidade de marcadores muda; toast com ~3,2s; nenhuma requisição adicional em trocas de aba  
**Constraints**: mesma origem da API (`API = ""`); `go:embed web/*`; Leaflet JS/CSS/ícones servidos localmente, sem CDN; nenhum novo endpoint; nenhuma alteração em Flutter, WebSocket de áudio, banco ou regras de admin  
**Scale/Scope**: 1 SPA administrativa; 2 estados de rota (login/casca), 2 abas internas, 6 endpoints REST consumidos, 7 cards, 1 mapa, 46 requisitos funcionais

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`specify` usa um arquivo de constituição com placeholders. Os gates abaixo foram derivados de `AGENTS.md`, `docs/requisitos.md`, `docs/decisoes.md` e da spec; a lacuna da constituição está registrada em [research.md](research.md).

| Gate | Status | Evidence |
|------|--------|----------|
| Monolito modular, sem microsserviços ou segundo servidor em produção | Pass | Build apenas gera estáticos para `backend/internal/admin/web/`; nenhum serviço novo |
| Backend intocado: endpoints, `RequireAdmin`, `is_admin` no PostgreSQL e regras de self-action | Pass | FR-044; contratos em [contracts/admin-rest.md](contracts/admin-rest.md) são somente consumidos |
| Privacidade: nenhum áudio persistido e nenhuma alteração em WebSocket/Flutter | Pass | FR-045; escopo exclusivamente frontend web |
| Token apenas em memória, sem persistência | Pass | FR-003/FR-011/FR-012; Pinia sem plugin de persistência |
| Mesma URL e entrega pelo `go:embed`/`SPAHandler` | Pass | FR-015/FR-043; `spa.go` sem alteração funcional |
| Leaflet local, sem CDN para JS/CSS/imagens | Pass | FR-025; `public/vendor/leaflet/images/` no build |
| Nenhuma funcionalidade nova ou melhoria de UX fora da troca de tecnologia | Pass | FR-001..FR-046 preservam rótulos, mensagens e comportamentos atuais |
| DEC-017: "sem projeto separado, sem nova dependência" | Pass with justification | Migração exige Vite/Vue/Vitest; registrada em Complexity Tracking e a ser formalizada como nova DEC antes da implementação |
| Testes escritos antes/durante a implementação | Pass | Vitest planejado para stores, API, router e componentes; `go test ./...` permanece |

**Post-design re-check**: todos os gates permanecem aprovados. A única tensão continua sendo DEC-017, tratada explicitamente como justificativa de complexidade e pendência de nova decisão arquitetural, não como falha de gate.

## Project Structure

### Documentation (this feature)

```text
specs/003-admin-vue-migration/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── admin-rest.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
admin-ui/
├── index.html                     # <base href="/admin/"> preservado
├── package.json
├── package-lock.json
├── vite.config.js                 # base '/admin/', outDir ../backend/internal/admin/web
├── public/
│   └── vendor/leaflet/images/     # ícones copiados do vendor atual
├── src/
│   ├── main.js
│   ├── App.vue
│   ├── router/
│   │   └── index.js
│   ├── stores/
│   │   └── admin.js
│   ├── services/
│   │   ├── api.js
│   │   └── adminApi.js
│   ├── views/
│   │   ├── LoginView.vue
│   │   └── AdminShell.vue
│   ├── components/
│   │   ├── AdminHeader.vue
│   │   ├── DashboardStats.vue
│   │   ├── AdminMap.vue
│   │   ├── UsersTable.vue
│   │   └── ToastMessage.vue
│   └── utils/
│       ├── format.js
│       └── escapeHtml.js
└── tests/
    ├── unit/
    │   ├── api.spec.js
    │   ├── adminStore.spec.js
    │   └── format.spec.js
    └── components/
        ├── LoginView.spec.js
        ├── UsersTable.spec.js
        └── AdminMap.spec.js

backend/internal/admin/web/       # saída do build, incorporada por go:embed
├── index.html                     # gerado
├── assets/                        # gerado
└── vendor/leaflet/images/         # copiado de admin-ui/public

backend/internal/admin/
├── handler.go                     # inalterado
├── handler_test.go                # inalterado
└── spa.go                         # sem alteração funcional
```

**Structure Decision**: criar `admin-ui/` na raiz para manter as fontes Vue separadas do diretório incorporado pelo Go e evitar que `node_modules` entre no `go:embed`. O build escreve diretamente em `backend/internal/admin/web/`, substituindo `index.html`, `app.js` e `style.css` atuais; os ícones do Leaflet são copiados de `admin-ui/public/vendor/leaflet/images/` a cada build. `frontend/` continua sendo exclusivamente o app Flutter.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Nova toolchain/build (Vite) e novo projeto `admin-ui/` contrariam DEC-017 ("sem projeto separado, sem nova dependência") | A stack Vue 3 foi exigida pelo solicitante; componentes, router e store precisam de transpilação/bundling para serem servidos como estáticos pelo `go:embed` | Continuar com JS vanilla atende DEC-017, mas não atende ao objetivo da feature; Flutter Web traria build separado e peso maior sem ganho para uso administrativo esporádico |
| Artefatos gerados precisam ser versionados em `backend/internal/admin/web/` | `go:embed` exige arquivos presentes no momento da compilação do binário | Servir diretório externo exigiria segundo servidor ou runtime de Node, quebrando o deploy de binário único e a DEC-003 |
