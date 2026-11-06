package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type InventarioHandler struct {
	svc *services.InventarioService
}

func NewInventarioHandler(svc *services.InventarioService) *InventarioHandler {
	return &InventarioHandler{svc: svc}
}

type especificacaoRequest struct {
	Nome  string `json:"nome"`
	Valor string `json:"valor"`
}

type coletaInventarioRequest struct {
	Hostname       string                 `json:"hostname"`
	NumeroSerie    string                 `json:"numero_serie"`
	Fabricante     string                 `json:"fabricante"`
	Modelo         string                 `json:"modelo"`
	Categoria      string                 `json:"categoria"`
	SetorID        *uint                  `json:"setor_id"`
	Especificacoes []especificacaoRequest `json:"especificacoes"`
}

// Coletar recebe a configuração de uma máquina (script de inventário) e faz o
// upsert do item. Responde 201 se criou, 200 se atualizou.
func (h *InventarioHandler) Coletar(c *gin.Context) {
	var req coletaInventarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	entrada := services.EntradaInventario{
		Hostname:    req.Hostname,
		NumeroSerie: req.NumeroSerie,
		Fabricante:  req.Fabricante,
		Modelo:      req.Modelo,
		Categoria:   req.Categoria,
		SetorID:     req.SetorID,
	}
	for _, e := range req.Especificacoes {
		entrada.Especificacoes = append(entrada.Especificacoes, services.EspecificacaoColeta{
			Nome:  e.Nome,
			Valor: e.Valor,
		})
	}

	item, criado, err := h.svc.Coletar(entrada)
	if err != nil {
		responderErro(c, err)
		return
	}
	status := http.StatusOK
	if criado {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"item": item, "criado": criado})
}

// Setores lista os departamentos (id/nome/sigla) para o admin mapear os IDs
// usados no parâmetro -SetorId do script de coleta.
func (h *InventarioHandler) Setores(c *gin.Context) {
	setores, err := h.svc.ListarSetores()
	if err != nil {
		responderErro(c, err)
		return
	}
	saida := make([]gin.H, 0, len(setores))
	for _, s := range setores {
		saida = append(saida, gin.H{"id": s.ID, "nome": s.Nome, "sigla": s.Sigla})
	}
	c.JSON(http.StatusOK, saida)
}
