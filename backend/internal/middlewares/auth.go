package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

const (
	ctxUsuarioID = "usuario_id"
	ctxPerfil    = "perfil"
	ctxNome      = "usuario_nome"
)

// Autenticacao valida o token JWT do header Authorization e injeta as claims
// no contexto. Bloqueia requisições sem token válido.
func Autenticacao(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortar(c, http.StatusUnauthorized, "Autenticação necessária.")
			return
		}
		partes := strings.SplitN(header, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			abortar(c, http.StatusUnauthorized, "Formato de token inválido. Use 'Bearer <token>'.")
			return
		}

		claims, err := auth.ValidarToken(strings.TrimSpace(partes[1]))
		if err != nil {
			abortar(c, http.StatusUnauthorized, "Token inválido ou expirado.")
			return
		}

		c.Set(ctxUsuarioID, claims.UsuarioID)
		c.Set(ctxPerfil, claims.Perfil)
		c.Set(ctxNome, claims.Nome)
		c.Next()
	}
}

// ExigirPerfil garante que o usuário tenha um dos perfis permitidos.
func ExigirPerfil(perfis ...models.Perfil) gin.HandlerFunc {
	permitidos := make(map[models.Perfil]bool, len(perfis))
	for _, p := range perfis {
		permitidos[p] = true
	}
	return func(c *gin.Context) {
		perfil, ok := PerfilDoContexto(c)
		if !ok || !permitidos[perfil] {
			abortar(c, http.StatusForbidden, "Você não tem permissão para esta operação.")
			return
		}
		c.Next()
	}
}

// SomenteAdministrador é um atalho para ExigirPerfil(administrador).
func SomenteAdministrador() gin.HandlerFunc {
	return ExigirPerfil(models.PerfilAdministrador)
}

// UsuarioIDDoContexto recupera o ID do usuário autenticado.
func UsuarioIDDoContexto(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ctxUsuarioID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// PerfilDoContexto recupera o perfil do usuário autenticado.
func PerfilDoContexto(c *gin.Context) (models.Perfil, bool) {
	v, ok := c.Get(ctxPerfil)
	if !ok {
		return "", false
	}
	p, ok := v.(models.Perfil)
	return p, ok
}

// NomeDoContexto recupera o nome do usuário autenticado (para auditoria).
func NomeDoContexto(c *gin.Context) string {
	v, ok := c.Get(ctxNome)
	if !ok {
		return ""
	}
	nome, _ := v.(string)
	return nome
}

func abortar(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"erro": msg})
}
