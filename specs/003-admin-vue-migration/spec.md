# Feature Specification: Migração do Painel Administrativo para Vue 3

**Feature Branch**: `feature/migracao-painel-adm-para-vue`  
**Created**: 2026-11-21  
**Status**: Implementado em 21/11/2026 (T043 pendente apenas por bloqueio externo de PostgreSQL)  
**Input**: User description: "Migrar o painel administrativo atual (HTML/CSS/JS puro) para Vue.js sem alterar funcionalidades, mantendo integração com o backend Go existente. Stack: Vue 3, Vite, JavaScript (sem TypeScript), Vue Router, Pinia e Leaflet 1.9.4."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Acesso administrativo preservado (Priority: P1)

Como administrador do RadioPX, eu quero entrar no painel com meu e-mail e senha e ser reconhecido como admin, para que continuar tendo acesso às mesmas funções de hoje sem que a origem da permissão mude.

**Why this priority**: Sem autenticação e autorização idênticas, nenhuma outra função do painel pode ser usada com segurança.

**Independent Test**: Fazer login com usuário admin e com usuário comum, verificando que apenas o admin entra e que o token nunca aparece em armazenamento persistente.

**Acceptance Scenarios**:

1. **Given** credenciais válidas de um usuário com `is_admin=true`, **When** o login é enviado, **Then** o token é guardado somente em memória e a casca autenticada do painel é exibida.
2. **Given** credenciais válidas de um usuário com `is_admin=false`, **When** o login é processado, **Then** o token é descartado e a mensagem "Acesso negado: apenas administradores" é exibida.
3. **Given** um token expirado ou inválido durante qualquer operação, **When** a API responde 401, **Then** a sessão em memória é limpa, o login reaparece e a mensagem "Sessão expirada. Entre novamente." é exibida.
4. **Given** a página é recarregada, **When** a aplicação reinicia, **Then** não há token persistido e a tela de login é exibida.

---

### User Story 2 - Gestão de usuários sem regressão (Priority: P2)

Como administrador, eu quero consultar, buscar e alterar usuários exatamente como faço hoje, para que a migração não mude resultados, mensagens nem regras de segurança.

**Why this priority**: É a operação administrativa mais sensível: suspensão, promoção e exclusão afetam diretamente outros usuários.

**Independent Test**: Carregar a lista, aplicar filtro e executar cada ação em um usuário diferente do admin atual, comparando requisições e mensagens com o comportamento atual.

**Acceptance Scenarios**:

1. **Given** a aba Usuários é aberta, **When** a lista é carregada, **Then** as colunas Nome, E-mail, Status, Papel, Localização e Ações são exibidas com os mesmos badges e formatos atuais.
2. **Given** um termo de busca, **When** o usuário digita no campo de busca, **Then** apenas nomes ou e-mails que contêm o termo (sem diferenciar maiúsculas) permanecem, sem nova chamada à API.
3. **Given** um usuário diferente do admin atual, **When** uma ação de suspender, reativar, promover, remover admin ou excluir é confirmada, **Then** a requisição correspondente é enviada, a mensagem atual aparece e a lista e o dashboard são recarregados.
4. **Given** a linha do próprio administrador, **When** a tabela é renderizada, **Then** a coluna Ações mostra "você" e nenhum botão de modificação.

---

### User Story 3 - Dashboard e mapa idênticos (Priority: P3)

Como administrador, eu quero ver as mesmas métricas e a última posição conhecida dos usuários, para que a migração não altere a leitura operacional do sistema.

**Why this priority**: O dashboard é a visão principal ao entrar; qualquer diferença em números ou no mapa seria percebida como regressão.

**Independent Test**: Comparar os sete cards, os subtítulos, os deltas e os marcadores do mapa antes e depois da migração com a mesma base de dados.

**Acceptance Scenarios**:

