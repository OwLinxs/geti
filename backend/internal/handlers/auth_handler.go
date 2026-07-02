package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/services"
)

type AuthHandler struct {
	auth           *services.AuthService
	usuarioService *services.UsuarioService
}

func NewAuthHandler(auth *services.AuthService, usuarioService *services.UsuarioService) *AuthHandler {
	return &AuthHandler{auth: auth, usuarioService: usuarioService}
}

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// Login autentica e devolve o token JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	// Contexto para a trilha de auditoria (registra a tentativa, com ou sem
	// sucesso — ver middleware Auditoria).
	c.Set(middlewares.CtxAuditEmail, req.Email)

	res, err := h.auth.Login(req.Email, req.Senha)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.Set(middlewares.CtxAuditUsuarioID, res.Usuario.ID)
	c.Set(middlewares.CtxAuditUsuarioNome, res.Usuario.Nome)
	c.JSON(http.StatusOK, res)
}

// EuMesmo devolve os dados do usuário autenticado (a partir do token).
func (h *AuthHandler) EuMesmo(c *gin.Context) {
	id, ok := middlewares.UsuarioIDDoContexto(c)
	if !ok {
		responderErro(c, services.ErrNaoAutorizado)
		return
	}
	u, err := h.usuarioService.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}
