# Data Model: Detecção de Presença RadioPX via Bluetooth (BLE)

## RadioPxPresenceState

Enum interno, sem persistência/backend:

| Valor | Texto UI | Condição |
|---|---|---|
| `nearby` | `Alguém por perto` | >=1 observação RadioPX compatível com `lastSeenAt` nos últimos 15s |
| `away` | `Ninguém por perto` | pré-condições OK, advertise+scan ativos, sem observação compatível recente |
| `unavailable` | `Indisponível` | BT off, permissão negada, hardware/adapter indisponível, falha, plataforma não Android, ou app fora de foreground autenticado |

## RadioPxAdvertisement (local)

| Campo | Tipo | Regra |
|---|---|---|
| `serviceUuid` | `String` | UUID fixo do protocolo RadioPX |
| `instanceId` | `String` | efêmero, aleatório, gerado em memória; não associa a usuário/conta/dispositivo |
| `startedAt` | `DateTime` | somente memória |

## NearbyRadioPxObservation (remota)

| Campo | Tipo | Regra |
|---|---|---|
| `instanceId` | `String` | chave efêmera recebida; dedupe apenas em memória |
| `lastSeenAt` | `DateTime` | atualizado a cada anúncio compatível |
| `rssi` | `int?` | opcional para debug; não persistir/logar como dado pessoal |

Regras: TTL 15s; expirar/remove no tick; limpar em background/logout/dispose.

## BluetoothAdapterPrecondition

| Campo | Tipo | Regra |
|---|---|---|
| `isAndroid` | `bool` | fora Android => unavailable |
| `hasPermission` | `bool` | scan/advertise/connect conforme plataforma |
| `adapterReady` | `bool` | poweredOn/ready |
| `isForegroundAuthenticated` | `bool` | requerido para advertise+scan |

## Transições

- precondições OK + observação compatível recente => `nearby`.
- precondições OK + sem observação recente => `away`.
- permissão/BT/hardware/falha/plataforma inválida/background => `unavailable`.
- receber anúncio compatível: upsert `instanceId`, `lastSeenAt=now`.
- após 15s sem anúncio compatível: remover observação; recalcular.

## Persistência

Nenhuma. Sem DB, SharedPreferences, arquivo, backend, WebSocket, analytics ou logs de identificadores pessoais.