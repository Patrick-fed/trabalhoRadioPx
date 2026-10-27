# Banco de Dados RadioPX

## Tecnologia

PostgreSQL 14+ (driver `jackc/pgx/v5` via `database/sql`; ver `backend/internal/db/`)

## Visão Geral

O banco de dados armazena apenas dados permanentes necessários para o funcionamento do sistema. Áudio de conversas e mensagens NÃO são persistidos (requisito crítico de privacidade). Schema aplicado pelas migrações em `backend/internal/db/migrations/` (estado atual: pós-`002_align_models.sql`).

## Dados Permanentemente Armazenados

### Tabela: users

Armazena informações dos usuários do sistema.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    cpf VARCHAR(14) UNIQUE,
    phone VARCHAR(15),
    password_hash VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Campos:**
- `id`: Identificador único (UUID)
- `name`: Nome completo do usuário
- `email`: E-mail único no sistema
- `cpf`: CPF único quando preenchido (opcional desde `002` — o app não coleta CPF)
- `phone`: Telefone (formato: (XX) XXXXX-XXXX)
- `password_hash`: Senha hasheada com bcrypt
- `provider`: Origem da conta (`local`; reservado para OAuth futuro)
- `provider_id`: ID no provedor externo (usado quando `provider ≠ local`)
- `is_active`: Flag de conta ativa
- `avatar`: Caminho/URL do avatar (vazio por padrão)
- `created_at`: Data de criação
- `updated_at`: Data da última atualização

---

### Tabela: channels

Armazena informações dos canais de comunicação.

```sql
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    max_users INTEGER DEFAULT 10,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Campos:**
- `id`: Identificador único (UUID)
- `name`: Nome do canal
- `description`: Descrição textual do canal
- `max_users`: Máximo de usuários permitidos (padrão: 10)
- `latitude`/`longitude`: Posição do canal para descoberta por proximidade (bounding-box + Haversine)
- `created_by`: Usuário criador (FK com `ON DELETE SET NULL`)
- `created_at`: Data de criação
- `updated_at`: Data da última atualização

---

### Tabela: user_channels

Armazena a relação entre usuários e canais.

```sql
CREATE TABLE user_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, channel_id)
);
```

**Campos:**
- `id`: Identificador único (UUID)
- `user_id`: Referência ao usuário
- `channel_id`: Referência ao canal
- `joined_at`: Data de entrada no canal
- `UNIQUE(user_id, channel_id)`: Um usuário só pode estar em um canal por vez

---

### Tabela: schema_migrations

Criada pelo runner de migrações (`backend/internal/db/migrate.go`) — não é parte do modelo de negócio.

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);
```

---

## Dados NÃO Persistidos (Crítico)

### ❌ Áudio

- Streaming em tempo real via WebSocket
- Armazenado apenas em memória durante transmissão
- NUNCA salvo em banco de dados
- Buffer temporário de 30 segundos

### ❌ Histórico de Mensagens

- Não há registro de conversas
- Buffer temporário apenas (30 segundos)
- Mensagens são descartadas após reprodução

### ❌ Gravações

- Nenhuma gravação é feita
- Áudio é descartado após reprodução
- Não há funcionalidade de gravação

---

## Índices

Criados pelas migrações (`001_initial.sql` + `002_align_models.sql`):

```sql
-- Índices para users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cpf ON users(cpf);

-- Índices para user_channels
CREATE INDEX idx_user_channels_user ON user_channels(user_id);
CREATE INDEX idx_user_channels_channel ON user_channels(channel_id);
CREATE INDEX idx_user_channels_joined ON user_channels(joined_at);

-- Índices para channels (002)
CREATE INDEX idx_channels_created ON channels(created_at);
CREATE INDEX idx_channels_location ON channels(latitude, longitude);
```

## Migrações

As migrações ficam em `backend/internal/db/migrations/` e são aplicadas por um runner embutido (`backend/internal/db/migrate.go`):

- `//go:embed migrations/*.sql` — migrações compiladas no binário
- cada migração roda em sua própria transação
- versões aplicadas registradas na tabela `schema_migrations`
- o servidor **falha rápido no boot** se o banco estiver inacessível (`log.Fatalf`)

### 001_initial.sql (idempotente)

