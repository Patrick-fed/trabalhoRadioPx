# Data Model: Migração do Painel Administrativo para Vue 3

**Feature**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)  
**Date**: 2026-11-21

## Visão geral

O painel não possui banco próprio nem novos modelos persistidos. Todos os dados abaixo vivem no store Pinia durante a sessão da página e são descartados no logout, em 401 ou no recarregamento. O PostgreSQL continua sendo acessado exclusivamente pela API Go existente.

## AdminSession

Estado da sessão administrativa em memória.

| Field | Type | Requirement | Notes |
|-------|------|-------------|-------|
| `token` | `string \| null` | Obrigatório após login | Nunca persistido; enviado como `Bearer` nas APIs admin |
| `me.id` | `string (UUID)` | Obrigatório após login | Usado para identificar a própria linha na tabela |
| `me.email` | `string` | Obrigatório após login | E-mail normalizado usado para localizar o admin na lista |
| `me.isAdmin` | `boolean` | Deve ser `true` | Se `false`, login é negado com mensagem padrão |
| `status` | `'anonymous' \| 'authenticating' \| 'authenticated'` | Obrigatório | Controla rota e exibição de loading |

**Validation rules**:
- Login sem `is_admin=true` → descarta token e lança "Acesso negado: apenas administradores".
- 401 em qualquer requisição → `status = 'anonymous'`, token/me limpos, mensagem "Sessão expirada. Entre novamente.".
- Recarga da página sempre inicia como `anonymous`.
- Logout limpa `token`, `me`, `users`, `stats`, `prevStats`, `activeTab`, `toast` e o timer.

**State transitions**:

```text
anonymous --login OK (is_admin)--> authenticating --> authenticated
authenticated --401/logout/reload--> anonymous
authenticated --429/rede/erro API--> authenticated (mantém token, mostra aviso)
```

## AdminUser

Representação do usuário retornado por `GET /api/v1/admin/users`.

| Field | Type | Requirement | Notes |
|-------|------|-------------|-------|
| `id` | `string (UUID)` | Obrigatório | Chave das ações e da comparação de "você" |
| `name` | `string` | Opcional na UI | Exibido como "—" quando ausente |
| `email` | `string` | Obrigatório na UI | Usado também na busca local |
| `isActive` | `boolean` | Obrigatório | Badge ativo/suspenso; define Suspender/Reativar |
| `isAdmin` | `boolean` | Obrigatório | Badge admin/usuário; define Tornar admin/Remover admin |
| `lastLat` | `number \| null` | Opcional | Coordenada da última posição conhecida |
| `lastLng` | `number \| null` | Opcional | Coordenada da última posição conhecida |
| `lastLocationAt` | `string (ISO) \| null` | Opcional | Formatado como `pt-BR` na tabela e no popup |

**Validation rules**:
- `lastLat`/`lastLng` nulos → sem marcador no mapa e "—" na coluna Localização.
- Latitude/longitude são exibidas com `toFixed(5)`.
- `name`, `email` e quaisquer strings da API são escapadas antes de virarem HTML do popup.
- A linha com `id === me.id` mostra "você" e nenhum botão de ação.

**Relationships**:
- `DashboardStats` é derivado em parte da lista de `AdminUser` (contagem de usuários com posição conhecida).
- `MapMarker` é derivado de cada `AdminUser` com coordenadas válidas.

## DashboardStats

Snapshot exibido nos sete cards do dashboard.

| Field | Type | Requirement | UI |
|-------|------|-------------|----|
| `usersTotal` | `number` | Obrigatório | Card Usuários |
| `usersActive` | `number` | Obrigatório | Card Ativos |
| `channelsTotal` | `number` | Obrigatório | Card Canais |
| `wsConnections` | `number` | Obrigatório | Card Conexões WS |
| `liveRooms` | `number` | Obrigatório | Card Salas ativas |
| `transmittingChannels` | `number` | Obrigatório | Card Transmitindo |
| `locatedUsers` | `number` | Derivado | Card Localização = contagem de `lastLat`/`lastLng` não nulos |

**Validation rules**:
- Subtítulos seguem as regras atuais (suspensos, percentual, conectado/silêncio etc.).
- Deltas são calculados somente para `usersTotal`, `usersActive`, `channelsTotal` e `locatedUsers`.
- Delta oculto quando não há snapshot anterior ou quando o valor não mudou.
- `prevStats` é sobrescrito a cada leitura bem-sucedida.

## MapMarker

Derivação de `AdminUser` para o Leaflet.

| Field | Type | Requirement | Notes |
|-------|------|-------------|-------|
| `id` | `string` | Obrigatório | Mesmo `id` do usuário |
| `position` | `[number, number]` | Obrigatório | `[lastLat, lastLng]` |
| `isActive` | `boolean` | Obrigatório | Classe `map-marker-active` ou `map-marker-inactive` |
| `popup` | `string` | Obrigatório | Nome, e-mail, status, data/hora e link OSM, com escape |

**Validation rules**:
- Criado apenas quando `lastLat != null && lastLng != null`.
- `fitBounds` executado na primeira renderização ou quando `markers.length` difere da contagem anterior.
- `invalidateSize` chamado ao voltar para a aba Dashboard.
- Instância do mapa e do layer group ficam fora do estado reativo (`markRaw`).

## ToastMessage

Aviso temporário exibido após ações e erros.

| Field | Type | Requirement | Notes |
|-------|------|-------------|-------|
| `text` | `string` | Obrigatório | Mensagem exata já existente |
| `kind` | `'ok' \| 'err'` | Opcional | Padrão `ok` |
| `visible` | `boolean` | Obrigatório | Controlado por timer de ~3200 ms |

**Validation rules**:
- Um toast por vez; nova mensagem substitui e reinicia o timer.
- Mensagens de sucesso preservadas: "Usuário suspenso", "Usuário reativado", "Usuário promovido a admin", "Privilégios de admin removidos", "Usuário excluído".
- Mensagens de erro vêm da camada de API.

## UiTabState

Estado local da casca autenticada; não é rota do Vue Router.

| Field | Type | Requirement | Notes |
|-------|------|-------------|-------|
| `activeTab` | `'dashboard' \| 'users'` | Obrigatório | Inicia em `dashboard` após login |
| `userSearch` | `string` | Obrigatório | Filtro local, sem persistência |

**Validation rules**:
- Troca de aba não altera URL nem histórico.
- Abrir Usuários dispara `refreshUsers`; abrir Dashboard dispara `refreshDashboard`.
- Busca usando `trim` e `toLowerCase` sobre `name` e `email`.
