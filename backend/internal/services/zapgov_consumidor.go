package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
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

	mu          sync.Mutex
	ultimoFetch map[int64]time.Time
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
	// Backfill: chamado achado por referência pode ter vindo de um evento sem
	// telefone; garante o contato para a equipe conseguir responder.
	c.osSvc.GarantirContato(os.ID, telefone, nome)
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

// onEvento trata um evento do SSE. O evento é só um AVISO (traz conversation_id
// e, às vezes, um preview) — a mensagem real é buscada no endpoint da conversa.
func (c *ZapGovConsumidor) onEvento(ev EventoZapGov) {
	tipo := strings.ToLower(ev.Type + ev.Event)
	if !strings.Contains(tipo, "mensag") && !strings.Contains(tipo, "message") {
		return // status, presença etc.
	}
	bruto := ev.Message
	if len(bruto) == 0 {
		bruto = ev.Data
	}
	var cab struct {
		ConversationID int64  `json:"conversation_id"`
		Name           string `json:"name"`
	}
	_ = json.Unmarshal(bruto, &cab)
	if cab.ConversationID == 0 {
		return
	}
	// Debounce: eventos chegam em rajada para a mesma conversa.
	if !c.podeSincronizar(cab.ConversationID) {
		return
	}
	c.sincronizarConversa(cab.ConversationID, cab.Name)
}

// podeSincronizar limita a 1 busca por conversa a cada 2s (anti-rajada).
func (c *ZapGovConsumidor) podeSincronizar(convID int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ultimoFetch == nil {
		c.ultimoFetch = map[int64]time.Time{}
	}
	agora := time.Now()
	if t, ok := c.ultimoFetch[convID]; ok && agora.Sub(t) < 2*time.Second {
		return false
	}
	c.ultimoFetch[convID] = agora
	return true
}

// sincronizarConversa busca as mensagens não-lidas da conversa e registra as
// que vieram do contato (entrada), com dedup por id externo.
func (c *ZapGovConsumidor) sincronizarConversa(convID int64, nomeFallback string) {
	msgs, err := c.cli.MensagensConversa(convID, 50, true)
	if err != nil {
		log.Printf("[zapgov] falha ao buscar conversa %d: %v", convID, err)
		return
	}
	ref := "zap:" + strconv.FormatInt(convID, 10)
	for _, m := range msgs {
		if !m.Recebida() {
			continue // eco de mensagem da equipe
		}
		idExterno := m.WaMessageID
		if idExterno == "" {
			idExterno = m.ID
		}
		nome := m.NomeContato()
		if nome == "" {
			nome = nomeFallback
		}
		_, _, err := c.ProcessarEntrada(ref, m.TelefoneContato(), nome,
			m.Conteudo(), m.Midia(), idExterno)
		if err != nil && !errors.Is(err, ErrDuplicado) {
			log.Printf("[zapgov] falha ao processar entrada (conv %d): %v", convID, err)
		}
	}
}
