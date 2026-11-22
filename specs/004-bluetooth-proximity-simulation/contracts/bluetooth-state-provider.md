# Contract: BluetoothStateProvider

Este é um contrato interno Dart. Não há contrato REST, WebSocket ou de banco nesta feature.

```dart
abstract class BluetoothStateProvider {
  Future<BluetoothProximityState> readState();
  Stream<BluetoothProximityState> watchState();
}
```

## Garantias

- `readState()` sempre completa com `nearby`, `away` ou `unavailable`; nunca lança para chamadores.
- `watchState()` nunca emite valor fora do enum; erros são tratados pelo controller como `unavailable`.
- Nenhuma implementação solicita permissão ao usuário.
- Nenhuma implementação liga/desliga o Bluetooth.
- Nenhuma implementação executa descoberta, advertise, pareamento, conexão, identificação, medição de distância, áudio, notificação ou ação automática.
- Em plataformas diferentes de Android, a factory retorna `UnsupportedBluetoothStateProvider` e o plugin não é chamado.

## Mapeamento do plugin Android

| Plugin (`BluetoothState`) | Estado do app |
|---|---|
| `on` | `nearby` |
| `off` | `away` |
| `uknow` | `unavailable` |
| valor não reconhecido | `unavailable` |
| `PlatformException` ou qualquer erro | `unavailable` |

## Uso

`MainScreen` aceita um `BluetoothStateProvider?` opcional para injeção em testes. Em produção, usa `createBluetoothStateProvider()`.
