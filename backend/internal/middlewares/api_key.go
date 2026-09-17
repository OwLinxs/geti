package middlewares

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChaveAPIIntegracao protege as rotas de integração externa. Autentica por
// header "X-API-Key" (ou "Authorization: Bearer <chave>"), comparado em tempo
// constante com a chave configurada. Se a chave não estiver configurada, a
// integração fica desabilitada (503), evitando exposição acidental.
func ChaveAPIIntegracao(chave string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if chave == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"erro": "Integração desabilitada. Configure INTEGRACAO_API_KEY no servidor.",
			})
			return
		}

		informada := c.GetHeader("X-API-Key")
		if informada == "" {
			// Aceita também via Bearer, por conveniência de alguns clientes.
			const prefixo = "Bearer "
			if h := c.GetHeader("Authorization"); len(h) > len(prefixo) && h[:len(prefixo)] == prefixo {
				informada = h[len(prefixo):]
			}
		}

		if subtle.ConstantTimeCompare([]byte(informada), []byte(chave)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"erro": "Chave de API inválida.",
			})
			return
		}

		// Marca a requisição como integração (para auditoria).
		c.Set("integracao", true)
		c.Next()
	}
}
