# Estado Atual do Projeto RadioPX

Este documento registra o estado atual do projeto para facilitar a continuidade do desenvolvimento.

---

## Fase

**Desenvolvimento - Protótipo Funcional + APK Android (10/10/2026 - revisão do login)**

Em 10/10/2026 o fluxo de login (RF02) foi auditado e endurecido: JWT consolidado em fonte única
(auth service + middleware injetado), refresh token passou a ser usado de verdade, e-mail
normalizado, validações em PT-BR, senha mínima 8 e correção do bug que impedia excluir a conta.

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

**10/10/2026 - Revisão do login: consolidação JWT, refresh token ativo e correção de bugs**

**Data**: 10/10/2026

**Descrição**:
- **Backend**: removido `internal/auth/middleware.go` (código morto); `middleware.AuthMiddleware`
  passou a receber um `TokenValidator` injetado (fonte única = `auth.AuthService`), com exigência de
  algoritmo **HS256**; claim `type` (`access`/`refresh`) e rejeição de token do tipo trocado em
  `/auth/refresh`; e-mail normalizado (`trim`+lowercase) em todo o fluxo de usuário; validação de
  e-mail (`net/mail`) e senha mínima de 8 caracteres; mensagens de erro em PT-BR; `DELETE
  /api/v1/user/profile` corrigido (usava `extractUserID`, que procurava o segmento `"users"` e
  sempre retornava 400 → conta nunca era excluída)
- **Frontend**: `isLoggedIn()` valida a expiração do JWT (`token_validity.dart`, `dart:convert`) e
  renova silenciosamente via refresh token; `getAuthHeaders()` renova token expirado antes das
  chamadas (liga o `refreshToken()` que era código morto) e desloga se a renovação falhar; e-mail
  com `trim()`; senha mínima 8; `"E-mail ou senha inválidos"` amigável no login; `UserModel`
  tolerante a `id`/`email` ausentes
- **Segurança**: `backend/.env` saiu do versionamento (`git rm --cached` + `backend/.gitignore`;
  `.env.example` segue rastreado como template). Backend sobe sem `.env` graças ao fallback dev do
  `main.go` e so ignora variáveis já definidas (parser stdlib, sem dependência nova).
- **Testes**: `cmd/server/main_test.go` (10 casos de handler) no backend; `token_validity_test.dart`
  (8 casos) + `UserModel` defensivo (3 casos) no frontend

**Verificação**:
- `go build ./...` + `go vet ./...` + `go test ./...` verdes
- `flutter analyze`: 15 issues (baseline, zero novos) · `flutter test`: 31/31 (era 20/20)
- E2E com servidor real do fonte: register 201 (e-mail normalizado), duplicado 409 PT, senha curta
  400 PT, login com maiúsculas/espaços 200, senha errada 401 PT, access token em `/auth/refresh`
  401, refresh token 200, `GET`/`DELETE /api/v1/user/profile` 200
- Detalhes completos em `CHANGELOG.md`

---

**06/10/2026 - Correção: salvar nome no perfil falhava com "type 'Null' is not a subtype of type 'String'"**

**Data**: 06/10/2026

**Descrição**:
- **Causa**: `PUT /api/v1/user/profile` respondia `200` com `{"message":"User updated"}` (`backend/internal/user/handler.go`), mas `ProfileService.updateProfile` fazia `ProfileModel.fromJson(jsonDecode(response.body))` — `json['id'] as String` recebia `null` → `type 'Null' is not a subtype of type 'String'`
- O dado salvava porque o `UpdateUser` roda **antes** de responder; o nome só aparecia ao voltar porque `app.dart` recria as telas na troca de aba e o `initState` refaz `_loadProfile()`
- "Não dava resposta": o `throw` ocorre **antes** do `Navigator.pop`, então a bottom sheet não fechava; e o `catch` mostrava o erro cru, já que `friendlyError()` só remove o prefixo `Exception: ` (`TypeError` não é `Exception`)
- **Pré-existente, não relacionado** ao trabalho de timeouts do mesmo dia (confirmado no `git diff`); varredura em todo o frontend mostrou que **só** este endpoint tinha corpo de resposta divergente (auth, channels, change-password e delete estavam corretos)
- **Backend**: `HandleUpdateUserProfile` agora chama `GetUser` após o update e devolve o objeto `User`, espelhando `HandleGetUserProfile`; `Password` segue `json:"-"`
- **Backend**: `backend/internal/user/handler_test.go` (novo, primeiro teste de handler do projeto) — `TestUserHandler_HandleUpdateUserProfile` exige `id`/`email`/`name` no `200`, confere persistência e ausência do `password`; `..._NotFound` cobre `404`. Usa `httptest` direto, pois a assinatura recebe `userID` por parâmetro
- **Frontend**: `_saveProfile` captura o `ScaffoldMessenger` **antes** do `await` e o reutiliza nos dois caminhos; sucesso agora fecha a sheet e mostra **"Perfil atualizado"** (verde)
- Backend reiniciado a partir do fonte; APK recompilado (50,8MB) e copiado para `RadioPX.apk`

