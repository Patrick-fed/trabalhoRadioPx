# Implementation Plan: Detecção de Presença RadioPX via Bluetooth (BLE)

**Branch**: `bluetooth` (alias SpecKit: `004-bluetooth-proximity-simulation`) | **Date**: 24/11/2026 | **Spec**: [spec.md](./spec.md)
**Input**: Especificação revisada em `/specs/004-bluetooth-proximity-simulation/spec.md`

## Summary

Implementar presença RadioPX↔RadioPX por BLE com **advertise + scan simultâneos** em foreground autenticado. O indicador mantém três estados, mas "Alguém por perto" só ocorre para anúncios com Service UUID fixo do RadioPX e identificador efêmero aleatório em memória. BLE genérico é ignorado. Lifecycle: iniciar em área autenticada/foreground; parar e limpar em background, saída autenticada, dispose ou fim de sessão. Sem GATT/pareamento/áudio/persistência/backend.

## Technical Context

**Language/Version**: Dart 3.13.4 / Flutter 3.47.5 (`sdk: >=3.3.0 <4.0.0`); Android `minSdk 24`, `compileSdk/targetSdk 36`, Java 17.
**Primary Dependencies**: candidato provisório validado em spike: `flutter_bluetooth_plugin: 4.0.0` (MIT, versão exata). Adoção definitiva em produção **bloqueada** até T005 (validação em 2 Androids físicos). `bluetooth_manager: 2.0.1` será removido na implementação.
**Storage**: nenhuma; observações e identificadores efêmeros apenas em memória, TTL 15s.
**Testing**: `flutter test`, `flutter analyze`, `flutter build apk --debug/release`, inspeção de manifesto mesclado, validação manual com 2 Androids.
**Target Platform**: Android; iOS/Web/Windows retornam `Indisponível` sem chamadas nativas.
**Project Type**: app Flutter; backend Go inalterado.
**Performance Goals**: presença compatível refletida em até 3s; expiração em 15s sem anúncios compatíveis.
**Constraints**: sem pareamento, GATT, connect, áudio, identificação pessoal, persistência, backend, notificação; sem `enable()` automático de BT; sem alterar canais/PTT/WebSocket/áudio/localização/raio 10km/buffer 30s.
**Scale/Scope**: um indicador compartilhado; uma instância de controller/scanner por área autenticada.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Regras efetivas: `AGENTS.md`, `docs/requisitos.md`, `docs/decisoes.md`, `politicaDeDepencdencias.md`.

| Gate | Resultado |
|---|---|
| RF10/RNF07/DEC alinhados ao novo escopo | **Pendente** em docs (T014-T016) antes de implementação |
| Backend/DB/áudio/WebSocket inalterados | Passa |
| Dependência nova: necessidade, licença, manutenção, transitivas | **Pendente T005**; spike inicial ok (MIT, analyze/build ok) |
| Permissões mínimas e Android isolado | Passa no design; inclui `BLUETOOTH_ADVERTISE` por necessidade do escopo |
| Sem identificação/persistência | Passa por contrato |
| Testes + validação manual 2 dispositivos | **Bloqueado por hardware** |

Re-check pós-design: passa somente após T005 e atualização de docs/política.

## Project Structure

### Documentation (this feature)

```text
specs/004-bluetooth-proximity-simulation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── bluetooth-peer-scanner.md        # atualizar para presença RadioPX advertise+scan
│   └── bluetooth-state-provider.md      # legado/superseded
├── spec.md
└── tasks.md
```

### Source Code (repository root)

```text
frontend/
├── lib/
│   ├── app.dart
│   └── features/bluetooth/
│       ├── models/
│       │   ├── bluetooth_presence_state.dart
│       │   └── radiopx_ble_protocol.dart
│       ├── services/
│       │   ├── radiopx_presence_scanner.dart
│       │   ├── radiopx_presence_scanner_io.dart
│       │   ├── radiopx_presence_scanner_stub.dart
│       │   ├── radiopx_presence_scanner_factory.dart
│       │   └── radiopx_presence_controller.dart
│       └── widgets/bluetooth_proximity_indicator.dart
├── test/
│   ├── bluetooth_radiopx_protocol_test.dart
│   ├── bluetooth_presence_controller_test.dart
│   ├── bluetooth_presence_expiry_test.dart
│   ├── bluetooth_presence_lifecycle_test.dart
│   ├── bluetooth_presence_privacy_test.dart
│   ├── bluetooth_isolation_test.dart
│   └── main_screen_bluetooth_indicator_test.dart
├── android/app/src/main/AndroidManifest.xml
└── pubspec.yaml
```

**Structure Decision**: isolar BLE atrás de `RadioPxPresenceScanner` injetável; controller próprio para estado/TTL/lifecycle; factory IO/stub; nenhum impacto em backend.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| `BLUETOOTH_ADVERTISE` + `BLUETOOTH_SCAN/CONNECT` | Presença RadioPX exige anunciar e descobrir simultaneamente | Scan-only genérico não identifica RadioPX e gera falsos positivos |
| Troca de dependência BLE | Lib atual não faz advertise+scan; precisa API única comprovada | Duas libs BLE aumentam superfície/conflito; nativo próprio custa mais |
| Identificador efêmero em anúncio | Diferenciar instâncias/sessões sem identidade pessoal | MAC/ID estável violaria privacidade; sem ID dificultaria dedupe/TTL |
| Lifecycle foreground-only | Privacidade e bateria | Background scanning/advertising amplia exposição sem necessidade |
