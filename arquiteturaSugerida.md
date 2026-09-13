
### Arquitetura inicial

```text
                         ┌──────────────────────┐
                         │      FLUTTER         │
                         │      Mobile App      │
                         │                      │
                         │ • Login/Cadastro     │
                         │ • Canais             │
                         │ • Push-to-Talk       │
                         │ • Microfone          │
                         │ • Reprodução de áudio│
                         │ • Localização        │
                         │ • Fila/Buffer        │
                         └──────────┬───────────┘
                                    │
                             HTTPS / WebSocket
                                    │
                                    ▼
                    ┌─────────────────────────────┐
                    │         GO BACKEND           │
                    │                              │
                    │      API / WebSocket         │
                    └──────────────┬───────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
      ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
      │ Auth Service  │     │ Channel      │     │ Voice        │
      │               │     │ Service      │     │ Service      │
      │ Login         │     │              │     │              │
      │ Cadastro      │     │ Criar canal  │     │ PTT          │
      │ Usuário       │     │ Entrar/Sair  │     │ Áudio em     │
      └──────────────┘     │ Usuários     │     │ tempo real   │
                           └──────────────┘     └──────┬───────┘
                                                       │
                                                       ▼
                                             ┌──────────────────┐
                                             │ Conexões WebSocket│
                                             │ dos usuários      │
                                             └──────────────────┘
```

### Como eu dividiria o backend Go

Eu **não faria vários microsserviços de verdade agora**. Para um TCC/trabalho de faculdade, começaria com um **monólito modular em Go**.

Algo nessa ideia:

```text
backend/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── user/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── channel/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── voice/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── websocket.go
│   │
│   ├── location/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   └── middleware/
│
├── pkg/
│   ├── websocket/
│   └── audio/
│
└── go.mod
```

A ideia é que cada módulo tenha uma responsabilidade clara.

### 🎙️ Parte mais importante: comunicação por voz

O fluxo principal seria:

```text
Motorista A
    │
    │ aperta PTT
    ▼
Flutter
    │
    │ áudio
    ▼
WebSocket
    │
    ▼
Go Backend
    │
    │ verifica canal
    │ verifica se está ocupado
    ▼
Canal
    │
    │ transmite
    ▼
WebSocket
    │
 ┌──┴───────────┐
 ▼              ▼
Motorista B   Motorista C
```

O ponto importante é que o backend **não precisa salvar a mensagem de voz em banco**.

Isso conversa diretamente com o RNF04, que determina que não deve existir histórico permanente das mensagens de voz. 

O áudio pode existir **somente em memória durante a transmissão** e ser descartado depois.

---

## 🗄️ Banco de dados

Aqui eu separaria:

**Dados permanentes:**

```text
Usuário
 ├── ID
 ├── Nome
 ├── Email
 ├── CPF
 └── Telefone

Canal
 ├── ID
 ├── Nome
 └── ...
```

**Dados que NÃO devem ser persistidos:**

```text
❌ Áudio
❌ Histórico de mensagens
❌ Gravações
```

Ou seja:

```text
                  BANCO
                    │
             ┌──────┴──────┐
             ▼             ▼
          Usuários       Canais
             │
             │
             X
          Áudio ❌
```

---

## 📱 E o Flutter?

No Flutter, eu também separaria por responsabilidades:

```text
lib/
│
├── core/
│   ├── network/
│   ├── websocket/
│   └── permissions/
│
├── features/
│   ├── auth/
│   │
│   ├── channels/
│   │
│   ├── voice/
│   │
│   ├── location/
│   │
│   └── profile/
│
└── main.dart
```

A parte de **voz** ficaria responsável por:

```text
Microfone
   ↓
Captura do áudio
   ↓
Codificação/compressão
   ↓
WebSocket
   ↓
Backend
   ↓
WebSocket
   ↓
Decodificação
   ↓
Fila de áudio
   ↓
Alto-falante
```

Isso também encaixa com o RF08, que prevê uma fila para ordenar os áudios, reproduzir um por vez e remover aquilo que já foi reproduzido. 

### 🔌 E quando cair a internet?

Esse é um ponto **muito importante nesse projeto**, porque vocês mesmos colocaram conectividade e reconexão como requisito. 

Eu faria:

```text
             Internet
                │
        ┌───────┴───────┐
        │               │
     Conectado       Caiu ❌
        │               │
        ▼               ▼
    WebSocket       Reconexão
                        │
                   tenta novamente
                        │
                  ┌─────┴─────┐
                  │           │
                OK          Falhou
                  │           │
                  ▼           ▼
              continua     tenta novamente
```

Mas tem uma decisão arquitetural importante aqui:

**não transformar o áudio perdido em uma gravação permanente.**

O buffer deve existir apenas temporariamente para lidar com oscilações, e depois ser descartado, conforme o RNF02. 

---

# 🧠 Então, resumindo a arquitetura

Eu colocaria no documento do trabalho algo próximo desta estrutura:

```text
                    ┌─────────────────┐
                    │   Flutter App   │
                    │                 │
                    │ Auth            │
                    │ Channels        │
                    │ Push-to-Talk     │
                    │ Audio            │
                    │ Location        │
                    └────────┬────────┘
                             │
                     HTTPS / WebSocket
                             │
                             ▼
                    ┌─────────────────┐
                    │   Go Backend    │
                    │                 │
                    │ Auth            │
                    │ Users           │
                    │ Channels        │
                    │ Voice / PTT     │
                    │ Location        │
                    └───────┬─────────┘
                            │
                 ┌──────────┴──────────┐
                 ▼                     ▼
          ┌──────────────┐      ┌──────────────┐
          │   Database   │      │  WebSocket   │
          │              │      │   Sessions   │
          │ Users        │      │              │
          │ Channels     │      │ Audio        │
          └──────────────┘      └──────────────┘
```

**Tecnologias principais:**

* **Frontend:** Flutter/Dart
* **Backend:** Go/Golang
* **Comunicação:** REST/HTTP para operações tradicionais
* **Tempo real:** WebSocket
* **Banco:** PostgreSQL
* **Áudio:** transmissão em fluxo, sem persistência
* **Autenticação:** JWT/OAuth (Google)/Casbin
* **Localização:** GPS do dispositivo
* **Infraestrutura:** inicialmente um único backend Go