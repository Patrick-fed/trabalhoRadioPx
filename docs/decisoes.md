# Decisões Técnicas RadioPX

Este documento registra as decisões arquiteturais e técnicas importantes do projeto.

---

## DEC-001 — Banco de Dados PostgreSQL

**Data**: 08/09/2026

**Decisão**: Utilizar PostgreSQL como banco de dados principal.

**Motivo**:
- Suporte nativo a UUID
- ACID compliance para integridade de dados
- Escalabilidade horizontal
- Comunidade ativa e documentação extensa
- Boa integração com Go via sqlx/GORM
- Suporte a JSON/JSONB para dados flexíveis

**Alternativas Consideradas**:
- **MySQL**: Menor suporte a tipos complexos, menos flexível para dados geográficos
- **MongoDB**: Menos adequado para dados relacionais, complexidade desnecessária
- **SQLite**: Inadequado para produção, sem suporte a concorrência

**Impacto**:
- Necessidade de ORM (sqlx ou GORM)
- Migrações de schema via goose/golang-migrate
- Configuração de connection pooling
- Backup e recuperação via pg_dump/pg_restore

---

## DEC-002 — Autenticação JWT + OAuth

**Data**: 08/09/2026

**Decisão**: Utilizar JWT para sessões e OAuth (Google) para login social.

**Motivo**:
- **JWT**: Stateless, escalável, padrão industry
- **OAuth**: Experiência do usuário, segurança, menos senhas para gerenciar
- **Casbin**: Controle de acesso flexível, RBAC

**Alternativas Consideradas**:
- **Sessões tradicionais**: Mais complexo para APIs, requer armazenamento server-side
- **Apenas OAuth**: Menor flexibilidade, usuários sem Google ficam de fora
- **Apenas JWT**: Menor experiência do usuário

**Impacto**:
- Necessidade de gerenciamento de tokens (access + refresh)
- Integração com Google API
- Configuração de Casbin para roles
- Armazenamento seguro de JWT Secret

---

## DEC-003 — Monolito Modular

**Data**: 08/09/2026

**Decisão**: Arquitetura monolítica modular (não microsserviços).

**Motivo**:
- Projeto acadêmico/TCC
- Menor complexidade operacional
- Facilidade de desenvolvimento e deploy
- Desempenho local (sem overhead de rede entre serviços)
- Equipe pequena (1-2 desenvolvedores)

**Alternativas Consideradas**:
- **Microsserviços**: Complexidade desnecessária para TCC, overhead de operação
- **Monolito sem módulos**: Difícil manutenção, acoplamento alto

**Impacto**:
- Estrutura de pastas clara com separação por responsabilidade
- Facilidade de testes unitários e de integração
- Deploy simplificado
- Escalabilidade futura pode requerer refactoring

---

## DEC-004 — Sem Persistência de Áudio

**Data**: 08/09/2026

**Decisão**: Áudio de conversas NUNCA será persistido em banco de dados.

**Motivo**:
- **Privacidade dos usuários**: Requisito crítico do projeto
- **Conformidade**: LGPD/GDPR
- **Requisito RNF01**: Especificação explícita
- **Redução de custos**: Sem necessidade de armazenamento massivo
- **Simplicidade**: Sem necessidade de gestão de arquivos de áudio

**Alternativas Consideradas**:
- **Armazenamento temporário**: Viola privacidade, risco de vazamento
- **Gravação opcional**: Complexidade desnecessária, aumenta superfície de ataque
- **Armazenamento criptografado**: Ainda persiste dados sensíveis

**Impacto**:
- Buffer apenas em memória (30 segundos)
- Sem histórico de conversas
- Replay de áudio perdido via buffer
- Necessidade de buffer robusto para resiliência

---

## DEC-005 — WebSocket para Áudio

**Data**: 08/09/2026

**Decisão**: Utilizar WebSocket para streaming de áudio em tempo real.

**Motivo**:
- **Baixa latência**: Conexão persistente, sem overhead HTTP
- **Bidirecional**: Cliente e servidor podem enviar
- **Suporte nativo**: Flutter e Go possuem suporte completo
- **Padrão industry**: Usado por Discord, Slack, etc.

**Alternativas Consideradas**:
- **HTTP polling**: Alta latência, desperdício de recursos
- **WebRTC**: Complexidade desnecessária, overkill para PTT
- **gRPC streaming**: Menor suporte mobile, complexidade

**Impacto**:
- Gestão de conexões (hub de conexões)
- Reconexão automática em caso de queda
- Buffer de mensagens para resiliência
- Monitoramento de conexões ativas

