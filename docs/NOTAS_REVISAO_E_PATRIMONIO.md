# Notas de revisão e evolução do patrimônio

Registro de contexto para as próximas mudanças do SIGE-TI (09/10/2026).

## Diagnóstico da revisão

- O Nginx encaminha apenas `/api/`, enquanto os endpoints ZapGov estão em `/v1/chamados` e `/integracao/zapgov/eventos`; o deploy publicado não os encaminha ao backend.
- O portal filtra notas internas ao listar mensagens, mas o download de anexo não verifica se a mensagem é interna.
- O middleware JWT usa o perfil do token sem revalidar conta ativa e perfil atual; desativação, alteração de perfil e troca de senha não revogam tokens emitidos.
- A deduplicação de `wa_message_id` é uma consulta seguida de inserção, sem unicidade no banco; eventos concorrentes podem duplicar. O webhook reabre a OS antes de verificar duplicidade.
- Status de entrega podem chegar antes da gravação assíncrona do `wa_message_id` na mensagem de saída.
- README principal e partes do README do backend descrevem backup SQLite, enquanto o Compose e os scripts atuais usam PostgreSQL.
- `go test ./...`, `go vet ./...` e o build do frontend passaram. `npm run lint` falhou porque `eslint` não está instalado/listado.
- Alterações locais preexistentes em `docs/INTEGRACAO-API.md` e `docs/INTEGRACAO_ZAPGOV.md` devem ser preservadas.

## Necessidade de uso relatada

Chegada de um lote de 48 monitores. Cada monitor precisa ser cadastrado e acompanhado individualmente antes de receber o número oficial de patrimônio e antes de definir seu setor ou responsável de destino.

## Fluxo implementado nesta etapa

- Manter um registro por unidade de bem permanente, com identificador interno gerado pelo SIGE, distinto do número oficial de patrimônio. Número oficial, série, setor e responsável podem começar vazios.
- A ação **Receber lote** cria até 200 bens permanentes individuais em uma transação, agrupados por `lote_codigo`, com códigos internos derivados dos IDs (`TI-000001`, etc.). O resultado permite baixar uma lista CSV dos códigos.
- Patrimônio, série, setor e responsável podem ficar pendentes; filtros da lista mostram bens sem patrimônio, sem setor e por lote.
- O menu agora separa **Itens** (categorias de consumo, por quantidade) e **Patrimônio** (categorias permanentes, por unidade). A API aceita `tipo=consumo` ou `tipo=patrimonio` para filtrar as listas; o painel mostra as contagens separadas.
- Transferência, empréstimo e devolução de bem permanente mantêm quantidade 1; descarte baixa a unidade. Mudanças de setor ou responsável por edição direta são recusadas, preservando o histórico da movimentação.
- CSV e PDF de inventário exibem o código interno. A implementação ainda não cadastra nota fiscal, fornecedor nem localização física mais detalhada que o setor; esses campos dependem da definição do fluxo real de recebimento.

## Pontos do código alterados

- `backend/internal/services/item_service.go`: patrimônio opcional, lote individual e validação de quantidade unitária.
- `backend/internal/models/item.go`: vínculo do bem ao lote.
- `backend/internal/services/movimentacao_service.go`: preservação da unidade ao mover bens permanentes.
- `frontend/src/components/CadastrarLoteDialog.tsx`: cadastro de lote e exportação dos códigos.

## Verificação desta etapa

- `go test ./...` passou, incluindo testes do endpoint HTTP, criação de 48 unidades, transferência sem perda da unidade e preenchimento posterior do patrimônio.
- `go vet ./...` e `npm run build` passaram.
- O novo campo `lote_codigo` será acrescentado ao banco pela migração automática no próximo início do backend. Ainda não houve teste com o PostgreSQL de produção nem implantação.
