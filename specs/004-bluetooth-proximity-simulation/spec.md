# Feature Specification: Detecção de Presença RadioPX via Bluetooth (BLE)

**Feature Branch**: `bluetooth` (alias SpecKit: `004-bluetooth-proximity-simulation`)
**Created**: 21/11/2026 | **Revised**: 24/11/2026
**Status**: Draft — escopo revisado para advertise+scan RadioPX↔RadioPX; spike de dependência pendente de validação em hardware
**Input**: Detectar presença de outros aparelhos com RadioPX aberto via BLE. Decisões aprovadas: Service UUID fixo do protocolo RadioPX + identificador efêmero aleatório em memória (sem associação com usuário/conta/dispositivo); spike Android antes da lib definitiva; Android 11- best-effort; RadioPX "aberto" = autenticado e primeiro plano; ao background/sair da área autenticada/encerrar tela responsável, parar advertise+scan e limpar observações em memória.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Detectar outro RadioPX próximo (Priority: P1)

Como usuário autenticado do RadioPX, quero ver "Alguém por perto" quando outro aparelho com RadioPX aberto estiver próximo, para saber que há presença compatível sem identificar pessoas.

**Why this priority**: É o valor central da revisão: presença só conta entre instâncias RadioPX.

**Independent Test**: Dois Androids com RadioPX autenticado em foreground, BT ligado e permissão concedida; um periférico BLE genérico próximo não deve alterar o indicador.

**Acceptance Scenarios**:

1. **Given** dois aparelhos estão autenticados, em foreground, com Bluetooth ligado e permissão concedida, **When** ambos anunciam e escaneiam o Service UUID RadioPX, **Then** cada um exibe "Alguém por perto".
2. **Given** o app está autenticado em foreground com scan ativo e nenhum outro RadioPX compatível é observado, **When** existe apenas BLE genérico por perto, **Then** o indicador exibe "Ninguém por perto".
3. **Given** Bluetooth desligado, sem hardware/permissão, ou falha de consulta, **When** o usuário abre a área autenticada, **Then** o indicador exibe "Indisponível" sem bloquear o app.

---

### User Story 2 - Atualização automática e expiração (Priority: P1)

Como usuário, quero que o indicador mude automaticamente quando outro RadioPX compatível aparecer ou desaparecer, sem reiniciar o app.

**Why this priority**: Demonstração depende de resposta temporal visível.

**Independent Test**: Aproximar/afastar o segundo aparelho RadioPX e observar atualização em até 3s e expiração em 15s.

**Acceptance Scenarios**:

1. **Given** o indicador mostra "Ninguém por perto", **When** um anúncio RadioPX compatível é recebido, **Then** muda para "Alguém por perto" em até 3 segundos.
2. **Given** o indicador mostra "Alguém por perto", **When** nenhum anúncio RadioPX compatível é recebido por 15 segundos, **Then** muda para "Ninguém por perto".
3. **Given** mudanças rápidas de presença, **When** múltiplos eventos chegam, **Then** o estado mais recente consistente vence, sem flicker perceptível.

---

### User Story 3 - Visibilidade global e lifecycle foreground-only (Priority: P2)

Como usuário, quero ver o mesmo indicador em Canais, Falar e Perfil, e que a detecção só ocorra enquanto o RadioPX estiver aberto.

**Why this priority**: Consistência visual e privacidade operacional.

**Independent Test**: Navegar entre abas; enviar app para background/sair da área autenticada e verificar parada de advertise+scan e limpeza de observações.

**Acceptance Scenarios**:

1. **Given** usuário autenticado na navegação principal, **When** alterna entre Canais, Falar e Perfil, **Then** o indicador permanece visível e coerente.
2. **Given** advertise+scan ativos, **When** app vai para background, sai da área autenticada ou fecha a tela responsável, **Then** advertise e scan param e observações em memória são limpas.
3. **Given** app retorna ao foreground autenticado com pré-condições válidas, **When** a tela responsável retoma, **Then** advertise+scan reiniciam e o indicador recalcula.

---

### User Story 4 - Isolamento, privacidade e compatibilidade (Priority: P2)

Como usuário, quero que a presença BLE não identifique pessoas, não persista dados e não afete canais/PTT/WebSocket/áudio/localização.

**Why this priority**: Evita regressão e risco de privacidade.

**Independent Test**: Usar canais/PTT/WebSocket durante mudanças de presença; inspecionar que não há pairing/GATT/persistência/API.

**Acceptance Scenarios**:

1. **Given** canais e comunicação por rede ativos, **When** presença BLE muda, **Then** nenhuma conexão GATT, pareamento, áudio, notificação, persistência ou chamada de backend é iniciada.
2. **Given** permissão negada ou Bluetooth desligado, **When** fluxo inicia, **Then** indicador exibe "Indisponível" e funcionalidades existentes permanecem utilizáveis.
3. **Given** observação RadioPX recebida, **When** processada, **Then** apenas identificador efêmero e timestamp em memória são usados; nenhum dado pessoal é exposto ou persistido.

