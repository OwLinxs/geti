package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

type CampoCategoriaHandler struct {
	svc *services.CampoCategoriaService
}

func NewCampoCategoriaHandler(svc *services.CampoCategoriaService) *CampoCategoriaHandler {
	return &CampoCategoriaHandler{svc: svc}
}

type campoRequest struct {
	Nome        string           `json:"nome"`
	Tipo        models.TipoCampo `json:"tipo"`
	Opcoes      string           `json:"opcoes"`
	Obrigatorio bool             `json:"obrigatorio"`
	Ordem       int              `json:"ordem"`
}

// Listar devolve os campos customizados de uma categoria (rota aninhada em
// /categorias/:id/campos).
func (h *CampoCategoriaHandler) Listar(c *gin.Context) {
	categoriaID, ok := parseID(c, "id")
	if !ok {
		return
	}
	campos, err := h.svc.ListarPorCategoria(categoriaID)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, campos)
}

func (h *CampoCategoriaHandler) Criar(c *gin.Context) {
	categoriaID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req campoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	campo, err := h.svc.Criar(services.EntradaCampo{
		CategoriaID: categoriaID,
		Nome:        req.Nome,
		Tipo:        req.Tipo,
		Opcoes:      req.Opcoes,
		Obrigatorio: req.Obrigatorio,
		Ordem:       req.Ordem,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, campo)
}

func (h *CampoCategoriaHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "campoId")
	if !ok {
		return
	}
	var req campoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	campo, err := h.svc.Atualizar(id, services.EntradaCampo{
		Nome:        req.Nome,
		Tipo:        req.Tipo,
		Opcoes:      req.Opcoes,
		Obrigatorio: req.Obrigatorio,
		Ordem:       req.Ordem,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, campo)
}

func (h *CampoCategoriaHandler) Remover(c *gin.Context) {
	id, ok := parseID(c, "campoId")
	if !ok {
		return
	}
	if err := h.svc.Remover(id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
