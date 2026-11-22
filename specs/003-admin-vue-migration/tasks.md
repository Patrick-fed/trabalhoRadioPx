# Tasks: Migração do Painel Administrativo para Vue 3

**Input**: Design documents from `/specs/003-admin-vue-migration/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

## Phase 1: Setup (Shared Infrastructure)

- [X] T001 Create `admin-ui/` skeleton and `admin-ui/package.json` with vue, vue-router, pinia, leaflet@1.9.4, vite, vitest, @vue/test-utils; run `npm install` to generate `admin-ui/package-lock.json`
- [X] T002 [P] Configure `admin-ui/vite.config.js` with `base: '/admin/'`, `outDir: '../backend/internal/admin/web'`, `emptyOutDir: true`, dev proxy `/api` → `http://localhost:8080`, and Vitest environment
- [X] T003 [P] Create `admin-ui/index.html` with `<div id="app">` and `<base href="/admin/">`, preserving the legacy base behavior
- [X] T004 [P] Copy `backend/internal/admin/web/vendor/leaflet/images/*` into `admin-ui/public/vendor/leaflet/images/` so builds emit local icons
- [X] T005 [P] Create `admin-ui/src/main.js` and `admin-ui/src/App.vue` mounting the Vue Router
- [X] T006 [P] Add `dev`, `build`, and `test` scripts to `admin-ui/package.json`

## Phase 2: Foundational (Blocking Prerequisites)

- [X] T007 Implement `admin-ui/src/services/api.js` with same-origin `fetch`, JSON headers, Bearer token, and exact 401/429/network/error mapping
- [X] T008 Implement `admin-ui/src/services/adminApi.js` with login, users, stats, patch, and delete calls defined in `specs/003-admin-vue-migration/contracts/admin-rest.md`
- [X] T009 Implement `admin-ui/src/stores/admin.js` core state/actions in `specs/003-admin-vue-migration/data-model.md`, with no persistence plugin and timer cleanup
- [X] T010 Implement `admin-ui/src/router/index.js` with `/login`, `/`, auth guard, and `router.replace` for auth transitions
- [X] T011 [P] Implement `admin-ui/src/utils/format.js` (pt-BR date, `toFixed(5)`) and `admin-ui/src/utils/escapeHtml.js`
- [X] T012 [P] Port legacy styles to `admin-ui/src/assets/admin.css`, preserving badge, toast, and `map-marker-*` class names
- [X] T013 [P] Implement `admin-ui/src/components/ToastMessage.vue` with single-toast replacement and ~3200ms timer
- [X] T014 Register DEC-022 in `docs/decisoes.md` authorizing Vite/Vue deps, the `admin-ui/` source directory, and versioned generated artifacts before story work

**Checkpoint**: Foundation ready — user story work can begin.

## Phase 3: User Story 1 - Acesso administrativo preservado (P1) 🎯 MVP

**Goal**: Login admin, non-admin rejection, 401/logout/reload behavior with token only in memory.  
**Independent Test**: Login with `is_admin=true` and `is_admin=false`, force a 401, reload, and inspect browser storage.

- [X] T015 [P] [US1] Write `admin-ui/tests/unit/api.spec.js` covering headers, 401, 429, network failure, and exact messages
- [X] T016 [P] [US1] Write `admin-ui/tests/unit/adminStore.spec.js` covering login, logout, 401 cleanup, and absence of persistence
- [X] T017 [P] [US1] Write `admin-ui/tests/components/LoginView.spec.js` covering admin login, non-admin denial, and error display
- [X] T018 [US1] Implement `admin-ui/src/views/LoginView.vue` with E-mail/Senha/Entrar, error area, and store login call
- [X] T019 [US1] Wire auth guard and `router.replace` in `admin-ui/src/router/index.js` using `admin-ui/src/stores/admin.js`
- [X] T020 [US1] Create authenticated shell skeleton with header, e-mail, and Sair in `admin-ui/src/views/AdminShell.vue`
- [X] T021 [US1] Validate login/logout/401 against `admin-ui/src/services/api.js` and `backend/internal/admin/handler.go` contracts without changing Go files

**Checkpoint**: US1 independently functional and testable.

## Phase 4: User Story 2 - Gestão de usuários sem regressão (P2)

**Goal**: List, filter, and mutate users exactly as current panel.  
**Independent Test**: Load list, filter by name/e-mail, run each action on a non-self user, and compare requests/messages.  
**Dependency**: Integração depende da US1; o componente `UsersTable.vue` pode ser desenvolvido/testado isoladamente após a Fundação.

- [X] T022 [P] [US2] Write `admin-ui/tests/components/UsersTable.spec.js` covering columns, local search, self-action, and action buttons
- [X] T023 [P] [US2] Write `admin-ui/tests/unit/adminApi.spec.js` covering methods, endpoints, headers, and request bodies only; do not assert success messages here
- [X] T024 [P] [US2] Write `admin-ui/tests/unit/adminActions.spec.js` covering confirm texts, exact success messages, and users/dashboard refresh after suspend, reactivate, promote, demote, and delete
- [X] T025 [P] [US2] Implement `admin-ui/src/components/UsersTable.vue` with six columns, badges, location cell, search input, and Atualizar button
- [X] T026 [US2] Implement user actions in `admin-ui/src/stores/admin.js` with `window.confirm` texts, toasts, and refresh of users/dashboard, using the authenticated session from US1
- [X] T027 [US2] Integrate `admin-ui/src/components/UsersTable.vue` as an internal tab in `admin-ui/src/views/AdminShell.vue`, depending on the US1 authenticated shell and without router changes
- [X] T028 [US2] Validate self-action guard and exact strings using `admin-ui/src/components/UsersTable.vue` and `docs/requisitos.md` RF09

