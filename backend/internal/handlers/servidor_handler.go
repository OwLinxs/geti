package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type ServidorHandler struct {
	svc *services.ServidorService
}

func NewServidorHandler(svc *services.ServidorService) *ServidorHandler {
	return &ServidorHandler{svc: svc}
}

// LGPD: somente nome e matrícula são coletados.
type servidorRequest struct {
	Nome      string `json:"nome"`
	Matricula string `json:"matricula"`
	SetorID   *uint  `json:"setor_id"`
	Ativo     *bool  `json:"ativo"`
}

func (r servidorRequest) toEntrada() services.EntradaServidor {
	ativo := true
	if r.Ativo != nil {
		ativo = *r.Ativo
	}
	return services.EntradaServidor{Nome: r.Nome, Matricula: r.Matricula, SetorID: r.SetorID, Ativo: ativo}
}

func (h *ServidorHandler) Listar(c *gin.Context) {
	ss, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ss)
}

func (h *ServidorHandler) Criar(c *gin.Context) {
	var req servidorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	srv, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, srv)
}

func (h *ServidorHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	srv, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, srv)
}

func (h *ServidorHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req servidorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	srv, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, srv)
}

func (h *ServidorHandler) Remover(c *gin.Context) {
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
