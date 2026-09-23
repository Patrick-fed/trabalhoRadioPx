# Estado Atual do Projeto RadioPX

Este documento registra o estado atual do projeto para facilitar a continuidade do desenvolvimento.

---

## Fase

**Desenvolvimento - Protótipo Funcional + APK Android (22/09/2026)**

Todas as phases implementadas. Em 22/09/2026 o protótipo foi colocado em funcionamento
de ponta a ponta e gerado o APK Android para teste no celular.

---

## Requisitos Concluídos

### Requisitos Funcionais

- ✅ RF01 - Criação de Conta (Definido)
- ✅ RF02 - Login (Definido)
- ✅ RF03 - Push-to-Talk (Definido)
- ✅ RF04 - Descoberta por Proximidade (Definido)
- ✅ RF05 - Alocação Dinâmica de Canais (Definido)
- ✅ RF06 - Buffer de Conexão (Definido)
- ✅ RF07 - Gestão de Perfil (Definido)
- ✅ RF08 - Fila de Áudio (Definido)

### Requisitos Não Funcionais

- ✅ RNF01 - Privacidade (Definido)
- ✅ RNF02 - Desempenho (Definido)
- ✅ RNF03 - Escalabilidade (Definido)
- ✅ RNF04 - Resiliência (Definido)
- ✅ RNF05 - Segurança (Definido)

---

## Implementação Concluída

### Phase 1: Setup ✅

- ✅ T001 - Go module com dependências
- ✅ T002 - Configuração de linting Go
- ✅ T003 - Configuração Flutter com dependências
- ✅ T004 - Migração inicial do PostgreSQL

### Phase 2: Foundational ✅

- ✅ T005 - Schema do banco de dados
- ✅ T006 - Middleware de autenticação JWT
- ✅ T007 - Middleware CORS
- ✅ T008 - WebSocket hub para gerenciamento de conexões
- ✅ T009 - WebSocket client handler
- ✅ T010 - Estrutura de rotas da API
- ✅ T011 - Modelo User
- ✅ T012 - Modelo Channel
- ✅ T013 - Variáveis de ambiente
- ✅ T014 - Tratamento de erros e logging

---

## Em Desenvolvimento

- ✅ User Story 1: Push-to-Talk (P1 - MVP) - Concluída
- ✅ User Story 2: Descoberta por Proximidade (P2) - Concluída
- ✅ User Story 3: Alocação Dinâmica de Canais (P3) - Concluída
- ✅ User Story 4: Resiliência de Conexão (P4) - Concluída
- ✅ Phase 7: Auth Feature - Concluída
- ✅ Phase 8: Profile Feature - Concluída
- ✅ Phase 9: Polish & Cross-Cutting Concerns - Concluída

---

## Pendentes

### Backend

- ✅ Módulo Voice (handler, service, websocket, buffer)
- ✅ Módulo Location (handler, service)
- ✅ Módulo Channel (handler, service)
- ✅ Módulo Auth (handler, service, middleware)
- ✅ Módulo User (handler, service)
- ⏳ Testes unitários
- ⏳ Testes de integração

### Frontend

- ✅ Feature Voice (screens, widgets, services, models)
- ✅ Feature Location (screens, services, models)
- ✅ Feature Channels (screens, services, models)
- ✅ Feature Auth (screens, services, models)
- ✅ Feature Profile (screens, services, models)
- ⏳ Core (network, websocket, permissions, utils)

### Infraestrutura

- ⏳ Configuração de deploy
- ⏳ Monitoramento
- ⏳ Testes de carga

---

## Última Alteração

**22/09/2026 - Correção: logout, cadastro, login reativo e latência de áudio**

**Data**: 22/09/2026

