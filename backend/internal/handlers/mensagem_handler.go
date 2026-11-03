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

type templateRequest struct {
	Template string   `json:"template"`
	Params   []string `json:"params"`
}

// EnviarTemplate envia um template aprovado ao contato do WhatsApp (fora da
// janela de 24h).
func (h *MensagemHandler) EnviarTemplate(c *gin.Context) {
	osID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req templateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	msg, err := h.svc.EnviarTemplateWhatsApp(osID, req.Template, req.Params)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, msg)
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
