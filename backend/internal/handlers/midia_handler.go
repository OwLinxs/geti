package handlers

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

// MidiaHandler serve anexos de mensagens por um link PÚBLICO assinado (sem
// login), para o gateway de WhatsApp (ZapGov) conseguir buscar a mídia que o
// técnico enviou. A assinatura (HMAC) impede acesso a anexos não autorizados.
type MidiaHandler struct {
	svc    *services.MensagemService
	secret string
}

func NewMidiaHandler(svc *services.MensagemService, secret string) *MidiaHandler {
	return &MidiaHandler{svc: svc, secret: secret}
}

// Servir valida a assinatura e entrega o arquivo do anexo.
func (h *MidiaHandler) Servir(c *gin.Context) {
	osID, ok := parseID(c, "osId")
	if !ok {
		return
	}
	msgID, ok := parseID(c, "msgId")
	if !ok {
		return
	}
	esperado := services.AssinaturaMidia(h.secret, osID, msgID)
	if subtle.ConstantTimeCompare([]byte(c.Query("sig")), []byte(esperado)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"erro": "assinatura inválida"})
		return
	}
	m, err := h.svc.BuscarAnexo(osID, msgID)
	if err != nil {
		responderErro(c, err)
		return
	}
	servirAnexo(c, m.AnexoTipo, m.AnexoNome, m.AnexoCaminho)
}
