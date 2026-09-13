# Feature Specification: Digital PX Radio Simulation

**Feature Branch**: `001-px-radio-simulation`  
**Created**: 2026-09-08  
**Status**: Draft  
**Input**: User description: "Aplicativo digital que simule um rádio PX para celular"

## User Scenarios & Testing

### User Story 1 - Push-to-Talk Communication (Priority: P1)

Como usuário do RadioPX, eu quero pressionar um botão para falar e soltar para ouvir, para que eu possa me comunicar com outros usuários no mesmo canal como em um rádio PX tradicional.

**Por que essa prioridade**: Funcionalidade principal do sistema - sem ela o app não existe

**Teste Independente**: Pode ser testado com dois usuários no mesmo canal verificando transmissão e recepção de áudio

**Cenários de Aceitação**:

1. **Dado** que o usuário está em um canal, **Quando** pressiona o botão PTT, **Então** o áudio é capturado e transmitido para outros usuários do canal
2. **Dado** que outro usuário está transmitindo, **Quando** o usuário tenta falar, **Então** recebe indicação de que o canal está ocupado
3. **Dado** que o usuário está ouvindo, **Quando** o transmissor solta o botão, **Então** a transmissão é encerrada e o canal fica disponível

---

### User Story 2 - Proximity-Based Discovery (Priority: P2)

Como usuário do RadioPX, eu quero que o sistema descubra automaticamente canais baseado na minha localização GPS, para que eu possa me comunicar com pessoas próximas (até 10km).

**Por que essa prioridade**: Funcionalidade essencial para o modelo de negócio baseado em proximidade

**Teste Independente**: Pode ser testado movendo-se fisicamente e verificando se canais são atualizados

**Cenários de Aceitação**:

1. **Dado** que o usuário está em uma área com outros usuários, **Quando** abre o app, **Então** canais disponíveis dentro de 10km são exibidos
2. **Dado** que o usuário está a 5km de outro grupo, **Quando** a localização é atualizada, **Então** canais da nova área aparecem
3. **Dado** que o usuário está a 15km de um grupo, **Quando** busca canais, **Então** não exibe canais fora do raio de 10km

---

### User Story 3 - Dynamic Channel Allocation (Priority: P3)

Como administrador do sistema, eu quero que canais sejam criados automaticamente a cada 10 usuários na mesma área, para que a comunicação seja organizada e eficiente.

**Por que essa prioridade**: Gerencia a escala e evita sobrecarga de canais

**Teste Independente**: Pode ser testado adicionando usuários e verificando criação automática de canais

**Cenários de Aceitação**:

1. **Dado** que há 9 usuários em uma área, **Quando** um novo usuário entra, **Então** um novo canal é criado automaticamente
2. **Dado** que um canal tem 10 usuários, **Quando** outro usuário tenta entrar, **Então** é redirecionado para um novo canal
3. **Dado** que o usuário está em um canal, **Quando** visualiza a lista, **Então** vê no máximo 15 canais

---

### User Story 4 - Connection Resilience (Priority: P4)

Como usuário do RadioPX, eu quero que mensagens sejam armazenadas temporariamente quando a conexão cai, para que eu não perca áudio importante durante oscilações de rede.

**Por que essa prioridade**: Garante experiência contínua mesmo com instabilidade de rede

**Teste Independente**: Pode ser testado simulando queda de conexão e verificando replay de mensagens

**Cenários de Aceitação**:

1. **Dado** que o usuário está ouvindo, **Quando** a conexão cai, **Então** o áudio recebido é armazenado em buffer temporário
2. **Dado** que há áudio em buffer, **Quando** a conexão é restaurada, **Então** as mensagens são reproduzidas automaticamente
3. **Dado** que o buffer está cheio (30 segundos), **Quando** nova mensagem chega, **Então** a mensagem mais antiga é descartada

---

### Edge Cases

- O que acontece quando o usuário nega permissão de localização?
- Como o sistema lida com múltiplos dispositivos no mesmo usuário?
- O que acontece quando o buffer de áudio está cheio durante desconexão?
- Como o sistema prioriza áudio quando há múltiplos transmissores tentando falar?

## Requirements

### Functional Requirements

- **FR-001**: Sistema DEVE suportar transmissão de áudio push-to-talk
- **FR-002**: Sistema DEVE impor um falante por vez por canal
- **FR-003**: Sistema DEVE calcular proximidade usando coordenadas GPS
- **FR-004**: Sistema DEVE limitar alcance de comunicação para raio de 10km
- **FR-005**: Sistema DEVE alocar dinamicamente usuários em canais (máximo 10 por canal)
- **FR-006**: Sistema DEVE exibir máximo de 15 canais para o usuário
- **FR-007**: Sistema NÃO DEVE persistir áudio de conversas
- **FR-008**: Sistema DEVE fornecer buffer temporário para quedas de conexão
- **FR-009**: Sistema DEVE reproduzir mensagens armazenadas quando conexão restaurada
- **FR-010**: Sistema DEVE suportar plataformas móveis (iOS/Android)
- **FR-011**: Sistema DEVE autenticar usuários via JWT
- **FR-012**: Sistema DEVE suportar login com Google OAuth
- **FR-013**: Sistema DEVE usar PostgreSQL para dados permanentes

### Key Entities

- **User**: Usuário do app mobile com localização GPS
- **Channel**: Grupo dinâmico de até 10 usuários
- **Transmission**: Pacote de áudio com timestamp e referência do usuário
- **MessageBuffer**: Armazenamento temporário para resiliência de conexão

## Success Criteria

### Measurable Outcomes

- **SC-001**: Usuários podem iniciar transmissão push-to-talk em até 2 segundos
- **SC-002**: Latência de áudio entre falantes < 500ms
- **SC-003**: Precisão de detecção de proximidade dentro de 100m
- **SC-004**: Alocação de canal completa em até 5 segundos após entrada
- **SC-005**: Mensagens armazenadas são reproduzidas em até 10 segundos após restauração
- **SC-006**: Nenhum áudio de conversa persiste após sessão encerrada
- **SC-007**: Sistema suporta 10.000 usuários concorrentes por região

## Assumptions

- Usuários possuem dispositivos móveis com GPS habilitado
- Conectividade de rede está disponível (com quedas intermitentes)
- Qualidade de áudio é aceitável para comunicação por voz
- Usuários consentem compartilhamento de localização para recursos de proximidade
- Plataformas móveis: iOS 15+ e Android 10+
- PostgreSQL está disponível para armazenamento de dados permanentes
- Google OAuth está configurado para login social
- Casbin é usado para controle de acesso baseado em papéis

## Clarifications

### Session 2026-09-08

- Q: Qual deve ser o codec de áudio utilizado? → A: Opus (recomendado para voz, baixa latência)
- Q: Deve haver suporte a múltiplos idiomas? → A: Não, apenas português brasileiro para v1
- Q: Como o sistema deve lidar com usuários offline? → A: Buffer temporário com replay ao reconectar