1. **Given** o dashboard está aberto, **When** a tela é carregada e a cada 10 segundos, **Then** as métricas e a lista de usuários são obtidas em paralelo e os sete cards são preenchidos com os mesmos valores de hoje.
2. **Given** duas leituras consecutivas com valores diferentes, **When** a segunda leitura termina, **Then** os deltas de Usuários, Ativos, Canais e Localização aparecem com o sinal e o valor atuais.
3. **Given** usuários com latitude e longitude preenchidas, **When** o mapa é renderizado, **Then** apenas a última posição de cada um aparece, com popup e cores iguais aos atuais.
4. **Given** nenhum usuário com posição conhecida, **When** o mapa é renderizado, **Then** a mensagem "Nenhum usuário com posição conhecida ainda." é exibida.

---

### User Story 4 - Entrega e navegação preservadas (Priority: P4)

Como responsável pelo projeto, eu quero que o painel migrado continue sendo servido pelo próprio backend e com a mesma URL, para que não seja necessário alterar deploy, rotas ou documentação de acesso.

**Why this priority**: Preserva a arquitetura atual e evita impacto em produção, sem criar um novo serviço.

**Independent Test**: Acessar `/admin`, `/admin/` e um caminho sob `/admin/*` no binário do servidor, alternar entre Dashboard e Usuários e conferir que a URL permanece `/admin/`.

**Acceptance Scenarios**:

1. **Given** o servidor Go compilado, **When** `/admin` é acessado, **Then** o backend redireciona para `/admin/` e serve a SPA.
2. **Given** um caminho sob `/admin/*`, **When** é solicitado, **Then** o `SPAHandler` devolve o HTML da aplicação como hoje.
3. **Given** a casca autenticada, **When** o administrador alterna entre Dashboard e Usuários, **Then** a URL não muda e nenhuma entrada é adicionada ao histórico do navegador.
4. **Given** o build de produção, **When** a aplicação é servida, **Then** Leaflet e seus ícones vêm de arquivos locais, sem CDN para JS, CSS ou imagens.

---

### Edge Cases

