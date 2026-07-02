package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

type UsuarioHandler struct {
	svc *services.UsuarioService
}

func NewUsuarioHandler(svc *services.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{svc: svc}
}

type usuarioRequest struct {
	Nome   string        `json:"nome"`
	Email  string        `json:"email"`
	Senha  string        `json:"senha"`
	Perfil models.Perfil `json:"perfil"`
	Ativo  *bool         `json:"ativo"`
}

func (h *UsuarioHandler) Listar(c *gin.Context) {
	us, err := h.svc.Listar()
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, us)
}

func (h *UsuarioHandler) Criar(c *gin.Context) {
	var req usuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	ativo := true
	if req.Ativo != nil {
		ativo = *req.Ativo
	}
	u, err := h.svc.Criar(services.EntradaUsuario{
		Nome:   req.Nome,
		Email:  req.Email,
		Senha:  req.Senha,
		Perfil: req.Perfil,
		Ativo:  ativo,
	})
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *UsuarioHandler) BuscarPorID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.BuscarPorID(id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

type redefinirSenhaRequest struct {
	Senha string `json:"senha"`
}

// RedefinirSenha troca a senha de um usuário (admin). PATCH /usuarios/:id/senha
func (h *UsuarioHandler) RedefinirSenha(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req redefinirSenhaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	u, err := h.svc.RedefinirSenha(id, req.Senha)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

type definirAtivoRequest struct {
	Ativo *bool `json:"ativo"`
}

// DefinirAtivo ativa/desativa um usuário (admin). PATCH /usuarios/:id/ativo
func (h *UsuarioHandler) DefinirAtivo(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req definirAtivoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		erroBind(c, err)
		return
	}
	if req.Ativo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Informe o campo 'ativo' (true ou false)."})
		return
	}
	u, err := h.svc.DefinirAtivo(id, *req.Ativo)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}
