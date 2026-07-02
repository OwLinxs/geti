package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

type RelatorioHandler struct {
	svc *services.RelatorioService
}

func NewRelatorioHandler(svc *services.RelatorioService) *RelatorioHandler {
	return &RelatorioHandler{svc: svc}
}

// ItensPorSetor: ?setor_id= opcional; ?formato=json|csv|pdf
func (h *RelatorioHandler) ItensPorSetor(c *gin.Context) {
	itens, err := h.svc.ItensPorSetor(queryUint(c, "setor_id"))
	if err != nil {
		responderErro(c, err)
		return
	}
	h.responderItens(c, "Relatório de Itens por Setor", "itens-por-setor", itens)
}

func (h *RelatorioHandler) ItensPorResponsavel(c *gin.Context) {
	itens, err := h.svc.ItensPorResponsavel(queryUint(c, "responsavel_id"))
	if err != nil {
		responderErro(c, err)
		return
	}
	h.responderItens(c, "Relatório de Itens por Responsável", "itens-por-responsavel", itens)
}

func (h *RelatorioHandler) EstoqueBaixo(c *gin.Context) {
	itens, err := h.svc.EstoqueBaixo()
	if err != nil {
		responderErro(c, err)
		return
	}
	h.responderItens(c, "Relatório de Estoque Abaixo do Mínimo", "estoque-baixo", itens)
}

func (h *RelatorioHandler) Inventario(c *gin.Context) {
	itens, err := h.svc.InventarioCompleto()
	if err != nil {
		responderErro(c, err)
		return
	}
	h.responderItens(c, "Inventário Geral", "inventario", itens)
}

// Movimentacoes por período: ?data_inicio=&data_fim=&tipo=&formato=
func (h *RelatorioHandler) Movimentacoes(c *gin.Context) {
	var inicio, fim *time.Time
	if di := c.Query("data_inicio"); di != "" {
		if t, err := time.Parse("2006-01-02", di); err == nil {
			inicio = &t
		}
	}
	if df := c.Query("data_fim"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			f := t.Add(24*time.Hour - time.Second)
			fim = &f
		}
	}
	var tipo *models.TipoMovimentacao
	if t := c.Query("tipo"); t != "" {
		tm := models.TipoMovimentacao(t)
		tipo = &tm
	}

	movs, err := h.svc.MovimentacoesPorPeriodo(inicio, fim, tipo)
	if err != nil {
		responderErro(c, err)
		return
	}

	switch c.Query("formato") {
	case "csv":
		csv, err := services.CSVMovimentacoes(movs)
		if err != nil {
			responderErro(c, err)
			return
		}
		baixarArquivo(c, "movimentacoes.csv", "text/csv", csv)
	default:
		c.JSON(http.StatusOK, movs)
	}
}

// responderItens devolve a lista no formato solicitado (json padrão / csv / pdf).
func (h *RelatorioHandler) responderItens(c *gin.Context, titulo, nomeArquivo string, itens []models.Item) {
	switch c.Query("formato") {
	case "csv":
		csv, err := services.CSVItens(itens)
		if err != nil {
			responderErro(c, err)
			return
		}
		baixarArquivo(c, nomeArquivo+".csv", "text/csv", csv)
	case "pdf":
		pdf, err := services.PDFRelatorioItens(titulo, itens, h.svc.Config())
		if err != nil {
			responderErro(c, err)
			return
		}
		baixarArquivo(c, nomeArquivo+".pdf", "application/pdf", pdf)
	default:
		c.JSON(http.StatusOK, itens)
	}
}

func baixarArquivo(c *gin.Context, nome, tipo string, dados []byte) {
	c.Header("Content-Disposition", "attachment; filename=\""+nome+"\"")
	c.Data(http.StatusOK, tipo, dados)
}