- Sessão expira (401) no meio de uma ação de usuário: limpar estado, voltar ao login e não aplicar mudança local não confirmada pela API.
- Limite de requisições (429) no login ou em qualquer ação: exibir a mensagem atual sem descartar o token, exceto no próprio login.
- Falha de rede: exibir "Não foi possível conectar ao servidor" sem trocar de tela.
- Usuário autenticado sem `is_admin=true`: bloquear o painel mesmo que a API de login tenha retornado token.
- Administrador tenta suspender, rebaixar ou excluir a si mesmo pela interface: não exibir botões, mantendo a proteção do backend.
- Lista vazia ou filtro sem correspondência: exibir "Nenhum usuário encontrado" em vez de tabela vazia.
- Usuário sem `last_lat` ou `last_lng`: mostrar "—" na tabela e não criar marcador.
- Coordenadas inválidas ou fora do intervalo: não quebrar o mapa; manter o comportamento de filtrar apenas valores não nulos como hoje.
- Container do mapa oculto ao trocar de aba: chamar `invalidateSize` ao voltar para o Dashboard.
- Atualização automática a cada 10s durante pan/zoom: não reajustar os limites do mapa se a quantidade de marcadores não mudou.
- Múltiplos cliques em "Atualizar" ou trocas rápidas de aba: evitar timers duplicados e requisições sobrepostas desnecessárias.
- Recarga com token em memória: token perdido é comportamento esperado; retornar ao login.
- Conteúdo de usuário (nome/e-mail) com caracteres especiais: continua escapado ao renderizar, preservando a proteção contra injeção de HTML.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O painel MUST exibir uma tela de login com os campos "E-mail" e "Senha", o botão "Entrar" e uma área de mensagem de erro.
- **FR-002**: O login MUST enviar `POST /api/v1/auth/login` com `{ "email", "password" }` e `Content-Type: application/json`, sem cabeçalho de autorização.
- **FR-003**: O token retornado MUST ser mantido apenas em estado em memória e MUST NOT ser gravado em `localStorage`, `sessionStorage`, cookies ou arquivos.
- **FR-004**: Após o login, o painel MUST consultar `GET /api/v1/admin/users` e localizar o administrador pelo e-mail normalizado (minúsculas e sem espaços nas bordas) para obter identidade, `id` e `is_admin`.
- **FR-005**: O acesso MUST ser negado com "Acesso negado: apenas administradores" e o token MUST ser descartado quando o usuário autenticado não tiver `is_admin=true`.
- **FR-006**: Toda chamada à API admin MUST enviar `Authorization: Bearer <token>` e `Content-Type: application/json`.
- **FR-007**: Uma resposta 401 MUST limpar token, identidade, cache, deltas e timer, exibir "Sessão expirada. Entre novamente." e retornar ao login.
- **FR-008**: Uma resposta 429 MUST exibir "Muitas tentativas. Aguarde um momento." sem descartar o token.
- **FR-009**: Uma falha de rede MUST exibir "Não foi possível conectar ao servidor".
- **FR-010**: Demais erros MUST exibir `error`/`message` do corpo da resposta ou `Erro <status>` quando não houver mensagem.
- **FR-011**: O logout MUST limpar todo o estado em memória, parar o timer, limpar o campo Senha e voltar à tela de login.
- **FR-012**: Um recarregamento de página MUST iniciar em estado não autenticado, sem renovação silenciosa de token.
- **FR-013**: O Vue Router MUST controlar apenas a transição entre a tela de login e a casca autenticada.
- **FR-014**: "Dashboard" e "Usuários" MUST permanecer como abas internas da casca autenticada, sem alterar a URL e sem criar entradas no histórico do navegador.
- **FR-015**: As URLs atuais `/admin`, `/admin/` e caminhos sob `/admin/*` MUST continuar sendo atendidos pelo `SPAHandler` existente.
- **FR-016**: O acesso direto a `/admin/` sem token em memória MUST exibir a tela de login.
- **FR-017**: Ao entrar no Dashboard e a cada 10 segundos, o painel MUST buscar em paralelo `GET /api/v1/admin/stats` e `GET /api/v1/admin/users`.
- **FR-018**: O Dashboard MUST exibir os sete cards atuais: Usuários, Ativos, Canais, Conexões WS, Salas ativas, Transmitindo e Localização.
- **FR-019**: Os cards MUST ser preenchidos respectivamente com `users_total`, `users_active`, `channels_total`, `ws_connections`, `live_rooms`, `transmitting_channels` e a contagem de usuários com `last_lat` e `last_lng` não nulos.
- **FR-020**: Os subtítulos dos cards MUST permanecer idênticos: suspensos/"todos ativos", "% do total"/"sem usuários", "canais criados"/"nenhum canal", "clientes conectados"/"nenhuma conexão", "salas com membros"/"nenhuma sala", "canais no ar"/"silêncio no ar" e "com posição conhecida".
- **FR-021**: Os deltas desde a leitura anterior MUST continuar aparecendo apenas para Usuários, Ativos, Canais e Localização, com sinal "+" ou "−" e ocultos quando não há variação.
- **FR-022**: O texto "Atualização automática a cada 10s. Variação desde a leitura anterior." MUST permanecer no Dashboard.
- **FR-023**: O timer de atualização MUST ser iniciado no login, reiniciado ao trocar para o Dashboard se necessário e interrompido no logout.
- **FR-024**: O mapa MUST mostrar apenas a última posição conhecida dos usuários, sem histórico de trajetos.
- **FR-025**: O mapa MUST usar tiles do OpenStreetMap e ícones/imagens do Leaflet 1.9.4 servidos localmente, preservando a atribuição atual.
- **FR-026**: Cada marcador MUST ter o estilo atual por status (ativo/suspenso) e popup com nome, e-mail, badge de status, data/hora em formato `pt-BR` e link "abrir no OpenStreetMap".
- **FR-027**: Os limites do mapa MUST ser reajustados apenas na primeira renderização ou quando a quantidade de marcadores mudar.
- **FR-028**: O mapa MUST chamar `invalidateSize` ao se tornar visível novamente após a troca de aba.
- **FR-029**: Quando não houver usuários com posição conhecida, o painel MUST exibir "Nenhum usuário com posição conhecida ainda.".
- **FR-030**: Os textos de ajuda sobre marcadores e origem dos dados do mapa MUST permanecer os atuais.
- **FR-031**: Ao abrir a aba Usuários, o painel MUST buscar `GET /api/v1/admin/users`, exibindo "Carregando…" até a resposta. **Concluído em 21/11/2026** — ver T047 e `admin-ui/tests/components/UsersTable.spec.js`.
- **FR-032**: A tabela MUST manter as colunas Nome, E-mail, Status, Papel, Localização e Ações, nessa ordem.
- **FR-033**: Os badges "ativo"/"suspenso" e "admin"/"usuário" MUST permanecer com os mesmos rótulos e significados.
- **FR-034**: A busca por nome ou e-mail MUST ser local, case-insensitive, com remoção de espaços nas bordas e sem nova requisição à API.
- **FR-035**: A coluna Localização MUST mostrar latitude e longitude com cinco casas decimais, link "ver no mapa" e data/hora em `pt-BR`, ou "—" quando a posição estiver ausente.
- **FR-036**: Quando a lista ou o filtro não retornar usuários, o painel MUST exibir "Nenhum usuário encontrado".
- **FR-037**: O botão "Atualizar" MUST recarregar a lista de usuários sob demanda.
- **FR-038**: A linha do administrador autenticado MUST exibir "você" na coluna Ações e MUST NOT exibir botões de suspensão, promoção, rebaixamento ou exclusão.
- **FR-039**: Para os demais usuários, o painel MUST manter os botões Suspender/Reativar, Tornar admin/Remover admin e Excluir, com os mesmos textos de confirmação e mensagens de sucesso atuais.
- **FR-040**: As ações MUST enviar, conforme o caso, `PATCH { "is_active": false|true }`, `PATCH { "is_admin": true|false }` ou `DELETE` para `/api/v1/admin/users/{id}` e, após sucesso, recarregar usuários e dashboard.
- **FR-041**: Os avisos temporários (toasts) MUST manter a duração de aproximadamente 3,2 segundos e as mensagens atuais, incluindo "Usuário suspenso", "Usuário reativado", "Usuário promovido a admin", "Privilégios de admin removidos" e "Usuário excluído".
- **FR-042**: A aplicação MUST continuar usando a mesma origem da API (`API = ""`), sem configurar uma URL base separada em produção.
- **FR-043**: O build de produção MUST ser incorporado ao binário Go existente por `go:embed` e servido por `SPAHandler`, sem exigir servidor estático adicional.
- **FR-044**: A migração MUST NOT introduzir novos endpoints, novos campos de API, mudanças de esquema no banco ou alterações nas regras de negócio do backend; `is_admin` no PostgreSQL continua sendo a única fonte da verdade.
- **FR-045**: A migração MUST NOT alterar Flutter, WebSocket de áudio, streaming de voz, armazenamento de áudio, Orbit ou Bluetooth.
- **FR-046**: Todo conteúdo originado da API MUST continuar sendo escapado ao ser exibido na interface.

