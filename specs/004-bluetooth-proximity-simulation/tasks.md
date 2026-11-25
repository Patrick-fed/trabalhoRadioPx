---
description: "Task list for feature implementation"
---

# Tasks: Simulação de Proximidade via Bluetooth

**Input**: Design documents from `/specs/004-bluetooth-proximity-simulation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Test tasks are included because spec.md defines Independent Tests and acceptance scenarios, and plan.md includes `flutter test`, `flutter analyze`, and manual validation.
**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story the task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Flutter app**: `frontend/lib/`, `frontend/test/`, `frontend/android/`
- **SpecKit docs**: `specs/004-bluetooth-proximity-simulation/`, `docs/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add the approved dependency and create the feature structure.

- [X] T001 [P] Add `bluetooth_manager: 2.0.1` and raise minimum SDK to `>=3.3.0` in `frontend/pubspec.yaml`
- [X] T002 [P] Create feature directories `frontend/lib/features/bluetooth/models/`, `frontend/lib/features/bluetooth/services/` and `frontend/lib/features/bluetooth/widgets/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create the contract, enum, factory, and minimal Android permissions. These tasks BLOCK all user stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 [P] Create `BluetoothProximityState` enum with exact labels in `frontend/lib/features/bluetooth/models/bluetooth_proximity_state.dart`
- [X] T004 [P] Create `BluetoothStateProvider` contract and `UnsupportedBluetoothStateProvider` in `frontend/lib/features/bluetooth/services/bluetooth_state_provider.dart`
- [X] T005 Create stub factory returning unsupported provider in `frontend/lib/features/bluetooth/services/bluetooth_state_provider_stub.dart`
- [X] T006 [P] Configure minimal Bluetooth permissions and merge removals in `frontend/android/app/src/main/AndroidManifest.xml`
- [X] T007 Create conditional provider factory in `frontend/lib/features/bluetooth/services/bluetooth_state_provider_factory.dart`

**Checkpoint**: Foundation ready - user story implementation can now begin.

---

## Phase 3: User Story 1 - Indicador reflete o estado inicial (Priority: P1) MVP

**Goal**: Ao abrir o app autenticado, exibir o texto correto para Bluetooth ligado, desligado e indisponível/desconhecido.
**Independent Test**: Abrir o app com Bluetooth ligado, desligado e em falha/estado desconhecido; verificar os três textos exatos.

### Tests for User Story 1

> Escrever os testes primeiro e garantir que falhem antes da implementação.

- [X] T008 [P] [US1] Add initial-state and error-fallback controller tests in `frontend/test/bluetooth_proximity_controller_test.dart`
- [X] T009 [P] [US1] Add widget tests for exact labels in `frontend/test/bluetooth_proximity_indicator_test.dart`

### Implementation for User Story 1

- [X] T010 [P] [US1] Implement `readState()` mapping and error fallback in `frontend/lib/features/bluetooth/services/bluetooth_state_provider_io.dart`
- [X] T011 [P] [US1] Create indicator widget with label, color and icon in `frontend/lib/features/bluetooth/widgets/bluetooth_proximity_indicator.dart`
- [X] T012 [US1] Implement controller initial read, notifications and dispose guard in `frontend/lib/features/bluetooth/services/bluetooth_proximity_controller.dart`
- [X] T013 [US1] Wire provider, controller and indicator into authenticated `MainScreen` with optional provider injection in `frontend/lib/app.dart`

**Checkpoint**: User Story 1 fully functional and independently testable.

---

## Phase 4: User Story 2 - Atualização automática durante o uso (Priority: P1)

**Goal**: Mudar o indicador sem reiniciar quando o Bluetooth é alterado externamente.
**Independent Test**: Com o app aberto, alternar o Bluetooth e observar a troca do indicador sem reinício.

### Tests for User Story 2

- [X] T014 [P] [US2] Add stream update and error tests in `frontend/test/bluetooth_proximity_controller_test.dart`

### Implementation for User Story 2

- [X] T015 [US2] Implement stream subscription, latest-state-wins and cancellation in `frontend/lib/features/bluetooth/services/bluetooth_proximity_controller.dart`
- [X] T016 [US2] Implement `watchState()` stream mapping and error handling in `frontend/lib/features/bluetooth/services/bluetooth_state_provider_io.dart`

**Checkpoint**: User Stories 1 and 2 independently functional.

---

## Phase 5: User Story 3 - Indicador global na navegação principal (Priority: P2)

**Goal**: Manter o mesmo indicador visível nas abas Canais, Falar e Perfil.
**Independent Test**: Alternar entre as três abas e verificar que o indicador permanece visível e consistente.

### Tests for User Story 3

- [X] T017 [P] [US3] Add navigation persistence test for Canais, Falar and Perfil in `frontend/test/bluetooth_proximity_indicator_test.dart`

### Implementation for User Story 3

- [X] T018 [P] [US3] Add stable indicator key for navigation assertions in `frontend/lib/features/bluetooth/widgets/bluetooth_proximity_indicator.dart`
- [X] T019 [US3] Ensure indicator stays outside per-tab body above `Expanded` in `frontend/lib/app.dart`

**Checkpoint**: User Stories 1, 2 and 3 independently functional.

---

## Phase 6: User Story 4 - Simulação isolada das comunicações existentes (Priority: P2)

**Goal**: Garantir que a simulação não altere canais, PTT, WebSocket, áudio ou permissões adicionais.
**Independent Test**: Alternar o Bluetooth com o app conectado e verificar ausência de descoberta, conexão, notificação, persistência ou regressão de comunicação.

### Tests for User Story 4

- [X] T020 [P] [US4] Add source audit test forbidding enable/disable/scan/advertise/connect APIs in `frontend/test/bluetooth_isolation_test.dart`
- [X] T021 [US4] Add provider-call isolation test allowing only read/watch in `frontend/test/bluetooth_isolation_test.dart`

### Implementation for User Story 4

- [X] T022 [US4] Verify and keep Bluetooth feature isolated from API/WebSocket/audio in `frontend/lib/features/bluetooth/services/bluetooth_state_provider_io.dart` and `frontend/lib/app.dart`

**Checkpoint**: All user stories independently functional.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentação, dependências e validação final.

- [X] T023 [P] Update RF10/RNF07 traceability in `docs/requisitos.md`
- [X] T024 [P] Add DEC-023 in `docs/decisoes.md` and Bluetooth feature note in `docs/arquitetura.md`
- [X] T025 [P] Update dependency registration, `CHANGELOG.md`, `PROJECT_STATE.md` and `politicaDeDepencdencias.md`
- [X] T026 Run `flutter pub get`, `flutter analyze` and `flutter test` in `frontend/` and record results in `specs/004-bluetooth-proximity-simulation/quickstart.md`
- [X] T027 Run `flutter build apk --debug` and verify merged permissions in `frontend/build/app/intermediates/merged_manifests/debug/AndroidManifest.xml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup completion; BLOCKS all user stories
- **User Stories (Phase 3+)**: Depend on Foundational
- **Polish (Phase 7)**: Depends on all desired user stories

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational; no dependencies on other stories; MVP scope
- **User Story 2 (P1)**: Can start after Foundational; stream integration uses T010 and T012 from US1
- **User Story 3 (P2)**: Can start after Foundational; visibility integration uses T011 and T013 from US1
- **User Story 4 (P2)**: Can start after Foundational; isolation audits use T010 and T016

