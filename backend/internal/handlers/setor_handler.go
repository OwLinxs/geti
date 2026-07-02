package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type SetorHandler struct {
	svc *services.SetorService
}

func NewSetorHandler(svc *services.SetorService) *SetorHandler {
	return &SetorHandler{svc: svc}
}

type setorRequest struct {
	Nome        string `json:"nome"`
	Sigla       string `json:"sigla"`
	Localizacao string `json:"localizacao"`
}

func (h *SetorHandler) Listar(c *gin.Context) {
	ss, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ss)
}

func (h *SetorHandler) Criar(c *gin.Context) {
	var req setorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	st, err := h.svc.Criar(services.EntradaSetor{Nome: req.Nome, Sigla: req.Sigla, Localizacao: req.Localizacao})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, st)
}

func (h *SetorHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	st, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

func (h *SetorHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req setorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	st, err := h.svc.Atualizar(id, services.EntradaSetor{Nome: req.Nome, Sigla: req.Sigla, Localizacao: req.Localizacao})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

func (h *SetorHandler) Remover(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Remover(id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