### Migration Constraints

- Stack obrigatória: Vue 3, Vite, JavaScript (sem TypeScript), Vue Router, Pinia e Leaflet 1.9.4.
- Navegação resolvida: Vue Router fica responsável por login/casca autenticada; Dashboard e Usuários continuam como abas internas com URL inalterada.
- Backend, modelo de dados, endpoints e autorização permanecem intactos.
- Assets do Leaflet (JS, CSS e imagens de marcadores) permanecem locais; apenas os tiles do OpenStreetMap continuam remotos, como hoje.
- Nenhuma funcionalidade nova, melhoria de UX ou alteração visual não estritamente necessária à troca de tecnologia entra neste escopo.

### Documented Conflicts

- DEC-017 descreve a SPA como "sem build, sem projeto separado, sem nova dependência"; a migração exige build e novas dependências. Registrar nova decisão arquitetural antes da implementação.
- DEC-019 define o Leaflet vendorizado em `backend/internal/admin/web/vendor/leaflet/`; a origem no repositório pode mudar com o build, mas o servido ao navegador deve continuar local.
- As diretrizes do SpecKit pedem specs sem tecnologias específicas; a stack foi mantida apenas em "Migration Constraints" porque foi exigida pelo solicitante.
- "Sem regressões perceptíveis" poderia conflitar com rotas de Dashboard/Usuários no Vue Router; resolvido com a opção C (URL inalterada).
- O nome original da branch (`feature: migracao Painel ADM para Vue`) era inválido no Git; substituído por `feature/migracao-painel-adm-para-vue` com traços.

