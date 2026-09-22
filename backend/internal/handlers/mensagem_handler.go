package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/services"
)

type MensagemHandler struct {
	svc *services.MensagemService
}

func NewMensagemHandler(svc *services.MensagemService) *MensagemHandler {
	return &MensagemHandler{svc: svc}
}

func (h *MensagemHandler) Listar(c *gin.Context) {
	osID, ok := parseID(c, "id")
	if !ok {
		return
	}
	msgs, err := h.svc.ListarPorOS(osID)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// Enviar registra uma mensagem da equipe (multipart: texto, interna, arquivo).
func (h *MensagemHandler) Enviar(c *gin.Context) {
	osID, ok := parseID(c, "id")
	if !ok {
		return
	}

	usuarioID, _ := middlewares.UsuarioIDDoContexto(c)
	arquivo, _ := c.FormFile("arquivo") // opcional

	entrada := services.EntradaMensagem{
		OrdemServicoID: osID,
		AutorID:        usuarioID,
		AutorNome:      middlewares.NomeDoContexto(c),
		Texto:          c.PostForm("texto"),
		Interna:        c.PostForm("interna") == "true",
		Arquivo:        arquivo,
	}

	msg, err := h.svc.Enviar(entrada)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// BaixarAnexo devolve o arquivo anexado a uma mensagem.
func (h *MensagemHandler) BaixarAnexo(c *gin.Context) {
	osID, ok := parseID(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseID(c, "msgId")
	if !ok {
		return
	}
	m, err := h.svc.BuscarAnexo(osID, msgID)
	if err != nil {
		responderErro(c, err)
		return
	}
	servirAnexo(c, m.AnexoTipo, m.AnexoNome, m.AnexoCaminho)
}
