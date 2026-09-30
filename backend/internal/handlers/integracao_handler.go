package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

// IntegracaoHandler expõe endpoints genéricos de sincronização de chamados para
// sistemas externos (autenticados por chave de API). O fluxo de WhatsApp em si
// vive em ZapGovHandler (webhooks + /v1/chamados).
type IntegracaoHandler struct {
	svc *services.OrdemServicoService
}

func NewIntegracaoHandler(svc *services.OrdemServicoService) *IntegracaoHandler {
	return &IntegracaoHandler{svc: svc}
}

type chamadoExternoRequest struct {
	ReferenciaExterna        string `json:"referencia_externa"`
	Origem                   string `json:"origem"`
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

// Criar cria/atualiza um chamado. abrir_chamado=true força um chamado novo;
// caso contrário reusa o chamado aberto da referência (idempotente).
func (h *IntegracaoHandler) Criar(c *gin.Context) {
	var req chamadoExternoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
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

// Listar devolve o board (chamados) para o sistema externo sincronizar.
func (h *IntegracaoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroOrdemServico{
		Status:  c.Query("status"),
		Contato: c.Query("telefone"),
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

// DefinirStatus move o chamado de coluna (sync do kanban externo).
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
