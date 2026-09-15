# Política de Gerenciamento de Dependências

A partir deste momento, trate o gerenciamento de dependências como uma atividade crítica do projeto.

O objetivo é manter o projeto seguro, previsível, reproduzível e com o menor número possível de dependências externas.

## 1. Regra principal

NÃO adicione, remova, atualize, substitua ou troque uma dependência sem antes analisar a necessidade e o impacto da alteração.

Antes de introduzir qualquer biblioteca ou pacote novo, responda internamente:

1. Essa dependência é realmente necessária?
2. A funcionalidade pode ser implementada utilizando recursos já existentes no projeto?
3. Já existe uma dependência instalada que resolve o mesmo problema?
4. A biblioteca é mantida atualmente?
5. Ela possui histórico relevante de vulnerabilidades?
6. A licença é compatível com o projeto?
7. Ela possui documentação oficial adequada?
8. Ela possui comunidade/ecossistema suficientemente confiável?
9. A dependência adicionará complexidade desnecessária?
10. Existe uma alternativa mais simples e madura?

Se a dependência não for necessária, NÃO a instale.

---

## 2. Bibliotecas desconhecidas

Não utilize uma biblioteca apenas porque ela apareceu em um exemplo, tutorial, resposta de IA, Stack Overflow, GitHub ou documentação de terceiros.

Antes de utilizá-la, faça uma verificação apropriada utilizando fontes confiáveis, priorizando:

* documentação oficial do projeto;
* repositório oficial;
* registro oficial do gerenciador de pacotes;
* advisories/bancos de dados de vulnerabilidades;
* informações oficiais sobre versões e manutenção.

Não considere uma biblioteca confiável simplesmente porque possui muitas estrelas ou downloads.

---

## 3. Segurança

Antes de adicionar uma dependência, verificar, quando possível:

* vulnerabilidades conhecidas;
* versão atual e versões anteriores relevantes;
* data da última atualização;
* atividade de manutenção;
* existência de mantenedores;
* dependências transitivas;
* histórico de problemas de segurança;
* compatibilidade com a versão atual do projeto.

Se houver vulnerabilidade conhecida relevante na versão que seria instalada, NÃO instalar essa versão sem uma justificativa explícita.

Se a vulnerabilidade puder ser evitada utilizando uma versão corrigida ou outra biblioteca, prefira a alternativa segura.

Nunca ignorar uma vulnerabilidade apenas para fazer o código funcionar.

---

## 4. Versões devem ser previsíveis

Evite dependências com versões flutuantes quando o gerenciador utilizado permitir especificar uma versão exata.

Prefira versões explicitamente fixadas, por exemplo:

"biblioteca": "1.4.2"

em vez de permitir atualizações automáticas através de intervalos de versão, quando isso puder ser evitado.

ATENÇÃO:

Não altere essa regra de forma cega. Respeite o funcionamento e as convenções do gerenciador de pacotes utilizado pelo projeto.

Além do arquivo de configuração das dependências, mantenha o arquivo de lock correspondente atualizado e versionado no repositório quando aplicável.

O objetivo é garantir que outra máquina consiga instalar as mesmas versões utilizadas no desenvolvimento.

---

## 5. Não atualizar dependências automaticamente

Ao executar uma tarefa que não esteja relacionada à atualização de dependências:

NÃO atualize pacotes simplesmente porque existe uma versão mais recente.

Exemplo:

Se a tarefa for:

"Corrigir o cadastro de usuários."

Não faça automaticamente:

* atualização do framework;
* atualização do ORM;
* atualização do banco;
* atualização de bibliotecas não relacionadas;
* atualização geral do projeto.

Alterações de dependências devem permanecer limitadas ao escopo da tarefa.

Se uma atualização for necessária para corrigir uma vulnerabilidade, incompatibilidade ou erro, informe claramente o motivo.

---

## 6. Não remover dependências sem análise

Antes de remover uma dependência:

1. Pesquise todas as referências à biblioteca.
2. Verifique se ela é utilizada direta ou indiretamente pelo código.
3. Verifique scripts, configurações e ferramentas de desenvolvimento.
4. Execute os testes relevantes.
5. Só então remova a dependência.

Não remova um pacote simplesmente porque ele não aparece em um único arquivo.

---

## 7. Evitar dependências duplicadas

Antes de adicionar uma biblioteca, procure no projeto se já existe outra dependência com função semelhante.

Evite manter duas bibliotecas diferentes para resolver o mesmo problema sem uma justificativa técnica.

Exemplo:

Se já existe uma biblioteca de validação utilizada pelo projeto, não adicione outra biblioteca de validação apenas porque ela possui uma API diferente.

Priorize consistência.

---

## 8. Dependências transitivas

Ao adicionar uma nova dependência, considere também as dependências que ela adicionará indiretamente.

