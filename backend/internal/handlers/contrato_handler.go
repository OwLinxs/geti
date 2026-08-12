package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type ContratoHandler struct {
	svc *services.ContratoService
}

func NewContratoHandler(svc *services.ContratoService) *ContratoHandler {
	return &ContratoHandler{svc: svc}
}

type contratoRequest struct {
	Numero       string   `json:"numero"`
	FornecedorID uint     `json:"fornecedor_id"`
	Objeto       string   `json:"objeto"`
	DataInicio   *string  `json:"data_inicio"`
	DataFim      *string  `json:"data_fim"`
	Valor        *float64 `json:"valor"`
	Status       string   `json:"status"`
	Observacao   string   `json:"observacao"`
}

func (r contratoRequest) toEntrada() services.EntradaContrato {
	return services.EntradaContrato{
		Numero:       r.Numero,
		FornecedorID: r.FornecedorID,
		Objeto:       r.Objeto,
		DataInicio:   parseDataCorpo(r.DataInicio),
		DataFim:      parseDataCorpo(r.DataFim),
		Valor:        r.Valor,
		Status:       r.Status,
		Observacao:   r.Observacao,
	}
}

func (h *ContratoHandler) Criar(c *gin.Context) {
	var req contratoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	ct, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, ct)
}

func (h *ContratoHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req contratoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	ct, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ct)
}

func (h *ContratoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroContrato{
		Q:               c.Query("q"),
		Status:          c.Query("status"),
		FornecedorID:    queryUint(c, "fornecedor_id"),
		VencendoAteDias: queryIntPtr(c, "vencendo"),
		Pagina:          queryInt(c, "pagina", 1),
		Tamanho:         queryInt(c, "tamanho", 20),
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

func (h *ContratoHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	ct, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ct)
}

func (h *ContratoHandler) Excluir(c *gin.Context) {
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

// parseDataCorpo aceita datas do corpo em RFC3339 (ISO) ou "YYYY-MM-DD".
func parseDataCorpo(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", *s); err == nil {
		return &t
	}
	return nil
}
