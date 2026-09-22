package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type RespostaRapidaHandler struct {
	svc *services.RespostaRapidaService
}

func NewRespostaRapidaHandler(svc *services.RespostaRapidaService) *RespostaRapidaHandler {
	return &RespostaRapidaHandler{svc: svc}
}

type respostaRapidaRequest struct {
	Titulo   string `json:"titulo"`
	Conteudo string `json:"conteudo"`
	Ordem    int    `json:"ordem"`
	Ativo    *bool  `json:"ativo"`
}

func (r respostaRapidaRequest) toEntrada() services.EntradaRespostaRapida {
	ativo := true
	if r.Ativo != nil {
		ativo = *r.Ativo
	}
	return services.EntradaRespostaRapida{
		Titulo:   r.Titulo,
		Conteudo: r.Conteudo,
		Ordem:    r.Ordem,
		Ativo:    ativo,
	}
}

func (h *RespostaRapidaHandler) Listar(c *gin.Context) {
	lista, err := h.svc.Listar(c.Query("ativas") == "true")
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *RespostaRapidaHandler) Criar(c *gin.Context) {
	var req respostaRapidaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	x, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, x)
}

func (h *RespostaRapidaHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req respostaRapidaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	x, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, x)
}

func (h *RespostaRapidaHandler) Excluir(c *gin.Context) {
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
