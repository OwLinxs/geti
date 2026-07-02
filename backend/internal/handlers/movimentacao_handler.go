package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type MovimentacaoHandler struct {
	svc *services.MovimentacaoService
}

func NewMovimentacaoHandler(svc *services.MovimentacaoService) *MovimentacaoHandler {
	return &MovimentacaoHandler{svc: svc}
}

type movimentacaoRequest struct {
	ItemID         uint                    `json:"item_id"`
	Tipo           models.TipoMovimentacao `json:"tipo"`
	Quantidade     int                     `json:"quantidade"`
	SetorOrigemID  *uint                   `json:"setor_origem_id"`
	SetorDestinoID *uint                   `json:"setor_destino_id"`
	ServidorID     *uint                   `json:"servidor_id"`
	Observacao     string                  `json:"observacao"`
	MotivoBaixa    string                  `json:"motivo_baixa"`
	DataEvento     *time.Time              `json:"data_evento"`
}

// Registrar cria uma movimentação (entrada/saída/baixa).
func (h *MovimentacaoHandler) Registrar(c *gin.Context) {
	var req movimentacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	usuarioID, _ := middlewares.UsuarioIDDoContexto(c)

	res, err := h.svc.Registrar(services.EntradaMovimentacao{
		ItemID:          req.ItemID,
		Tipo:            req.Tipo,
		Quantidade:      req.Quantidade,
		SetorOrigemID:   req.SetorOrigemID,
		SetorDestinoID:  req.SetorDestinoID,
		ServidorID:      req.ServidorID,
		Observacao:      req.Observacao,
		MotivoBaixa:     req.MotivoBaixa,
		RegistradoPorID: usuarioID,
		DataEvento:      req.DataEvento,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"movimentacao":   res.Movimentacao,
		"item":           res.Item,
		"alerta_estoque": res.AlertaEstoque,
	})
}

// Listar consulta o histórico com filtros (item, tipo, período).
func (h *MovimentacaoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroMovimentacao{
		ItemID:        queryUint(c, "item_id"),
		Pagina:        queryInt(c, "pagina", 1),
		TamanhoPagina: queryInt(c, "tamanho", 50),
	}
	if t := c.Query("tipo"); t != "" {
		tipo := models.TipoMovimentacao(t)
		f.Tipo = &tipo
	}
	if di := c.Query("data_inicio"); di != "" {
		if t, err := time.Parse("2006-01-02", di); err == nil {
			f.DataInicio = &t
		}
	}
	if df := c.Query("data_fim"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			// inclui o dia inteiro
			fim := t.Add(24*time.Hour - time.Second)
			f.DataFim = &fim
		}
	}

	movs, total, err := h.svc.Listar(f)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"dados":   movs,
		"total":   total,
		"pagina":  f.Pagina,
		"tamanho": f.TamanhoPagina,
	})
}
