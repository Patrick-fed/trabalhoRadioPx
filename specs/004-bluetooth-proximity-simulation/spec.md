# Feature Specification: Simulação de Proximidade via Bluetooth

**Feature Branch**: `feature/migracao-painel-adm-para-vue` (alias SpecKit para validação: `004-bluetooth-proximity-simulation`)
**Created**: 21/11/2026
**Status**: Draft
**Input**: User description: "Indicador visual de proximidade simulado pelo estado do adaptador Bluetooth do próprio aparelho, para demonstração da presença ou ausência de alguém próximo, sem descoberta de dispositivos, identificação, distância, conexão, comunicação, notificação ou ações automáticas."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Indicador reflete o estado inicial (Priority: P1)

Como apresentador do RadioPX, quero abrir o aplicativo e ver imediatamente um indicador de proximidade simulada, para demonstrar presença ou ausência de alguém próximo sem precisar realizar nenhuma ação adicional.

**Why this priority**: É o comportamento essencial da demonstração: o indicador precisa ser correto já na abertura do app.

**Independent Test**: Pode ser testado abrindo o app autenticado com o Bluetooth ligado, desligado e indisponível/desconhecido, verificando o texto exibido em cada caso.

**Acceptance Scenarios**:

1. **Given** o dispositivo Android está com Bluetooth ligado, **When** o usuário autenticado abre o aplicativo, **Then** o indicador exibe "Alguém por perto".
2. **Given** o dispositivo Android está com Bluetooth desligado, **When** o usuário autenticado abre o aplicativo, **Then** o indicador exibe "Ninguém por perto".
3. **Given** o estado do Bluetooth não pode ser determinado, é indisponível ou a consulta falha, **When** o usuário autenticado abre o aplicativo, **Then** o indicador exibe "Indisponível" sem bloquear o uso do aplicativo.

---

### User Story 2 - Atualização automática durante o uso (Priority: P1)

Como apresentador, quero que o indicador mude automaticamente quando eu ligar ou desligar o Bluetooth nas configurações do aparelho, para simular a mudança de proximidade sem reiniciar o aplicativo.

**Why this priority**: A apresentação depende da mudança visível em tempo real; reiniciar o app quebraria a demonstração.

**Independent Test**: Pode ser testado com o aplicativo aberto, alternando o Bluetooth externamente e observando a troca do indicador sem reinício.

**Acceptance Scenarios**:

1. **Given** o aplicativo está aberto e o indicador mostra "Ninguém por perto", **When** o Bluetooth é ligado nas configurações do aparelho, **Then** o indicador passa a exibir "Alguém por perto" sem reiniciar o aplicativo.
2. **Given** o aplicativo está aberto e o indicador mostra "Alguém por perto", **When** o Bluetooth é desligado nas configurações do aparelho, **Then** o indicador passa a exibir "Ninguém por perto" sem reiniciar o aplicativo.

---

### User Story 3 - Indicador global na navegação principal (Priority: P2)

Como usuário do aplicativo, quero ver o mesmo indicador em todas as abas principais, para que a simulação continue visível independentemente de onde eu esteja no app.

**Why this priority**: Garante consistência visual durante a apresentação, sem exigir indicadores separados por tela.

**Independent Test**: Pode ser testado alternando entre Canais, Falar e Perfil com o aplicativo aberto e verificando que o indicador permanece visível e coerente.

**Acceptance Scenarios**:

1. **Given** o indicador está visível na área principal do aplicativo, **When** o usuário alterna entre as abas Canais, Falar e Perfil, **Then** o indicador permanece visível em todas elas com o mesmo estado.

---

### User Story 4 - Simulação isolada das comunicações existentes (Priority: P2)

Como usuário, quero que a simulação por Bluetooth não influencie canais, chamadas ou conexões, para que as funcionalidades atuais do RadioPX continuem funcionando normalmente.

**Why this priority**: A feature é apenas uma simulação visual; qualquer efeito colateral em comunicação comprometeria as demais funcionalidades.

**Independent Test**: Pode ser testado alternando o Bluetooth enquanto o app está conectado aos fluxos existentes e verificando que canais, WebSocket e PTT permanecem inalterados.

**Acceptance Scenarios**:

1. **Given** o aplicativo está usando canais e comunicação por rede, **When** o estado do Bluetooth muda, **Then** nenhuma descoberta, conexão, transmissão, notificação ou ação automática é iniciada.
2. **Given** o indicador muda de estado, **When** o usuário usa as funcionalidades existentes, **Then** a comunicação por Wi-Fi e WebSocket permanece inalterada.

