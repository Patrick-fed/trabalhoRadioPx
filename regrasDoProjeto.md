/projeto
│
├── README.md
│   └── Visão geral do sistema
│
├── AGENTS.md
│   └── Regras que o OpenCode deve seguir
│
├── docs/
│   ├── requisitos.md
│   ├── arquitetura.md
│   ├── banco-de-dados.md
│   └── decisoes.md
│
└── CHANGELOG.md
    └── Histórico das alterações


1. AGENTS.md 
Esse provavelmente seria mais importante que o README.

Nele colocamos regras como:
Você está trabalhando em um projeto acadêmico.
Não altere arquivos sem entender a arquitetura existente.
Não crie funcionalidades que não estejam nos requisitos.
Antes de implementar algo grande, explique o que pretende fazer.
Após alterações importantes, atualize a documentação correspondente.

Também podemos colocar:

tecnologias utilizadas;
estrutura das pastas;
padrão de código;
como rodar o projeto;
regras do banco;
regras de API;
como tratar autenticação;
coisas que não devem ser alteradas.

2. README.md → apresentação e visão geral

Aqui eu deixaria algo mais estável:

# Nome do projeto

## Descrição
...

## Objetivo
...

## Tecnologias
...

## Como executar
...

## Estrutura do projeto
...

## Status atual
...

Ele pode ser atualizado quando houver mudanças relevantes, mas não precisa registrar cada alteração minúscula.

3. CHANGELOG.md → diário das alterações

Aqui entra exatamente a sua ideia.

Por exemplo:

# Histórico do Projeto

## 08/09/2026

### Alterações
- Criado sistema inicial de autenticação.
- Criada estrutura da API.
- Criada tabela de usuários.

### Motivo
Implementação dos requisitos RF01 e RF02.

### Próximo passo
Implementar recuperação de senha.

Isso é MUITO útil.

Porque se amanhã você perguntar:

"OpenCode, onde paramos?"

ele pode consultar o histórico.

4. docs/requisitos.md → nossa fonte da verdade

Esse aqui eu considero obrigatório para o nosso projeto.

Colocaríamos algo tipo:

# Requisitos Funcionais

RF01 - Criação de conta
RF02 - Login
RF03 - Perfil do usuário
...

# Requisitos Não Funcionais

RNF01 - Segurança
RNF02 - Desempenho
...

E futuramente podemos relacionar:

RF01
↓
Backend
↓
Banco de dados
↓
Frontend
↓
Testes


5. docs/decisoes.md → MUITO interessante
Toda vez que tomarmos uma decisão arquitetural importante:

# Decisões Técnicas

## DEC-001 — Banco PostgreSQL

Data: 08/09/2026

Decisão:
Utilizar PostgreSQL como banco de dados.

Motivo:
...

Alternativas consideradas:
...

Impacto:
...

Isso cria uma memória de por que fizemos determinada coisa.


PROJECT_STATE.md


Exemplo:

# Estado Atual do Projeto

## Fase
Desenvolvimento do backend

## Requisitos concluídos
- RF01
- RF02

## Em desenvolvimento
- RF03

## Pendentes
- RF04
- RF05
- RF06
- RF07
- RF08

## Última alteração
Implementação inicial da autenticação.

## Próxima tarefa
Criar estrutura de perfil do usuário.

## Problemas conhecidos
Nenhum.

## Última atualização
08/09/2026

Aí temos:

README → "O que é o projeto?"

AGENTS → "Como o OpenCode deve trabalhar?"

Requisitos → "O que o sistema precisa fazer?"

Decisões → "Por que fizemos assim?"

CHANGELOG → "O que mudou?"

PROJECT_STATE → "Onde estamos agora?"