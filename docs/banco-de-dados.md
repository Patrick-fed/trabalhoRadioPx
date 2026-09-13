# Banco de Dados RadioPX

## Tecnologia

PostgreSQL 14+

## Visão Geral

O banco de dados armazena apenas dados permanentes necessários para o funcionamento do sistema. Áudio de conversas e mensagens NÃO são persistidos (requisito crítico de privacidade).

## Dados Permanentemente Armazenados

### Tabela: users

Armazena informações dos usuários do sistema.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    phone VARCHAR(15),
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Campos:**
- `id`: Identificador único (UUID)
- `name`: Nome completo do usuário
- `email`: E-mail único no sistema
- `cpf`: CPF único (formato: XXX.XXX.XXX-XX)
- `phone`: Telefone (formato: (XX) XXXXX-XXXX)
- `password_hash`: Senha hasheada com bcrypt
- `created_at`: Data de criação
- `updated_at`: Data da última atualização

---

### Tabela: channels

Armazena informações dos canais de comunicação.

```sql
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    max_users INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Campos:**
- `id`: Identificador único (UUID)
- `name`: Nome do canal
- `max_users`: Máximo de usuários permitidos (padrão: 10)
- `created_at`: Data de criação

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

```sql
-- Índices para users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cpf ON users(cpf);
CREATE INDEX idx_users_phone ON users(phone);

-- Índices para user_channels
CREATE INDEX idx_user_channels_user ON user_channels(user_id);
CREATE INDEX idx_user_channels_channel ON user_channels(channel_id);
CREATE INDEX idx_user_channels_joined ON user_channels(joined_at);
```

## Migrações

### 001_initial.sql

```sql
-- Criação das tabelas iniciais
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de usuários
CREATE TABLE users (
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
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    max_users INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabela de relação usuário-canal
CREATE TABLE user_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, channel_id)
);

-- Índices
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cpf ON users(cpf);
CREATE INDEX idx_user_channels_user ON user_channels(user_id);
CREATE INDEX idx_user_channels_channel ON user_channels(channel_id);
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
2. **Cascade delete**: Ao deletar usuário, remove relações
3. **Unique constraints**: E-mail e CPF são únicos
4. **Timestamps**: created_at e updated_at automáticos

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