---

### Edge Cases

- Consulta ao estado do Bluetooth falha ou retorna valor desconhecido: o indicador deve exibir "Indisponível".
- Dispositivo sem adaptador Bluetooth utilizável: o indicador deve exibir "Indisponível".
- Permissão necessária não concedida: o indicador deve exibir "Indisponível", sem exibir mensagem bloqueante ou solicitar automaticamente que o usuário ligue o Bluetooth.
- Mudanças rápidas de estado: o indicador deve refletir o estado observado mais recente.
- Estados transitórios do adaptador: se não puderem ser mapeados para um dos três textos definidos, devem ser tratados como "Indisponível".
- A funcionalidade não é aplicável a iOS, Web ou Windows.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O aplicativo MUST exibir um indicador visual de proximidade simulada na área principal autenticada, compartilhada pelas abas Canais, Falar e Perfil.
- **FR-002**: O indicador MUST exibir exatamente "Alguém por perto" quando o adaptador Bluetooth estiver ligado.
- **FR-003**: O indicador MUST exibir exatamente "Ninguém por perto" quando o adaptador Bluetooth estiver desligado.
- **FR-004**: O indicador MUST exibir exatamente "Indisponível" quando o Bluetooth estiver indisponível, o estado for desconhecido ou a consulta falhar.
- **FR-005**: Ao abrir o aplicativo, o sistema MUST consultar o estado atual do Bluetooth e exibir o indicador correspondente.
- **FR-006**: Enquanto o aplicativo estiver aberto, o sistema MUST acompanhar mudanças no estado do Bluetooth e atualizar o indicador sem exigir reinício.
- **FR-007**: O sistema MUST NOT solicitar automaticamente que o usuário ligue o Bluetooth nem tentar ligá-lo por conta própria.
- **FR-008**: O estado do Bluetooth MUST ser usado somente para atualizar o indicador visual.
- **FR-009**: A funcionalidade MUST NOT realizar descoberta de dispositivos, identificação de usuários ou aparelhos, medição de distância, pareamento, conexão Bluetooth/BLE, comunicação entre dispositivos, transmissão de áudio, notificações ou ações automáticas.
- **FR-010**: A funcionalidade MUST NOT alterar a comunicação existente por Wi-Fi e WebSocket.
- **FR-011**: A funcionalidade MUST ser limitada ao Android; suporte para iOS, Web ou Windows está fora de escopo.

### Key Entities *(include if feature involves data)*

- **Indicador de Proximidade Simulada**: elemento visual com três estados observáveis — presença simulada, ausência simulada e indisponível — associados aos textos "Alguém por perto", "Ninguém por perto" e "Indisponível".
- **Estado do adaptador Bluetooth**: entrada externa e temporária, usada apenas para determinar qual dos três estados do indicador deve ser exibido; não é persistida nem enviada ao backend.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% dos sete cenários de aceitação listados são verificados com sucesso em um dispositivo Android.
- **SC-002**: Após o usuário ligar ou desligar o Bluetooth com o aplicativo aberto, o indicador atualiza em até 3 segundos, sem reinício.
- **SC-003**: Em 100% das verificações de navegação, o indicador permanece visível e consistente nas abas Canais, Falar e Perfil.
- **SC-004**: Nenhuma funcionalidade existente de canais, PTT ou WebSocket apresenta regressão quando o estado do Bluetooth é alterado.
- **SC-005**: Nenhuma ação automática de Bluetooth, notificação, conexão, descoberta ou transmissão é iniciada pela mudança do indicador.
- **SC-006**: Em 100% dos cenários de falha ou estado desconhecido, o aplicativo exibe "Indisponível" e continua utilizável.

## Assumptions

- A funcionalidade existe para fins de demonstração acadêmica: ligar/desligar o Bluetooth representa presença/ausência simulada, não uma medição real de proximidade.
- O público-alvo desta versão é o aplicativo Android autenticado; o indicador aparece na área principal após o login.
- A biblioteca `bluetooth_manager 2.0.1` foi pré-selecionada e teve compatibilidade com o ambiente local verificada; a definição das permissões mínimas fica para a etapa de planejamento.
- Nenhuma informação de estado do Bluetooth é persistida, enviada ao backend ou usada para identificar usuários.
- Permissões de descoberta e anúncio de dispositivos são evitadas por estarem fora do escopo.
