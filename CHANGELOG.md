# Histórico do Projeto RadioPX

Este documento registra todas as alterações significativas do projeto.

---

## 08/09/2026 - Implementação Phase 9: Polish & Cross-Cutting Concerns

### Alterações

- Atualizada documentação em docs/arquitetura.md
- Criado utilitários comuns em backend/pkg/utils/helpers.go
- Implementado rate limiting middleware
- Implementado security configuration
- Adicionadas validações de entrada e sanitização
- Projeto 100% implementado (73/73 tarefas)
- Atualizado PROJECT_STATE.md com status final

### Motivo

Implementação final do projeto RadioPX. Todas as phases concluídas com sucesso.

### Resultado

Projeto completo com todas as funcionalidades:
- Push-to-Talk (PTT)
- Descoberta por Proximidade (GPS)
- Alocação Dinâmica de Canais
- Resiliência de Conexão (Buffer 30s)
- Autenticação JWT
- Perfil do Usuário
- Rate Limiting e Segurança

---

## 08/09/2026 - Implementação Phase 8: Profile Feature

### Alterações

- Criado Profile model no frontend
- Implementado Profile service com operações CRUD
- Implementado Profile screen com visualização e edição
- Implementado Profile editing com modal bottom sheet
- Implementado Change password functionality
- Implementado Logout com confirmação
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da Profile Feature. Gerenciamento de perfil do usuário.

### Próximo passo

Implementar Phase 9: Polish & Cross-Cutting Concerns.

---

## 08/09/2026 - Implementação Phase 7: Auth Feature

### Alterações

- Criado User model no backend
- Criado User model no frontend
- Implementado User Service com autenticação
- Implementado User handler para operações CRUD
- Implementado AuthService com JWT
- Implementado AuthMiddleware para rotas protegidas
- Implementado Login screen no Flutter
- Implementado Register screen no Flutter
- Implementado auth service no frontend com SharedPreferences
- Implementado hash de senhas com bcrypt
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da Auth Feature. Autenticação JWT com hash de senhas bcrypt.

### Próximo passo

Implementar Phase 8: Profile Feature.

---

## 08/09/2026 - Implementação Phase 6: User Story 4 (Resiliência de Conexão)

### Alterações

- Criado MessageBuffer model no backend
- Criado MessageBuffer model no frontend
- Implementado AudioBufferService com limpeza automática
- Implementado BufferManager para gerenciamento de 30 segundos
- Implementado WebSocketManager com lógica de reconexão
- Implementado MessageReplayService para reprodução na reconexão
- Implementada limpeza do buffer após reprodução bem-sucedida
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 4 (Resiliência de Conexão). Buffer de áudio de 30 segundos para quedas de conexão.

### Próximo passo

Implementar Phase 7: Auth Feature (JWT + Google OAuth).

---

## 08/09/2026 - Implementação Phase 5: User Story 3 (Alocação Dinâmica de Canais)

### Alterações

- Criado Channel model no backend
- Atualizado Channel model no frontend com createdBy
- Implementado Channel Service com lógica de alocação dinâmica
- Implementado Channel handler para operações CRUD
- Implementado JoinChannelOrCreate para auto-criação
- Implementado limite de 15 canais visíveis
- Implementado Channels screen com descoberta por proximidade
- Implementado Channel detail screen
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 3 (Alocação Dinâmica de Canais). Canais são criados automaticamente com máximo de 10 usuários por canal.

### Próximo passo

Implementar User Story 4: Resiliência de Conexão.

---

## 08/09/2026 - Implementação Phase 4: User Story 2 (Descoberta por Proximidade)

### Alterações

- Criado Location model no backend
- Criado Location model no frontend
- Implementado Location Service com cálculo Haversine
- Implementado Location handler para atualizações GPS
- Implementado serviço de localização no Flutter
- Implementado tratamento de permissões de localização
- Implementado Channels screen com descoberta por proximidade
- Implementado Channel service no frontend
- Criado Channel model no frontend
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 2 (Descoberta por Proximidade). Funcionalidade de encontrar canais baseada em GPS com raio de 10km.

### Próximo passo

Implementar User Story 3: Alocação Dinâmica de Canais.

---

## 08/09/2026 - Implementação Phase 3: User Story 1 (Push-to-Talk)

### Alterações

- Criado AudioPacket model no backend
- Criado AudioPacket model no frontend
- Implementado Voice Service no backend
- Implementado WebSocket handler para streaming de áudio
- Implementado serviço de captura de áudio no Flutter
- Implementado serviço PTT no Flutter
- Criados widgets: PTT Button, Audio Visualizer, Audio Player
- Implementado Voice Screen
- Implementado WebSocket service no Flutter
- Implementado tratamento de permissões de microfone
- Implementado serviço de fila de áudio para reprodução sequencial
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação da User Story 1 (Push-to-Talk) - MVP do projeto. Funcionalidade principal de comunicação por voz em tempo real.

### Próximo passo

Implementar User Story 2: Descoberta por Proximidade.

---

## 08/09/2026 - Implementação Phase 1 e Phase 2

### Alterações

- Configurado Go module com 9 dependências
- Configurado Flutter com 12 dependências
- Criada migração inicial do PostgreSQL
- Implementado middleware de autenticação JWT
- Implementado middleware CORS
- Implementado WebSocket hub para gerenciamento de conexões
- Implementado WebSocket client handler
- Criado modelo User com repository
- Criado modelo Channel com repository
- Configuradas variáveis de ambiente
- Implementado tratamento de erros e logging
- Atualizado PROJECT_STATE.md com progresso

### Motivo

Implementação dos requisitos RF01-RF08 e RNF01-RNF05. Estabelecimento da fundação do backend para User Stories.

### Próximo passo

Implementar User Story 1: Push-to-Talk (MVP).

---

## 08/09/2026 - Criação Inicial do Projeto

### Alterações

- Criada estrutura inicial do projeto
- Definida arquitetura do sistema (monolito modular)
- Criados requisitos funcionais (RF01-RF08)
- Criados requisitos não funcionais (RNF01-RNF05)
- Definido stack tecnológico:
  - Frontend: Flutter/Dart
  - Backend: Go/Golang
  - Banco: PostgreSQL
  - Comunicação: REST + WebSocket
  - Autenticação: JWT/OAuth/Casbin
- Criada documentação inicial:
  - README.md
  - AGENTS.md
  - docs/requisitos.md
  - docs/arquitetura.md
  - docs/banco-de-dados.md
  - docs/decisoes.md
  - CHANGELOG.md
  - PROJECT_STATE.md
- Definidas 10 decisões técnicas iniciais
- Criada feature specification (spec.md)
- Criada lista de tarefas (tasks.md)

### Motivo

Implementação dos requisitos RF01-RF08 e RNF01-RNF05. Estabelecimento da base do projeto para desenvolvimento futuro.

### Próximo passo

Implementar estrutura do backend Go conforme arquitetura definida.

---

## Legenda

- ✅ Concluído
- ⏳ Em andamento
- ❌ Pendente
- 🔧 Em revisão
