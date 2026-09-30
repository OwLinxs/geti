package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

// ZapGovHandler recebe os webhooks do gateway de WhatsApp (ZapGov) e expõe a
// listagem de chamados por telefone (usada no "Retomar chamado" do bot).
type ZapGovHandler struct {
	osSvc  *services.OrdemServicoService
	msgSvc *services.MensagemService
}

func NewZapGovHandler(osSvc *services.OrdemServicoService, msgSvc *services.MensagemService) *ZapGovHandler {
	return &ZapGovHandler{osSvc: osSvc, msgSvc: msgSvc}
}

// zapgovEvento cobre os campos dos três eventos (o campo "evento" roteia).
type zapgovEvento struct {
	Evento string `json:"evento"`

	// chamado.abrir
	Telefone   string `json:"telefone"`
	Nome       string `json:"nome"`
	Referencia string `json:"referencia"`
	Dados      struct {
		Descricao string `json:"descricao"`
	} `json:"dados"`

	// mensagem.recebida
	TicketID    uint   `json:"ticket_id"`
	WaMessageID string `json:"wa_message_id"`
	Tipo        string `json:"tipo"`
	Texto       string `json:"texto"`
	MidiaURL    string `json:"midia_url"`
	MidiaMime   string `json:"midia_mime"`
	Timestamp   string `json:"timestamp"`

	// mensagem.status
	Status string `json:"status"`
	Erro   string `json:"erro"`
}

// Eventos processa o webhook único do ZapGov, roteando pelo campo "evento".
func (h *ZapGovHandler) Eventos(c *gin.Context) {
	var ev zapgovEvento
	if err := c.ShouldBindJSON(&ev); err != nil {
		erroBind(c, err)
		return
	}
	switch ev.Evento {
	case "chamado.abrir":
		h.chamadoAbrir(c, ev)
	case "mensagem.recebida":
		h.mensagemRecebida(c, ev)
	case "mensagem.status":
		h.mensagemStatus(c, ev)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"erro": "evento desconhecido: " + ev.Evento})
	}
}

func (h *ZapGovHandler) chamadoAbrir(c *gin.Context, ev zapgovEvento) {
	// Idempotência: mesma referência devolve o mesmo chamado.
	if ref := strings.TrimSpace(ev.Referencia); ref != "" {
		if os, err := h.osSvc.BuscarPorReferenciaExterna(ref); err == nil && os != nil {
			c.JSON(http.StatusOK, gin.H{"ticket_id": os.ID, "numero": os.Numero})
			return
		}
	}
	os, err := h.osSvc.AbrirNovoViaIntegracao(services.EntradaOS{
		Origem:             "whatsapp",
		ReferenciaExterna:  ev.Referencia,
		SolicitanteContato: ev.Telefone,
		SolicitanteNome:    ev.Nome,
		DefeitoRelatado:    primeiroNaoVazio(ev.Dados.Descricao, "Chamado aberto via WhatsApp"),
		Prioridade:         "normal",
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket_id": os.ID, "numero": os.Numero})
}

func (h *ZapGovHandler) mensagemRecebida(c *gin.Context, ev zapgovEvento) {
	os, err := h.osSvc.BuscarPorID(ev.TicketID)
	if err != nil {
		responderErro(c, err)
		return
	}
	// Chamado fechado → reabrir (a pessoa voltou a falar).
	if os.Status == models.OSConcluida || os.Status == models.OSCancelada {
		if _, err := h.osSvc.DefinirStatus(os.ID, models.OSEmAndamento); err != nil {
			responderErro(c, err)
			return
		}
	}
	quando, _ := time.Parse(time.RFC3339, ev.Timestamp)
	_, dup, err := h.msgSvc.RegistrarRecebida(services.EntradaRecebida{
		OrdemServicoID: os.ID,
		AutorNome:      os.SolicitanteNomeSnapshot,
		Tipo:           ev.Tipo,
		Texto:          ev.Texto,
		MidiaURL:       ev.MidiaURL,
		MidiaMime:      ev.MidiaMime,
		IdExterno:      ev.WaMessageID,
		Quando:         quando,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	_ = dup // resposta é a mesma (idempotente)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ZapGovHandler) mensagemStatus(c *gin.Context, ev zapgovEvento) {
	st := mapStatusWhatsApp(ev.Status)
	if st != "" {
		_ = h.msgSvc.AtualizarStatusExterno(ev.WaMessageID, st)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func mapStatusWhatsApp(s string) models.StatusMensagem {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "enviado":
		return models.MsgEnviada
	case "entregue":
		return models.MsgEntregue
	case "lido":
		return models.MsgLida
	case "falhou":
		return models.MsgFalhou
	}
	return ""
}

// ListarChamados atende GET /v1/chamados?telefone=<E164>&limite=10 — usado no
// menu "Retomar chamado" do bot.
func (h *ZapGovHandler) ListarChamados(c *gin.Context) {
	telefone := strings.TrimSpace(c.Query("telefone"))
	if telefone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe o telefone"})
		return
	}
	limite := queryInt(c, "limite", 10)
	lista, err := h.osSvc.ListarPorContato(telefone, limite)
	if err != nil {
		responderErro(c, err)
		return
	}
	type chamadoResumo struct {
		ID           uint   `json:"id"`
		Numero       string `json:"numero"`
		Resumo       string `json:"resumo"`
		Status       string `json:"status"`
		Fechado      bool   `json:"fechado"`
		AtualizadoEm string `json:"atualizado_em"`
	}
	out := make([]chamadoResumo, 0, len(lista))
	for _, os := range lista {
		fechado := os.Status == models.OSConcluida || os.Status == models.OSCancelada
		out = append(out, chamadoResumo{
			ID:           os.ID,
			Numero:       os.Numero,
			Resumo:       resumoChamado(os),
			Status:       string(os.Status),
			Fechado:      fechado,
			AtualizadoEm: os.AtualizadoEm.UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"chamados": out})
}

// resumoChamado devolve um resumo curto (<= 60 chars) do chamado.
func resumoChamado(os models.OrdemServico) string {
	txt := strings.TrimSpace(os.Assunto)
	if txt == "" {
		txt = strings.TrimSpace(os.DefeitoRelatado)
	}
	txt = strings.ReplaceAll(txt, "\n", " ")
	if len(txt) > 60 {
		txt = strings.TrimSpace(txt[:60])
	}
	return txt
}