---

## DEC-006 — Flutter para Frontend

**Data**: 08/09/2026

**Decisão**: Utilizar Flutter/Dart para o aplicativo mobile.

**Motivo**:
- **Cross-platform**: iOS e Android com mesma codebase
- **Performance**: Compilação nativa
- **Hot reload**: Produtividade elevada
- **Suporte Google**: Comunidade ativa, atualizações frequentes
- **WebSocket**: Suporte nativo

**Alternativas Consideradas**:
- **React Native**: Performance inferior, bridge overhead
- **Nativo (Swift/Kotlin)**: Duas codebases, maior esforço
- **Ionic/Cordova**: Performance inferior, UX ruim

**Impacto**:
- Necessidade de conhecimento em Dart
- Estrutura de pastas feature-based
- Widgets reutilizáveis
- Testes unitários e de widget

---

## DEC-007 — Go para Backend

**Data**: 08/09/2026

**Decisão**: Utilizar Go/Golang para o servidor backend.

**Motivo**:
- **Performance**: Compilação nativa, baixa latência
- **Concurrency**: Goroutines para milhares de conexões WebSocket
- **Simplicidade**: Sintaxe clara, fácil manutenção
- **Deploy**: Binário único, sem dependências
- **WebSocket**: Suporte excelente via gorilla/websocket

**Alternativas Consideradas**:
- **Node.js**: Performance inferior, single-threaded
- **Python**: Performance inferior, GIL
- **Java/Kotlin**: Verboso, overhead de JVM

**Impacto**:
- Necessidade de conhecimento em Go
- Estrutura de pastas cmd/internal/pkg
- Testes unitários nativos
- Deploy simplificado

---

## DEC-008 — Buffer de 30 Segundos

**Data**: 08/09/2026

**Decisão**: Buffer temporário de áudio com duração máxima de 30 segundos.

**Motivo**:
- **Resiliência**: Cobrir quedas curtas de conexão
- **Memória**: 30s é suficiente para reconexão típica
- **Privacidade**: Dados temporários, descartados após uso
- **Performance**: Buffer limitado, sem consumo excessivo de memória

**Alternativas Consideradas**:
- **15 segundos**: Muito curto, pode não cobrir reconexão
- **60 segundos**: Muito longo, consome memória, risco de privacidade
- **Ilimitado**: Risco de privacidade, consumo excessivo

**Impacto**:
- Necessidade de FIFO queue em memória
- Descarte automático de mensagens antigas
- Replay automático ao reconectar
- Monitoramento de uso de memória

---

## DEC-009 — Max 10 Usuários por Canal

**Data**: 08/09/2026

**Decisão**: Limite de 10 usuários por canal de comunicação.

**Motivo**:
- **Experiência**: Canais grandes são confusos para PTT
- **Performance**: Menos conexões por canal = menor latência
- **Controle**: Facilita gestão de quem está falando
- **Escalabilidade**: Canais menores = mais canais, melhor distribuição

**Alternativas Consideradas**:
- **5 usuários**: Muito restritivo, canais demais
- **20 usuários**: Muitos usuários, experiência ruim
- **Ilimitado**: Caótico, impossível gerenciar PTT

**Impacto**:
- Criação automática de novos canais
- Lógica de redistribuição
- Limitação de 15 canais visíveis
- Monitoramento de canais ativos

---

## DEC-010 — Raio de Proximidade 10km

**Data**: 08/09/2026

**Decisão**: Raio de proximidade de 10km para descoberta de canais.

**Motivo**:
- **Urbano**: Cobertura adequada para cidades
- **Rural**: Alcance razoável para áreas abertas
- **Balanceamento**: Nem muito grande (confuso), nem muito pequeno (poucos usuários)
- **GPS**: Precisão adequada para esse raio

**Alternativas Consideradas**:
- **5km**: Muito restritivo, poucos usuários por canal
- **20km**: Muito amplo, muitos canais, experiência ruim
- **50km**: Excessivo, não faz sentido para PX

**Impacto**:
- Cálculo de distância via Haversine
- Atualização de localização a cada 30 segundos
- Filtragem de canais por distância
- Índices geográficos no banco

---

## DEC-011 — Escopo da v1: protótipo em memória

**Data**: 13/10/2026

**Decisão**: A v1 (protótipo funcional entregue para teste em campo) roda **em memória** e **não
implementa** PostgreSQL, Google OAuth, Casbin nem TLS. Esses itens ficam como **trabalho futuro
fora de escopo da v1**.

