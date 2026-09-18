package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

// MeusChamadosHandler é o portal do solicitante (usuário final): abre e
// acompanha os PRÓPRIOS chamados e conversa. Todo acesso é escopado ao usuário
// autenticado — não enxerga chamados de terceiros.
type MeusChamadosHandler struct {
	osSvc  *services.OrdemServicoService
	msgSvc *services.MensagemService
}

func NewMeusChamadosHandler(
	osSvc *services.OrdemServicoService,
	msgSvc *services.MensagemService,
) *MeusChamadosHandler {
	return &MeusChamadosHandler{osSvc: osSvc, msgSvc: msgSvc}
}

// carregarProprio busca o chamado e garante que pertence ao usuário. Em caso
// contrário responde 404 (não revela existência de chamados alheios).
func (h *MeusChamadosHandler) carregarProprio(c *gin.Context) (*models.OrdemServico, bool) {
	id, ok := parseID(c, "id")
	if !ok {
		return nil, false
	}
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	os, err := h.osSvc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return nil, false
	}
	if os.AbertoPorID != uid {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Chamado não encontrado."})
		return nil, false
	}
	return os, true
}

type abrirChamadoRequest struct {
	Assunto              string `json:"assunto"`
	DefeitoRelatado      string `json:"defeito_relatado"`
	EquipamentoDescricao string `json:"equipamento_descricao"`
	CategoriaChamadoID   *uint  `json:"categoria_chamado_id"`
	Prioridade           string `json:"prioridade"`
}

// Abrir cria um chamado em nome do próprio solicitante.
func (h *MeusChamadosHandler) Abrir(c *gin.Context) {
	var req abrirChamadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	uid, _ := middlewares.UsuarioIDDoContexto(c)

	os, err := h.osSvc.Criar(services.EntradaOS{
		Assunto:              req.Assunto,
		DefeitoRelatado:      req.DefeitoRelatado,
		EquipamentoDescricao: req.EquipamentoDescricao,
		CategoriaChamadoID:   req.CategoriaChamadoID,
		Prioridade:           req.Prioridade,
		Origem:               "portal",
		AbertoPorID:          uid,
		SolicitanteNome:      middlewares.NomeDoContexto(c),
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, os)
}

// Listar devolve os chamados do próprio solicitante (paginado, filtro status).
func (h *MeusChamadosHandler) Listar(c *gin.Context) {
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	f := repositories.FiltroOrdemServico{
		AbertoPorID: &uid,
		Status:      c.Query("status"),
		Pagina:      queryInt(c, "pagina", 1),
		Tamanho:     queryInt(c, "tamanho", 20),
	}
	lista, total, err := h.osSvc.Listar(f)
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

func (h *MeusChamadosHandler) BuscarPorID(c *gin.Context) {
	os, ok := h.carregarProprio(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, os)
}

func (h *MeusChamadosHandler) ListarMensagens(c *gin.Context) {
	os, ok := h.carregarProprio(c)
	if !ok {
		return
	}
	msgs, err := h.msgSvc.ListarPublicasPorOS(os.ID)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, msgs)
}

func (h *MeusChamadosHandler) EnviarMensagem(c *gin.Context) {
	os, ok := h.carregarProprio(c)
	if !ok {
		return
	}
	uid, _ := middlewares.UsuarioIDDoContexto(c)
	arquivo, _ := c.FormFile("arquivo")

	msg, err := h.msgSvc.Enviar(services.EntradaMensagem{
		OrdemServicoID: os.ID,
		AutorID:        uid,
		AutorNome:      middlewares.NomeDoContexto(c),
		AutorTipo:      models.AutorServidor,
		Texto:          c.PostForm("texto"),
		Arquivo:        arquivo,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *MeusChamadosHandler) BaixarAnexo(c *gin.Context) {
	os, ok := h.carregarProprio(c)
	if !ok {
		return
	}
	msgID, ok := parseID(c, "msgId")
	if !ok {
		return
	}
	m, err := h.msgSvc.BuscarAnexo(os.ID, msgID)
	if err != nil {
		responderErro(c, err)
		return
	}
	if m.AnexoTipo != "" {
		c.Header("Content-Type", m.AnexoTipo)
	}
	c.Header("Content-Disposition", "inline; filename=\""+m.AnexoNome+"\"")
	c.File(m.AnexoCaminho)
}
