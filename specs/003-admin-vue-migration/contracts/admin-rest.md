# Interface Contracts: Admin SPA ↔ Go Backend

**Feature**: [spec.md](../spec.md)  
**Plan**: [plan.md](../plan.md)  
**Date**: 2026-11-21

## Status

Todos os contratos abaixo já existem e são **imutáveis** nesta feature. A migração apenas troca o cliente HTML/JS por Vue 3; nenhum endpoint, campo, status ou regra de negócio é acrescentado.

## Convenções comuns

- Base: mesma origem da SPA (`API = ""`).
- Todo request envia `Content-Type: application/json`.
- Todo request autenticado envia `Authorization: Bearer <token>`.
- Exceção: `POST /api/v1/auth/login` não envia `Authorization`.
- Erros são lidos de `body.error` ou `body.message`; na ausência, usa-se `Erro <status>`.

### Mapeamento de erros observável

| Situação | Comportamento do cliente |
|----------|--------------------------|
| `401` | limpa sessão em memória, volta ao login, mostra "Sessão expirada. Entre novamente." |
| `429` | mostra "Muitas tentativas. Aguarde um momento." e mantém a sessão |
| Falha de rede / `fetch` rejeitado | mostra "Não foi possível conectar ao servidor" |
| `403` no fluxo de login | descarta token e mostra "Acesso negado: apenas administradores" |
| Outros `!ok` | mostra mensagem do corpo ou `Erro <status>` |

## `POST /api/v1/auth/login`

Autentica credenciais e retorna token JWT.

**Request**:
```json
{ "email": "admin@radiopx.local", "password": "senha" }
```

**Response `200`**:
```json
{ "token": "<jwt>" }
```

**Uso na SPA**: guarda `token` em Pinia (memória), depois chama `GET /api/v1/admin/users` para localizar o administrador por e-mail normalizado e confirmar `is_admin=true`.

**Erros**: `401` credenciais inválidas; `429` limite de tentativas.

## `GET /api/v1/admin/users`

Lista usuários sem `password_hash`.

**Request**: sem body.

**Response `200`** (array):
```json
[
  {
    "id": "uuid",
    "name": "Nome",
    "email": "admin@radiopx.local",
    "is_active": true,
    "is_admin": true,
    "last_lat": -23.5505,
    "last_lng": -46.6333,
    "last_location_at": "2026-11-21T12:34:56Z"
  }
]
```

**Uso na SPA**:
- popular a tabela e o cache local;
- calcular o card Localização;
- alimentar o mapa;
- localizar `me` após o login.

**Erros**: `401`, `403` quando o usuário não é admin.

## `GET /api/v1/admin/stats`

Retorna métricas do sistema.

**Request**: sem body.

**Response `200`**:
```json
{
  "users_total": 12,
  "users_active": 10,
  "channels_total": 4,
  "ws_connections": 8,
  "live_rooms": 3,
  "transmitting_channels": 1
}
```

**Uso na SPA**: preencher seis dos sete cards; o sétimo é derivado da contagem de usuários com posição conhecida em `GET /api/v1/admin/users`.

**Erros**: `401`, `403`.

## `PATCH /api/v1/admin/users/{id}`

Altera status ativo ou papel administrativo.

**Request — suspender/reativar**:
```json
{ "is_active": false }
```

**Request — promover/demover admin**:
```json
{ "is_admin": true }
```

**Response `200`**: usuário atualizado (ou payload equivalente aceito pelo cliente atual).

**Uso na SPA**: após sucesso, exibe o toast correspondente e recarrega usuários e dashboard.

**Regras preservadas**:
- o backend impede ação do admin sobre si mesmo;
- suspensão passa a valer no próximo login;
- alteração de `is_admin` vale imediatamente porque `RequireAdmin` consulta o banco.

**Erros**: `400` payload inválido; `401`; `403` self-action ou não-admin; `404`.

## `DELETE /api/v1/admin/users/{id}`

Exclui definitivamente um usuário.

**Request**: sem body.

**Response `200` / `204`**: sucesso sem conteúdo relevante para a SPA.

**Uso na SPA**: confirmação prévia, toast "Usuário excluído", recarga de usuários e dashboard.

**Regras preservadas**: backend impede excluir a si mesmo.

**Erros**: `401`, `403`, `404`.

## Contrato de entrega estática (`/admin`)

| Path | Comportamento esperado |
|------|------------------------|
| `GET /admin` | redirect `302` para `/admin/` |
| `GET /admin/` | serve `index.html` gerado pelo Vite |
| `GET /admin/assets/*` | serve bundles JS/CSS pelo `http.FileServer` do `SPAHandler` |
| `GET /admin/vendor/leaflet/images/*` | serve ícones locais copiados de `admin-ui/public/` |
| `GET /admin/<desconhecido>` | fallback para `index.html`, permitindo que o Router trate o caminho |

**Restrição de build**: todos os arquivos acima precisam existir fisicamente em `backend/internal/admin/web/` antes de `go build`, por causa de `//go:embed web/*`. O `SPAHandler` e o `index.html` devem manter `<base href="/admin/">` para que ícones relativos do Leaflet continuem resolvendo corretamente.

**Fora de escopo**: novos endpoints, alteração de CORS, alteração de status codes, alteração de mensagens, modificação funcional de `backend/internal/admin/handler.go` ou `backend/internal/admin/spa.go`.
