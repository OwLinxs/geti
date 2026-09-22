package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type ConfiguracaoHandler struct {
	svc *services.ConfiguracaoService
}

func NewConfiguracaoHandler(svc *services.ConfiguracaoService) *ConfiguracaoHandler {
	return &ConfiguracaoHandler{svc: svc}
}

// mascarar remove a senha SMTP da resposta (não trafega o segredo pro cliente).
func mascarar(cfg services.ConfigChamados) services.ConfigChamados {
	cfg.SMTPSenha = ""
	return cfg
}

func (h *ConfiguracaoHandler) ObterChamados(c *gin.Context) {
	cfg, err := h.svc.ObterChamados()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, mascarar(cfg))
}

func (h *ConfiguracaoHandler) SalvarChamados(c *gin.Context) {
	var req services.ConfigChamados
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	// Senha em branco = manter a atual (o GET nunca devolve a senha).
	if req.SMTPSenha == "" {
		if atual, err := h.svc.ObterChamados(); err == nil {
			req.SMTPSenha = atual.SMTPSenha
		}
	}
	cfg, err := h.svc.SalvarChamados(req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, mascarar(cfg))
}
