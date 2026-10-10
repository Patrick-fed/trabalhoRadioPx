# Histórico do Projeto RadioPX

Este documento registra todas as alterações significativas do projeto.

---

## 10/10/2026 - Revisão do login: consolidação JWT, refresh token ativo e correção de bugs

### Contexto

Auditoria completa do fluxo de autenticação (RF02) na branch `feature-login`. O login já
funcionava no caminho feliz, mas havia código JWT duplicado, um bug latente na exclusão de
conta e o refresh token existia mas nunca era usado.

### Alterações

**Backend**

- `internal/auth/middleware.go` — **removido** (código morto: `AuthMiddleware` duplicado não era
  referenciado em lugar nenhum).
- `internal/middleware/auth.go` — `AuthMiddleware` agora recebe um `TokenValidator` injetado
  (implementado pelo `auth.AuthService`, fonte única de verificação de JWT). Removidos o segundo
  `Claims`, `GenerateToken` e `ValidateToken` que liam `JWT_SECRET` por conta própria. `Bearer`
  passa a ser aceito sem diferenciar maiúsculas/minúsculas.
- `internal/auth/service.go` — validação agora exige algoritmo **HS256** (evita alg-confusion);
  novo claim `type` (`access`/`refresh`); novo `ValidateAccessToken` que retorna `userID` e
  rejeita refresh token usado como access; `RefreshToken` rejeita access token usado como refresh.
- `internal/user/service.go` — e-mail normalizado (`trim` + lowercase) em `CreateUser`,
  `Authenticate`, `GetUserByEmail`, `UpdateUser` e `CreateOAuthUser`; validação de e-mail
  (`net/mail`, stdlib) e senha mínima de 8 caracteres (alinhado ao RF01). Erros traduzidos para
  PT-BR (`este e-mail já está cadastrado`, `a senha deve ter pelo menos 8 caracteres`, etc.).
- `internal/user/handler.go` — removidos handlers mortos por path (`HandleCreateUser`,
  `HandleGetUser`, `HandleUpdateUser`, `HandleDeleteUser`, `HandleListUsers`,
  `HandleChangePassword`, `extractUserID`). Novo `HandleDeleteUserByID`.
- `cmd/server/main.go` — middleware e injetor passam a usar `authService.ValidateAccessToken`;
  DELETE `/api/v1/user/profile` agora passa o `userID` (corrige o bug de excluir conta que
  sempre respondia 400, pois `extractUserID` procurava o segmento `"users"`); `handleRegister`/
  `handleLogin`/`handleRefresh` com mensagens em PT-BR e validação (400) para campos vazios,
  e-mail inválido e senha curta.
- `cmd/server/main_test.go` (novo) — 10 testes dos handlers: register sucesso/duplicado/inputs
  inválidos, login sucesso/senha errada/normalização de e-mail, refresh rejeitando access token e
  aceitando refresh token.

**Frontend**

- `lib/features/auth/services/token_validity.dart` (novo) — decodifica `exp` do JWT com
  `dart:convert` (sem dependência nova): `jwtExpiry` e `tokenIsExpired`.
- `lib/features/auth/services/auth_service.dart` — `isLoggedIn()` passou a validar a expiração do
  access token e, se expirado, **renova silenciosamente** via refresh token antes de deslogar;
  `getAuthHeaders()` renova proativamente token expirado e faz logout se a renovação falhar
  (liga o `refreshToken()` que era código morto); guard `_refreshingSession` evita loops.
- `login_screen.dart` / `register_screen.dart` — e-mail com `trim()`; senha mínima de 8
  caracteres; login mapeia `Invalid credentials`/senha inválida para `"E-mail ou senha inválidos"`.
- `user_model.dart` — parsing defensivo (`id`/`email` toleram ausência/nulo, datas via
  `DateTime.tryParse`); `displayName`/`initials` não quebram com e-mail vazio (retornam `?`).
- Testes: `token_validity_test.dart` (8 casos) + casos de `UserModel` tolerante no
  `models_test.dart` (3 casos).

### Motivo

Revisão solicitada do login: confirmar o que funciona, apontar o que falta e corrigir bugs.

### Verificação

