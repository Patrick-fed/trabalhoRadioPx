# Implementation Plan: Simulação de Proximidade via Bluetooth

**Branch**: `feature/migracao-painel-adm-para-vue` (alias SpecKit: `004-bluetooth-proximity-simulation`) | **Date**: 22/11/2026 | **Spec**: [spec.md](./spec.md)
**Input**: Especificação em `/specs/004-bluetooth-proximity-simulation/spec.md`

## Summary

Implementar um indicador visual global no app Flutter autenticado, exibindo "Alguém por perto", "Ninguém por perto" ou "Indisponível" a partir exclusivamente do estado ligado/desligado do adaptador Bluetooth Android. A leitura e o stream serão feitos por `bluetooth_manager 2.0.1` através de um `BluetoothStateProvider` injetável; o controller manterá apenas o último estado em memória, sem persistência, rede, permissões adicionais em runtime, descoberta, pareamento, conexão ou ações automáticas.

## Technical Context

**Language/Version**: Dart 3.x / Flutter (projeto usa `sdk: >=3.2.0 <4.0.0`; `bluetooth_manager 2.0.1` exige `>=3.3.0`)
**Primary Dependencies**: Flutter SDK, `provider: 6.1.1` (já declarado), `bluetooth_manager: 2.0.1` (nova dependência exata)
**Storage**: Nenhuma; estado apenas em memória
**Testing**: `flutter test`, `flutter analyze`, validação manual em Android
**Target Platform**: Android; iOS/Web/Windows fora de escopo e devem exibir "Indisponível" sem chamar o plugin
**Project Type**: aplicativo mobile Flutter com backend Go separado, inalterado nesta feature
**Performance Goals**: atualização do indicador em até 3 segundos após mudança externa do Bluetooth
**Constraints**: sem áudio persistido, sem WebSocket/API novas, sem permissões Nearby Devices, sem `enable()`/`disable()`, sem chamadas de scan/advertise/connect;raio de proximidade, canais e buffer permanecem inalterados
**Scale/Scope**: um indicador na área autenticada compartilhada pelas abas Canais, Falar e Perfil

## Constitution Check

| Gate | Resultado |
|---|---|
| Requisitos RF10/RNF07 rastreáveis | Passa |
| Monolito modular e backend inalterado | Passa |
| PostgreSQL e áudio não alterados | Passa |
| Sem nova permissão de runtime Nearby Devices | Passa |
| Android isolado; outras plataformas sem chamada nativa | Passa |
| Testes unitários, widget e validação manual previstos | Passa |
| Documentação e CHANGELOG atualizados | Passa |

## Project Structure

### Documentation (this feature)

```text
specs/004-bluetooth-proximity-simulation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── bluetooth-state-provider.md
├── spec.md
└── tasks.md (não criado nesta etapa)
```

### Source Code (repository root)

```text
frontend/
├── lib/
│   ├── app.dart
│   └── features/bluetooth/
│       ├── models/bluetooth_proximity_state.dart
│       ├── services/bluetooth_state_provider.dart
│       ├── services/bluetooth_state_provider_factory.dart
│       ├── services/bluetooth_state_provider_io.dart
│       ├── services/bluetooth_state_provider_stub.dart
│       ├── services/bluetooth_proximity_controller.dart
│       └── widgets/bluetooth_proximity_indicator.dart
├── test/
│   ├── bluetooth_proximity_controller_test.dart
│   └── bluetooth_proximity_indicator_test.dart
└── android/app/src/main/AndroidManifest.xml
```

**Structure Decision**: seguir a estrutura `features/<domínio>/` já usada por `auth`, `channels`, `voice` e `profile`. O adapter nativo fica isolado atrás de `BluetoothStateProvider`; `MainScreen` recebe o provider por parâmetro opcional para testes e usa a factory padrão em produção.

## Complexity Tracking

Nenhuma violação de constituição identificada.

Fonte exata e decisão de permissões registradas em [research.md](./research.md).Contrato interno em [contracts/bluetooth-state-provider.md](./contracts/bluetooth-state-provider.md).