**Checkpoint**: US1 and US2 independently functional; US2 integration requires US1.

## Phase 5: User Story 3 - Dashboard e mapa idênticos (P3)

**Goal**: Seven cards, subtitles, deltas, 10s refresh, and Leaflet map identical to current behavior.  
**Independent Test**: Compare cards, deltas, and markers with the same database before and after migration.  
**Dependency**: Integração e refresh dependem da US1; `DashboardStats.vue`, `AdminMap.vue` e utilitários podem ser desenvolvidos/testados isoladamente após a Fundação.

- [X] T029 [P] [US3] Write `admin-ui/tests/unit/format.spec.js` covering pt-BR dates and five-decimal coordinates
- [X] T030 [P] [US3] Write `admin-ui/tests/components/DashboardStats.spec.js` covering seven cards, subtitles, and deltas
- [X] T031 [P] [US3] Write `admin-ui/tests/components/AdminMap.spec.js` covering marker count, `fitBounds` rule, and `invalidateSize`
- [X] T032 [P] [US3] Implement `admin-ui/src/components/DashboardStats.vue` with exact card labels, subtitles, delta badges, and the refresh hint text `Atualização automática a cada 10s. Variação desde a leitura anterior.` preserved from `backend/internal/admin/web/index.html:89`
- [X] T033 [P] [US3] Implement `admin-ui/src/components/AdminMap.vue` with Leaflet init, OSM tiles, local icons, popups, marker classes, and teardown
- [X] T034 [US3] Implement `refreshDashboard` action and 10s timer in `admin-ui/src/stores/admin.js` using `Promise.all` for stats/users, using the authenticated session from US1
- [X] T035 [US3] Integrate dashboard components into `admin-ui/src/views/AdminShell.vue`, depending on the US1 authenticated shell, and trigger `invalidateSize` when the tab becomes visible

**Checkpoint**: All user stories independently functional; US3 integration requires US1.

## Phase 6: User Story 4 - Entrega e navegação preservadas (P4)

**Goal**: Same `/admin/` URL, tabs without history, build embedded by existing `SPAHandler`.  
**Independent Test**: Access `/admin`, `/admin/`, unknown subpath, switch tabs, and verify served assets.  
**Dependency**: Validação final depende de US1, US2 e US3 integrados.

- [X] T036 [P] [US4] Write `admin-ui/tests/unit/router.spec.js` verifying only two routes, auth redirect, and no tab routes/history
- [X] T037 [US4] Finalize `admin-ui/vite.config.js` build output into `backend/internal/admin/web/` and verify `backend/internal/admin/web/index.html`
- [X] T038 [US4] Run `npm run build` in `admin-ui/` and `go build ./cmd/server` in `backend/`; verify generated `backend/internal/admin/web/assets/` and `backend/internal/admin/web/vendor/leaflet/images/` are embedded by functionally unchanged `backend/internal/admin/spa.go`
- [X] T039 [US4] Validate `/admin` redirect, `/admin/`, unknown subpath fallback, and local asset responses against `backend/internal/admin/spa.go`
- [X] T040 [US4] Validate tab switches keep URL `/admin/` and add no history entries in `admin-ui/src/views/AdminShell.vue`

**Checkpoint**: Migration deliverable complete with unchanged backend.

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T041 [P] Update `specs/003-admin-vue-migration/quickstart.md` with final build/test commands and `admin-ui/` paths
- [X] T042 [P] Update `docs/arquitetura.md` and `docs/requisitos.md` to describe the Vue admin SPA without changing RF09 semantics
- [ ] T043 Run `admin-ui/package.json` test script, `backend/go test ./...`, and `scripts/test-backend.ps1`
  - Status: `npm test` (39/39) e `go test ./...` OK; `scripts/test-backend.ps1` segue bloqueado por PostgreSQL local (`SQLSTATE 28P01`).
- [X] T044 Audit generated `backend/internal/admin/web/index.html` and `backend/internal/admin/web/assets/` for CDN references
- [X] T045 Audit `admin-ui/src/stores/admin.js` for `localStorage`, `sessionStorage`, or cookie usage
- [X] T046 Compare SC-001..SC-010 in `specs/003-admin-vue-migration/spec.md` with observed behavior and record result in `specs/003-admin-vue-migration/quickstart.md`
- [X] T047 [US2] Restore `Carregando…` on the first Users tab load after Dashboard preload (FR-031), with regression coverage in `admin-ui/tests/components/UsersTable.spec.js`

## Dependencies & Execution Order

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: depends on Setup; blocks all stories.
- **US1 (P1)**: after Foundational; MVP; no story dependency.
- **US2 (P2)**: after Foundational and US1 for authenticated integration; isolated component/test work may proceed in parallel after Foundational.
- **US3 (P3)**: after Foundational and US1 for authenticated refresh/integration; isolated component/test work may proceed in parallel after Foundational.
- **US4 (P4)**: after US1, US2 and US3 integration.
- **Polish**: after all desired stories.

**Parallel opportunities**: T002-T006; T011-T013; T015-T017; T022-T025; T029-T033; T036.

**Independent tests**: US1 login/401/storage; US2 users CRUD/filter/self-action; US3 cards/map/deltas; US4 `/admin` delivery/URL.

**Suggested MVP**: Phase 1 + Phase 2 + Phase 3 (US1 only).

**Format validation**: todos os 47 itens seguem `- [ ] T### [P]? [USn]? Descrição`, com caminho de arquivo explícito.
