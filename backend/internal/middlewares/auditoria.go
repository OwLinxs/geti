package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

// Chaves de contexto que os handlers podem preencher para enriquecer a
// auditoria (opcional). O middleware usa fallbacks quando ausentes.
const (
	CtxAuditUsuarioID   = "audit_usuario_id"   // uint  — autor (usado no login)
	CtxAuditUsuarioNome = "audit_usuario_nome" // string
	CtxAuditEmail       = "audit_email"        // string
	CtxAuditRecursoID   = "audit_recurso_id"   // uint  — id do recurso criado
	CtxAuditDetalhe     = "audit_detalhe"      // string — descrição amigável
	CtxAuditSkip        = "audit_skip"         // bool  — não auditar esta requisição
)

// singularRecurso mapeia o segmento de rota (plural) para um rótulo singular.
var singularRecurso = map[string]string{
	"itens":         "item",
	"usuarios":      "usuario",
	"servidores":    "servidor",
	"categorias":    "categoria",
	"setores":       "setor",
	"movimentacoes": "movimentacao",
	"termos":        "termo",
}

// Auditoria registra, após cada requisição sensível, QUEM fez O QUÊ e QUANDO.
// Audita: login (sucesso/falha) e toda operação de escrita (POST/PUT/PATCH/
// DELETE) sob /api/v1. Leituras (GET) não são auditadas.
func Auditoria(svc *services.AuditoriaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handlers podem sinalizar que a requisição não deve ser auditada
		// (ex.: simulação/dry-run que não altera dados).
		if c.GetBool(CtxAuditSkip) {
			return
		}

		path := c.Request.URL.Path
		ehLogin := strings.HasSuffix(path, "/auth/login")

		if !ehLogin && !metodoDeEscrita(c.Request.Method) {
			return
		}
		// Não audita pré-flight CORS.
		if c.Request.Method == http.MethodOptions {
			return
		}

		reg := &models.RegistroAuditoria{
			Metodo:  c.Request.Method,
			Caminho: path,
			Status:  c.Writer.Status(),
			IP:      c.ClientIP(),
		}
		preencherAutor(c, reg, ehLogin)

		if ehLogin {
			reg.Acao = "login"
			if reg.Status == http.StatusOK {
				reg.Detalhe = "Login bem-sucedido."
			} else {
				reg.Detalhe = "Tentativa de login malsucedida."
			}
		} else {
			preencherAcaoRecurso(c, reg)
		}

		svc.Registrar(reg)
	}
}

func metodoDeEscrita(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// preencherAutor define quem realizou a ação. Em rotas autenticadas vem do
// contexto (JWT); no login vem dos valores que o handler injeta ao autenticar.
func preencherAutor(c *gin.Context, reg *models.RegistroAuditoria, ehLogin bool) {
	if id, ok := UsuarioIDDoContexto(c); ok {
		reg.UsuarioID = &id
		reg.UsuarioNome = NomeDoContexto(c)
	}
	if v, ok := c.Get(CtxAuditUsuarioID); ok {
		if id, ok := v.(uint); ok && id != 0 {
			reg.UsuarioID = &id
		}
	}
	if v := c.GetString(CtxAuditUsuarioNome); v != "" {
		reg.UsuarioNome = v
	}
	if v := c.GetString(CtxAuditEmail); v != "" {
		reg.UsuarioEmail = v
	}
}

// preencherAcaoRecurso deriva ação/recurso/id a partir do método e da rota,
// com casos especiais para sub-rotas conhecidas.
func preencherAcaoRecurso(c *gin.Context, reg *models.RegistroAuditoria) {
	seg := segmentosAPI(c.Request.URL.Path)
	if len(seg) == 0 {
		reg.Acao = acaoPorMetodo(c.Request.Method)
		return
	}

	reg.Recurso = singular(seg[0])

	// id do recurso: parâmetro de rota :id, ou valor injetado pelo handler.
	if idStr := c.Param("id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			u := uint(id)
			reg.RecursoID = &u
		}
	}
	if v, ok := c.Get(CtxAuditRecursoID); ok {
		if id, ok := v.(uint); ok && id != 0 {
			reg.RecursoID = &id
		}
	}

	// Casos especiais por sub-rota.
	sub := ""
	if len(seg) >= 3 {
		sub = seg[2]
	}
	switch {
	case seg[0] == "usuarios" && sub == "senha":
		reg.Acao = "redefiniu senha"
	case seg[0] == "usuarios" && sub == "ativo":
		reg.Acao = "alterou acesso"
	case seg[0] == "itens" && len(seg) >= 2 && seg[1] == "importar":
		reg.Acao = "importou"
	default:
		reg.Acao = acaoPorMetodo(c.Request.Method)
	}

	if d := c.GetString(CtxAuditDetalhe); d != "" {
		reg.Detalhe = d
	}
}

func acaoPorMetodo(m string) string {
	switch m {
	case http.MethodPost:
		return "criou"
	case http.MethodPut:
		return "atualizou"
	case http.MethodPatch:
		return "alterou"
	case http.MethodDelete:
		return "excluiu"
	}
	return "operou"
}

// segmentosAPI devolve os segmentos de caminho após /api/v1.
func segmentosAPI(path string) []string {
	partes := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range partes {
		if p == "v1" {
			return partes[i+1:]
		}
	}
	return nil
}

func singular(seg string) string {
	if s, ok := singularRecurso[seg]; ok {
		return s
	}
	return seg
}
