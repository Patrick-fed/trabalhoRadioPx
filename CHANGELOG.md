# Histórico do Projeto RadioPX

Este documento registra todas as alterações significativas do projeto.

---

## 22/09/2026 - Correção: Joining de canal e voz entre 2 celulares na mesma conta

### Alterações

- `backend/internal/channel/handler.go`: join idempotente — reentrar em canal que já pertence agora retorna 200 em vez de 409 (corrige "Failed to join channel" ao retocar o canal ou usar a mesma conta em 2 celulares)
- `backend/pkg/websocket/hub.go`: broadcast de áudio não envia mais o pacote de volta ao próprio remetente (elimina o eco)
- `frontend/lib/features/channels/services/channel_service.dart`: exceções de join/leave agora incluem status HTTP e corpo da resposta para facilitar diagnóstico
- `frontend/lib/features/voice/screens/voice_screen.dart`: removida a supressão de eco por `user_id` (a supressão agora é feita no servidor) — permite que 2 celulares logados na mesma conta conversem entre si
- `scripts/test-backend.ps1`: valida que join repetido retorna 200 (idempotente) e que o remetente NÃO recebe o próprio eco
- APK release recompilado com as correções e copiado para `C:\RadioPx\RadioPX.apk`

### Motivo

Teste em campo com 2 celulares na mes ma conta falhava ao entrar no canal (409) e, se entrasse, cada aparelho silenciava o outro (supressão por user_id).

### Resultado

- Script de validação 28/28 verdes (incluindo idempotência e ausência de eco)
- `go test ./...` verde; `flutter test` 20/20; `flutter analyze` sem erros

---

## 22/09/2026 - Protótipo Funcional + APK Android

### Alterações

- Backend Go compilado e rodando (`server.exe`) sem depender de PostgreSQL (dados em memória)
- Adicionado carregamento automático do `.env` no `main.go` (JWT_SECRET/SERVER_PORT agora funcionam sem export manual)
- Corrigido JWT secret: middleware de autenticação usa o mesmo fallback do `main.go`
- Corrigidas rotas de canais: `/api/v1/channels/{id}/join` e `/api/v1/channels/{id}/leave`
- Corrigidos endpoints de perfil: `/api/v1/user/profile` e `/api/v1/user/change-password` agora alcançáveis
- Frontend: criado `lib/core/config.dart` com `API_BASE_URL` configurável via `--dart-define`
- Frontend: services (auth/channels/profile) usam a config central
- Frontend: `websocket_service.dart` corrigido (protocolo `room`, áudio em base64) e ligado na VoiceScreen
- Voz real de ponta a ponta: captura PCM16 via `record`, streaming de reprodução via `just_audio` (StreamAudioSource + WAV)
- PTT funcional: segurar grava e transmite via WebSocket, soltar para; supressão de eco próprio; reconexão automática
- VoiceScreen: sair do canal e ver membros funcionais
- Permissões Android adicionadas (INTERNET, RECORD_AUDIO, localização) + cleartext liberado
- Pasta `android/` gerada e APK release compilado (assinado com debug key = instalável)
- Testes Flutter 20/20 verdes; `flutter analyze` sem erros; backend `go test ./...` verde

### Resultado

- APK em `frontend/build/app/outputs/flutter-apk/app-release.apk` (cópia em `RadioPX.apk`)
- Backend em `:8080`. Celular acessa via IP LAN da máquina (ex.: `http://172.16.158.222:8080`)
- Observação: regra de firewall para porta 8080 exige execução como administrador

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
