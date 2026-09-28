package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
)

// ZapGovConsumidor recebe mensagens do WhatsApp (via SSE ou webhook) e as
// roteia para o chamado certo — reaproveitado pelo handler de integração e pelo
// stream de eventos, para não duplicar a lógica de resolução do chamado.
type ZapGovConsumidor struct {
	cli    *ZapGovClient
	osSvc  *OrdemServicoService
	msgSvc *MensagemService
}

func NewZapGovConsumidor(cli *ZapGovClient, osSvc *OrdemServicoService, msgSvc *MensagemService) *ZapGovConsumidor {
	return &ZapGovConsumidor{cli: cli, osSvc: osSvc, msgSvc: msgSvc}
}

func primeiroNaoVazio(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// ProcessarEntrada acha o chamado por referência externa, senão por telefone
// (chamado aberto), senão abre um novo; depois registra a mensagem. Devolve
// ErrDuplicado quando a mensagem (id_externo) já foi processada.
func (c *ZapGovConsumidor) ProcessarEntrada(ref, telefone, nome, texto, midia, idExterno string) (*models.OrdemServico, *models.MensagemChamado, error) {
	var os *models.OrdemServico
	if ref != "" {
		if o, err := c.osSvc.BuscarPorReferenciaExterna(ref); err == nil {
			os = o
		}
	}
	if os == nil && telefone != "" {
		if o, err := c.osSvc.BuscarAbertaPorContato(telefone); err == nil {
			os = o
		}
	}
	if os == nil {
		novo, _, err := c.osSvc.UpsertViaIntegracao(EntradaOS{
			ReferenciaExterna:  ref,
			Origem:             "whatsapp",
			DefeitoRelatado:    primeiroNaoVazio(texto, "Mensagem recebida via WhatsApp"),
			SolicitanteNome:    nome,
			SolicitanteContato: telefone,
			Prioridade:         "normal",
		})
		if err != nil {
			return nil, nil, err
		}
		os = novo
	}
	msg, err := c.msgSvc.RegistrarEntradaExterna(os.ID, nome, texto, midia, idExterno)
	if err != nil {
		return os, nil, err
	}
	return os, msg, nil
}

// Iniciar sobe o consumo por SSE em background (não bloqueia). Reconecta com
// backoff até o contexto ser cancelado. Não faz nada se o ZapGov não estiver
// configurado.
func (c *ZapGovConsumidor) Iniciar(ctx context.Context) {
	if c.cli == nil || !c.cli.Configurado() {
		log.Printf("[zapgov] consumo de entrada desativado (sem credenciais)")
		return
	}
	go c.loop(ctx)
}

func (c *ZapGovConsumidor) loop(ctx context.Context) {
	backoff := 2 * time.Second
	const maxBackoff = 60 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		log.Printf("[zapgov] conectando ao stream de eventos")
		err := c.cli.StreamEventos(ctx, c.onEvento)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("[zapgov] stream caiu: %v (reconecta em %s)", err, backoff)
		} else {
			log.Printf("[zapgov] stream encerrado (reconecta em %s)", backoff)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (c *ZapGovConsumidor) onEvento(ev EventoZapGov) {
	tipo := strings.ToLower(ev.Type + ev.Event)
	// Ignora eventos que claramente não são mensagem recebida.
	if tipo != "" && !strings.Contains(tipo, "mensag") && !strings.Contains(tipo, "message") {
		return
	}
	bruto := ev.Message
	if len(bruto) == 0 {
		bruto = ev.Data
	}
	if len(bruto) == 0 {
		return
	}
	var m MensagemZapGov
	if err := json.Unmarshal(bruto, &m); err != nil {
		return
	}
	if !m.Recebida() {
		return // mensagem de saída (eco da própria equipe)
	}
	ref := ""
	if m.ConversationID != 0 {
		ref = "zap:" + strconv.FormatInt(m.ConversationID, 10)
	}
	idExterno := m.WaMessageID
	if idExterno == "" {
		idExterno = m.ID
	}
	_, _, err := c.ProcessarEntrada(ref, m.TelefoneContato(), m.NomeContato(),
		m.Conteudo(), m.Midia(), idExterno)
	if err != nil && !errors.Is(err, ErrDuplicado) {
		log.Printf("[zapgov] falha ao processar entrada: %v", err)
	}
}
