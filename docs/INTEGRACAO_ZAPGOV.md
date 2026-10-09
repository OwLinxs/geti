# Integração ZapGov (gateway de WhatsApp)

O **ZapGov** cuida só do WhatsApp (enviar/receber, mídia, janela de 24h, status
de entrega). O **SIGE-TI é o dono do chamado e do chat**: o técnico conversa com
a pessoa dentro da tela do chamado, e as mensagens vão e voltam por API.

Arquitetura = **webhook push** (o ZapGov chama o SIGE). A antiga integração por
login JWT + SSE foi removida.

## Configuração (`.env`, só no servidor)

| Variável | Uso |
|----------|-----|
| `ZAPGOV_BASE_URL` | Base da API do ZapGov, sem barra no final (o client soma `/v1/...`). |
| `ZAPGOV_API_KEY` | Chave do SIGE no ZapGov (header `X-API-Key` nas chamadas de saída). |
| `INTEGRACAO_API_KEY` | Chave do sistema; o ZapGov manda em `X-API-Key` ao chamar o SIGE. |
| `INTEGRACAO_HMAC_SECRET` | Segredo HMAC-SHA256 p/ validar `X-Signature` dos webhooks. Vazio = valida só `X-API-Key`. |

Vazio (`ZAPGOV_BASE_URL`/`ZAPGOV_API_KEY`) = WhatsApp desabilitado: as mensagens
ficam só registradas no SIGE, sem sair para o WhatsApp.

## Fluxo geral

1. Pessoa manda WhatsApp → bot cru pergunta "Abrir chamado" ou "Retomar chamado".
2. **Abrir** → ZapGov chama `chamado.abrir` → SIGE cria o chamado.
3. **Retomar** → ZapGov chama `GET /v1/chamados?telefone=` → pessoa escolhe um
   (pode ser FECHADO → reabre) → a conversa passa a cair no chat daquele chamado.
4. Cada mensagem da pessoa → ZapGov chama `mensagem.recebida` → entra no chat.
5. Técnico responde no SIGE → SIGE chama `POST /v1/mensagens` → vai ao WhatsApp.
6. Técnico encerra o chamado → SIGE avisa a pessoa + `POST /v1/relay/encerrar`
   (próxima mensagem da pessoa volta ao menu do bot).

## Endpoints que o SIGE EXPÕE (o ZapGov chama)

### `GET /v1/chamados?telefone=<E164>&limite=10`
Auth: `X-API-Key: <INTEGRACAO_API_KEY>`. Lista os chamados do telefone
(`solicitante_contato`), mais recentes primeiro, **abertos e fechados**.

```json
{ "chamados": [
  {"id":123,"numero":"OS-2026-0123","resumo":"Impressora não liga",
   "status":"aberta","fechado":false,"atualizado_em":"2026-10-06T12:00:00Z"},
  {"id":118,"numero":"OS-2026-0118","resumo":"Sem internet",
   "status":"concluida","fechado":true,"atualizado_em":"2026-09-20T09:00:00Z"}
]}
```
`resumo` ≤ 60 chars. Sem chamados → `{"chamados":[]}`. `fechado` = status
`concluida` ou `cancelada`.

### `POST /integracao/zapgov/eventos`
Auth: `X-API-Key: <INTEGRACAO_API_KEY>` **e** `X-Signature` = HMAC-SHA256 do corpo
cru (segredo `INTEGRACAO_HMAC_SECRET`; aceita prefixo `sha256=`). 401 se não bater.
Um endpoint só; o campo `evento` roteia.

**a) `chamado.abrir`** — cria um chamado (idempotente por `referencia`).
```json
{"evento":"chamado.abrir","telefone":"5546...","nome":"João da Silva",
 "referencia":"wa-<uuid>","dados":{"descricao":"texto do problema"}}
→ 200 {"ticket_id":123,"numero":"OS-2026-0123"}
```