```sql
-- Extensão para UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de usuários
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    phone VARCHAR(15),
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tabela de canais
CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    max_users INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabela de relação usuário-canal
CREATE TABLE IF NOT EXISTS user_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, channel_id)
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_cpf ON users(cpf);
CREATE INDEX IF NOT EXISTS idx_user_channels_user ON user_channels(user_id);
CREATE INDEX IF NOT EXISTS idx_user_channels_channel ON user_channels(channel_id);
CREATE INDEX IF NOT EXISTS idx_user_channels_joined ON user_channels(joined_at);
CREATE INDEX IF NOT EXISTS idx_channels_created ON channels(created_at);
```

### 002_align_models.sql (20/10/2026)

Alinha o schema aos modelos Go da Fase 1:

```sql
-- users: CPF opcional (app não coleta) + colunas para provider/conta
ALTER TABLE users ALTER COLUMN cpf DROP NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50) NOT NULL DEFAULT 'local';
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255) NOT NULL DEFAULT '';

-- channels: localização e metadados
ALTER TABLE channels ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Índice geográfico para bounding-box
CREATE INDEX IF NOT EXISTS idx_channels_location ON channels(latitude, longitude);
```

## Regras de Negócio

### Alocação de Canais

1. **Máximo 10 usuários por canal**: Quando um canal atinge 10 usuários, um novo canal é criado
2. **Máximo 15 canais visíveis**: Usuário só vê até 15 canais
3. **Raio de proximidade**: 10km para descoberta de canais
4. **Alocação dinâmica**: Canais são criados/conforme demanda

### Privacidade

1. **NENHUM áudio persistido**: Áudio existe apenas em memória
2. **Buffer temporário**: 30 segundos máximo
3. **Descarte automático**: Mensagens antigas são removidas
4. **Sem histórico**: Não há registro de conversas

### Integridade

1. **UUIDs**: Chaves primárias são UUIDs
2. **Cascade delete**: Ao deletar usuário/canal, remove relações em `user_channels`
3. **Unique constraints**: E-mail é único; CPF é único quando preenchido (opcional)
4. **Timestamps**: `created_at`/`updated_at` automáticos

## Consultas Frequentes

### Buscar canais do usuário

```sql
SELECT c.*
FROM channels c
JOIN user_channels uc ON c.id = uc.channel_id
WHERE uc.user_id = :user_id;
```

### Buscar usuários no canal

```sql
SELECT u.*
FROM users u
JOIN user_channels uc ON u.id = uc.user_id
WHERE uc.channel_id = :channel_id;
```

### Verificar se usuário está em canal

```sql
SELECT COUNT(*)
FROM user_channels
WHERE user_id = :user_id AND channel_id = :channel_id;
```

### Contar usuários no canal

```sql
SELECT COUNT(*)
FROM user_channels
WHERE channel_id = :channel_id;
```

### Buscar canais próximos (bounding-box, raio 10 km)

```sql
SELECT c.*, (
    6371 * acos(
        cos(radians(:lat)) * cos(radians(c.latitude)) *
        cos(radians(c.longitude) - radians(:lng)) +
        sin(radians(:lat)) * sin(radians(c.latitude))
    )
) AS distance
FROM channels c
WHERE c.latitude BETWEEN :lat - 0.09 AND :lat + 0.09
  AND c.longitude BETWEEN :lng - 0.09 AND :lng + 0.09
HAVING distance <= 10
ORDER BY distance;
```

> A distância final é recalculada em Go (`internal/location/distance.go`, Haversine); o filtro SQL usa bounding-box para reduzir candidatos.

## Backup e Recuperação

### Estratégia

- **Backup diário**: Completo às 02:00
- **Backup incremental**: A cada 6 horas
- **Retenção**: 30 dias
- **Teste de recuperação**: Semanal

### Comandos

```bash
# Backup completo
pg_dump -U postgres -d radiopx -F c -f backup_$(date +%Y%m%d).dump

# Restaurar
pg_restore -U postgres -d radiopx backup_20260908.dump
```

## Monitoramento

### Métricas

- Conexões ativas
- Queries por segundo
- Tamanho do banco
- Índices não utilizados
- Locks aguardando

### Alertas

- Conexões > 80% do máximo
- Queries lentas > 1s
- Espaço em disco > 80%
- Replicação atrasada
