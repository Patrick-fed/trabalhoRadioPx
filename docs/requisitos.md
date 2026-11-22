# Requisitos do Sistema RadioPX

## Visão Geral

Este documento lista todos os requisitos funcionais e não funcionais do sistema RadioPX.

> **Status de implementação (21/11/2026)**: este documento define **o que** o sistema deve fazer.
> Nem todos os RF/RNF abaixo estão totalmente implementados na v1.
> O mapeamento de lacunas e de escopo fica em `PROJECT_STATE.md` →
> *Lacunas / Aderência aos Requisitos*. Em especial: RF02 (OAuth é trabalho futuro),
> RNF05 (Casbin fora do escopo v1; TLS via túnel ngrok) e RNF02/RNF03 (sem testes de carga).
> Já resolvidos: RF03 (um falante por vez, `voiceGuard`), RF06 (buffer/replay via
> `request_replay`, 27/10), RF08 (fila com teto de 30s, 27/10), persistência
> PostgreSQL (Fase 1), RNF06 (URL do servidor editável, Fase 3) e RF09
> (painel admin com métricas, gestão de usuários e localização, 02/11; SPA
> migrada para Vue 3 em 21/11, sem mudança de RF09 — DEC-020).

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
- Transmissão é feita via WebSocket autenticado por JWT (access token)
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

**Implementação (27/10/2026):** o relay do WebSocket grava cada chunk no
`AudioBufferService` (memória, 30s/10MB) e retransmite `sequence`; após
reconexão o app envia `request_replay {room, since_sequence}` e recebe os
segmentos perdidos por usuário, encerrados por um pacote `is_ptt=false`.

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

**Implementação (27/10/2026):** `AudioQueue` (FIFO, ~960KB ≈ 30s de PCM16)
entre o WebSocket e o player; fim de segmento entra como sentinel na própria
fila. "Fila de prioridade" é atendida por ordem de chegada (FIFO) — prioridade
por timestamp ficou fora de escopo.

---

### RF09 - Painel Administrativo

O sistema deve oferecer um painel web de administração com:
- Dashboard de métricas do sistema
- Gestão de usuários (listar, suspender, reativar, promover/demover admin, excluir)
- Visualização da última localização conhecida de cada usuário

**Critérios de Aceitação:**
- Painel acessível em `/admin` no mesmo host do backend (SPA embutida)
- Acesso restrito a usuários com `is_admin=true` (bootstrap via `ADMIN_EMAILS`)
- Dashboard mostra: total de usuários, ativos, canais, conexões WebSocket,
  salas ativas e canais transmitindo (refresh a cada 10s), com subtexto de
  detalhe e badge de variação desde a leitura anterior
- Dashboard exibe **mapa** (Leaflet + tiles OSM) com um marcador por usuário
  que tenha posição conhecida; popup com nome, data, status e link OSM;
  suspenso = marcador âmbar; sem posição = sem marcador
- Card "Localização" mostra quantos usuários têm posição conhecida
- Lista de usuários **nunca** expõe `password_hash`
- Suspensão invalida o login (`Authenticate` rejeita `!IsActive`)
- Admin não pode suspender, demover nem excluir **a si mesmo**
- Cada usuário exibe a última posição (`last_lat`/`last_lng`/`last_location_at`)
  com link para OpenStreetMap, ou "—" se nunca enviou
- Localização capturada de forma **passiva** (requests que já enviam lat/lng:
  `/channels/nearby` e `POST /channels`), sem histórico — só a última posição

**Implementação (02/11/2026, SPA migrada em 21/11/2026):** pacote `internal/admin`
(API `/api/v1/admin/{stats,users}`), middleware `RequireAdmin`, migrações
`003_admin_users` + `004_user_last_location`, SPA em `internal/admin/web/`
(`go:embed`). A origem da SPA passou a ser `admin-ui/` (Vue 3 + Vite + Pinia +
Leaflet); o build continua gerando os mesmos artefatos embutidos e os critérios
acima permanecem inalterados. Ver DEC-017, DEC-018 e DEC-020.

---

### RF10 - Simulação de Proximidade via Bluetooth

O aplicativo deve exibir um indicador visual de proximidade simulada baseado
exclusivamente no estado do adaptador Bluetooth do próprio dispositivo Android.

