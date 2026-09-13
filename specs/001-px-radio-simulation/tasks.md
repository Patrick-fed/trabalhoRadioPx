---

description: "Task list for RadioPX feature implementation"
---

# Tasks: Digital PX Radio Simulation

**Input**: Design documents from `/specs/001-px-radio-simulation/`
**Prerequisites**: spec.md (user stories with priorities), architecture docs

**Tests**: Tests are OPTIONAL - not explicitly requested in specification

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/internal/`, `backend/pkg/`, `backend/cmd/`
- **Frontend**: `frontend/lib/core/`, `frontend/lib/features/`
- **Database**: `backend/migrations/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module with dependencies in backend/go.mod
- [x] T002 [P] Configure Go linting and formatting tools
- [x] T003 [P] Configure Flutter project with pubspec.yaml dependencies
- [x] T004 [P] Setup PostgreSQL database and create initial migration in backend/migrations/001_initial.sql

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Setup database schema with users, channels, user_channels tables in backend/migrations/001_initial.sql
- [x] T006 [P] Implement authentication middleware in backend/internal/middleware/auth.go
- [x] T007 [P] Setup CORS middleware in backend/internal/middleware/cors.go
- [x] T008 [P] Implement WebSocket hub for connection management in backend/pkg/websocket/hub.go
- [x] T009 [P] Create WebSocket client handler in backend/pkg/websocket/client.go
- [x] T010 Setup API routing structure in backend/cmd/server/main.go
- [x] T011 Create base User model in backend/internal/user/repository.go
- [x] T012 Create base Channel model in backend/internal/channel/repository.go
- [x] T013 [P] Configure environment variables and configuration management
- [x] T014 [P] Setup error handling and logging infrastructure

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Push-to-Talk Communication (Priority: P1) 🎯 MVP

**Goal**: Users can communicate via push-to-talk within a channel

**Independent Test**: Two users in same channel can transmit and receive audio

### Implementation for User Story 1

- [x] T015 [P] [US1] Create AudioPacket model in backend/internal/voice/websocket.go
- [x] T016 [P] [US1] Create AudioPacket model in frontend/lib/features/voice/models/audio_packet.dart
- [x] T017 [US1] Implement Voice Service in backend/internal/voice/service.go
- [x] T018 [US1] Implement WebSocket handler for audio streaming in backend/internal/voice/handler.go
- [x] T019 [US1] Implement audio capture service in frontend/lib/features/voice/services/audio_service.dart
- [x] T020 [US1] Implement PTT service for push-to-talk in frontend/lib/features/voice/services/ptt_service.dart
- [x] T021 [US1] Create PTT button widget in frontend/lib/features/voice/widgets/ptt_button.dart
- [x] T022 [US1] Create audio visualizer widget in frontend/lib/features/voice/widgets/audio_visualizer.dart
- [x] T023 [US1] Create audio player widget in frontend/lib/features/voice/widgets/audio_player.dart
- [x] T024 [US1] Implement Voice screen in frontend/lib/features/voice/screens/voice_screen.dart
- [x] T025 [US1] Implement WebSocket service for audio in frontend/lib/core/websocket/websocket_service.dart
- [x] T026 [US1] Add microphone permission handling in frontend/lib/core/permissions/microphone_permission.dart
- [x] T027 [US1] Implement audio queue for sequential playback in frontend/lib/features/voice/services/audio_queue.dart

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Proximity-Based Discovery (Priority: P2)

**Goal**: Users discover channels based on GPS location within 10km radius

**Independent Test**: User can see channels within 10km and channels outside radius are filtered

### Implementation for User Story 2

- [x] T028 [P] [US2] Create Location model in backend/internal/location/model.go
- [x] T029 [P] [US2] Create Location model in frontend/lib/features/location/models/location_model.dart
- [x] T030 [US2] Implement Location Service in backend/internal/location/service.go
- [x] T031 [US2] Implement location handler for GPS updates in backend/internal/location/handler.go
- [x] T032 [US2] Implement location service for GPS in frontend/lib/features/location/services/location_service.dart
- [x] T033 [US2] Add location permission handling in frontend/lib/core/permissions/location_permission.dart
- [x] T034 [US2] Implement proximity calculation using Haversine formula in backend/internal/location/service.go
- [x] T035 [US2] Integrate location service with channel discovery

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Dynamic Channel Allocation (Priority: P3)

**Goal**: Channels are created automatically with max 10 users per channel

**Independent Test**: Adding 11th user creates new channel, user sees max 15 channels

### Implementation for User Story 3

- [x] T036 [P] [US3] Create Channel model in backend/internal/channel/model.go
- [x] T037 [P] [US3] Create Channel model in frontend/lib/features/channels/models/channel_model.dart
- [x] T038 [US3] Implement Channel Service in backend/internal/channel/service.go
- [x] T039 [US3] Implement channel handler for CRUD operations in backend/internal/channel/handler.go
- [x] T040 [US3] Implement channel service in frontend/lib/features/channels/services/channel_service.dart
- [x] T041 [US3] Implement Channels screen in frontend/lib/features/channels/screens/channels_screen.dart
- [x] T042 [US3] Implement Channel detail screen in frontend/lib/features/channels/screens/channel_detail_screen.dart
- [x] T043 [US3] Implement dynamic channel creation logic (10 users per channel)
- [x] T044 [US3] Implement channel limit (max 15 visible channels)

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: User Story 4 - Connection Resilience (Priority: P4)