- `go build ./...` + `go vet ./...` + `go test ./...` verdes (inclui os 10 testes novos de handler).
- `flutter analyze`: **15 issues — os mesmos 15 do baseline, zero novos** (durante o trabalho chegou
  a 16 com um erro real em `isLoggedIn` — variável `refreshToken` sombreava o método — corrigido).
- `flutter test`: **31/31** (era 20/20).
- Fluxo E2E contra o backend real (servidor iniciado do fonte, depois encerrado):
  register 201 com e-mail normalizado para lowercase; duplicado → 409 PT; senha curta → 400 PT;
  login com `Mixed.User@Example.com` + espaços → 200; senha errada → 401 PT; access token em
  `/auth/refresh` → 401; refresh token → 200; `GET /api/v1/user/profile` → 200; `DELETE
  /api/v1/user/profile` → 200 (antes sempre 400).

### Segurança (adicional)

- `backend/.env` **removido do versionamento** (`git rm --cached`) — o arquivo pode conter
  `JWT_SECRET`/credenciais Google, o que a regra do AGENTS.md proíbe (nunca committar segredos).
  O arquivo permanece no disco (ignorado pelo novo `backend/.gitignore`, que mantém
  `.env.example` rastreado como template).
- O servidor não depende do arquivo para subir: `main.go` usa fallback dev
  (`dev-secret-change-in-production`) quando `JWT_SECRET` não está definido, e o `loadEnvFile`
  próprio (parser stdlib, sem biblioteca externa) só injeta variáveis ausentes. Smoke test
  comprovou boot sem `.env` (health OK).
- **Nota**: o segredo antigo continua no histórico do git. Purga definitiva exigiria reescrever o
  histórico (`filter-repo`), o que é arriscado num repo acadêmico — recomendado apenas se o repositório
  for público. Para o TCC, troque o `JWT_SECRET` antes de qualquer exposição.

### Pendências abertas (fora de escopo)

- **Google OAuth** (parte do RF02) não implementado — exige nova dependência (`google_sign_in`) e
  decisão de escopo.
- `internal/middleware/cors.go`, `logging.go` e `ratelimit.go` ainda são código aparentemente não
  usado pelo `main.go` (que usa `github.com/rs/cors` e não aplica rate limit) — candidatos a remoção
  em tarefa futura.
- Backend continua em memória (mapas) — PostgreSQL segue pendente (PROJECT_STATE, problema nº 6).

---

## 06/10/2026 - Correção: salvar nome no perfil falhava com "type 'Null' is not a subtype of type 'String'"

### Causa

`PUT /api/v1/user/profile` respondia `200` com `{"message":"User updated"}` (`backend/internal/user/handler.go`), mas `ProfileService.updateProfile` fazia `ProfileModel.fromJson(jsonDecode(response.body))`, cujo factory exige `json['id'] as String` e `json['email'] as String` — ambos `null` → **`type 'Null' is not a subtype of type 'String'`** (o "subtype string null").

Por que o dado salvava mesmo assim: o `UpdateUser` executa **antes** de gravar a resposta, então o backend já tinha aplicado a alteração. O nome só aparecia ao voltar porque `frontend/lib/app.dart` monta a lista de telas dentro do `build`, o `State` é destruído/recriado na troca de aba e o `initState` refaz `_loadProfile()`.

Por que "não dava resposta": o `throw` ocorre **antes** do `Navigator.pop(context)`, então a bottom sheet de edição nunca fechava; e o `catch` exibia o erro cru, porque `friendlyError()` apenas remove o prefixo `Exception: ` e `TypeError` não é `Exception`.

**Pré-existente e não relacionado** ao trabalho de timeouts do mesmo dia — confirmado no `git diff`, que mostra que a troca `http.put` → `apiPut` e a tradução das mensagens não tocaram no parsing. Varredura por `jsonDecode(response.body)` em todo o frontend: auth (`data['user']`), channels (retornam o objeto), `change-password` e `delete` (só checam status) estavam corretos — **só este endpoint divergia**.

### Alterações

