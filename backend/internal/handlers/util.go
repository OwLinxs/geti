package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// servirAnexo entrega um arquivo anexado com cabeçalhos de segurança:
// nosniff sempre, e exibição inline apenas para imagem/PDF (demais tipos são
// baixados como anexo, evitando execução/render de conteúdo no navegador).
func servirAnexo(c *gin.Context, tipo, nome, caminho string) {
	c.Header("X-Content-Type-Options", "nosniff")
	if tipo != "" {
		c.Header("Content-Type", tipo)
	}
	disp := "attachment"
	if strings.HasPrefix(tipo, "image/") || tipo == "application/pdf" {
		disp = "inline"
	}
	c.Header("Content-Disposition", disp+"; filename=\""+nome+"\"")
	c.File(caminho)
}

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

// queryIntPtr lê um parâmetro de query como *int (nil quando ausente/inválido).
func queryIntPtr(c *gin.Context, nome string) *int {
	raw := c.Query(nome)
	if raw == "" {
		return nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &v
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
