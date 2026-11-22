# Data Model: Simulação de Proximidade via Bluetooth

## BluetoothProximityState

Enum interno do app, sem persistência e sem vínculo com o backend:

| Valor | Texto exato | Origem |
|---|---|---|
| `nearby` | `Alguém por perto` | `BluetoothState.on` / `isEnabled() == true` |
| `away` | `Ninguém por perto` | `BluetoothState.off` / `isEnabled() == false` |
| `unavailable` | `Indisponível` | `BluetoothState.uknow`, adapters nulos, erro de plataforma, plataformas fora do Android |

## BluetoothStateProvider

Contrato injetável com duas operações:

- `Future<BluetoothProximityState> readState()` — retorna sempre um dos três estados; nunca lança para a UI.
- `Stream<BluetoothProximityState> watchState()` — emite estados mapeados; erros de stream são convertidos em `unavailable` pelo controller.

Implementações:

- `PluginBluetoothStateProvider`: usa `bluetooth_manager 2.0.1` somente em Android; nunca chama `enableBluetooth()`/`disableBluetooth()`.
- `UnsupportedBluetoothStateProvider`: usado em iOS/Web/Windows; retorna `unavailable` e stream vazia.

## BluetoothProximityController

- `ChangeNotifier` com estado inicial `unavailable`.
- Faz leitura inicial ao ser criado e assina o stream.
- Converte qualquer erro em `unavailable`.
- Mantém somente o último estado observado.
- Cancela a subscrição no `dispose()`.
- Não armazena em SharedPreferences, banco, arquivo, backend ou WebSocket.

## Regras de transição

- `on` → `nearby`.
- `off` → `away`.
- `uknow` → `unavailable`.
- exceção/erro → `unavailable`.
- mudança rápida: último evento recebido vence.
- estados transitórios do adaptador são ignorados pelo plugin; se chegarem como desconhecidos, viram `unavailable`.

## Persistência

Nenhuma. Nenhuma tabela, coluna, DTO REST ou payload WebSocket é alterado.