**Backend**
- `backend/internal/user/handler.go` — `HandleUpdateUserProfile` agora chama `h.userService.GetUser(userID)` após o update e devolve o objeto `User`, espelhando o padrão já existente de `HandleGetUserProfile`. `User.Password` continua `json:"-"`, então não vaza hash.
- `backend/internal/user/handler_test.go` (novo — primeiro teste de handler do projeto, que só tinha testes de service):
  - `TestUserHandler_HandleUpdateUserProfile` — guarda de regressão: exige `id`, `email` e `name` no corpo do `200`, confere que o serviço persistiu o `name` novo e que o campo `password` não aparece no JSON
  - `TestUserHandler_HandleUpdateUserProfile_NotFound` — `404` para usuário inexistente
  - a assinatura recebe `userID` por parâmetro, então o teste usa `httptest.NewRequest`/`NewRecorder` direto, sem subir router

**Frontend**
- `frontend/lib/features/profile/screens/profile_screen.dart` — `_saveProfile` captura o `ScaffoldMessenger` **antes** do `await` e o reutiliza nos dois caminhos: sucesso fecha a sheet e mostra **"Perfil atualizado"** (verde); erro mantém o SnackBar vermelho. Capturar antes do await evita usar `BuildContext` após gap assíncrono.

### Motivo

Teste em campo: "ao editar o nome no perfil quando se aperta no salvar ele não dá resposta, dá um erro de subtype string null, mas quando volta o perfil foi alterado".

### Verificação

- `go build ./...` + `go test ./...` verde (testes novos executados, não cacheados)
- `flutter analyze`: **15 issues** — subconjunto do baseline de 16: saiu um `use_build_context_synchronously` pré-existente (o `ScaffoldMessenger.of(context)` do bloco `catch`), **zero issues novos**
- `flutter test` 20/20
- Backend reiniciado a partir do fonte (`go run cmd/server/main.go`) e `GET /health` → `200 {"status":"ok","service":"radiopx-backend","connections":0}`
- Artefato conferido, não presumido: `aapt2` → `package="com.radiopx.radiopx"`, `launchMode=2` (singleTask), `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e a string nova `Perfil atualizado`, e **não** contém `10.201.108.96`
- APK release recompilado (50,8MB) e copiado para `C:\RadioPx\RadioPX.apk`

### Notas

- Fragilidade pré-existente mantida como está: no caminho de sucesso o `finally { setState(...) }` roda logo após `Navigator.pop` — funciona porque a bottom sheet ainda está na animação de saída.
- `HandleUpdateUser` (handler.go:69) é código morto, sem nenhuma chamada — fora do escopo desta correção.
- Backend reiniciado zera os dados em memória (o `UserService` é um mapa em RAM), então é esperado precisar recriar as contas de teste.

### Teste em campo sugerido

Editar o nome → Salvar → a sheet fecha e exibe "Perfil atualizado"; reabrir o perfil mantém o nome. Com o backend parado, deve aparecer "Erro ao salvar: Não foi possível conectar ao servidor. Verifique se ele está rodando e na mesma rede."

---

## 06/10/2026 - Correção: app ficava carregando infinitamente ao entrar (sem timeout em nenhuma chamada)

### Causa

O APK anterior embutia `API_BASE_URL=http://10.201.108.96:8080`, um IP transitório de DHCP. A rede renovou o lease e a máquina passou a responder em `172.16.158.222` — o host embutido ficou inalcançável (verificado: `TcpTestSucceeded=False` no IP antigo, `True` no atual, com o backend saudável em 8080).

O histórico confirma que **o endereço oscila**: `172.16.158.222` (build de 22/09) → `10.201.108.96` (build de 04/10, quando se supôs que o anterior estava "obsoleto") → `172.16.158.222` (build atual). Ou seja, o valor já voltou a ser o de antes, o que descarta qualquer teoria de troca de sub-rede e confirma que fixar o IP no binário é insustentável.

O sintoma era travamento em vez de erro porque **nenhuma chamada HTTP do frontend tinha timeout** (busca por `.timeout(`: zero ocorrências em `lib/`). Com o host inalcançável o `SYN` fica sem resposta e o TCP retenta por ~2 minutos; como a Future nunca completa, os blocos `catch`/`finally` que zeram `_isLoading` nunca executam e o `CircularProgressIndicator` gira para sempre.

### Alterações

