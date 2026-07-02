package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/services"
)

type TermoHandler struct {
	svc *services.TermoService
}

func NewTermoHandler(svc *services.TermoService) *TermoHandler {
	return &TermoHandler{svc: svc}
}

type termoRequest struct {
	ItemID         uint   `json:"item_id"`
	ServidorID     uint   `json:"servidor_id"`
	MovimentacaoID *uint  `json:"movimentacao_id"`
	Observacao     string `json:"observacao"`
}

func (h *TermoHandler) Emitir(c *gin.Context) {
	var req termoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	usuarioID, _ := middlewares.UsuarioIDDoContexto(c)
	termo, err := h.svc.Emitir(services.EntradaTermo{
		ItemID:         req.ItemID,
		ServidorID:     req.ServidorID,
		MovimentacaoID: req.MovimentacaoID,
		Observacao:     req.Observacao,
		EmitidoPorID:   usuarioID,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, termo)
}

func (h *TermoHandler) Listar(c *gin.Context) {
	ts, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, ts)
}

func (h *TermoHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	t, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// PDF gera e devolve o termo em PDF para download.
func (h *TermoHandler) PDF(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	pdf, termo, err := h.svc.GerarPDF(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	nome := fmt.Sprintf("termo-%s.pdf", termo.Numero)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", nome))
	c.Data(http.StatusOK, "application/pdf", pdf)
}
