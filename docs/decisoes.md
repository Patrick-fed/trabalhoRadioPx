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