**Motivo**:
- Priorizar o fluxo principal (auth + PTT + canais + GPS + perfil) e a validação em campo com APK.
- Reduzir escopo e dependências para o prazo do trabalho acadêmico.

**Relação com decisões anteriores**:
- **DEC-001 (PostgreSQL)**: decidido, mas **não implementado na v1**. Nenhum driver de banco no
  `go.mod`; os dados vivem em `map` + `sync.RWMutex` e são zerados a cada reinício. A migração
  `backend/migrations/001_initial.sql` existe mas não é aplicada.
- **DEC-002 (JWT + OAuth)**: JWT **implementado** (access + refresh, HS256); **OAuth (Google) não**.
- **RNF05 (Casbin/TLS)**: **não implementados** na v1 (só JWT, bcrypt e rate limiting).

**Impacto**:
- Divergência assumida entre `docs/arquitetura.md`, `docs/banco-de-dados.md` e o código.
- O mapeamento consolidado das lacunas fica em `PROJECT_STATE.md` →
  *Lacunas / Aderência aos Requisitos*.

**Atualização (20/10/2026)**: PostgreSQL **implementado** na Fase 1 (ver DEC-012) —
stores PG, runner de migrações em `backend/internal/db/migrations/` e persistência
de `users`/`channels`/`user_channels`. O restante do texto acima permanece válido
(OAuth, Casbin e TLS seguem fora do escopo v1; TLS hoje é entregue pelo túnel
ngrok, ver DEC-013).

---

## DEC-012 — Persistência PostgreSQL com stores injetados (DIP)

**Data**: 20/10/2026

**Decisão**: Implementar a persistência da Fase 1 com interfaces `UserStore` e
`ChannelStore`, implementações `PGUserStore`/`PGChannelStore` (driver
`jackc/pgx/v5` via `database/sql`) e `MemoryUserStore`/`MemoryChannelStore`
mantidas para testes, injetadas nos services (injeção de dependência).

**Motivo**:
- Concretizar a DEC-001 sem acoplar o domínio ao driver de banco.
- Testes de unidade rápidos com store em memória; integração real via
  `RADIOPX_TEST_DB=1` no `pg_integration_test.go`.
- Runner de migrações embutido (`go:embed`), idempotente e com falha rápida no boot.

**Alternativas Consideradas**:
- GORM/ORM: mais convenção que o necessário para 3 tabelas.
- sqlc/geração de tipos: sobreposição para o tamanho do modelo.
- Store concreto nos services: acoplaria domínio e infra (violaria DIP).

**Impacto**:
- `internal/user/repository.go` e `internal/channel/repository.go` definem as interfaces; `cmd/server/main.go` escolhe PG (com banco) ou memória (fallback de teste).
- `001_initial.sql` movido para `internal/db/migrations/` + `002_align_models.sql` alinhando os modelos.

---

## DEC-013 — Exposição do backend via ngrok (TLS + CORS wildcard)

**Data**: 27/10/2026

**Decisão**: Expor o backend na porta 8080 por túnel **ngrok** (plano free,
`ngrok http 8080`), com URL HTTPS/WSS por sessão; CORS aceitando
`https://*.ngrok-free.app` e `https://*.ngrok-free.dev` (wildcard por origem,
`rs/cors` v1.10.1) para permitir testes de navegador.

**Motivo**:
- O IP da máquina oscila a cada lease de DHCP e quebrava o APK (Problema nº 4).
- O túnel entrega TLS e um hostname estável **por sessão** sem configurar
  certificados localmente (RNF05 parcial: transporte cifrado de ponta a ponta
  no fluxo do app).
- Wildcard é suportado nativamente pelo `rs/cors` (verificado na fonte).

**Alternativas Consideradas**:
- Cloudflare Tunnel: exige conta/CLI separada, sem vantagem para o protótipo.
- IP estático/reserva DHCP: depende do roteador, não resolve para fora da LAN.
- Deploy em nuvem: custo e escopo maiores que o necessário para o TCC.

**Impacto**:
- URL muda a cada reinício do túnel no plano free — mitigado pela DEC-015
  (URL editável em runtime no app).
- Todo tráfego REST chega com `RemoteAddr` do agente local → ver DEC-014.
- Scripts de teste usam o header `ngrok-skip-browser-warning: 1` para não
  receber a página de aviso (`ERR_NGROK_6024`).

---

## DEC-014 — Rate limit por RemoteAddr mantido atrás do túnel

**Data**: 27/10/2026

