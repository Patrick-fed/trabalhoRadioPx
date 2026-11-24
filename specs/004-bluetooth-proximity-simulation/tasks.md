# Tasks: Detecção de Presença RadioPX via Bluetooth BLE (Advertise + Scan)

**Input**: Design documents from `/specs/004-bluetooth-proximity-simulation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Scope revision (approved 2026-11-24)**: This task list supersedes the obsolete scan-only/generic-BLE draft. Target behavior is **RadioPX-to-RadioPX presence**: authenticated foreground app simultaneously advertises and scans BLE using a fixed RadioPX Service UUID plus an ephemeral random in-memory identifier (no user/account/device association). Generic BLE devices MUST NOT count as presence. No pairing, GATT, personal identification, persistence, backend, or changes to channels/PTT/WebSocket/audio. Android 11 or lower is best-effort with documented limitations. Production code MUST NOT change before dependency spike validation.

**Tests**: Included where needed for regression/safety and contract confidence; manual hardware validation is required because two physical Android devices are needed for the full advertise+scan proof.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently after the blocking spike/design phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (`[US1]`, `[US2]`, `[US3]`, `[US4]`)
- Every task includes an exact file path

## Phase 1: Setup (Spike & Evidence)

**Purpose**: Complete the approved dependency spike outside production code and capture evidence for the final library decision.

- [ ] T001 Create/refresh the Android BLE spike workspace with a minimal Flutter app in C:\Users\User\AppData\Local\Temp\opencode\radiopx_ble_spike/pubspec.yaml
- [ ] T002 Implement RadioPX advertise+scan POC with fixed Service UUID, ephemeral in-memory instance ID, 15s expiry, and debug logs in C:\Users\User\AppData\Local\Temp\opencode\radiopx_ble_spike\lib\main.dart
- [ ] T003 [P] Capture spike dependency/license/maintenance/transitive-dependency audit for evaluated candidates in specs/004-bluetooth-proximity-simulation/research.md
- [ ] T004 [P] Run spike static/build checks and record results (`flutter analyze`, `flutter build apk --debug`, merged manifest inspection) in specs/004-bluetooth-proximity-simulation/quickstart.md
- [ ] T005 Record two-device Android hardware validation evidence (foreground, BT on, permission granted, Android 11- best-effort case) in specs/004-bluetooth-proximity-simulation/quickstart.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Update design artifacts and dependency policy before any production implementation. **⚠️ CRITICAL**: No user story implementation starts before this phase completes.

- [ ] T006 Update feature scope, acceptance criteria, edge cases, and FR/RNF mapping for RadioPX advertise+scan in specs/004-bluetooth-proximity-simulation/spec.md
- [ ] T007 [P] Update technical context, constitution/gate check, structure, and dependency decision in specs/004-bluetooth-proximity-simulation/plan.md
- [ ] T008 [P] Document spike outcome, selected library, rejected alternatives, Android 11- limitations, and privacy rationale in specs/004-bluetooth-proximity-simulation/research.md
- [ ] T009 [P] Redefine ephemeral presence entities/state transitions for advertise+scan and RadioPX-only observations in specs/004-bluetooth-proximity-simulation/data-model.md
- [ ] T010 [P] Replace scan-only contract with RadioPX presence contract covering advertise+scan lifecycle and isolation guarantees in specs/004-bluetooth-proximity-simulation/contracts/bluetooth-peer-scanner.md
- [ ] T011 [P] Update or supersede legacy state-provider contract to reflect adapter state as precondition only in specs/004-bluetooth-proximity-simulation/contracts/bluetooth-state-provider.md
- [ ] T012 [P] Update manual validation steps for two RadioPX devices, permissions, merged manifest, and Android 11- limitations in specs/004-bluetooth-proximity-simulation/quickstart.md
- [ ] T013 Update dependency policy entry for the approved BLE library, exact version, license, and transitive audit in politicaDeDepencdencias.md
- [ ] T014 [P] Update RF10/RNF07 and traceability for RadioPX-to-RadioPX BLE presence in docs/requisitos.md
- [ ] T015 [P] Record DEC-022 (superseding DEC-021 scope) with privacy/scope/lifecycle decisions in docs/decisoes.md
- [ ] T016 [P] Update Bluetooth module architecture description for advertise+scan foreground presence in docs/arquitetura.md
- [ ] T017 Add the approved BLE dependency and remove obsolete Bluetooth dependency in frontend/pubspec.yaml
- [ ] T018 Update Android BLE permissions/feature declarations for advertise+scan while keeping least privilege in frontend/android/app/src/main/AndroidManifest.xml
- [ ] T019 Create RadioPX BLE protocol constants (fixed Service UUID, presence window, scan/advertise config) in frontend/lib/features/bluetooth/models/radiopx_ble_protocol.dart

**Checkpoint**: Design approved, docs/contracts updated, dependency registered, and production dependency manifest ready.

---

## Phase 3: User Story 1 - Detectar outro RadioPX próximo (Priority: P1) 🎯 MVP

**Goal**: When two authenticated RadioPX apps are open in foreground with Bluetooth/permission ready, they detect each other through RadioPX BLE advertisements and show the correct proximity indicator.

**Independent Test**: With two Android devices running authenticated RadioPX in foreground, device A shows "Alguém por perto" only when device B is advertising the RadioPX Service UUID; a generic BLE peripheral does not trigger nearby.

### Tests for User Story 1

- [ ] T020 [P] [US1] Add protocol/filter unit tests for RadioPX Service UUID matching and generic BLE rejection in frontend/test/bluetooth_radiopx_protocol_test.dart
- [ ] T021 [P] [US1] Add controller unit tests for nearby/away/unavailable mapping from advertise+scan observations in frontend/test/bluetooth_presence_controller_test.dart

### Implementation for User Story 1

- [ ] T022 [P] [US1] Create ephemeral presence/domain models for RadioPX observations and indicator state in frontend/lib/features/bluetooth/models/bluetooth_presence_state.dart
- [ ] T023 [P] [US1] Define `RadioPxPresenceScanner` interface for permission, adapter state, advertise, scan, and dispose lifecycle in frontend/lib/features/bluetooth/services/radiopx_presence_scanner.dart
- [ ] T024 [US1] Implement Android plugin-backed advertise+scan scanner filtered by RadioPX Service UUID in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_io.dart
- [ ] T025 [P] [US1] Implement unsupported stub returning unavailable without native calls in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_stub.dart
- [ ] T026 [P] [US1] Add conditional factory selecting IO/stub implementation in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_factory.dart
- [ ] T027 [US1] Implement `RadioPxPresenceController` to combine adapter/permission preconditions with filtered observations in frontend/lib/features/bluetooth/services/radiopx_presence_controller.dart
- [ ] T028 [US1] Update proximity indicator widget to consume new controller/state contract in frontend/lib/features/bluetooth/widgets/bluetooth_proximity_indicator.dart
- [ ] T029 [US1] Wire controller/scanner creation into authenticated main screen lifecycle in frontend/lib/app.dart

**Checkpoint**: US1 is demonstrable on two Android devices with RadioPX-only detection.

---

## Phase 4: User Story 2 - Atualização automática e expiração (Priority: P1)

**Goal**: The indicator updates automatically during use when compatible RadioPX presence appears/disappears, without restart.

**Independent Test**: With app open on both devices, bringing a RadioPX device into range updates to "Alguém por perto" within 3s; removing it returns to "Ninguém por perto" after 15s without new compatible advertisements.

### Tests for User Story 2

- [ ] T030 [P] [US2] Add timing tests for 3s appearance and 15s expiry/reconciliation behavior in frontend/test/bluetooth_presence_expiry_test.dart

### Implementation for User Story 2

- [ ] T031 [US2] Add observation dedupe/last-seen refresh and expiry timer reconciliation in frontend/lib/features/bluetooth/services/radiopx_presence_controller.dart
- [ ] T032 [US2] Ensure scan/advertise event streams update state idempotently and handle rapid changes in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_io.dart
- [ ] T033 [US2] Add UI-safe state transition handling to avoid flicker/regression in frontend/lib/features/bluetooth/widgets/bluetooth_proximity_indicator.dart

**Checkpoint**: US2 passes timed appearance/expiry validation independently.

---

## Phase 5: User Story 3 - Visibilidade global e lifecycle foreground-only (Priority: P2)

**Goal**: The indicator remains visible and consistent across authenticated tabs, and advertise/scan only run while RadioPX is authenticated and in foreground.

**Independent Test**: Navigate Canais/Falar/Perfil and observe same indicator state; backgrounding, leaving authenticated area, or closing responsible screen stops advertise/scan and clears in-memory observations.

### Tests for User Story 3

- [ ] T034 [P] [US3] Add widget/navigation tests proving indicator visibility and consistent state across Canais/Falar/Perfil in frontend/test/main_screen_bluetooth_indicator_test.dart
- [ ] T035 [P] [US3] Add lifecycle tests for foreground resume, background pause, and authenticated-exit cleanup in frontend/test/bluetooth_presence_lifecycle_test.dart

### Implementation for User Story 3

- [ ] T036 [US3] Integrate app lifecycle observer to start/stop advertise+scan on foreground/background transitions in frontend/lib/app.dart
- [ ] T037 [US3] Ensure controller dispose clears ephemeral observations and stops scanner resources in frontend/lib/features/bluetooth/services/radiopx_presence_controller.dart
- [ ] T038 [US3] Keep indicator placement shared in authenticated navigation without affecting tab behavior in frontend/lib/features/home/main_screen.dart

**Checkpoint**: US3 validates global visibility and no background scanning/advertising.

---

## Phase 6: User Story 4 - Isolamento, privacidade e compatibilidade (Priority: P2)

**Goal**: Presence detection is anonymous, local-only, and does not regress existing RadioPX communication or permissions behavior.

**Independent Test**: With channels/PTT/WebSocket/location flows active, Bluetooth presence changes cause no pairing/GATT/audio/API/WebSocket side effects; no identifiers are persisted; denied permission/BT off show "Indisponível".

### Tests for User Story 4

- [ ] T039 [P] [US4] Add isolation/regression tests ensuring no channel/WebSocket/audio/location coupling in frontend/test/bluetooth_isolation_test.dart
- [ ] T040 [P] [US4] Add privacy/cleanup tests proving no persistence or personal identifier exposure in frontend/test/bluetooth_presence_privacy_test.dart

### Implementation for User Story 4

- [ ] T041 [US4] Enforce anonymous ephemeral identifier generation and in-memory-only observation storage in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_io.dart
- [ ] T042 [US4] Map permission denied, Bluetooth off, unsupported hardware, and plugin errors to "Indisponível" without blocking app in frontend/lib/features/bluetooth/services/radiopx_presence_controller.dart
- [ ] T043 [US4] Add defensive guards/comments-free code paths preventing GATT/connect/persist/network calls in frontend/lib/features/bluetooth/services/radiopx_presence_scanner.dart

**Checkpoint**: US4 proves privacy/isolation and graceful unavailable behavior.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final documentation, validation, and cleanup across all stories.

- [ ] T044 [P] Update feature changelog entry with scope revision, dependency, and validation status in CHANGELOG.md
- [ ] T045 [P] Update project status/risks/blockers for BLE hardware validation and scope change in PROJECT_STATE.md
- [ ] T046 [P] Refresh obsolete legacy Bluetooth implementation references after migration in docs/arquitetura.md
- [ ] T047 Run static analysis and tests for touched frontend scope in frontend/pubspec.yaml
- [ ] T048 Run manual two-device Android quickstart validation and record final evidence in specs/004-bluetooth-proximity-simulation/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup/Spike)**: Starts immediately; produces evidence for library decision.
- **Phase 2 (Foundational)**: Depends on Phase 1 evidence; **blocks all user stories**.
- **Phase 3 (US1)**: Depends on Phase 2; MVP.
- **Phase 4 (US2)**: Depends on US1 controller/scanner baseline.
- **Phase 5 (US3)**: Depends on US1 integration; can proceed after T029.
- **Phase 6 (US4)**: Depends on US1/US2 behavior; can overlap with US3 after core controller exists.
- **Phase 7 (Polish)**: Depends on desired user stories complete.

### User Story Dependencies

- **US1 (P1)**: First deliverable after foundation - no dependency on other stories.
- **US2 (P1)**: Depends on US1 presence pipeline; independently testable for timing/expiry.
- **US3 (P2)**: Depends on US1 UI integration; independently testable for lifecycle/navigation.
- **US4 (P2)**: Depends on US1/US2 implementation details; independently testable for isolation/privacy/unavailable mapping.

### Within Each User Story

- Tests first where included, confirm they fail before implementation.
- Models/contracts before services.
- Services before UI wiring.
- Story checkpoint before moving on.

## Parallel Opportunities

- Phase 1: T003 and T004 can run in parallel after T002.
- Phase 2: T007-T012, T014-T016 can run in parallel after T006 defines scope.
- US1: T020/T021 parallel; T022/T023/T025/T026 parallel; then T024/T027/T028/T029.
- US2: T030 parallel with T031 prep; T032/T033 after controller contract stable.
- US3: T034/T035 parallel; T037/T038 parallel after T036 touchpoints identified.
- US4: T039/T040 parallel; T041-T043 mostly sequential in shared Bluetooth service/controller files.
- Polish: T044-T046 parallel.

## Parallel Example: User Story 1

```bash
Task: "T020 [US1] Add protocol/filter unit tests in frontend/test/bluetooth_radiopx_protocol_test.dart"
Task: "T021 [US1] Add controller unit tests in frontend/test/bluetooth_presence_controller_test.dart"