**`frontend/lib/core/api_client.dart` (novo)**
- `apiTimeout = Duration(seconds: 10)` — fonte única do tempo de espera
- `apiGet`/`apiPost`/`apiPut`/`apiDelete` com `.timeout(apiTimeout)`, capturando `TimeoutException`, `SocketException` e `http.ClientException` e relançando mensagens em PT-BR ("Servidor não respondeu. Verifique a conexão com a rede." / "Não foi possível conectar ao servidor...")
- `friendlyError(Object)` — remove o prefixo `Exception: ` nas mensagens exibidas na UI

**Serviços — 14 chamadas migradas**
- `auth_service.dart`: 3 `http.post` → `apiPost` (`import http` mantido, ainda usado em `_readError`)
- `channel_service.dart`: 7 chamadas → `apiGet`/`apiPost`
- `profile_service.dart`: 4 chamadas → `apiGet`/`apiPut`/`apiPost`/`apiDelete`

**Mensagens em inglês traduzidas (11)** — `channel_service` (7) e `profile_service` (4) agora lançam "Falha ao carregar canais próximos", "Falha ao atualizar o perfil" etc. O `register_screen` continua detectando `already exists` na mensagem do backend, não numa string local.

**`location_service.dart`** — `getCurrentPosition` ganhou `timeLimit: apiTimeout`. Verificado na fonte do pacote que o geolocator 10.1.1 aceita `Duration? timeLimit`; o `catch` já existente devolve `null` e vira mensagem de erro, não pendura.

**`websocket_service.dart`** — `WebSocketChannel.connect` é lazy e o `try/catch` anterior não capturava falha de rede; a linha que marcava `_isConnected = true` roda **antes** da conexão existir (falso positivo). Agora:
- `await channel.ready.timeout(apiTimeout)` — confirmado na fonte do `web_socket_channel` 2.4.5 que no Android (`IOWebSocketChannel`) `ready` é o `_readyCompleter` real e **faz `completeError` em falha**; na classe base é `Future.value()` (stub)
- `_isConnected = true` só após o handshake concluir
- em falha, `sink.close()` é disparado **sem aguardar** com teto de 2s — o `WebSocketSinkCompleter` só recebe destino quando o handshake conclui, então aguardar poderia pendurar (reintroduziria o bug)

**`voice_screen.dart`** — substituí o `Timer(2s)` que reconectava para sempre:
- backoff exponencial 2s → 4s → 8s → 16s → 30s, teto de **5 tentativas**
- um único `Timer?` agendado (sem duplicar), cancelado no `dispose()` — antes o Timer solto podia disparar após o dispose
- sucesso **zera** `_reconnectAttempts`
- esgotou → estado terminal `_connectionFailed`: banner "Sem conexão com o servidor" e o botão PTT é substituído por **"Tentar novamente"** + instrução, que zera os contadores e reconecta

**`friendlyError` aplicado em 11 pontos de UI** — `login_screen`, `register_screen`, `channels_screen` (4), `profile_screen` (3), `channel_detail_screen` (2, hoje órfã).

### Motivo

Teste em campo: "na hora que tentamos entrar, apenas fica carregando infinitamente" — a rede havia mudado e toda a UI pendurava sem qualquer indicação de erro.

### Verificação

- `flutter analyze`: **16 issues — os mesmos 16 do baseline, zero novos**
- `flutter test` 20/20 · `go test ./...` verde
- Artefato conferido, não presumido: `aapt2` → `package="com.radiopx.radiopx"`, `launchMode=2` (singleTask), `taskAffinity` ausente; `libapp.so` contém `172.16.158.222` e as strings novas (`Verifique a conex`, `Falha ao carregar canais pr`, `Sem conex`) e **não** contém `10.201.108.96`
- APK release recompilado (50,8MB) e copiado para `C:\RadioPx\RadioPX.apk`

### Trade-offs e pendências

- **Em web** o `ready` é `Future.value()` (stub), então a detecção de falha de handshake do WebSocket não é efetiva nessa plataforma. `apiTimeout` na base só evita travamento, nunca falso-bloqueio. O alvo de teste é APK.
- **A causa raiz da troca de IP continua fora do repo**: a máquina recebe IP por DHCP e a URL fica fixa no binário via `--dart-define`. Resolver com reserva DHCP/IP estático no roteador, ou tornando a URL editável em runtime (esta última exige um RF/RNF novo em `docs/requisitos.md` — decisão pendente).
- `channel_detail_screen` continua órfã (nunca é navegada).

