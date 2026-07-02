package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type AuditoriaHandler struct {
	svc *services.AuditoriaService
}

func NewAuditoriaHandler(svc *services.AuditoriaService) *AuditoriaHandler {
	return &AuditoriaHandler{svc: svc}
}

// Listar devolve a trilha de auditoria com filtros e paginação (admin-only).
// Query: usuario_id, recurso, acao, de (YYYY-MM-DD), ate (YYYY-MM-DD),
// pagina, tamanho.
func (h *AuditoriaHandler) Listar(c *gin.Context) {
	f := repositories.FiltroAuditoria{
		UsuarioID: queryUint(c, "usuario_id"),
		Recurso:   c.Query("recurso"),
		Acao:      c.Query("acao"),
		De:        parseDataQuery(c.Query("de"), false),
		Ate:       parseDataQuery(c.Query("ate"), true),
		Pagina:    queryInt(c, "pagina", 1),
		Tamanho:   queryInt(c, "tamanho", 30),
	}

	regs, total, err := h.svc.Listar(f)
	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dados":   regs,
		"total":   total,
		"pagina":  f.Pagina,
		"tamanho": f.Tamanho,
	})
}

// parseDataQuery aceita "YYYY-MM-DD". Quando fimDoDia=true, ajusta para o fim
// do dia para que o filtro "ate" seja inclusivo.
func parseDataQuery(s string, fimDoDia bool) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	if fimDoDia {
		t = t.Add(24*time.Hour - time.Second)
	}
	return &t
}
