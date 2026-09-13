# Requisitos do Sistema RadioPX

## Visão Geral

Este documento lista todos os requisitos funcionais e não funcionais do sistema RadioPX.

## Requisitos Funcionais

### RF01 - Criação de Conta

O sistema deve permitir que usuários criem contas com:
- Nome completo
- E-mail
- CPF
- Telefone
- Senha

**Critérios de Aceitação:**
- E-mail deve ser único no sistema
- CPF deve ser válido e único
- Senha deve ter no mínimo 8 caracteres
- Confirmação de e-mail enviada após cadastro

---

### RF02 - Login

O sistema deve autenticar usuários via:
- E-mail + Senha
- Google OAuth

**Critérios de Aceitação:**
- Token JWT gerado após autenticação válida
- Sessão expira após 24 horas
- Refresh token disponível para renovação

---

### RF03 - Push-to-Talk

O sistema deve permitir transmissão de áudio via botão PTT:
- Pressionar para falar
- Soltar para ouvir
- Indicação de canal ocupado

**Critérios de Aceitação:**
- Apenas um usuário pode falar por vez
- Transmissão é feita via WebSocket
- Latência de áudio < 500ms
- Indicação visual de canal ocupado

---

### RF04 - Descoberta por Proximidade

O sistema deve descobrir canais baseado em GPS:
- Raio de 10km
- Atualização automática de localização
- Filtragem por distância

**Critérios de Aceitação:**
- Localização atualizada a cada 30 segundos
- Canais fora do raio não são exibidos
- Precisão de 100m na detecção

---

### RF05 - Alocação Dinâmica de Canais

O sistema deve gerenciar canais automaticamente:
- Máximo 10 usuários por canal
- Criação automática de novos canais
- Máximo 15 canais visíveis

**Critérios de Aceitação:**
- Canal criado automaticamente ao atingir 10 usuários
- Usuário excedente redirecionado para novo canal
- Lista de canais limitada a 15 itens

---

### RF06 - Buffer de Conexão

O sistema deve armazenar áudio temporariamente:
- Buffer de 30 segundos
- Replay automático ao reconectar
- Descarte de mensagens antigas

**Critérios de Aceitação:**
- Buffer armazenado apenas em memória
- Mensagens anteriores a 30s são descartadas
- Replay automático ao reconectar

---

### RF07 - Gestão de Perfil

O sistema deve permitir:
- Visualizar perfil
- Editar informações
- Alterar senha

**Critérios de Aceitação:**
- Dados exibidos corretamente
- Edição validada antes de salvar
- Confirmação de senha atual para alteração

---

### RF08 - Fila de Áudio

O sistema deve ordenar áudios:
- Reproduzir um por vez
- Remover após reprodução
- Fila de prioridade

**Critérios de Aceitação:**
- Áudios reproduzidos em ordem de chegada
- Áudio reproduzido é removido da fila
- Fila limitada a 30 segundos de áudio

---

## Requisitos Não Funcionais

### RNF01 - Privacidade

- NENHUM áudio de conversa deve ser persistido
- Apenas dados de usuário e canal são salvos
- Conformidade com LGPD/GDPR

**Critérios de Verificação:**
- Auditoria de banco de dados não deve encontrar áudio
- Logs não devem conter conteúdo de áudio
- Política de retenção de dados clara

---

### RNF02 - Desempenho

- Latência de áudio < 500ms
- Alocação de canal < 5 segundos
- Detecção de proximidade com 100m de precisão

**Critérios de Verificação:**
- Testes de carga com 1000 usuários
- Monitoramento de latência em produção
- Benchmark de alocação de canais

---

### RNF03 - Escalabilidade

- Suporte a 10.000 usuários concorrentes por região
- Canais dinâmicos conforme demanda

**Critérios de Verificação:**
- Testes de estresse com múltiplos canais
- Escalabilidade horizontal do backend
- Balanceamento de carga

---

### RNF04 - Resiliência

- Buffer temporário para quedas de conexão
- Reconexão automática
- Replay de mensagens perdidas

**Critérios de Verificação:**
- Testes de desconexão simulada
- Verificação de replay automático
- Monitoramento de reconexão

---

### RNF05 - Segurança

- Autenticação JWT
- Criptografia em trânsito (TLS)
- Controle de acesso via Casbin

**Critérios de Verificação:**
- Testes de penetração
- Auditoria de autenticação
- Verificação de TLS em produção

---

## Rastreabilidade

### RF01 → Backend → Database → Frontend → Testes

- **Backend**: `internal/auth/`
- **Database**: `users` table
- **Frontend**: `features/auth/`
- **Testes**: unit, integration

### RF02 → Backend → Database → Frontend → Testes

- **Backend**: `internal/auth/`
- **Database**: `users` table
- **Frontend**: `features/auth/`
- **Testes**: unit, integration

### RF03 → Backend → Database → Frontend → Testes

- **Backend**: `internal/voice/`
- **Database**: N/A (sem persistência)
- **Frontend**: `features/voice/`
- **Testes**: unit, integration, e2e

### RF04 → Backend → Database → Frontend → Testes

- **Backend**: `internal/location/`
- **Database**: N/A (GPS do dispositivo)
- **Frontend**: `features/location/`
- **Testes**: unit, integration

### RF05 → Backend → Database → Frontend → Testes

- **Backend**: `internal/channel/`
- **Database**: `channels`, `user_channels` tables
- **Frontend**: `features/channels/`
- **Testes**: unit, integration

### RF06 → Backend → Database → Frontend → Testes

- **Backend**: `internal/voice/`
- **Database**: N/A (memória apenas)
- **Frontend**: `features/voice/`
- **Testes**: unit, integration

### RF07 → Backend → Database → Frontend → Testes

- **Backend**: `internal/user/`
- **Database**: `users` table
- **Frontend**: `features/profile/`
- **Testes**: unit, integration

### RF08 → Backend → Database → Frontend → Testes

- **Backend**: `internal/voice/`
- **Database**: N/A (memória apenas)
- **Frontend**: `features/voice/`
- **Testes**: unit, integration
