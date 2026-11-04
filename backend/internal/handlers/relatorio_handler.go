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

// Catalogo lista os tipos de relatório e os filtros aplicáveis a cada um.
func (h *RelatorioHandler) Catalogo(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Catalogo())
}

// Gerar é o endpoint único do construtor de relatórios.
//
//	GET /relatorios?tipo=&secretaria_id=&setor_id=&responsavel_id=&categoria_id=
//	    &estado=&data_inicio=&data_fim=&formato=json|csv|pdf
func (h *RelatorioHandler) Gerar(c *gin.Context) {
	p := services.ParametrosRelatorio{
		Tipo:          c.Query("tipo"),
		SecretariaID:  queryUint(c, "secretaria_id"),
		SetorID:       queryUint(c, "setor_id"),
		ResponsavelID: queryUint(c, "responsavel_id"),
		CategoriaID:   queryUint(c, "categoria_id"),
		DataInicio:    queryData(c, "data_inicio", false),
		DataFim:       queryData(c, "data_fim", true),
	}
	if e := c.Query("estado"); e != "" {
		est := models.EstadoConservacao(e)
		p.Estado = &est
	}
	if tm := c.Query("tipo_mov"); tm != "" {
		mv := models.TipoMovimentacao(tm)
		p.TipoMov = &mv
	}

	res, err := h.svc.Gerar(p)
	if err != nil {
		responderErro(c, err)
		return
	}

	switch c.Query("formato") {
	case "csv":
		h.baixarCSV(c, res)
	case "pdf":
		h.baixarPDF(c, res)
	default:
		c.JSON(http.StatusOK, res)
	}
}

func (h *RelatorioHandler) baixarCSV(c *gin.Context, res *services.ResultadoRelatorio) {
	if res.EhEvento {
		dados, err := services.CSVMovimentacoes(res.Movimentacoes)
		if err != nil {
			responderErro(c, err)
			return
		}
		baixarArquivo(c, res.Tipo+".csv", "text/csv", dados)
		return
	}
	dados, err := services.CSVItens(res.Itens)
	if err != nil {
		responderErro(c, err)
		return
	}
	baixarArquivo(c, res.Tipo+".csv", "text/csv", dados)
}

func (h *RelatorioHandler) baixarPDF(c *gin.Context, res *services.ResultadoRelatorio) {
	if res.EhEvento {
		dados, err := services.PDFRelatorioMovimentacoes(res.Titulo, res.Movimentacoes, h.svc.Config())
		if err != nil {
			responderErro(c, err)
			return
		}
		baixarArquivo(c, res.Tipo+".pdf", "application/pdf", dados)
		return
	}
	dados, err := services.PDFRelatorioItens(res.Titulo, res.Itens, h.svc.Config())
	if err != nil {
		responderErro(c, err)
		return
	}
	baixarArquivo(c, res.Tipo+".pdf", "application/pdf", dados)
}

// queryData faz o parse de uma data YYYY-MM-DD. Quando fimDoDia é true, ajusta
// para o último instante do dia (filtro inclusivo em data_fim).
func queryData(c *gin.Context, chave string, fimDoDia bool) *time.Time {
	v := c.Query(chave)
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	if fimDoDia {
		t = t.Add(24*time.Hour - time.Second)
	}
	return &t
}

func baixarArquivo(c *gin.Context, nome, tipo string, dados []byte) {
	c.Header("Content-Disposition", "attachment; filename=\""+nome+"\"")
	c.Data(http.StatusOK, tipo, dados)
}