**Verificação** (conferida, não presumida): `go build ./...` + `go test ./...` verde com os testes novos; `flutter analyze` = **15 issues**, subconjunto do baseline de 16 (saiu um `use_build_context_synchronously` pré-existente, **zero novos**); `flutter test` 20/20; `GET /health` → `200 {"status":"ok",...}`; `aapt2` → `launchMode=2` e `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e `Perfil atualizado`, não contém `10.201.108.96`.

**Motivo**: teste em campo — "ao editar o nome no perfil quando se aperta no salvar ele não dá resposta, dá um erro de subtype string null, mas quando volta o perfil foi alterado".

**Fragilidade mantida**: no caminho de sucesso o `finally { setState(...) }` roda logo após `Navigator.pop` — funciona porque a bottom sheet ainda está na animação de saída.

**Teste em campo sugerido**: editar nome → Salvar → sheet fecha e exibe "Perfil atualizado"; reabrir o perfil mantém o nome. Com o backend parado deve aparecer "Erro ao salvar: Não foi possível conectar ao servidor…". Observação: reiniciar o backend zera os usuários em memória (o `UserService` é um mapa em RAM), então é esperado recriar as contas de teste.

---

## Alteração Anterior

**06/10/2026 - Correção: app ficava carregando infinitamente ao entrar (sem timeout)**

**Data**: 06/10/2026

**Descrição**:
- **Causa**: o APK embutia `API_BASE_URL=http://10.201.108.96:8080` e a máquina voltou a responder em `172.16.158.222` → host inalcançável
- **Por pendurar em vez de dar erro**: nenhuma das 14 chamadas HTTP do frontend tinha timeout (busca por `.timeout(` em `lib/`: zero) → com o `SYN` sem resposta o TCP retenta ~2min → a Future nunca completa → os `catch`/`finally` que zeram `_isLoading` nunca rodam → `CircularProgressIndicator` eterno
- **O IP oscila**: `172.16.158.222` (build 22/09) → `10.201.108.96` (build 04/10, quando se supôs que o anterior estava "obsoleto") → `172.16.158.222` (build atual). Já voltou a ser o valor antigo, então não é troca de sub-rede — embarcar IP transitório no binário é insustentável
- Novo `frontend/lib/core/api_client.dart`: `apiTimeout = 10s`, `apiGet`/`apiPost`/`apiPut`/`apiDelete` com timeout que traduz `TimeoutException`/`SocketException`/`http.ClientException` para mensagens em PT-BR, e `friendlyError()` que remove o prefixo `Exception: ` nas mensagens da UI
- 14 chamadas migradas: `auth_service` (3), `channel_service` (7), `profile_service` (4)
- 11 mensagens de serviço traduzidas de EN para PT-BR (`channel_service` 7, `profile_service` 4); `register_screen` segue detectando `already exists` na mensagem do backend
- `location_service.dart`: `timeLimit: apiTimeout` no `getCurrentPosition` (verificado que o geolocator 10.1.1 aceita `Duration? timeLimit`)
- `websocket_service.dart`: `await channel.ready.timeout()` passa a detectar falha real de handshake — antes a linha `_isConnected = true` rodava antes de existir conexão (falso positivo) e o `try/catch` não capturava nada, porque `WebSocketChannel.connect` é lazy. `sink.close()` em falha é disparado sem aguardar (teto de 2s), já que o `WebSocketSinkCompleter` só recebe destino após handshake
- `voice_screen.dart`: backoff exponencial 2/4/8/16/30s com teto de **5 tentativas**; um único `Timer?` cancelado no `dispose()` (antes o Timer solto podia disparar depois); sucesso zera os contadores. Esgotou → estado terminal `_connectionFailed`: banner "Sem conexão com o servidor" e o botão PTT é substituído por **"Tentar novamente"** + instrução
- `friendlyError` aplicado em 11 pontos de UI: `login_screen`, `register_screen`, `channels_screen` (4), `profile_screen` (3), `channel_detail_screen` (2)
- APK recompilado (50,8MB) com `API_BASE_URL=http://172.16.158.222:8080` e copiado para `RadioPX.apk`