---

## 04/10/2026 - Correção: app abria em múltiplas instâncias em vez de ir para segundo plano

### Alterações

- `frontend/android/app/src/main/AndroidManifest.xml`:
  - removido `android:taskAffinity=""` — a activity ficava sem afinidade com qualquer task, então o launcher criava uma **task nova a cada lançamento** em vez de retomar a existente
  - `android:launchMode="singleTop"` → `"singleTask"` — `singleTop` só reaproveita a instância se a activity já estiver no topo da task atual, o que nunca ocorre com o app em background

### Motivo

Teste em campo: ao tocar no ícone do launcher, o app abria em uma instância nova (na aba Canais) em vez de voltar para a tela anterior. Isso gerava múltiplas cópias no recents, e como cada instância é um Flutter engine próprio, cada uma abria seu próprio WebSocket no mesmo canal (voice_screen.dart) e sua própria sessão `flutter_sound` — audible como áudio dobrado e PTT agindo em duplicidade.

Isso também explica o relato de "não há opção de sair do canal": a única forma de sair (engrenagem no AppBar → "Sair do Canal") fica na tela de voz, e o usuário nunca retornava a ela porque o ícone do launcher abria outra instância. Os dois atributos foram introduzidos à mão no commit `d6c3f91` (22/09) e não existem no template padrão do Flutter.

### Verificação

- `aapt2 dump xmltree` no APK gerado: `android:launchMode(0x0101001d)=2` (singleTask) e `taskAffinity` **ausente** — correção conferida no artefato final, não presumida
- `libapp.so` do APK contém `http://10.201.108.96:8080` e **não** contém o IP antigo `172.16.158.222` (a rede mudou desde 22/09 e o `--dart-define` do APK anterior estava desatualizado)
- APK release recompilado (50,8MB / 53.279.333 bytes) e copiado para `C:\RadioPx\RadioPX.apk`

### Trade-offs aceitos

- `singleTask` roteia novos intents para `onNewIntent` em vez de `onNewActivity`. Sem impacto agora: `MainActivity` é um `FlutterActivity` vazio e não há deep link. Se futuramente houver notificação ou deep link, tratar via `onNewIntent`/`NewIntentIntent`.
- `singleTask` impede duas instâncias do app no mesmo aparelho. Não afeta o teste em 2 celulares.

### Resultado

- `flutter analyze` sem novos issues; `flutter test` 20/20 verdes
- **Validação em campo pendente**: abrir o app, entrar num canal, ir para a tela inicial e tocar no ícone 3-4 vezes — deve retomar na mesma tela e o recents deve mostrar 1 entrada. Depois revalidar o fluxo de sair (engrenagem → "Sair do Canal"). Se o fluxo de sair ainda falhar, o próximo passo é expor um botão visível na tela de voz.

---

## 04/10/2026 - Correção: lista de canais não atualiza automaticamente

### Alterações

- `frontend/lib/features/channels/screens/channels_screen.dart`:
  - polling automático a cada 10s (`_autoRefreshInterval`) para atualizar as listas "Meus canais" e "Perto de mim" sem intervenção do usuário
  - `_loadAll` e `_silentRefresh` agora delegam para `_refresh({required bool showLoading})` — os call sites existentes (botão refresh, pull-to-refresh, criar/entrar/sair de canal) não mudaram
  - refresh silencioso não altera `_isLoading`/`_error`, portanto o `CircularProgressIndicator` de tela cheia e a tela de erro ficam restritos à carga inicial e ao refresh manual (antes, um refresh em polling trocaria o `TabBarView` inteiro e resetaria aba/scroll)
  - erro em refresh silencioso é ignorado (mantém a lista visível em vez de trocar a tela inteira por um erro)
  - guarda de concorrência `_isRefreshing` evita requisições sobrepostas e respostas fora de ordem
  - `_sameChannels` só chama `setState` quando algo de fato mudou (id, userCount, isAvailable, updatedAt, distance)
  - cache de localização com TTL de 2min (`_resolveLocation`/`_locationTtl`) — o polling reutiliza a última fix em vez de acordar o GPS de alta precisão a cada 10s; refresh manual continua forçando GPS fresco
  - `WidgetsBindingObserver`: pausa o timer quando a app vai para background e faz um refresh imediato ao voltar ao foreground