**Goal**: Audio messages are buffered temporarily during connection drops

**Independent Test**: Simulate connection drop and verify replay of buffered messages

### Implementation for User Story 4

- [x] T045 [P] [US4] Create MessageBuffer model in backend/internal/voice/buffer_model.go
- [x] T046 [P] [US4] Create MessageBuffer model in frontend/lib/features/voice/models/message_buffer_model.dart
- [x] T047 [US4] Implement audio buffer service in backend/internal/voice/buffer_service.go
- [x] T048 [US4] Implement buffer manager for 30-second capacity in backend/internal/voice/buffer_manager.go
- [x] T049 [US4] Implement reconnection logic in frontend/lib/core/websocket/websocket_manager.dart
- [x] T050 [US4] Implement message replay on reconnection in frontend/lib/features/voice/services/message_replay_service.dart
- [x] T051 [US4] Add buffer cleanup after successful replay

**Checkpoint**: All user stories should now be independently functional

---

## Phase 7: Auth Feature (Supporting Feature)

**Goal**: User authentication via JWT and Google OAuth

**Independent Test**: User can register, login, and access protected routes

### Implementation for Auth Feature

- [x] T052 [P] Create User model in backend/internal/user/model.go
- [x] T053 [P] Create User model in frontend/lib/features/auth/models/user_model.dart
- [x] T054 Implement User Service in backend/internal/user/service.go
- [x] T055 Implement User handler for CRUD operations in backend/internal/user/handler.go
- [x] T056 Implement authentication service with JWT in backend/internal/auth/service.go
- [x] T057 Implement auth middleware for protected routes in backend/internal/auth/middleware.go
- [x] T058 Implement Login screen in frontend/lib/features/auth/screens/login_screen.dart
- [x] T059 Implement Register screen in frontend/lib/features/auth/screens/register_screen.dart
- [x] T060 Implement auth service in frontend/lib/features/auth/services/auth_service.dart
- [x] T061 Add password hashing with bcrypt

---

## Phase 8: Profile Feature (Supporting Feature)

**Goal**: User profile management

**Independent Test**: User can view and edit profile information

### Implementation for Profile Feature

- [x] T062 [P] Create Profile service in frontend/lib/features/profile/services/profile_service.dart
- [x] T063 Implement Profile screen in frontend/lib/features/profile/screens/profile_screen.dart
- [x] T064 Implement profile edit functionality
- [x] T065 Implement password change functionality

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T066 [P] Update documentation in docs/
- [x] T067 Code cleanup and refactoring
- [x] T068 Performance optimization across all stories
- [x] T069 Security hardening (JWT secret, HTTPS, input validation)
- [x] T070 Add rate limiting middleware in backend/internal/middleware/ratelimit.go
- [x] T071 Run end-to-end validation of all features
- [x] T072 Update CHANGELOG.md with implementation details
- [x] T073 Update PROJECT_STATE.md with current status

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Auth Feature (Phase 7)**: Can start after Foundational - needed for protected routes
- **Profile Feature (Phase 8)**: Depends on Auth Feature
- **Polish (Phase 9)**: Depends on all desired features being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Depends on US1 for audio buffer

### Within Each User Story

- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all models for User Story 1 together:
Task: "Create AudioPacket model in backend/internal/voice/websocket.go"
Task: "Create AudioPacket model in frontend/lib/features/voice/models/audio_packet.dart"

# Launch all widgets for User Story 1 together:
Task: "Create PTT button widget in frontend/lib/features/voice/widgets/ptt_button.dart"
Task: "Create audio visualizer widget in frontend/lib/features/voice/widgets/audio_visualizer.dart"
Task: "Create audio player widget in frontend/lib/features/voice/widgets/audio_player.dart"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Push-to-Talk)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Voice/PTT)
   - Developer B: User Story 2 (Location)
   - Developer C: User Story 3 (Channels)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Summary

| Phase | Tasks | Parallel [P] |
|-------|-------|--------------|
| Phase 1: Setup | 4 | 3 |
| Phase 2: Foundational | 10 | 6 |
| Phase 3: US1 (PTT) | 13 | 2 |
| Phase 4: US2 (Location) | 8 | 2 |
| Phase 5: US3 (Channels) | 9 | 2 |
| Phase 6: US4 (Buffer) | 7 | 2 |
| Phase 7: Auth | 10 | 2 |
| Phase 8: Profile | 4 | 1 |
| Phase 9: Polish | 8 | 1 |
| **Total** | **73** | **21** |

**Parallel Opportunities**: 21 tasks can run in parallel across different files

**Independent Test Criteria**:
- US1: Two users can transmit/receive audio via PTT
- US2: Channels within 10km are discovered, outside are filtered
- US3: 11th user creates new channel, max 15 channels visible
- US4: Connection drop triggers buffer, reconnection triggers replay