**Verificação** (conferida no artefato, não presumida): `flutter analyze` = **16 issues, os mesmos 16 do baseline** (zero novos); `flutter test` 20/20; `go test ./...` verde; `aapt2` → `launchMode=2` (singleTask) e `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e as strings novas (`Verifique a conex`, `Falha ao carregar canais pr`, `Sem conex`) e **não** contém `10.201.108.96`.

**Motivo**: teste em campo — "na hora que tentamos entrar, apenas fica carregando infinitamente".

**Pendências abertas**:
- URL editável em runtime = **decisão pendente** (exige RF/RNF novo em `docs/requisitos.md`)
- Reserva DHCP/IP estático no roteador = causa raiz da troca, vive fora do repo
- Em web o `ready` é `Future.value()` (stub) → a detecção de falha de handshake do WebSocket não é efetiva nessa plataforma; o alvo de teste é APK
- Validar em campo: com o backend parado, entrar num canal deve exibir a mensagem de erro em ~10s (não spinner) e o login deve cair no retry da reconexão

**Relacionado**: corrige/substitui a entrada de 04/10/2026, que registrou o build com `10.201.108.96` acreditando que `172.16.158.222` estava obsoleto — hoje o valor correto é `172.16.158.222` de novo.

---

## Alterações Anteriores

### 04/10/2026 - Correção: app abria em múltiplas instâncias em vez de ir para segundo plano

**Data**: 04/10/2026

**Descrição**:
- `frontend/android/app/src/main/AndroidManifest.xml`: removido `android:taskAffinity=""` e `launchMode` alterado de `singleTop` para `singleTask`
- Consequência corrigida: launcher criava uma task nova por lançamento (múltiplas cópias no recents), cada uma com WebSocket e sessão `flutter_sound` próprios → áudio dobrado e PTT duplicado
- APK release recompilado com `API_BASE_URL=http://10.201.108.96:8080` (o IP do APK anterior, `172.16.158.222`, está obsoleto — a rede mudou) e copiado para `RadioPX.apk` (50,8MB)
- Correção conferida no artefato: `aapt2` reporta `launchMode=2` (singleTask) e ausência de `taskAffinity`; `libapp.so` contém o IP novo e não contém o antigo

**Relacionado**: este era provavelmente o mesmo bug reportado como "não há opção de sair do canal" — a única forma de sair fica na tela de voz (engrenagem → "Sair do Canal") e o ícone do launcher abria outra instância, então o usuário nunca retornava àquela tela.

**Validação em campo pendente**: abrir o app, entrar num canal, ir para a tela inicial e tocar no ícone 3-4 vezes (deve retomar na mesma tela, 1 entrada no recents); depois revalidar o fluxo de sair. Se ainda falhar, expor botão visível de sair na tela de voz.

### 04/10/2026 - Correção: lista de canais não atualiza automaticamente

**Data**: 04/10/2026

**Descrição**:
- `ChannelsScreen` agora faz polling automático a cada 10s nas listas "Meus canais" e "Perto de mim"
- Refresh de polling é **silencioso**: não mexe em `_isLoading`/`_error`, logo não troca o `TabBarView` por spinner de tela cheia nem reseta aba/scroll; erro de rede é ignorado em vez de virar tela de erro
- Refresh inicial, botão refresh e pull-to-refresh seguem mostrando loading via `_refresh({required bool showLoading})` — nenhum call site mudou
- Guarda de concorrência `_isRefreshing` + `_sameChannels` (só chama `setState` se algo mudou)
- Localização em cache com TTL de 2min; refresh manual continua forçando GPS fresco
- `WidgetsBindingObserver`: pausa o timer em background e atualiza imediatamente ao voltar ao foreground

**Motivo**: ao criar um canal no celular B, o celular A só exibia o canal novo depois de trocar de aba. O `ChannelsScreen` só buscava dados em `initState`/refresh manual, e o reload ao trocar de aba era acidental (o `State` é destruído e recriado por `app.dart`).

**Resultado**: `flutter analyze` sem novos issues, `flutter test` 20/20 verdes, `go test ./...` verde.

**Pendências abertas** (não alteradas aqui):
- `backend/internal/channel/model.go` — `Channel` não tem campo `distance`; `GetNearbyChannels` só usa a distância para ordenar e nunca a serializa → a linha de distância do card de canal nunca aparece (exige mudança de backend)
- `frontend/lib/app.dart` — telas construídas dentro do `build` destroem o `State` ao trocar de aba (perda de estado); irrelevante para a atualização da lista, mas é a causa raiz do comportamento original