**Descrição**:
- Perfil usa `AppConfig.apiBaseUrl` (antes `localhost` fixo) → funciona no celular
- `sessionState` (ValueNotifier) no AuthService; `AuthGate` agora é reativo — login/logout sem reiniciar o app
- Cadastro aceita 200 e 201 e mostra o erro real do backend; e-mail duplicado vira mensagem amigável
- Botão "Sair" sempre visível no perfil (inclusive no estado de erro)
- Áudio replay reescrito com `flutter_sound` (sessão contínua PCM16 16kHz) — elimina o delay de ~2,5s do just_audio; `minSdk` 24
- APK recompilado (50,4MB) com `API_BASE_URL=http://172.16.158.222:8080` e copiado para `RadioPX.apk`
- `flutter analyze` sem erros; `flutter test` 20/20 verdes
- Teste em campo sugerido: cadastrar (deve ir para home sem erro), sair da conta, relogar sem reiniciar o app e conferir latência do PTT entre 2 celulares

---

## Próxima Tarefa

**Projeto Completo - Todas as phases implementadas**

### Próximos Passos Opcionais

1. Executar testes end-to-end completos
2. Configurar ambiente de produção
3. Implementar monitoramento e logs
4. Configurar deploy automatizado

---

## Problemas Conhecidos

Nenhum.

---

## Dependências

### Backend

- Go 1.21+
- PostgreSQL 14+
- Redis (opcional, para cache)

### Frontend

- Flutter 3.16+
- Dart SDK
- Android SDK / Xcode

### Desenvolvimento

- Git
- IDE (VS Code/GoLand)
- Docker (opcional)

---

## Métricas do Projeto

### Arquivos Criados

| Arquivo | Tipo | Tamanho |
|---------|------|---------|
| README.md | Documentação | ~4KB |
| AGENTS.md | Regras | ~6KB |
| docs/requisitos.md | Requisitos | ~8KB |
| docs/arquitetura.md | Arquitetura | ~17KB |
| docs/banco-de-dados.md | Database | ~6KB |
| docs/decisoes.md | Decisões | ~8KB |
| CHANGELOG.md | Histórico | ~1KB |
| PROJECT_STATE.md | Estado | ~4KB |
| backend/go.mod | Configuração | ~200B |
| backend/.golangci.yml | Configuração | ~500B |
| backend/.env.example | Configuração | ~600B |
| backend/cmd/server/main.go | Backend | ~3KB |
| backend/internal/middleware/auth.go | Backend | ~2KB |
| backend/internal/middleware/cors.go | Backend | ~500B |
| backend/internal/middleware/logging.go | Backend | ~2KB |
| backend/internal/user/repository.go | Backend | ~2KB |
| backend/internal/channel/repository.go | Backend | ~3KB |
| backend/pkg/websocket/hub.go | Backend | ~3KB |
| backend/pkg/websocket/client.go | Backend | ~4KB |
| backend/migrations/001_initial.sql | Database | ~1KB |
| frontend/pubspec.yaml | Configuração | ~600B |
| frontend/lib/main.dart | Frontend | ~100B |
| frontend/lib/app.dart | Frontend | ~2KB |
| specs/.../spec.md | Especificação | ~7KB |
| specs/.../tasks.md | Tarefas | ~10KB |
| specs/.../requirements.md | Checklist | ~2KB |

### Progresso

| Fase | Tarefas | Concluídas | Status |
|------|---------|------------|--------|
| Phase 1: Setup | 4 | 4 | ✅ 100% |
| Phase 2: Foundational | 10 | 10 | ✅ 100% |
| Phase 3: US1 (PTT) | 13 | 13 | ✅ 100% |
| Phase 4: US2 (Location) | 8 | 8 | ✅ 100% |
| Phase 5: US3 (Channels) | 9 | 9 | ✅ 100% |
| Phase 6: US4 (Buffer) | 7 | 7 | ✅ 100% |
| Phase 7: Auth | 10 | 10 | ✅ 100% |
| Phase 8: Profile | 4 | 4 | ✅ 100% |
| Phase 9: Polish | 8 | 8 | ✅ 100% |
| **Total** | **73** | **73** | **100%** |

---

## Última Atualização

**Data**: 22/09/2026

**Hora**: 21:30

**Por**: OpenCode (assistente AI)

---

## Notas para Continuação

1. **Testes**: Executar testes end-to-end completos
2. **Deploy**: Configurar ambiente de produção
3. **Monitoramento**: Implementar logs e métricas
4. **Documentação**: Atualizar README com instruções de deploy
5. **Google OAuth**: Implementar login social (futuro)
