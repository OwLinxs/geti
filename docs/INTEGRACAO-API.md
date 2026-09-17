# API de Integração — Chamados / Kanban de Manutenção (SIGE-TI)

API para sistemas externos (ex.: plataforma de WhatsApp) criarem, atualizarem,
lerem e moverem chamados técnicos, que aparecem no **Kanban de Manutenção** do
SIGE-TI. Um "chamado" é uma **Ordem de Serviço (OS)**.

## Autenticação

Todas as rotas exigem a chave de API no cabeçalho:

```
X-API-Key: <sua-chave>
```

(Também aceita `Authorization: Bearer <sua-chave>`.)

A chave é definida no servidor via `INTEGRACAO_API_KEY` (arquivo `.env`).
**Sem a chave configurada, a integração fica desabilitada** e as rotas
respondem `503`.

Base da URL: `http(s)://<servidor>/api/v1/integracao`

## Colunas do Kanban (status)

| status            | Coluna           |
|-------------------|------------------|
| `aberta`          | Aberta           |
| `em_andamento`    | Em andamento     |
| `aguardando_peca` | Aguardando peça  |
| `concluida`       | Concluída        |
| `cancelada`       | (fora do board)  |

Prioridade: `baixa` \| `normal` \| `alta`.

---

## 1. Criar / atualizar chamado (idempotente)

`POST /integracao/ordens-servico`

Idempotência: envie sempre o `referencia_externa` (o id do card na sua
plataforma). Se já existir, o chamado é **atualizado** (resposta `200`); se não,
é **criado** (resposta `201`). Assim o sync não duplica.

Campos:

| Campo                       | Tipo   | Obrigatório | Observação |
|-----------------------------|--------|-------------|------------|
| `referencia_externa`        | string | recomendado | id do card na sua plataforma (idempotência) |
| `origem`                    | string | não         | ex.: `"whatsapp"` (padrão: `"externo"`) |
| `defeito_relatado`          | string | **sim**     | descrição do problema |
| `equipamento_descricao`     | string | condicional | descreva o equipamento quando **não** for item do inventário |
| `item_id`                   | number | condicional | id do item do inventário quando o chamado tem **computador linkado** |
| `equipamento_identificacao` | string | não         | série / patrimônio / tag |
| `prioridade`                | string | não         | `baixa`\|`normal`\|`alta` (padrão `normal`) |
| `solicitante_nome`          | string | não         | nome de quem abriu (aparece no card) |
| `solicitante_contato`       | string | não         | telefone/WhatsApp do solicitante |
| `setor_id`                  | number | não         | id do departamento |

> Regra: informe **`item_id`** (computador do inventário) **ou**
> **`equipamento_descricao`** (chamado sem equipamento linkado). Um dos dois.

Exemplo — chamado sem computador linkado:

```bash
curl -X POST http://SEU_SERVIDOR/api/v1/integracao/ordens-servico \
  -H "X-API-Key: SUA_CHAVE" \
  -H "Content-Type: application/json" \
  -d '{
    "referencia_externa": "WA-1001",
    "origem": "whatsapp",
    "equipamento_descricao": "Impressora da sala 3",
    "defeito_relatado": "Não imprime, luz piscando",
    "prioridade": "alta",
    "solicitante_nome": "Joana (RH)",
    "solicitante_contato": "46 99999-0000"
  }'
```

Exemplo — chamado com computador do inventário:

```json
{
  "referencia_externa": "WA-1002",
  "item_id": 42,
  "defeito_relatado": "Computador não liga",
  "prioridade": "normal"
}
```

Resposta (`201`/`200`): o objeto da OS, incluindo `id`, `numero` (ex.:
`OS-2026-0001`) e `status`.

---

## 2. Mover de coluna (mudar status)

`PATCH /integracao/ordens-servico/{id}/status`

```bash
curl -X PATCH http://SEU_SERVIDOR/api/v1/integracao/ordens-servico/1/status \
  -H "X-API-Key: SUA_CHAVE" \
  -H "Content-Type: application/json" \
  -d '{ "status": "em_andamento" }'
```

`{id}` é o `id` retornado na criação. Ao concluir (`concluida`), o sistema
carimba a data de conclusão automaticamente.

---

## 3. Ler o board (sync de volta)

`GET /integracao/ordens-servico?status=&pagina=1&tamanho=50`

- `status` (opcional): filtra por coluna.
- Paginação: `pagina`, `tamanho`.

```bash
curl "http://SEU_SERVIDOR/api/v1/integracao/ordens-servico?status=aberta" \
  -H "X-API-Key: SUA_CHAVE"
```

Resposta:

```json
{ "dados": [ /* ...OS... */ ], "total": 12, "pagina": 1, "tamanho": 50 }
```

Ler um único chamado: `GET /integracao/ordens-servico/{id}`.

---

## Códigos de resposta

| Código | Significado |
|--------|-------------|
| 200    | Atualizado / OK |
| 201    | Criado |
| 401    | Chave de API ausente/inválida |
| 422    | Campos inválidos (ver `campos` na resposta) |
| 503    | Integração desabilitada (`INTEGRACAO_API_KEY` não configurada) |

## Segurança

- Trafegue **sempre por HTTPS** em produção — a chave vai no cabeçalho.
- Guarde a chave como segredo na sua plataforma; nunca a exponha no cliente.
- Para revogar/rotacionar: troque `INTEGRACAO_API_KEY` no `.env` e reinicie o
  backend.
