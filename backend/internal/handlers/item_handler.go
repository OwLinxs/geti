package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type ItemHandler struct {
	svc    *services.ItemService
	movSvc *services.MovimentacaoService
}

func NewItemHandler(svc *services.ItemService, movSvc *services.MovimentacaoService) *ItemHandler {
	return &ItemHandler{svc: svc, movSvc: movSvc}
}

type itemRequest struct {
	Descricao         string                   `json:"descricao"`
	CategoriaID       uint                     `json:"categoria_id"`
	NumeroPatrimonio  *string                  `json:"numero_patrimonio"`
	NumeroSerie       *string                  `json:"numero_serie"`
	Marca             string                   `json:"marca"`
	Modelo            string                   `json:"modelo"`
	EstadoConservacao models.EstadoConservacao `json:"estado_conservacao"`
	Quantidade        int                      `json:"quantidade"`
	EstoqueMinimo     int                      `json:"estoque_minimo"`
	SetorID           *uint                    `json:"setor_id"`
	ResponsavelID     *uint                    `json:"responsavel_id"`
	DataAquisicao     *time.Time               `json:"data_aquisicao"`
	Valor             *float64                 `json:"valor"`
}

func (r itemRequest) toEntrada() services.EntradaItem {
	return services.EntradaItem{
		Descricao:         r.Descricao,
		CategoriaID:       r.CategoriaID,
		NumeroPatrimonio:  r.NumeroPatrimonio,
		NumeroSerie:       r.NumeroSerie,
		Marca:             r.Marca,
		Modelo:            r.Modelo,
		EstadoConservacao: r.EstadoConservacao,
		Quantidade:        r.Quantidade,
		EstoqueMinimo:     r.EstoqueMinimo,
		SetorID:           r.SetorID,
		ResponsavelID:     r.ResponsavelID,
		DataAquisicao:     r.DataAquisicao,
		Valor:             r.Valor,
	}
}

// Listar suporta busca e filtros via query string.
func (h *ItemHandler) Listar(c *gin.Context) {
	f := repositories.FiltroItem{
		Texto:         c.Query("q"),
		CategoriaID:   queryUint(c, "categoria_id"),
		SetorID:       queryUint(c, "setor_id"),
		ResponsavelID: queryUint(c, "responsavel_id"),
		Pagina:        queryInt(c, "pagina", 1),
		TamanhoPagina: queryInt(c, "tamanho", 20),
	}
	if estado := c.Query("estado"); estado != "" {
		e := models.EstadoConservacao(estado)
		f.Estado = &e
	}
	if c.Query("baixado") != "" {
		b := c.Query("baixado") == "true"
		f.SomenteBaixados = &b
	}
	if c.Query("abaixo_minimo") == "true" {
		f.AbaixoDoMinimo = true
	}

	itens, total, err := h.svc.Listar(f)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"dados":   itens,
		"total":   total,
		"pagina":  f.Pagina,
		"tamanho": f.TamanhoPagina,
	})
}

func (h *ItemHandler) Criar(c *gin.Context) {
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	item, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ItemHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ItemHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	item, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// Excluir remove (soft delete) um item criado por engano. Admin only (rota).
func (h *ItemHandler) Excluir(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Excluir(id); err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": "Item excluído com sucesso."})
}

// Historico devolve o histórico de movimentações do item.
func (h *ItemHandler) Historico(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	movs, err := h.movSvc.HistoricoPorItem(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, movs)
}

// AlertasEstoqueBaixo lista consumíveis abaixo do mínimo.
func (h *ItemHandler) AlertasEstoqueBaixo(c *gin.Context) {
	itens, err := h.svc.ListarAbaixoDoMinimo()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, itens)
}

// Importar recebe um CSV (campo "arquivo" em multipart, ou o corpo bruto) e
// cadastra itens em massa. Query ?validar=true faz apenas simulação (dry-run),
// devolvendo o relatório de erros sem gravar nada.
func (h *ItemHandler) Importar(c *gin.Context) {
	dryRun := c.Query("validar") == "true"
	if dryRun {
		// Simulação não altera dados: não polui a trilha de auditoria.
		c.Set(middlewares.CtxAuditSkip, true)
	}

	reader := c.Request.Body
	if arquivo, cabecalho, err := c.Request.FormFile("arquivo"); err == nil {
		defer arquivo.Close()
		_ = cabecalho
		res, err := h.svc.ImportarCSV(arquivo, dryRun)
		h.responderImportacao(c, res, err, dryRun)
		return
	}
	res, err := h.svc.ImportarCSV(reader, dryRun)
	h.responderImportacao(c, res, err, dryRun)
}

func (h *ItemHandler) responderImportacao(c *gin.Context, res *services.ResultadoImportacao, err error, dryRun bool) {
	if err != nil {
		responderErro(c, err)
		return
	}
	if !dryRun {
		// Enriquecimento da auditoria.
		c.Set(middlewares.CtxAuditDetalhe,
			fmt.Sprintf("Importou %d item(ns) via CSV (%d erro(s)).", res.Importados, len(res.Erros)))
	}
	c.JSON(http.StatusOK, res)
}

// ModeloCSV devolve um arquivo CSV modelo (cabeçalho + linha de exemplo) para
// o usuário preencher e reimportar.
func (h *ItemHandler) ModeloCSV(c *gin.Context) {
	exemplo := []string{
		"Notebook Dell Latitude 5440", "Notebook", "PM-2026-00123", "SN123456",
		"Dell", "Latitude 5440", "novo", "1", "0", "Departamento de T.I.",
		"", "2026-01-15", "4500.00",
	}
	conteudo := strings.Join(services.CabecalhoCSVItens, ",") + "\n" +
		strings.Join(exemplo, ",") + "\n"

	c.Header("Content-Disposition", "attachment; filename=modelo-itens.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(conteudo))
}