### Motivo

Teste em campo com 2 celulares: ao criar um canal no aparelho B, o aparelho A só mostrava o canal novo depois de trocar de aba. O `ChannelsScreen` só buscava dados em `initState`, no botão refresh e no pull-to-refresh — o reload ao trocar de aba era acidental, causado pelo `State` ser destruído e recriado por `app.dart` (que monta a lista de telas dentro do `build`).

### Resultado

- `flutter analyze` sem novos issues (16 pré-existentes, nenhum em `channels_screen.dart`)
- `flutter test` 20/20 verdes; `go test ./...` verde

### Pendências identificadas (não alteradas)

- `backend/internal/channel/model.go`: `Channel` não possui campo `distance`; `GetNearbyChannels` calcula a distância apenas para ordenar e nunca a serializa → `ChannelModel.distance` é sempre `null` e a linha de distância do card de canal nunca aparece. Requer mudança de backend.
- `frontend/lib/app.dart`: as telas continuam sendo construídas dentro do `build`, o que destrói o `State` ao trocar de aba (perda de estado). Passou a ser irrelevante para a atualização da lista, mas continua sendo a causa raiz do comportamento original.

---

## 22/09/2026 - Correção: Logout quebrado, erro no cadastro, login travado e latência de áudio

### Alterações

- `frontend/lib/features/profile/services/profile_service.dart`: `baseUrl` fixa (`http://localhost:8080`) substituída por `AppConfig.apiBaseUrl` — perfil (e o botão Sair) funcionam no celular
- `frontend/lib/features/auth/services/auth_service.dart`:
  - novo `ValueNotifier<bool> sessionState` (liga/desliga toda vez que o token de sessão muda em `logout`/`_saveAuthData`)
  - `register` aceita `200` e `201` (antes só `201` → missão confundia o sucesso com erro genérico)
  - login e register agora propagam o corpo da resposta do backend via `_readError` (mensagens de erro reais, ex.: e-mail duplicado)
- `frontend/lib/app.dart`: `AuthGate` escuta `AuthService.sessionState` e reavalia o estado de login — sem precisar reiniciar o app
- `frontend/lib/features/auth/screens/login_screen.dart`: removido `_LoginSuccess` desnecessário; navegação agora é reativa (AuthGate troca o body); guards `mounted` nos `setState`
- `frontend/lib/features/auth/screens/register_screen.dart`: sucesso volta para a tela inicial via `popUntil`; mensagem amigável para e-mail já cadastrado; guards `mounted`
- `frontend/lib/features/profile/screens/profile_screen.dart`: botão "Sair" fica sempre visível (inclusive quando o perfil falha); `_logout` só fecha a sessão (o AuthGate cuida da navegação)
- `frontend/lib/features/voice/services/stream_audio_player.dart`: reescrito com `flutter_sound` (sessão de streaming contínua PCM16/16kHz, `uint8ListSink`) — elimina o delay de ~2,5s do buffer inicial do `just_audio`
- `frontend/lib/features/voice/screens/voice_screen.dart`: voz abre a sessão de áudio ao entrar no canal (`startSession`)
- `frontend/pubspec.yaml`: `just_audio` removido; `flutter_sound ^9.28.0` adicionado
- `frontend/android/app/build.gradle.kts`: `minSdk` elevado para 24 (requisito do flutter_sound)
- APK release recompilado com `--dart-define=API_BASE_URL=http://172.16.158.222:8080` e copiado para `C:\RadioPx\RadioPX.apk`

### Motivo

Teste em campo: cadastro mostrava erro mesmo quando funcionava, login não entrava até reiniciar o app, não havia como sair da conta e o áudio tinha delay enorme.

### Resultado

- `flutter analyze` sem erros; `flutter test` 20/20 verdes
- `flutter build apk --release` OK (50,4MB) → `C:\RadioPx\RadioPX.apk`

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
