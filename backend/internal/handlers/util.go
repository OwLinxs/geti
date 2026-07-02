package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseID extrai e valida um parâmetro de rota numérico.
func parseID(c *gin.Context, nome string) (uint, bool) {
	raw := c.Param(nome)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Identificador inválido."})
		return 0, false
	}
	return uint(id), true
}

// queryUint lê um parâmetro de query opcional como *uint.
func queryUint(c *gin.Context, nome string) *uint {
	raw := c.Query(nome)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	u := uint(v)
	return &u
}

// queryInt lê um parâmetro de query como int com valor padrão.
func queryInt(c *gin.Context, nome string, padrao int) int {
	raw := c.Query(nome)
	if raw == "" {
		return padrao
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return padrao
	}
	return v
}
