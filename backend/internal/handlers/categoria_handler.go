package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type CategoriaHandler struct {
	svc *services.CategoriaService
}

func NewCategoriaHandler(svc *services.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{svc: svc}
}

type categoriaRequest struct {
	Nome       string `json:"nome"`
	Descricao  string `json:"descricao"`
	Consumivel bool   `json:"consumivel"`
}

func (h *CategoriaHandler) Listar(c *gin.Context) {
	cs, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, cs)
}

func (h *CategoriaHandler) Criar(c *gin.Context) {
	var req categoriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	cat, err := h.svc.Criar(services.EntradaCategoria{Nome: req.Nome, Descricao: req.Descricao, Consumivel: req.Consumivel})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *CategoriaHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	cat, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *CategoriaHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req categoriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	cat, err := h.svc.Atualizar(id, services.EntradaCategoria{Nome: req.Nome, Descricao: req.Descricao, Consumivel: req.Consumivel})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *CategoriaHandler) Remover(c *gin.Context) {
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
