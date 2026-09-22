package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/services"
)

type NotificacaoHandler struct {
	svc *services.NotificacaoService
}

func NewNotificacaoHandler(svc *services.NotificacaoService) *NotificacaoHandler {
	return &NotificacaoHandler{svc: svc}
}

func (h *NotificacaoHandler) Listar(c *gin.Context) {
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	apenasNaoLidas := c.Query("nao_lidas") == "true"
	lista, err := h.svc.Listar(uid, apenasNaoLidas, queryInt(c, "limite", 30))
	if err != nil {
		responderErro(c, err)
		return
	}
	total, _ := h.svc.ContarNaoLidas(uid)
	c.JSON(http.StatusOK, gin.H{"dados": lista, "nao_lidas": total})
}

func (h *NotificacaoHandler) ContarNaoLidas(c *gin.Context) {
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	total, err := h.svc.ContarNaoLidas(uid)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"nao_lidas": total})
}

func (h *NotificacaoHandler) MarcarLida(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	if err := h.svc.MarcarLida(id, uid); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificacaoHandler) MarcarTodasLidas(c *gin.Context) {
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	if err := h.svc.MarcarTodasLidas(uid); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
