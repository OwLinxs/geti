package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type CategoriaChamadoHandler struct {
	svc *services.CategoriaChamadoService
}

func NewCategoriaChamadoHandler(svc *services.CategoriaChamadoService) *CategoriaChamadoHandler {
	return &CategoriaChamadoHandler{svc: svc}
}

type categoriaChamadoRequest struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Ordem     int    `json:"ordem"`
	Ativo     *bool  `json:"ativo"`
}

func (r categoriaChamadoRequest) toEntrada() services.EntradaCategoriaChamado {
	ativo := true
	if r.Ativo != nil {
		ativo = *r.Ativo
	}
	return services.EntradaCategoriaChamado{
		Nome:      r.Nome,
		Descricao: r.Descricao,
		Ordem:     r.Ordem,
		Ativo:     ativo,
	}
}

// Listar devolve as categorias. ?ativas=true retorna só as ativas (uso nos
// formulários de abertura de chamado).
func (h *CategoriaChamadoHandler) Listar(c *gin.Context) {
	lista, err := h.svc.Listar(c.Query("ativas") == "true")
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *CategoriaChamadoHandler) Criar(c *gin.Context) {
	var req categoriaChamadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	cat, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *CategoriaChamadoHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req categoriaChamadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	cat, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *CategoriaChamadoHandler) Excluir(c *gin.Context) {
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
