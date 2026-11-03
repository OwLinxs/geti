package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type PermissaoHandler struct {
	svc *services.PermissaoService
}

func NewPermissaoHandler(svc *services.PermissaoService) *PermissaoHandler {
	return &PermissaoHandler{svc: svc}
}

// Catalogo lista as permissões delegáveis. GET /permissoes/catalogo
func (h *PermissaoHandler) Catalogo(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Catalogo())
}

// DoUsuario devolve as permissões de um usuário. GET /usuarios/:id/permissoes
func (h *PermissaoHandler) DoUsuario(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	est, err := h.svc.DoUsuario(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, est)
}

type permissoesRequest struct {
	Chaves []string `json:"chaves"`
}

// Salvar substitui as permissões de um usuário. PUT /usuarios/:id/permissoes
func (h *PermissaoHandler) Salvar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req permissoesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	est, err := h.svc.Salvar(id, req.Chaves)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, est)
}
