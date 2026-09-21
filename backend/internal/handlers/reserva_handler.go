package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type ReservaHandler struct {
	svc *services.ReservaService
}

func NewReservaHandler(svc *services.ReservaService) *ReservaHandler {
	return &ReservaHandler{svc: svc}
}

type reservaRequest struct {
	ItemID        uint    `json:"item_id"`
	SolicitanteID *uint   `json:"solicitante_id"`
	DataInicio    *string `json:"data_inicio"`
	DataFim       *string `json:"data_fim"`
	Finalidade    string  `json:"finalidade"`
	LocalDestino  string  `json:"local_destino"`
	Status        string  `json:"status"`
	Observacao    string  `json:"observacao"`
}

func (r reservaRequest) toEntrada() services.EntradaReserva {
	return services.EntradaReserva{
		ItemID:        r.ItemID,
		SolicitanteID: r.SolicitanteID,
		DataInicio:    parseDataCorpo(r.DataInicio),
		DataFim:       parseDataCorpo(r.DataFim),
		Finalidade:    r.Finalidade,
		LocalDestino:  r.LocalDestino,
		Status:        r.Status,
		Observacao:    r.Observacao,
	}
}

// ListarEquipamentos devolve o pool de equipamentos reserváveis com o estado
// atual (alocado ou no departamento).
func (h *ReservaHandler) ListarEquipamentos(c *gin.Context) {
	lista, err := h.svc.ListarEquipamentos()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *ReservaHandler) Criar(c *gin.Context) {
	var req reservaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	entrada := req.toEntrada()
	entrada.AprovadoPorID, _ = middlewares.UsuarioIDDoContexto(c)

	r, err := h.svc.Criar(entrada)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *ReservaHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req reservaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	r, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *ReservaHandler) Listar(c *gin.Context) {
	f := repositories.FiltroReserva{
		ItemID:  queryUint(c, "item_id"),
		Status:  c.Query("status"),
		Pagina:  queryInt(c, "pagina", 1),
		Tamanho: queryInt(c, "tamanho", 20),
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

func (h *ReservaHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *ReservaHandler) DefinirStatus(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	r, err := h.svc.DefinirStatus(id, models.StatusReserva(req.Status))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *ReservaHandler) Excluir(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Excluir(id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