### Within Each User Story

- Tests included here must be written and fail before implementation
- Models and contracts before services
- Services before widget/app integration
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T001 and T002 can run in parallel
- T003, T004 and T006 can run in parallel within Foundational
- T008, T009, T010 and T011 can run in parallel within US1 after Foundational
- T014 can run in parallel with T015 after US1 controller exists
- T016 must run after T010 because both touch `bluetooth_state_provider_io.dart`
- T017 and T018 can run in parallel within US3
- T023, T024 and T025 can run in parallel within Polish
- Do not run tasks in parallel when they edit the same file

---

## Parallel Example: User Story 1

```bash
Task: "T008 Add initial-state and error-fallback controller tests in frontend/test/bluetooth_proximity_controller_test.dart"
Task: "T009 Add widget tests for exact labels in frontend/test/bluetooth_proximity_indicator_test.dart"
Task: "T010 Implement readState() mapping and error fallback in frontend/lib/features/bluetooth/services/bluetooth_state_provider_io.dart"
Task: "T011 Create indicator widget with label, color and icon in frontend/lib/features/bluetooth/widgets/bluetooth_proximity_indicator.dart"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. STOP and VALIDATE: test User Story 1 independently
5. Demo the initial-state indicator

### Incremental Delivery

1. Setup + Foundational
2. US1 - initial indicator
3. US2 - automatic updates
4. US3 - global navigation visibility
5. US4 - isolation guarantees
6. Polish - docs, dependency validation and APK manifest verification

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Then split US1, US2, US3 and US4 by owner where file conflicts allow
3. Merge through checkpoints in priority order

---

## Notes

- `[P]` tasks use different files and have no incomplete dependencies
- `[Story]` labels map tasks to user stories in `spec.md`
- Every task includes an exact file path
- Tests follow first-fail-then-implement ordering within each story
- Do not execute optional git hooks or commits automatically
- Stop at any checkpoint to validate the story independently
- Avoid vague tasks, same-file conflicts and cross-story dependencies that break independence
