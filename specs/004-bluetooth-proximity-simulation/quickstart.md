# Quickstart: Presença RadioPX via Bluetooth (BLE)

## Pré-condições

- 2 dispositivos Android físicos com RadioPX instalado.
- Usuário autenticado em ambos, app em foreground.
- Bluetooth ligado e permissões concedidas (`BLUETOOTH_SCAN`, `BLUETOOTH_CONNECT`, `BLUETOOTH_ADVERTISE` no Android 12+).
- Android 11-: best-effort; localização do sistema pode ser necessária para scan, conforme comportamento do OS.

## Spike (antes de produção)

```bash
cd C:\Users\User\AppData\Local\Temp\opencode\radiopx_ble_spike
flutter analyze lib/main.dart
flutter build apk --debug
```

Validar com 2 aparelhos: iniciar advertise+scan, confirmar detecção apenas com UUID RadioPX, expiração 15s, parar ao sair/background.

## Implementação (após T005)

```bash
cd frontend
flutter pub get
flutter analyze
flutter test
flutter build apk --debug
```

## Manifesto mesclado

Conferir `frontend/build/app/intermediates/merged_manifest/debug/processDebugMainManifest/AndroidManifest.xml`:
- presente: `BLUETOOTH_SCAN` com `neverForLocation`, `BLUETOOTH_CONNECT`, `BLUETOOTH_ADVERTISE`.
- presentes legadas maxSdk30 quando aplicável.
- `bluetooth`/`bluetooth_le` `required=false`.
- sem permissões desnecessárias.

## Validação manual

1. A+B autenticados foreground, BT on, permissão ok: ambos "Alguém por perto".
2. BLE genérico próximo: permanece "Ninguém por perto".
3. Afastar/desligar B: A volta para "Ninguém por perto" após 15s.
4. Background/saída autenticada: advertise+scan param; observações limpas.
5. BT off/permissão negada: "Indisponível", app utilizável.
6. Navegar Canais/Falar/Perfil: indicador consistente.
7. Confirmar sem GATT/pairing/áudio/persistência/backend e sem regressão de canais/PTT/WebSocket/localização.

## Evidência

Registrar data, aparelhos, Android versions, resultado, logs/análise/build e limitações Android 11-.