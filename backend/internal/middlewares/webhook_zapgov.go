package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CtxRawBody guarda o corpo cru da requisição (lido pela validação HMAC) para o
// handler reaproveitar sem reler o Body.
const CtxRawBody = "rawBody"

// WebhookZapGov protege o endpoint de webhooks do ZapGov. Exige X-API-Key igual
// à chave do sistema e, quando o segredo HMAC está configurado, valida a
// assinatura X-Signature (HMAC-SHA256 do corpo cru). Rejeita com 401.
func WebhookZapGov(apiKey, hmacSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"erro": "Integração desabilitada. Configure INTEGRACAO_API_KEY no servidor.",
			})
			return
		}

		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"erro": "corpo inválido"})
			return
		}
		// Restaura o Body para o ShouldBindJSON do handler e guarda o cru.
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		c.Set(CtxRawBody, raw)

		if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-API-Key")), []byte(apiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "Chave de API inválida."})
			return
		}

		if hmacSecret != "" {
			sig := strings.TrimPrefix(c.GetHeader("X-Signature"), "sha256=")
			mac := hmac.New(sha256.New, []byte(hmacSecret))
			mac.Write(raw)
			esperado := hex.EncodeToString(mac.Sum(nil))
			if subtle.ConstantTimeCompare([]byte(sig), []byte(esperado)) != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "Assinatura inválida."})
				return
			}
		}

		c.Set("integracao", true)
		c.Next()
	}
}