**Teste em campo sugerido**: 2 celulares na mesma conta, criar canal no B e observar "Perto de mim" no A em até ~10s sem tocar em nada; depois testar ida ao background → foreground (deve atualizar na hora) e modo avião (não pode virar tela de erro).

### 22/09/2026 - Correção: logout, cadastro, login reativo e latência de áudio

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

1. **Distância dos canais nunca aparece** — `backend/internal/channel/model.go`: a struct `Channel` não possui campo `distance`. `GetNearbyChannels` calcula a distância apenas para ordenar e nunca a serializa, então `ChannelModel.distance` é sempre `null` e o card de canal não renderiza a linha de distância (channels_screen.dart). Correção exige adicionar o campo ao modelo e popular no handler.
2. **Estado das telas perdido ao trocar de aba** — `frontend/lib/app.dart` monta a lista de telas dentro do `build` (`body: _screens[_currentIndex]`), então o `State` é destruído e recriado a cada troca de aba. Não afeta mais a atualização da lista de canais (que agora é por polling), mas faz scroll e aba interna voltarem ao padrão. Correção sugerida: `IndexedStack` + atualização explícita ao selecionar a aba.
3. **Validação em campo pendente da correção de instância única** — a mudança no manifesto (`taskAffinity` removido, `singleTask`) foi conferida no APK gerado via `aapt2`, mas o comportamento no aparelho ainda precisa ser validado: tocar no ícone do launcher 3-4 vezes com o app em background deve retomar na mesma tela e manter 1 entrada no recents. Se o fluxo "sair do canal" (engrenagem → "Sair do Canal") continuar falhando depois disso, o próximo passo é expor um botão visível de sair na tela de voz — hoje ele está 2 níveis abaixo de um ícone de engrenagem.
4. **IP do servidor embutido no APK fica obsoleto a cada troca de DHCP** — `API_BASE_URL` entra no binário pelo `--dart-define` do build, e o endereço da máquina oscila entre `172.16.158.222` e `10.201.108.96` (histórico: 22/09 → 04/10 → 06/10), quebrando o APK a cada troca. Mitigação aplicada em 06/10/2026: timeout de 10s + mensagem clara de erro (antes era spinner infinito). Correção real exige reserva DHCP/IP estático no roteador (fora do repo) ou URL editável em runtime (esta exige RF/RNF novo em `docs/requisitos.md`).
5. **Validação em campo pendente do timeout/reconexão** — a detecção de falha de handshake do WebSocket foi conferida na fonte do `web_socket_channel 2.4.5` e no código, mas não em aparelho: com o backend parado, entrar num canal deve exibir a mensagem de erro em ~10s e a tela de voz deve passar por 5 tentativas com backoff até o botão "Tentar novamente". Em web esse comportamento não é garantido (o `ready` lá é stub).
6. **Backend não usa PostgreSQL — tudo é em memória** — `internal/user/service.go`, `internal/channel/service.go` e `internal/location/service.go` guardam os dados em `map` protegidos por `sync.RWMutex`; não existe `database/sql`, pgx ou gorm em lugar nenhum do código. Isso **diverge** de `docs/decisoes.md` (DEC-001) e `docs/banco-de-dados.md`, que definem PostgreSQL 14+ como banco principal. Consequência prática: reiniciar o backend zera usuários, canais e localizações, então as contas de teste precisam ser recriadas a cada reinício. Correção exige implementar a persistência já prevista na documentação.
7. **`setState` logo após `Navigator.pop` no salvamento de perfil** — em `profile_screen.dart`, o `finally { setState(...) }` de `_saveProfile` executa logo após o `pop` do caminho de sucesso. Hoje funciona porque a bottom sheet ainda está na animação de saída, mas é frágil: se a animação mudar, pode virar `setState() called after dispose()`.

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

**Data**: 10/10/2026

**Hora**: 2026-10-10

**Por**: OpenCode (assistente AI)

---

## Notas para Continuação

1. **Testes**: Executar testes end-to-end completos
2. **Deploy**: Configurar ambiente de produção
3. **Monitoramento**: Implementar logs e métricas
4. **Documentação**: Atualizar README com instruções de deploy
5. **Google OAuth**: Implementar login social (futuro)
6. **IP do servidor**: a URL do backend entra no binário via `--dart-define` e oscila junto com o DHCP (ver Problemas Conhecidos nº 4). Resolver com reserva DHCP no roteador, ou tornando a URL editável em runtime — esta última exige RF/RNF novo em `docs/requisitos.md` (decisão pendente)
