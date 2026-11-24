# Research: Detecção de Presença RadioPX via Bluetooth BLE

## 0. Revisão de escopo (24/11/2026)

**Decision**: substituir detecção BLE genérica por **presença RadioPX↔RadioPX** com **advertise + scan simultâneos** em foreground autenticado. Identidade do anúncio usa Service UUID fixo do protocolo RadioPX + identificador efêmero aleatório em memória, sem associação com usuário/conta/dispositivo. Android 11 ou inferior: best-effort com limitações documentadas.

**Rationale**: o usuário aprovou que apenas outro RadioPX aberto conta como presença. Isso elimina falsos positivos de BLE genérico e exige prova de viabilidade de advertise+scan na mesma sessão.

## 1. Spike de viabilidade (T001-T005)

**Status**: parcial, sem hardware físico.

| Item | Resultado |
|---|---|
| Workspace | `C:\Users\User\AppData\Local\Temp\opencode\radiopx_ble_spike` criado fora do repo |
| Candidato inicial | `flutter_bluetooth_plugin 4.0.0` (MIT, `yuanzhihong1/flutter_bluetooth_plugin`) |
| API | `startAdvertising`, `startScan(serviceUuids)`, `scanEvents`, `advertisingStateChanges`, `isPeripheralModeSupported`, `requestPermissions`, `getAdapterState` |
| POC | `lib/main.dart`: UUID fixo `0000c0de-0000-1000-8000-00805f9b34fb`, ID efêmero 8 bytes em memória, filtro por UUID, janela 15s, logs/timestamps, sem GATT/connect |
| `flutter analyze lib/main.dart` | **PASS** |
| `flutter build apk --debug` | **PASS** (`build/app/outputs/flutter-apk/app-debug.apk`) |
| Manifesto mesclado | contém `BLUETOOTH_SCAN neverForLocation`, `BLUETOOTH_CONNECT`, `BLUETOOTH_ADVERTISE`, legadas maxSdk30, `bluetooth/bluetooth_le required=false` |
| Dependências transitivas (pub deps) | `flutter_bluetooth_plugin`, `plugin_platform_interface`, `flutter_web_plugins`, `web`; SDK Flutter/Dart padrão |
| Licença | MIT (`LICENSE` no pacote) |
| Hardware 2 Androids | **BLOCKED** - `adb devices -l` vazio; sem comprovação física advertise+scan simultâneo |

**Risco conhecido**: plugin não tem publisher verificado no pub.dev; changelog ativo até 4.0.0 (2026-11-03). Necessária validação em 2 aparelhos antes de adoção definitiva em produção.

## 2. Decisão de dependência (provisória, pendente T005)

**Decision**: adotar provisoriamente `flutter_bluetooth_plugin: 4.0.0` para spike/design; confirmar após validação em dois Androids físicos.

**Rationale**: única dependência avaliada até aqui que cobre advertise+scan/permissões/adapter state na mesma API; licença MIT; manifesto compatível com least privilege; build/analyze passaram. `flutter_reactive_ble` 5.6.0 (decisão anterior) cobre scan, mas não advertise.

**Alternatives considered**:
- `flutter_reactive_ble` + `flutter_ble_peripheral`: split de responsabilidades, mas duas deps BLE, maior superfície e coordenação.
- `flutter_blue_plus`: licença comercial/restrições de uso comercial - evitar.
- MethodChannel/Kotlin próprio: controle total, maior custo e risco de manutenção.
- `bluetooth_manager` 2.0.1 (legado): só estado do adaptador, sem scan/advertise.

## 3. Permissões e plataforma

**Decision**: Android 12+ requer `BLUETOOTH_SCAN` (`neverForLocation`), `BLUETOOTH_CONNECT`, `BLUETOOTH_ADVERTISE`; Android <=11 usa permissões legadas/location conforme exigência do sistema. Solicitar runtime antes de iniciar advertise/scan. Não solicitar ativação automática do Bluetooth.

**Rationale**: advertise+scan simultâneos exigem `BLUETOOTH_ADVERTISE` além de scan/connect. `neverForLocation` preserva privacidade por não derivar localização.

**Android 11- best-effort**: scan/advertise podem depender de estado do sistema/localização; documentar como limitação operacional, não promessa de paridade.

## 4. Protocolo de presença RadioPX

**Decision**: anunciar e filtrar por Service UUID fixo RadioPX. Carregar apenas identificador efêmero aleatório gerado em memória por sessão/instância ativa. Não incluir userId, deviceId estável, MAC, nome pessoal, localização, áudio ou payload persistente.

**Rationale**: atende privacidade aprovada e evita contagem de BLE genérico. Observação expira em 15s sem anúncios compatíveis.

## 5. Lifecycle

**Decision**: iniciar advertise+scan apenas com usuário autenticado e app em primeiro plano. Em background, saída da área autenticada, fechamento da tela responsável ou fim da sessão, parar advertise+scan e limpar mapa de observações em memória.

**Rationale**: reduz superfície de privacidade, bateria e exposição fora do contexto de uso aprovado.

## 6. Fontes

- Android Developers — Bluetooth permissions / BLE advertise & scan.
- pub.dev / GitHub — `flutter_bluetooth_plugin` 4.0.0 (MIT, API e changelog).
- pub.dev — `flutter_reactive_ble` 5.6.0 (scan-only para este requisito).
- Código local do spike em `C:\Users\User\AppData\Local\Temp\opencode\radiopx_ble_spike`.