**Decisão**: Manter o rate limit (10 req/min em register/login/refresh)
identificando o cliente por `RemoteAddr`, **ignorando `X-Forwarded-For`**, mesmo
com o tráfego vindo do ngrok — aceitando que a cota passa a valer para o
**somatório de todos os clientes** do túnel.

**Motivo**:
- Confiar em XFF sem validar a cadeia permite forjar IP e burlar o limite
  (o mesmo problema do bypass corrigido em 11/10/2026).
- Para o protótipo com poucos testers, cota global é aceitável; um 429
  compartilhado é mais tolerável do que um rate limit contornável.

**Alternativas Consideradas**:
- Confiar no XFF quando `RemoteAddr` for o agente local: quebra se o túnel for
  exposto publicamente sem proxy de confiança — decisão de segurança adiada.
- Autenticação prévia no rate limit: mudaria a ordem das camadas e o escopo.

**Impacto**:
- Registrado como **Problema Conhecido nº 9** em `PROJECT_STATE.md` — reavaliar
  se aparecer 429 em campo (contornos: subir cota ou confiar XFF só do agente local).

---

## DEC-015 — URL do servidor editável em runtime no app (RNF06)

**Data**: 27/10/2026

**Decisão**: A URL da API passa a ser configurável **dentro do app** (RNF06):
semente do build (`--dart-define=API_BASE_URL`) como default, URL salva pelo
usuário em `shared_preferences` com precedência, sheet com **Testar**
(`GET /health`) e **Salvar** acessível no Login e no Perfil. A **sessão é
mantida** ao trocar a URL.

**Motivo**:
- Sem isso, cada nova URL do ngrok exigiria recompilar e reinstalar o APK.
- A sessão segue válida no mesmo banco/secret; se o usuário apontar para outro
  servidor, o fluxo existente de 401/refresh-fail já limpa a sessão.

**Alternativas Consideradas**:
- Só recompilar com `--dart-define`: era o estado anterior (Fase 2) — inviável em demonstração.
- Arquivo de config externo ao lado do APK: complicação extra no Android sem ganho.
- Forçar logout ao trocar URL: derrubava a sessão mesmo no mesmo servidor.

**Impacto**:
- `AppConfig` (runtime) + 3 services lendo `baseUrl` via getter; `main()` carrega
  as preferências antes do `runApp`.
- Fim do Problema Conhecido nº 4 (DHCP) e da pendência de URL por sessão da Fase 2.

## DEC-016 — Replay por usuário via sequence (RF06/RNF04)

**Data**: 27/10/2026

**Decisão**: o buffer de 30s do RF06 e o replay de reconexão são entregues com
(1) relay gravando no **AudioBufferService** (memória, 30s/10MB) e anexando
**sequence** ao payload de áudio, (2) o **cliente solicitando** o catch-up com
`request_replay {room, since_sequence}` após reconectar (stateless, por
usuário) e (3) fila FIFO no app (**AudioQueue**) com teto de ~960KB e sentinel
de fim de segmento. O `BufferManager` permanece fora do fluxo; as flags
`IsReplayed`/`MarkAsReplayed` não são usadas.

**Motivo**:
- `IsReplayed` é **global** por canal: marcaria a mensagem como vista para
  todos, quebrando o replay individual de outro usuário reconectado.
- Replay pedido pelo cliente evita estado de sessão no servidor e só busca o
  que falta (`since_sequence` por cliente).
- FIFO por ordem de chegada atende o RF08 (prioridade por timestamp ficou de fora).
- RNF01/RNF04: o buffer só existe em memória e expira por idade (30s), sem
  persistência de áudio.

**Alternativas Consideradas**:
- `BufferManager.HandleReconnection`/`ConfirmReplay`: exige estado por sessão
  e traria o mesmo problema das flags globais.
- Replay automático sem pedido: enviaria áudio a clientes que não reconectaram
  de um canal em que já estavam atrasados.
- Remover `BufferManager`: mantido como código morto deliberado (documentado)
  para eventual uso futuro, evitando diff grande fora do escopo.

**Impacto**:
- `pkg/websocket/client.go` (relay + case `request_replay` + `safeSend`),
  `internal/voice/buffer_adapter.go` (DIP), injeção em `main.go`/`handler.go`.
- Frontend: `AudioQueue` (reescrita), `voice_screen` (fila + `_lastSequence`
  + replay), `websocket_service.requestReplay`; fixes de disconnect e ociosidade.
- Pendente: validação E2E com 2 clientes (`test-backend.ps1` não cobre WS).
