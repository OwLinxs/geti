package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

type OrdemServicoHandler struct {
	svc *services.OrdemServicoService
}

func NewOrdemServicoHandler(svc *services.OrdemServicoService) *OrdemServicoHandler {
	return &OrdemServicoHandler{svc: svc}
}

type ordemServicoRequest struct {
	ItemID                   uint   `json:"item_id"`
	EquipamentoDescricao     string `json:"equipamento_descricao"`
	EquipamentoIdentificacao string `json:"equipamento_identificacao"`
	SetorID                  *uint  `json:"setor_id"`
	CategoriaChamadoID       *uint  `json:"categoria_chamado_id"`
	SolicitanteID            *uint  `json:"solicitante_id"`
	DefeitoRelatado          string `json:"defeito_relatado"`
	Diagnostico              string `json:"diagnostico"`
	SolucaoAplicada          string `json:"solucao_aplicada"`
	Prioridade               string `json:"prioridade"`
	TecnicoID                *uint  `json:"tecnico_id"`
}

func (r ordemServicoRequest) toEntrada() services.EntradaOS {
	return services.EntradaOS{
		ItemID:                   r.ItemID,
		EquipamentoDescricao:     r.EquipamentoDescricao,
		EquipamentoIdentificacao: r.EquipamentoIdentificacao,
		SetorID:                  r.SetorID,
		CategoriaChamadoID:       r.CategoriaChamadoID,
		SolicitanteID:            r.SolicitanteID,
		DefeitoRelatado:          r.DefeitoRelatado,
		Diagnostico:              r.Diagnostico,
		SolucaoAplicada:          r.SolucaoAplicada,
		Prioridade:               r.Prioridade,
		TecnicoID:                r.TecnicoID,
	}
}

func (h *OrdemServicoHandler) Criar(c *gin.Context) {
	var req ordemServicoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	entrada := req.toEntrada()
	entrada.AbertoPorID, _ = middlewares.UsuarioIDDoContexto(c)

	os, err := h.svc.Criar(entrada)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, os)
}

func (h *OrdemServicoHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req ordemServicoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	os, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}

func (h *OrdemServicoHandler) Listar(c *gin.Context) {
	f := repositories.FiltroOrdemServico{
		Status:    c.Query("status"),
		TecnicoID: queryUint(c, "tecnico_id"),
		De:        parseDataQuery(c.Query("de"), false),
		Ate:       parseDataQuery(c.Query("ate"), true),
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

func (h *OrdemServicoHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	os, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *OrdemServicoHandler) DefinirStatus(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	os, err := h.svc.DefinirStatus(id, models.StatusOS(req.Status))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}

type passoRequest struct {
	Descricao  string `json:"descricao"`
	Concluido  bool   `json:"concluido"`
	Observacao string `json:"observacao"`
}

type passosRequest struct {
	Passos []passoRequest `json:"passos"`
}

func (h *OrdemServicoHandler) SalvarPassos(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req passosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	entradas := make([]services.PassoEntrada, 0, len(req.Passos))
	for _, p := range req.Passos {
		entradas = append(entradas, services.PassoEntrada{
			Descricao:  p.Descricao,
			Concluido:  p.Concluido,
			Observacao: p.Observacao,
		})
	}
	os, err := h.svc.SalvarPassos(id, entradas)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, os)
}

func (h *OrdemServicoHandler) Excluir(c *gin.Context) {
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

type documentoRequest struct {
	PassosExtra []string `json:"passos_extra"`
}

// Documento gera o PDF da OS. Aceita passos_extra no corpo para acrescentar
// itens ao checklist apenas na impressão.
func (h *OrdemServicoHandler) Documento(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req documentoRequest
	// Corpo opcional: sem body, gera com o checklist salvo.
	_ = c.ShouldBindJSON(&req)

	pdf, os, err := h.svc.GerarPDF(id, req.PassosExtra)
	if err != nil {
		responderErro(c, err)
		return
	}
	nome := fmt.Sprintf("ordem-servico-%s.pdf", os.Numero)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", nome))
	c.Data(http.StatusOK, "application/pdf", pdf)
}