Task: "T022 [US1] Create presence/domain models in frontend/lib/features/bluetooth/models/bluetooth_presence_state.dart"
Task: "T023 [US1] Define RadioPxPresenceScanner interface in frontend/lib/features/bluetooth/services/radiopx_presence_scanner.dart"
Task: "T025 [US1] Implement unsupported stub in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_stub.dart"
Task: "T026 [US1] Add conditional factory in frontend/lib/features/bluetooth/services/radiopx_presence_scanner_factory.dart"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 spike evidence and library decision.
2. Complete Phase 2 design/docs/dependency updates.
3. Complete Phase 3 (US1).
4. **STOP and VALIDATE**: two physical Android devices, RadioPX-only detection, generic BLE ignored.
5. Demo MVP before timing/lifecycle/privacy hardening.

### Incremental Delivery

1. Spike + docs/contracts/dependency policy ready.
2. US1: basic RadioPX-to-RadioPX nearby detection.
3. US2: 3s appearance / 15s expiry stability.
4. US3: global indicator + foreground-only lifecycle.
5. US4: privacy/isolation/unavailable hardening and regression checks.

### Parallel Team Strategy

1. One developer owns spike evidence and dependency decision (Phase 1-2).
2. After Phase 2 checkpoint:
   - Developer A: US1 scanner/controller.
   - Developer B: US1/US3 UI + lifecycle integration.
   - Developer C: US2/US4 tests and hardening.
3. Rejoin for Phase 7 hardware validation and docs.

## Notes

- [P] tasks use different files and have no incomplete dependency.
- User story labels appear only in user story phases.
- Do not implement production code before T001-T005/T013 evidence/decision is complete.
- Keep RadioPX identifiers ephemeral, in-memory, and anonymous.
- Do not add pairing, GATT, backend, persistence, or audio/network coupling.