### Key Entities

- **AdminSession**: estado em memória da sessão atual; contém token, identidade do administrador (`id`, `email`, `is_admin`) e e-mail de exibição; descartado no logout ou recarga.
- **AdminUser**: usuário retornado pela API admin; atributos relevantes: `id`, `name`, `email`, `is_active`, `is_admin`, `last_lat`, `last_lng`, `last_location_at`.
- **DashboardStats**: snapshot exibido nos cards; contém `users_total`, `users_active`, `channels_total`, `ws_connections`, `live_rooms`, `transmitting_channels`.
- **MapMarker**: derivado de `AdminUser` apenas quando há coordenadas; guarda posição, status visual e conteúdo do popup.
- **ToastMessage**: aviso temporário em memória com texto e tipo (sucesso/erro), exibido por aproximadamente 3,2 segundos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% das interações existentes no painel permanecem disponíveis após a migração, sem nenhuma funcionalidade adicional.
- **SC-002**: Login admin, login não-admin, 401, 429 e falha de rede produzem exatamente os mesmos resultados e mensagens de hoje em 100% dos cenários de teste.
- **SC-003**: Nenhuma inspeção de `localStorage`, `sessionStorage` ou cookies encontra token após login, ação, logout ou recarga.
- **SC-004**: A URL permanece `/admin/` após 100% das trocas entre Dashboard e Usuários, com zero entradas novas no histórico do navegador.
- **SC-005**: Os sete cards e seus subtítulos apresentam valores idênticos aos da API em todas as leituras, com atualização a cada 10 segundos.
- **SC-006**: A quantidade de marcadores do mapa é igual à quantidade de usuários com posição conhecida, e pan/zoom não são redefinidos quando essa quantidade não muda.
- **SC-007**: Todas as ações de usuário geram as mesmas requisições, confirmações e mensagens atuais; a própria linha do admin não oferece ações.
- **SC-008**: O painel continua disponível no mesmo endereço `/admin/` do servidor atual, sem servidor estático adicional, e nenhum recurso visual do mapa é buscado em CDN.
- **SC-009**: Todos os cenários de aceitação desta especificação passam em teste de regressão comparativo com a versão atual.
- **SC-010**: Nenhum arquivo de backend, Flutter, WebSocket, banco de dados ou contrato de API é alterado como parte desta migração.

## Assumptions

- Os endpoints e formatos de resposta atuais permanecem disponíveis e estáveis durante a migração.
- A aplicação continuará sendo servida na mesma origem da API em produção; em desenvolvimento, proxy local poderá reproduzir esse comportamento.
- Os artefatos do build serão disponibilizados no diretório incorporado por `go:embed` antes da compilação do binário.
- Administradores utilizam navegadores modernos com suporte a Vue 3 e Leaflet 1.9.4.
- Os tiles do OpenStreetMap permanecem remotos, pois já são assim na versão atual.
- A expiração do token em 24 horas e a ausência de renovação silenciosa continuam válidas.
- Não há necessidade de internacionalização: a interface permanece em português brasileiro.
- A migração é apenas de tecnologia de frontend; decisões arquiteturais novas serão registradas separadamente em `docs/decisoes.md` antes da implementação.
