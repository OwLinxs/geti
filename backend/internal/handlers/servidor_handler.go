package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/services"
)

type ServidorHandler struct {
	svc *services.ServidorService
}

func NewServidorHandler(svc *services.ServidorService) *ServidorHandler {
	return &ServidorHandler{svc: svc}
}

// LGPD: somente nome e matrícula são coletados.
type servidorRequest struct {
	Nome      string `json:"nome"`
	Matricula string `json:"matricula"`
	SetorID   *uint  `json:"setor_id"`
	Ativo     *bool  `json:"ativo"`
}

func (r servidorRequest) toEntrada() services.EntradaServidor {
	ativo := true
	if r.Ativo != nil {
		ativo = *r.Ativo
	}
	return services.EntradaServidor{Nome: r.Nome, Matricula: r.Matricula, SetorID: r.SetorID, Ativo: ativo}
}

func (h *ServidorHandler) Listar(c *gin.Context) {
	ss, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ss)
}

func (h *ServidorHandler) Criar(c *gin.Context) {
	var req servidorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	srv, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, srv)
}

func (h *ServidorHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	srv, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, srv)
}

func (h *ServidorHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req servidorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	srv, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, srv)
}

func (h *ServidorHandler) Remover(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Remover(id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Importar recebe um CSV (campo "arquivo" em multipart, ou o corpo bruto) e
// cadastra servidores em massa. Query ?validar=true faz apenas simulação.
func (h *ServidorHandler) Importar(c *gin.Context) {
	dryRun := c.Query("validar") == "true"
	if dryRun {
		c.Set(middlewares.CtxAuditSkip, true)
	}

	reader := c.Request.Body
	if arquivo, _, err := c.Request.FormFile("arquivo"); err == nil {
		defer arquivo.Close()
		res, err := h.svc.ImportarCSV(arquivo, dryRun)
		h.responderImportacao(c, res, err, dryRun)
		return
	}
	res, err := h.svc.ImportarCSV(reader, dryRun)
	h.responderImportacao(c, res, err, dryRun)
}

func (h *ServidorHandler) responderImportacao(c *gin.Context, res *services.ResultadoImportacao, err error, dryRun bool) {
	if err != nil {
		responderErro(c, err)
		return
	}
	if !dryRun {
		c.Set(middlewares.CtxAuditDetalhe,
			fmt.Sprintf("Importou %d servidor(es) via CSV (%d erro(s)).", res.Importados, len(res.Erros)))
	}
	c.JSON(http.StatusOK, res)
}

// ModeloCSV devolve um CSV modelo (cabeçalho + linha de exemplo).
func (h *ServidorHandler) ModeloCSV(c *gin.Context) {
	exemplo := []string{"Maria da Silva", "12345", "Departamento de T.I.", "sim"}
	conteudo := strings.Join(services.CabecalhoCSVServidores, ",") + "\n" +
		strings.Join(exemplo, ",") + "\n"

	c.Header("Content-Disposition", "attachment; filename=modelo-servidores.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(conteudo))
}
