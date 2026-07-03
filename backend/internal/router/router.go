package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/container"
	"github.com/pmfb/sige-ti/internal/middlewares"
)

// Setup constrói o engine Gin, aplica middlewares globais e registra todas as
// rotas a partir do container de dependências.
func Setup(cfg *config.Config, ct *container.Container) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg))
	// Trilha de auditoria: registra login e operações de escrita (após o
	// handler executar, para capturar o status e o usuário autenticado).
	r.Use(middlewares.Auditoria(ct.AuditoriaService))

	// Health (público, fora do prefixo de API).
	r.GET("/health", ct.HealthHandler.Health)

	api := r.Group("/api/v1")
	api.GET("/health", ct.HealthHandler.Health)

	// Autenticação (pública) com rate limiting por IP contra força bruta.
	api.POST("/auth/login",
		middlewares.RateLimitLogin(cfg.LoginRateLimite, cfg.LoginRateJanela),
		ct.AuthHandler.Login,
	)

	// Rotas autenticadas.
	auth := api.Group("")
	auth.Use(middlewares.Autenticacao(ct.AuthService))
	{
		auth.GET("/auth/eu", ct.AuthHandler.EuMesmo)

		registrarUsuarios(auth, ct)
		registrarCategorias(auth, ct)
		registrarSetores(auth, ct)
		registrarServidores(auth, ct)
		registrarItens(auth, ct)
		registrarMovimentacoes(auth, ct)
		registrarTermos(auth, ct)
		registrarRelatorios(auth, ct)
		registrarAuditoria(auth, ct)
	}

	return r
}

// adminOnly é atalho para o middleware de perfil administrador.
func adminOnly() gin.HandlerFunc { return middlewares.SomenteAdministrador() }

func registrarUsuarios(g *gin.RouterGroup, ct *container.Container) {
	// Gestão de usuários é exclusiva de administradores.
	u := g.Group("/usuarios", adminOnly())
	u.GET("", ct.UsuarioHandler.Listar)
	u.POST("", ct.UsuarioHandler.Criar)
	u.GET("/:id", ct.UsuarioHandler.BuscarPorID)
	// Gestão de contas: reset de senha e ativação/desativação (sem exclusão
	// física — usuário inativo não consegue logar).
	u.PATCH("/:id/senha", ct.UsuarioHandler.RedefinirSenha)
	u.PATCH("/:id/ativo", ct.UsuarioHandler.DefinirAtivo)
	u.PATCH("/:id/perfil", ct.UsuarioHandler.DefinirPerfil)
}

func registrarCategorias(g *gin.RouterGroup, ct *container.Container) {
	c := g.Group("/categorias")
	// Leitura: ambos os perfis.
	c.GET("", ct.CategoriaHandler.Listar)
	c.GET("/:id", ct.CategoriaHandler.BuscarPorID)
	// Escrita: administrador.
	c.POST("", adminOnly(), ct.CategoriaHandler.Criar)
	c.PUT("/:id", adminOnly(), ct.CategoriaHandler.Atualizar)
	c.DELETE("/:id", adminOnly(), ct.CategoriaHandler.Remover)
}

func registrarSetores(g *gin.RouterGroup, ct *container.Container) {
	s := g.Group("/setores")
	s.GET("", ct.SetorHandler.Listar)
	s.GET("/:id", ct.SetorHandler.BuscarPorID)
	s.POST("", adminOnly(), ct.SetorHandler.Criar)
	s.PUT("/:id", adminOnly(), ct.SetorHandler.Atualizar)
	s.DELETE("/:id", adminOnly(), ct.SetorHandler.Remover)
}

func registrarServidores(g *gin.RouterGroup, ct *container.Container) {
	s := g.Group("/servidores")
	s.GET("", ct.ServidorHandler.Listar)
	s.GET("/:id", ct.ServidorHandler.BuscarPorID)
	// Operador pode cadastrar/editar servidores (fluxo operacional).
	s.POST("", ct.ServidorHandler.Criar)
	s.PUT("/:id", ct.ServidorHandler.Atualizar)
	s.DELETE("/:id", adminOnly(), ct.ServidorHandler.Remover)
}

func registrarItens(g *gin.RouterGroup, ct *container.Container) {
	i := g.Group("/itens")
	// Leitura/consulta/busca/filtros e alertas: ambos os perfis.
	i.GET("", ct.ItemHandler.Listar)
	i.GET("/alertas/estoque-baixo", ct.ItemHandler.AlertasEstoqueBaixo)
	// Importação em massa via CSV (operador e administrador).
	i.GET("/modelo-csv", ct.ItemHandler.ModeloCSV)
	i.POST("/importar", ct.ItemHandler.Importar)
	i.GET("/:id", ct.ItemHandler.BuscarPorID)
	i.GET("/:id/historico", ct.ItemHandler.Historico)
	// Cadastro/edição: operador e administrador.
	i.POST("", ct.ItemHandler.Criar)
	i.PUT("/:id", ct.ItemHandler.Atualizar)
	// Exclusão (correção de cadastro errado): só administrador, e somente se o
	// item não tiver histórico (caso contrário, usar baixa patrimonial).
	i.DELETE("/:id", adminOnly(), ct.ItemHandler.Excluir)
}

func registrarMovimentacoes(g *gin.RouterGroup, ct *container.Container) {
	m := g.Group("/movimentacoes")
	// Operador e administrador podem registrar e consultar movimentações.
	m.GET("", ct.MovimentacaoHandler.Listar)
	m.POST("", ct.MovimentacaoHandler.Registrar)
}

func registrarTermos(g *gin.RouterGroup, ct *container.Container) {
	t := g.Group("/termos")
	t.GET("", ct.TermoHandler.Listar)
	t.POST("", ct.TermoHandler.Emitir)
	t.GET("/:id", ct.TermoHandler.BuscarPorID)
	t.GET("/:id/pdf", ct.TermoHandler.PDF)
}

func registrarRelatorios(g *gin.RouterGroup, ct *container.Container) {
	rel := g.Group("/relatorios")
	rel.GET("/itens-por-setor", ct.RelatorioHandler.ItensPorSetor)
	rel.GET("/itens-por-responsavel", ct.RelatorioHandler.ItensPorResponsavel)
	rel.GET("/estoque-baixo", ct.RelatorioHandler.EstoqueBaixo)
	rel.GET("/inventario", ct.RelatorioHandler.Inventario)
	rel.GET("/movimentacoes", ct.RelatorioHandler.Movimentacoes)
}

func registrarAuditoria(g *gin.RouterGroup, ct *container.Container) {
	// Consulta da trilha de auditoria é exclusiva de administradores.
	a := g.Group("/auditoria", adminOnly())
	a.GET("", ct.AuditoriaHandler.Listar)
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
