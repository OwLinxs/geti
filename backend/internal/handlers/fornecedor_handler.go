package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

type FornecedorHandler struct {
	svc *services.FornecedorService
}

func NewFornecedorHandler(svc *services.FornecedorService) *FornecedorHandler {
	return &FornecedorHandler{svc: svc}
}

type fornecedorRequest struct {
	Nome       string `json:"nome"`
	CNPJ       string `json:"cnpj"`
	Email      string `json:"email"`
	Telefone   string `json:"telefone"`
	Endereco   string `json:"endereco"`
	Observacao string `json:"observacao"`
}

func (r fornecedorRequest) toEntrada() services.EntradaFornecedor {
	return services.EntradaFornecedor{
		Nome:       r.Nome,
		CNPJ:       r.CNPJ,
		Email:      r.Email,
		Telefone:   r.Telefone,
		Endereco:   r.Endereco,
		Observacao: r.Observacao,
	}
}

func (h *FornecedorHandler) Criar(c *gin.Context) {
	var req fornecedorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	f, err := h.svc.Criar(req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, f)
}

func (h *FornecedorHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req fornecedorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	f, err := h.svc.Atualizar(id, req.toEntrada())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *FornecedorHandler) Listar(c *gin.Context) {
	lista, err := h.svc.Listar(c.Query("q"))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *FornecedorHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	f, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *FornecedorHandler) Excluir(c *gin.Context) {
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