---

### Edge Cases

- Permissão negada/revogada: "Indisponível", sem bloquear app.
- Bluetooth desligado, adaptador ausente, hardware BLE indisponível ou falha de plugin: "Indisponível".
- Android 11 ou inferior: best-effort; limitações de sistema/localização podem impedir descoberta; documentar no quickstart.
- BLE genérico, beacon, headset, carro etc.: ignorados se não anunciarem Service UUID RadioPX.
- Anúncio RadioPX sem sessão autenticada/foreground: não deve ocorrer; se ocorrer por estado inválido, parar e limpar.
- Mudanças rápidas: último estado consistente vence.
- Identificadores efêmeros e observações: apenas memória, TTL 15s, limpos no dispose/background/logout.
- Fora de escopo: iOS, Web, Windows.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O app MUST exibir indicador visual de proximidade na área principal autenticada, compartilhada por Canais, Falar e Perfil.
- **FR-002**: O indicador MUST exibir "Alguém por perto" somente quando ao menos uma observação RadioPX compatível for recebida nos últimos 15 segundos.
- **FR-003**: O indicador MUST exibir "Ninguém por perto" quando advertise+scan estiverem ativos e não houver observações RadioPX compatíveis recentes.
- **FR-004**: O indicador MUST exibir "Indisponível" quando Bluetooth/permissão/hardware/estado estiver indisponível ou consulta falhar.
- **FR-005**: Na área autenticada em foreground, o sistema MUST consultar pré-condições, solicitar permissão necessária uma vez e iniciar advertise+scan BLE.
- **FR-006**: O sistema MUST atualizar o indicador em até 3 segundos após mudança observada, sem reinício.
- **FR-007**: O sistema MUST NOT solicitar ativação automática do Bluetooth.
- **FR-008**: O estado do adaptador MUST ser apenas pré-condição; nunca sinal direto de presença.
- **FR-009**: O app MUST anunciar e filtrar pelo Service UUID fixo do protocolo RadioPX; BLE genérico MUST ser ignorado.
- **FR-010**: O anúncio MUST usar identificador efêmero aleatório em memória, sem associação com usuário, conta, dispositivo estável, MAC, nome pessoal ou localização.
- **FR-011**: O sistema MUST NOT fazer pareamento, conexão GATT, troca de áudio, identificação pessoal, persistência, backend ou notificação.
- **FR-012**: Ao background, saída da área autenticada, dispose da tela responsável ou fim de sessão, o sistema MUST parar advertise+scan e limpar observações em memória.
- **FR-013**: A funcionalidade MUST NOT alterar canais, PTT, WebSocket, áudio ou permissões/localização existentes.
- **FR-014**: A funcionalidade MUST ser limitada ao Android; Android 11- é best-effort com limitações documentadas.

### Key Entities *(include if feature involves data)*

- **RadioPxPresenceState**: `nearby`, `away`, `unavailable` mapeados para os textos exatos da UI.
- **RadioPxAdvertisement**: anúncio BLE local com Service UUID fixo RadioPX e identificador efêmero aleatório em memória.
- **NearbyRadioPxObservation**: observação remota compatível com `instanceId` efêmero e `lastSeenAt`; TTL 15s; apenas memória.
- **BluetoothAdapterPrecondition**: estado/permissão/hardware; habilita ou bloqueia advertise+scan.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% dos cenários de aceitação passam em dois Androids físicos com RadioPX autenticado em foreground.
- **SC-002**: Presença compatível atualiza em até 3s; ausência expira em 15s sem novos anúncios.
- **SC-003**: BLE genérico não altera o indicador em 100% dos testes de filtro.
- **SC-004**: Em 100% dos ciclos background/saída autenticada/dispose, advertise+scan param e memória de observações é limpa.
- **SC-005**: Nenhum pareamento, GATT, áudio, persistência, backend, ou regressão em canais/PTT/WebSocket/localização.
- **SC-006**: Permissão negada/BT off/hardware indisponível resultam em "Indisponível" sem bloquear o app.

## Assumptions

- A finalidade é demonstração acadêmica de proximidade entre instâncias RadioPX, não rastreamento, distância precisa ou identificação.
- "Aparelho próximo" conta apenas quando anuncia o Service UUID RadioPX; dispositivos sem anúncio compatível não contam.
- A dependência BLE definitiva será escolhida após spike com evidência em dois Androids físicos e auditoria de licença/manutenção/transitivas.
- Nenhum identificador efêmero é persistido, enviado ao backend, logado como dado pessoal ou associado a usuário/conta/dispositivo.
