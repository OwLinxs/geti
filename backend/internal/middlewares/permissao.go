package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/services"
)

// ExigirPermissao garante que o usuário autenticado possua a permissão
// informada. O administrador sempre passa (super-admin); os demais dependem de
// concessão explícita por usuário. A verificação é por requisição (sem cache),
// então mudanças de permissão têm efeito imediato, sem necessidade de novo login.
func ExigirPermissao(svc *services.PermissaoService, chave string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perfil, ok := PerfilDoContexto(c)
		if !ok {
			abortar(c, http.StatusForbidden, "Você não tem permissão para esta operação.")
			return
		}
		id, ok := UsuarioIDDoContexto(c)
		if !ok {
			abortar(c, http.StatusForbidden, "Você não tem permissão para esta operação.")
			return
		}
		tem, err := svc.Tem(id, perfil, chave)
		if err != nil {
			abortar(c, http.StatusInternalServerError, "Falha ao verificar permissões.")
			return
		}
		if !tem {
			abortar(c, http.StatusForbidden, "Você não tem permissão para esta operação.")
			return
		}
		c.Next()
	}
}