**b) `mensagem.recebida`** — anexa a mensagem ao chat (direção entrada). Dedup por
`wa_message_id`. Se o chamado estiver fechado, **reabre**. Abre a janela de 24h.
```json
{"evento":"mensagem.recebida","ticket_id":123,"telefone":"5546...",
 "wa_message_id":"wamid.XXX","tipo":"texto|imagem|audio|documento|video",
 "texto":"...","midia_url":"https://.../assinada","midia_mime":"image/jpeg",
 "timestamp":"2026-10-06T12:00:00Z"}
→ 200 {"ok":true}
```

**c) `mensagem.status`** — atualiza o status de entrega da mensagem de SAÍDA
(casa por `wa_message_id` devolvido no envio).
```json
{"evento":"mensagem.status","wa_message_id":"wamid.YYY",
 "status":"enviado|entregue|lido|falhou","erro":""}
→ 200 {"ok":true}
```

## Endpoints que o SIGE CHAMA (para responder a pessoa)

Base `ZAPGOV_BASE_URL`, header `X-API-Key: <ZAPGOV_API_KEY>`.

### `POST /v1/mensagens`
```json
Texto:    {"telefone":"5546...","tipo":"texto","texto":"..."}
Mídia:    {"telefone":"5546...","tipo":"midia","midia_url":"...","midia_mime":"..."}
Template: {"telefone":"5546...","tipo":"template","template":"nome","params":["..."]}
→ 200 {"wa_message_id":"wamid.YYY"}
```
`401` chave inválida; `422 {"erro":"janela_fechada"}` → texto fora da janela de 24h
(o SIGE marca a mensagem como `falhou`; a UI deve oferecer envio por template).

### `POST /v1/relay/encerrar`
```json
{"telefone":"5546..."}
```
Chamado quando o chamado é fechado no SIGE: tira a conversa do modo relay.

## Janela de 24h

O WhatsApp só deixa texto livre até 24h após a última mensagem da pessoa. Cada
`mensagem.recebida` reabre a janela (`ultima_msg_solicitante_em`). Fora da janela,
`POST /v1/mensagens` (texto) devolve `janela_fechada` → usar template aprovado.

## Modelo de dados (reaproveitado)

- **Chamado** = `ordens_servico`: `solicitante_contato` (indexado), `origem`
  (`whatsapp`), `referencia_externa`, `ultima_msg_solicitante_em`.
- **Mensagens** = `mensagens_chamado`: `direcao` (entrada/saida/interna),
  `id_externo` (=`wa_message_id`, dedup), `tipo`, `texto`, `midia_url`, `anexo_*`,
  `status` (enviado/entregue/lido/recebido/falhou), `enviada_em`.
- Dedup por `id_externo` em nível de aplicação (`ExistePorIdExterno`).

## Arquivos principais (backend)

- `internal/services/zapgov_client.go` — client do ZapGov (saída).
- `internal/handlers/zapgov_webhook_handler.go` — webhook + `GET /v1/chamados`.
- `internal/middlewares/webhook_zapgov.go` — X-API-Key + HMAC.
- `internal/services/mensagem_service.go` — `RegistrarRecebida`,
  `AtualizarStatusExterno`, `EncerrarRelayWhatsApp`, `enviarWhatsApp` (saída).
- `internal/services/ordem_servico_service.go` — `ListarPorContato`,
  `AbrirNovoViaIntegracao`, `BuscarAbertaPorReferencia`, hook de encerramento.

## Contrato-chave para o ZapGov

- `referencia` estável por conversa (ex.: `wa-<uuid>`) → idempotência da abertura.
- `solicitante_contato` = telefone **real** E164 (formato que o `/v1/mensagens`
  aceita) → a resposta e o encerramento voltam certo.
- `wa_message_id` em toda mensagem → dedup (entrada) e casamento de status (saída).

## Pendências (Fase 3/4)

- **Saída de mídia**: técnico anexa arquivo → hoje só o TEXTO vai ao WhatsApp
  (`EnviarMidia` existe no client, falta ligar na UI).
- **Envio por template** fora da janela: `EnviarTemplate` existe, falta o botão na
  tela do chamado.
- Download da mídia recebida é best-effort (a `midia_url` é temporária/assinada).
