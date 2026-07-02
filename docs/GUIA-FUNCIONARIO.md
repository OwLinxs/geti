# SIGE-TI — Guia rápido do funcionário

Como cadastrar equipamentos e registrar movimentações. Leva 5 minutos.

---

## 1. Entrar no sistema

1. Abra o navegador e acesse o endereço do sistema (peça ao administrador — algo como `http://192.168.0.x`).
2. Informe seu **e-mail** e **senha**.
3. Na primeira vez, troque a senha se o administrador pedir.

> Esqueceu a senha? Peça ao administrador para redefinir (menu **Usuários**).

---

## 2. Cadastrar um item (equipamento ou material)

Menu **Itens → Novo item**.

Há dois tipos, e o formulário se ajusta conforme a **categoria** escolhida:

| Tipo | Exemplos | O que preencher |
|------|----------|-----------------|
| **Patrimoniado** (controle unitário) | Notebook, Monitor, Impressora | Categoria + **Número de patrimônio** (obrigatório), marca, modelo, nº de série, estado, setor, responsável |
| **Consumo** (controle por quantidade) | Toner, Cabo de rede, Mouse | Categoria + **Quantidade** + **Estoque mínimo** (dispara alerta quando acaba) |

Campos principais:

- **Descrição** — nome claro do item. Ex.: `Notebook Dell Latitude 5440`.
- **Categoria** — escolha na lista. Se faltar, peça ao administrador para criar.
- **Número de patrimônio** — a plaquinha do bem (só para patrimoniados).
- **Estado de conservação** — novo, bom, regular ou inservível.
- **Setor / Responsável** — onde está e quem responde pelo item (opcional).

Clique em **Salvar**. Pronto.

---

## 3. Cadastrar MUITOS itens de uma vez (CSV)

Menu **Itens → Importar CSV**.

1. Clique em **Baixar modelo** — abre uma planilha com as colunas certas.
2. Preencha uma linha por item (no Excel/LibreOffice, salve como **CSV**).
3. Volte, **escolha o arquivo**. O sistema mostra uma **prévia**: quantas linhas estão OK e quais têm erro (com o motivo em cada linha).
4. Corrija os erros no arquivo, se houver, e reenvie.
5. Clique em **Importar** — só as linhas válidas são gravadas.

> Categoria e setor são reconhecidos pelo **nome** (precisa já existir no sistema). O responsável, pela **matrícula**.

---

## 4. Registrar uma movimentação (entrada / saída / baixa)

Menu **Movimentações → Nova movimentação**.

1. **Item** — escolha o item.
2. **Tipo**:
   - *Entradas*: compra, doação, devolução (aumentam o estoque).
   - *Saídas*: empréstimo, transferência, descarte (reduzem o estoque).
3. **Quantidade**.
4. **Setor de origem/destino** e **Servidor responsável** (opcionais).
   - O responsável não está cadastrado? Clique em **Novo** ao lado do campo e cadastre na hora (nome + matrícula).
5. Se for **baixa/descarte**, informe o **motivo**.
6. **Registrar**.

O estoque é recalculado automaticamente. Se ficar abaixo do mínimo, aparece um **alerta**.

---

## 5. Emitir termo de responsabilidade

Menu **Termos → Novo termo**. Escolha o item e o servidor. O sistema gera um **PDF** com o nome de quem recebe e a linha de assinatura — imprima e colha a assinatura.

---

## Dúvidas frequentes

- **Cadastrei errado, e agora?** Se o item ainda não tem histórico, o administrador pode **excluir**. Se já tem movimentações, use **baixa** (não some do sistema, mantém o rastro).
- **Onde vejo o que está acabando?** Menu **Alertas de Estoque**.
- **Preciso de relatório?** Menu **Relatórios** — exporta CSV/PDF por setor, responsável, inventário etc.

---

Em caso de erro na tela, anote a mensagem e avise o Departamento de T.I.
