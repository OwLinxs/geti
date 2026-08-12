package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type ConhecimentoHandler struct {
	svc *services.ConhecimentoService
}

func NewConhecimentoHandler(svc *services.ConhecimentoService) *ConhecimentoHandler {
	return &ConhecimentoHandler{svc: svc}
}

type artigoRequest struct {
	Titulo    string `json:"titulo"`
	Categoria string `json:"categoria"`
	Conteudo  string `json:"conteudo"`
	Publicado *bool  `json:"publicado"`
}

func (r artigoRequest) toEntrada() services.EntradaArtigo {
	// Publicado ausente = publicado por padrão.
	pub := true
	if r.Publicado != nil {
		pub = *r.Publicado
	}
	return services.EntradaArtigo{
		Titulo:    r.Titulo,
		Categoria: r.Categoria,
		Conteudo:  r.Conteudo,
		Publicado: pub,
	}
}

func (h *ConhecimentoHandler) Criar(c *gin.Context) {
	var req artigoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	entrada := req.toEntrada()
	entrada.AutorID, _ = middlewares.UsuarioIDDoContexto(c)

	artigo, err := h.svc.Criar(entrada)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, artigo)
}

func (h *ConhecimentoHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req artigoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	artigo, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, artigo)
}

func (h *ConhecimentoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroConhecimento{
		Q:         c.Query("q"),
		Categoria: c.Query("categoria"),
		Publicado: parseBoolQuery(c.Query("publicado")),
		Pagina:    queryInt(c, "pagina", 1),
		Tamanho:   queryInt(c, "tamanho", 20),
	}
	lista, total, err := h.svc.Listar(f)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"dados":   lista,
		"total":   total,
		"pagina":  f.Pagina,
		"tamanho": f.Tamanho,
	})
}

func (h *ConhecimentoHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	artigo, err := h.svc.Visualizar(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, artigo)
}

func (h *ConhecimentoHandler) Excluir(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Excluir(id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// parseBoolQuery interpreta "true"/"false" em *bool (nil quando ausente).
func parseBoolQuery(s string) *bool {
	if s == "" {
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}
	return &b
}
