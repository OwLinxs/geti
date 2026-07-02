package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

// responderErro mapeia erros de domínio para respostas HTTP padronizadas,
// sempre com mensagens em português.
func responderErro(c *gin.Context, err error) {
	// Erros de validação por campo.
	var ev *services.ErroValidacao
	if errors.As(err, &ev) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"erro":   "Há campos inválidos.",
			"campos": ev.Campos,
		})
		return
	}

	switch {
	case errors.Is(err, services.ErrNaoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"erro": "Registro não encontrado."})
	case errors.Is(err, services.ErrCredenciaisInvalidas):
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "E-mail ou senha inválidos."})
	case errors.Is(err, services.ErrNaoAutorizado):
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Acesso não autorizado."})
	case errors.Is(err, services.ErrEstoqueInsuficiente):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	case errors.Is(err, services.ErrItemBaixado):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	case errors.Is(err, services.ErrItemComHistorico):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	case errors.Is(err, services.ErrConflito):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	case errors.Is(err, services.ErrRegraNegocio):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno ao processar a solicitação."})
	}
}

// erroBind responde a falhas de desserialização do corpo da requisição.
func erroBind(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"erro":     "Corpo da requisição inválido.",
		"detalhe":  err.Error(),
	})
}