**Critérios de Aceitação:**
- Bluetooth ligado: exibir "Alguém por perto"
- Bluetooth desligado: exibir "Ninguém por perto"
- Bluetooth indisponível, estado desconhecido ou falha na consulta: exibir "Indisponível"
- Consultar o estado atual ao abrir o aplicativo autenticado
- Atualizar o indicador automaticamente quando o estado do Bluetooth mudar com o aplicativo aberto, sem reinício
- Manter o indicador visível na área principal compartilhada pelas abas Canais, Falar e Perfil
- Não solicitar ao usuário que ligue o Bluetooth nem tentar ligá-lo automaticamente
- Não descobrir dispositivos, identificar usuários/aparelhos, medir distância, parear, conectar via Bluetooth/BLE, comunicar-se entre dispositivos, transmitir áudio, notificar ou iniciar ações automáticas
- Não alterar a comunicação existente por Wi-Fi e WebSocket
- Aplicar somente ao Android; iOS, Web e Windows ficam fora de escopo

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
- Rate limiting (10 req/min por IP) nos endpoints de autenticação (register/login/refresh)
- Streaming de áudio via WebSocket autenticado por JWT, com identidade derivada do token

**Critérios de Verificação:**
- Testes de penetração
- Auditoria de autenticação
- Verificação de TLS em produção
- Excesso de requisições de autenticação retorna `429`

---

### RNF06 - Configurabilidade do servidor

- A URL da API é editável em runtime pelo próprio app (tela/sheet acessível no
  login e no perfil), sem exigir novo build para trocar de servidor
- Valor inicial (semente) vem do build (`--dart-define=API_BASE_URL`); a URL
  salva pelo usuário persiste localmente (`shared_preferences`) e tem precedência
  sobre a semente
- Entrada validada: apenas `http://` ou `https://`, com `trim` e remoção de `/`
  final; a URL WebSocket é derivada automaticamente (`https:` → `wss:`,
  `http:` → `ws:`)
- Botão de teste de conexão contra `GET /health` antes de salvar

**Critérios de Verificação:**
- Trocar a URL no app e conectar sem recompilar o APK
- URL inválida (sem `http(s)://`) é recusada com mensagem clara
- Persistência sobrevive ao reinício do app; sem valor salvo, usa a semente do build
- `flutter test` cobre normalização, persistência e o formulário

---

### RNF07 - Minimização de Permissões e Isolamento da Simulação Bluetooth

- A simulação deve utilizar o menor conjunto de permissões Android necessário apenas para consultar o estado do adaptador Bluetooth
- Permissões de descoberta de dispositivos e anúncio/devisibilidade devem ser evitadas
- Qualquer permissão adotada deve ter justificativa técnica registrada na etapa de planejamento
- O estado do Bluetooth não deve ser persistido nem enviado ao backend
- Falhas na consulta do estado do Bluetooth devem degradar apenas o indicador, sem bloquear funcionalidades existentes

**Critérios de Verificação:**
- Revisão das permissões declaradas no manifesto Android
- Verificação de que nenhuma permissão de descoberta/anúncio é necessária para o escopo
- Teste de consulta com estado desconhecido/falha sem afetar canais, PTT ou WebSocket
- Auditoria de que nenhum dado de estado Bluetooth é enviado à API

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

- **Backend**: `internal/channel/` (descoberta via `GetNearbyChannels`) + `internal/location/` (cálculo de distância)
- **Database**: N/A (GPS do dispositivo; coordenadas trafegam na API de canais)
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

### RF09 → Backend → Database → Frontend → Testes

- **Backend**: `internal/admin/` + `middleware.RequireAdmin` + bootstrap `ADMIN_EMAILS`
- **Database**: `users.is_admin`, `users.last_lat/last_lng/last_location_at` (migrações 003/004)
- **Frontend**: origem `admin-ui/` (Vue 3); artefato buildado em
  `internal/admin/web/` (servido em `/admin`; sem tela no app mobile)
- **Testes**: `admin-ui/tests/**` (39 testes Vue), `internal/admin/handler_test.go`
  (11 casos), `user/service_test.go` (suspend/promote/bootstrap/location), E2E
  manual + `test-backend.ps1` (regressão)

### RF10 → Backend → Database → Frontend → Testes

- **Backend**: N/A (simulação local)
- **Database**: N/A (sem persistência)
- **Frontend**: aplicativo Flutter, área principal autenticada compartilhada pelas três abas
- **Testes**: `frontend/test/bluetooth_proximity_controller_test.dart`,
  `frontend/test/bluetooth_proximity_indicator_test.dart` e validação manual em
  dispositivo Android
