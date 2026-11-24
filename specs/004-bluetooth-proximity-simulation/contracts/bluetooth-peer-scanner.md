# Contract: RadioPxPresenceScanner

Contrato interno Dart (sem REST/WebSocket/DB).

```dart
abstract class RadioPxPresenceScanner {
  Future<bool> ensurePermissions();
  Stream<BluetoothAdapterPrecondition> watchPreconditions();
  Future<void> start({required RadioPxAdvertisement advertisement});
  Stream<NearbyRadioPxObservation> watchObservations();
  Future<void> stopAndClear();
  Future<void> dispose();
}
```

## Garantias

- Só inicia advertise+scan quando autenticado e foreground.
- Filtra por Service UUID fixo RadioPX; ignora BLE genérico.
- Não usa GATT, connect, pairing, áudio, backend, persistência.
- `instanceId` efêmero, aleatório, só memória.
- `stopAndClear()` para advertise+scan e limpa observações.
- Erros viram estado `Indisponível` no controller.

## Mapeamento

| Entrada | Estado |
|---|---|
| Precondição OK + observação compatível <=15s | `nearby` |
| Precondição OK + sem observação compatível | `away` |
| Permissão negada / BT off / erro / não Android / background | `unavailable` |