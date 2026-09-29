package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

// IntegracaoHandler expõe a API de chamados para sistemas externos (ex.: a
// plataforma de WhatsApp). Autenticação por chave de API (middleware).
type IntegracaoHandler struct {
	svc        *services.OrdemServicoService
	msgSvc     *services.MensagemService
	consumidor *services.ZapGovConsumidor
}

func NewIntegracaoHandler(svc *services.OrdemServicoService, msgSvc *services.MensagemService, consumidor *services.ZapGovConsumidor) *IntegracaoHandler {
	return &IntegracaoHandler{svc: svc, msgSvc: msgSvc, consumidor: consumidor}
}

type chamadoExternoRequest struct {
	ReferenciaExterna        string `json:"referencia_externa"`
	Origem                   string `json:"origem"`
	// AbrirChamado vem do menu do bot: true = "Abrir Chamado" (força um chamado
	// NOVO, mesmo com um aberto); false/ausente = "Retomar" (reusa o aberto).
	AbrirChamado             bool   `json:"abrir_chamado"`
	ItemID                   uint   `json:"item_id"`
	EquipamentoDescricao     string `json:"equipamento_descricao"`
	EquipamentoIdentificacao string `json:"equipamento_identificacao"`
	SetorID                  *uint  `json:"setor_id"`
	SolicitanteNome          string `json:"solicitante_nome"`
	SolicitanteContato       string `json:"solicitante_contato"`
	DefeitoRelatado          string `json:"defeito_relatado"`
	Prioridade               string `json:"prioridade"`
}

func (r chamadoExternoRequest) toEntrada() services.EntradaOS {
	origem := r.Origem
	if origem == "" {
		origem = "externo"
	}
	return services.EntradaOS{
		ReferenciaExterna:        r.ReferenciaExterna,
		Origem:                   origem,
		ItemID:                   r.ItemID,
		EquipamentoDescricao:     r.EquipamentoDescricao,
		EquipamentoIdentificacao: r.EquipamentoIdentificacao,
		SetorID:                  r.SetorID,
		SolicitanteNome:          r.SolicitanteNome,
		SolicitanteContato:       r.SolicitanteContato,
		DefeitoRelatado:          r.DefeitoRelatado,
		Prioridade:               r.Prioridade,
	}
}

// Criar cria ou atualiza (idempotente por referencia_externa) um chamado.
func (h *IntegracaoHandler) Criar(c *gin.Context) {
	var req chamadoExternoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	// "Abrir Chamado" (true) força um chamado novo; "Retomar" (false) reusa o
	// chamado aberto da conversa, se houver.
	if req.AbrirChamado {
		os, err := h.svc.AbrirNovoViaIntegracao(req.toEntrada())
		if err != nil {
			responderErro(c, err)
			return
		}
		c.JSON(http.StatusCreated, os)
		return
	}
	os, criado, err := h.svc.UpsertViaIntegracao(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	status := http.StatusOK
	if criado {
		status = http.StatusCreated
	}
	c.JSON(status, os)
}

type mensagemExternaRequest struct {
	ReferenciaExterna string `json:"referencia_externa"`
	Telefone          string `json:"telefone"`
	Nome              string `json:"nome"`
	Texto             string `json:"texto"`
	MidiaURL          string `json:"midia_url"`
	IdExterno         string `json:"id_externo"`
}

// ReceberMensagem processa uma mensagem recebida do WhatsApp (entrada). Acha o
// chamado por referencia_externa, senão por telefone (chamado aberto), senão
// cria um novo. Depois anexa a mensagem na conversa.
func (h *IntegracaoHandler) ReceberMensagem(c *gin.Context) {
	var req mensagemExternaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}

	os, msg, err := h.consumidor.ProcessarEntrada(
		req.ReferenciaExterna, req.Telefone, req.Nome,
		req.Texto, req.MidiaURL, req.IdExterno)
	if err != nil {
		// Mensagem repetida (mesmo id_externo): idempotente, responde OK.
		if errors.Is(err, services.ErrDuplicado) {
			c.JSON(http.StatusOK, gin.H{"duplicada": true})
			return
		}
		responderErro(c, err)
		return
	}
	// Sem chamado aberto na conversa: mensagem ignorada (só o bot cria chamado).
	if os == nil || msg == nil {
		c.JSON(http.StatusOK, gin.H{"ignorada": true})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"ordem_servico_id": os.ID,
		"numero":           os.Numero,
		"mensagem_id":      msg.ID,
	})
}

// Listar devolve o board (chamados) para o sistema externo sincronizar. Aceita
// `conversa` (id da conversa do WhatsApp) ou `referencia_externa` para listar o
// histórico de uma pessoa (usado no menu "Retomar chamado" do bot).
func (h *IntegracaoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroOrdemServico{
		Status:  c.Query("status"),
		Pagina:  queryInt(c, "pagina", 1),
		Tamanho: queryInt(c, "tamanho", 50),
	}
	if ref := c.Query("referencia_externa"); ref != "" {
		f.ReferenciaExterna = ref
	} else if conv := c.Query("conversa"); conv != "" {
		f.ReferenciaExterna = "zap:" + conv
	}
	lista, total, err := h.svc.Listar(f)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"dados":   lista,
		"total":   total,
		"pagina":  f.Pagina,
		"tamanho": f.Tamanho,
	})
}

func (h *IntegracaoHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	os, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}

// DefinirStatus move o chamado de coluna (sincroniza o kanban do sistema
// externo com o daqui).
func (h *IntegracaoHandler) DefinirStatus(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	os, err := h.svc.DefinirStatus(id, models.StatusOS(req.Status))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}