Uma biblioteca pequena que introduz dezenas de dependências pode aumentar desnecessariamente:

* superfície de ataque;
* tamanho da aplicação;
* complexidade;
* possibilidade de conflitos;
* quantidade de vulnerabilidades futuras.

Quando duas soluções forem tecnicamente adequadas, considere a quantidade e a qualidade das dependências transitivas como um dos fatores da decisão.

---

## 9. Licença

Antes de adicionar uma dependência, verificar sua licença quando essa informação estiver disponível.

Não adicionar bibliotecas cuja licença seja incompatível ou gere dúvida relevante sobre o uso no projeto sem primeiro informar o problema.

Registrar a licença da dependência quando isso for relevante para a documentação do projeto.

---

## 10. Não substituir bibliotecas por preferência pessoal

Não substitua uma biblioteca existente apenas porque existe outra considerada "melhor", "mais moderna" ou "mais popular".

Uma substituição deve possuir motivo técnico claro, como:

* vulnerabilidade;
* abandono do projeto;
* incompatibilidade;
* necessidade funcional;
* problema de desempenho comprovado;
* problema de manutenção;
* requisito arquitetural.

Não faça migrações desnecessárias durante tarefas não relacionadas.

---

## 11. Registro das alterações

Sempre que uma nova dependência for adicionada, removida ou substituída, registre:

* nome;
* versão;
* finalidade;
* motivo da alteração;
* alternativa considerada, quando relevante;
* impacto esperado;
* informações de segurança relevantes;
* data da alteração.

Atualize o histórico de alterações do projeto quando a mudança for relevante.

---

## 12. Comandos de instalação

Antes de executar um comando que altere dependências, confirme qual gerenciador de pacotes o projeto utiliza.

Não misture gerenciadores sem necessidade.

Exemplos:

* npm;
* pnpm;
* yarn;
* pip;
* poetry;
* cargo;
* Maven;
* Gradle;
* outro definido pela arquitetura do projeto.

Respeite os arquivos de lock e as convenções já existentes.

Não crie um novo arquivo de lock utilizando outro gerenciador sem necessidade.

---

## 13. Não utilizar comandos destrutivos sem necessidade

Evite comandos que removam ou recriem grande parte das dependências quando uma alteração localizada for suficiente.

Antes de executar comandos potencialmente destrutivos, entenda seu efeito e verifique se são realmente necessários.

Não apague arquivos de configuração ou lockfiles simplesmente para "resolver" um problema sem investigar sua causa.

---

## 14. Validação após alterações

Depois de alterar dependências:

1. Instale/resolva as dependências.
2. Verifique conflitos.
3. Execute o build.
4. Execute lint, quando existir.
5. Execute os testes disponíveis.
6. Verifique se a aplicação continua funcionando.
7. Verifique alterações inesperadas no lockfile.
8. Informe qualquer problema encontrado.

Não considere a alteração concluída apenas porque o comando de instalação terminou sem erro.

---

## 15. Princípio da menor dependência

Prefira:

"resolver com o que o projeto já possui"

antes de:

"adicionar uma nova biblioteca".

E prefira:

"biblioteca pequena, madura e bem mantida"

antes de:

"biblioteca enorme que resolve dezenas de problemas que não precisamos".

Não adicione dependências apenas para economizar algumas linhas de código.

---

## 16. Quando houver dúvida

Se houver dúvida significativa sobre a segurança, manutenção, licença, compatibilidade ou necessidade de uma dependência:

NÃO tome uma decisão arbitrária.

Explique:

* qual é a dúvida;
* qual é o risco;
* quais alternativas existem;
* qual informação está faltando.

Quando não houver informação suficiente para tomar uma decisão segura, pare a alteração relacionada à dependência e solicite confirmação.

---

## 17. Regra contra "overengineering"

Não introduza bibliotecas, frameworks ou ferramentas apenas para tornar uma implementação aparentemente mais sofisticada.

O projeto deve utilizar a menor quantidade razoável de tecnologias necessárias para cumprir seus requisitos.

A complexidade deve ser justificada pela necessidade do sistema, não pela capacidade da ferramenta.

---

## 18. Princípio final

Considere qualquer dependência externa como uma nova peça permanente do projeto.

Antes de adicioná-la, pergunte:

"Se eu adicionar isso hoje, estarei disposto a manter, atualizar, testar e proteger essa dependência durante toda a vida do projeto?"

Se a resposta for não, procure uma solução melhor.

Nunca invente informações sobre uma biblioteca.

Nunca declare uma biblioteca segura sem evidências suficientes.

Nunca atualize dependências sem necessidade.

Nunca introduza uma dependência apenas porque ela parece conveniente.

Quando houver conflito entre velocidade de implementação e segurança/manutenibilidade, priorize a solução segura, simples, justificável e reproduzível.