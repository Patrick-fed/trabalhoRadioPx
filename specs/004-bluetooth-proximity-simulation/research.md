# Research: Simulação de Proximidade via Bluetooth

## 1. Versão exata do plugin

- `bluetooth_manager 2.0.1`, publicado em `2026-05-29T14:03:36.382881Z`.
- Commit exato identificado: `9ad0adfe366cb3c907a49d2ca5c1d688158ba6ee` (`fix gradle namespace`).
- O commit `70859913b882b62d969848ccd5f9c8abbd785406` contém `2.0.0` e não é a versão publicada; foi usado apenas como verificação cruzada.
- Tags `2.0.1` no repositório retornam 404; a versão foi confirmada pelo `pubspec.yaml` no commit exato e pela API do pub.dev.
- `pubspec.yaml` 2.0.1: `sdk >=3.3.0 <4.0.0`, `flutter >=3.19.0`, `permission_handler: ^12.0.0`.
- `permission_handler` é dependência transitiva, mas nenhum arquivo Dart 2.0.1 importa ou chama `permission_handler`.

## 2. Caminho de código usado pela feature

Dart:

- `lib/bluetooth_manager.dart` → delega para `BluetoothManagerPath`.
- `lib/path/conf.dart` → `Platform.isAndroid` seleciona `BluetoothManagerAndroid`; plataformas não suportadas lançam erro.
- `lib/path/android/bluetooth_manager_android.dart`:
  - `MethodChannel('bluetooth_manager')` → `getBluetoothState`.
  - `EventChannel('bluetooth_manager/events')` → `receiveBroadcastStream()`.
  - resultado desconhecido/erro → `BluetoothState.uknow`.
- `lib/models/bluetooth_models.dart` → `BluetoothState { on, off, uknow }`.

Android (commit exato):

- `BluetoothManagerPlugin.java` (212 linhas) implementa `FlutterPlugin`, `MethodCallHandler`, `EventChannel.StreamHandler`.
- `getBluetoothState()`: exige API ≥ 19, chama `BluetoothAdapter.getDefaultAdapter()` e `adapter.isEnabled()`; sem adaptador ou exceção retorna `uknow`.
- `onListen()`: envia estado inicial via `events.success(getBluetoothState())` e registra `BroadcastReceiver` para `BluetoothAdapter.ACTION_STATE_CHANGED`.
- `onCancel()`: remove o receiver.
- Não há `checkSelfPermission`, `requestPermissions`, `ActivityCompat`, `ContextCompat`, `BLUETOOTH_*` em lógica de permissão, scan, advertise, discovery, pairing, conexão, áudio ou persistência.
- `enableBluetooth()`/`disableBluetooth()` existem no plugin, mas a feature não deve chamá-los.

## 3. Fundamentação Android para permissões mínimas

- Documentação oficial `bt-permissions` (target Android 12+):
  - `BLUETOOTH_SCAN`: apenas para procurar dispositivos Bluetooth.
  - `BLUETOOTH_ADVERTISE`: apenas para tornar o dispositivo detectável.
  - `BLUETOOTH_CONNECT`: apenas para comunicar com dispositivos já pareados.
  - permissões legados devem usar `android:maxSdkVersion="30"`.
- Referência oficial `BluetoothAdapter`:
  - `isEnabled()`: para apps com target R ou inferior exige `BLUETOOTH`; nenhuma exigência adicional em S+.
  - `ACTION_STATE_CHANGED`: para apps com target R ou inferior exige `BLUETOOTH`; nenhuma exigência adicional em S+.
  - `ACTION_CONNECTION_STATE_CHANGED`, `ACTION_DISCOVERY_*`, `ACTION_REQUEST_ENABLE` etc. são diferentes e têm exigências próprias; nenhum é usado.
- AOSP Android 14 `BluetoothAdapter.java`:
  - `getDefaultAdapter()`: `@RequiresNoPermission`.
  - `getState()`: `@RequiresLegacyBluetoothPermission` + `@RequiresNoPermission`.
  - `isEnabled()`: `@RequiresLegacyBluetoothPermission` + `@RequiresNoPermission`.
  - `ACTION_STATE_CHANGED`: `@RequiresLegacyBluetoothPermission`.
- AOSP Android 14 `core/res/AndroidManifest.xml`:
  - `android.bluetooth.adapter.action.STATE_CHANGED` é `<protected-broadcast>`, portanto receiver registrado em runtime recebe apenas broadcast do sistema e se enquadra na exceção de flags `RECEIVER_EXPORTED`/`RECEIVER_NOT_EXPORTED` do Android 14.
- Decisão: manter apenas `BLUETOOTH` com `maxSdkVersion="30"`; remover do manifesto mesclado `BLUETOOTH_ADMIN`, `BLUETOOTH_SCAN`, `BLUETOOTH_ADVERTISE` e `BLUETOOTH_CONNECT`.
- `ACCESS_FINE_LOCATION`/`ACCESS_COARSE_LOCATION` já existem para geolocalização RF04 e não são removidas; não são permissões adotadas por RF10.
- Manifestos de teste devem ser validados no merge final com `flutter build apk --debug`.

## 4. Alternativas rejeitadas

- Custom MethodChannel nativo: exigiria código Java próprio e manutenção de receiver; plugin pré-selecionado já cobre exatamente a leitura/stream.
- `permission_handler`: não é necessário; o plugin não solicita runtime permission e a feature não deve pedir Nearby Devices.
- Chamar `enableBluetooth()`/`disableBluetooth()`: viola RF007 e altera configuração do usuário.
- Usar `getProfileConnectionState()` ou `getBondedDevices()`: exigiria `BLUETOOTH_CONNECT` e sairia do escopo.
