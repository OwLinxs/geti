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
	// Auto-cadastro público de solicitante (usuário final). Sempre cria com
	// perfil "solicitante"; rate-limited por IP.
	api.POST("/auth/registrar",
		middlewares.RateLimitLogin(cfg.LoginRateLimite, cfg.LoginRateJanela),
		ct.AuthHandler.Registrar,
	)

	// API de integração externa (chamados via WhatsApp/plataforma), autenticada
	// por chave de API — fora do fluxo de login por JWT.
	registrarIntegracao(api, cfg, ct)

	// Rotas autenticadas.
	auth := api.Group("")
	auth.Use(middlewares.Autenticacao(ct.AuthService))
	{
		// Disponível para qualquer perfil autenticado (inclusive solicitante).
		auth.GET("/auth/eu", ct.AuthHandler.EuMesmo)

		// Portal do solicitante (usuário final): abre e acompanha os PRÓPRIOS
		// chamados + conversa. Acesso escopado ao próprio usuário nos handlers.
		registrarMeusChamados(auth, ct)

		// Categorias de chamado: leitura para qualquer perfil (o solicitante
		// precisa escolher ao abrir); escrita só administrador (configuração).
		registrarCategoriasChamado(auth, ct)

		// Área da equipe de T.I. (administrador/operador). Solicitantes são
		// barrados aqui.
		equipe := auth.Group("", middlewares.SomenteEquipe())
		registrarUsuarios(equipe, ct)
		registrarCategorias(equipe, ct)
		registrarSetores(equipe, ct)
		registrarServidores(equipe, ct)
		registrarItens(equipe, ct)
		registrarMovimentacoes(equipe, ct)
		registrarTermos(equipe, ct)
		registrarRelatorios(equipe, ct)
		registrarAuditoria(equipe, ct)
		registrarOrdensServico(equipe, ct)
		registrarConhecimento(equipe, ct)
		registrarFornecedores(equipe, ct)
		registrarContratos(equipe, ct)
		registrarReservas(equipe, ct)
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

func registrarOrdensServico(g *gin.RouterGroup, ct *container.Container) {
	// Módulo de manutenção: operador e administrador podem operar; exclusão
	// (correção de OS aberta por engano) é exclusiva de administradores.
	o := g.Group("/ordens-servico")
	o.GET("", ct.OrdemServicoHandler.Listar)
	o.POST("", ct.OrdemServicoHandler.Criar)
	o.GET("/:id", ct.OrdemServicoHandler.BuscarPorID)
	o.PUT("/:id", ct.OrdemServicoHandler.Atualizar)
	o.PATCH("/:id/status", ct.OrdemServicoHandler.DefinirStatus)
	o.PUT("/:id/passos", ct.OrdemServicoHandler.SalvarPassos)
	o.POST("/:id/documento", ct.OrdemServicoHandler.Documento)
	o.DELETE("/:id", adminOnly(), ct.OrdemServicoHandler.Excluir)
	// Conversa do chamado (mensagens + anexos).
	o.GET("/:id/mensagens", ct.MensagemHandler.Listar)
	o.POST("/:id/mensagens", ct.MensagemHandler.Enviar)
	o.GET("/:id/mensagens/:msgId/anexo", ct.MensagemHandler.BaixarAnexo)
}

func registrarConhecimento(g *gin.RouterGroup, ct *container.Container) {
	// Base de conhecimento: leitura para qualquer perfil; escrita para
	// operador/administrador; exclusão exclusiva de administradores.
	k := g.Group("/conhecimento")
	k.GET("", ct.ConhecimentoHandler.Listar)
	k.GET("/:id", ct.ConhecimentoHandler.BuscarPorID)
	k.POST("", ct.ConhecimentoHandler.Criar)
	k.PUT("/:id", ct.ConhecimentoHandler.Atualizar)
	k.DELETE("/:id", adminOnly(), ct.ConhecimentoHandler.Excluir)
}

func registrarFornecedores(g *gin.RouterGroup, ct *container.Container) {
	// Leitura para qualquer perfil; escrita/exclusão para administradores.
	f := g.Group("/fornecedores")
	f.GET("", ct.FornecedorHandler.Listar)
	f.GET("/:id", ct.FornecedorHandler.BuscarPorID)
	f.POST("", adminOnly(), ct.FornecedorHandler.Criar)
	f.PUT("/:id", adminOnly(), ct.FornecedorHandler.Atualizar)
	f.DELETE("/:id", adminOnly(), ct.FornecedorHandler.Excluir)
}

func registrarContratos(g *gin.RouterGroup, ct *container.Container) {
	// Leitura para qualquer perfil; escrita/exclusão para administradores.
	ct2 := g.Group("/contratos")
	ct2.GET("", ct.ContratoHandler.Listar)
	ct2.GET("/:id", ct.ContratoHandler.BuscarPorID)
	ct2.POST("", adminOnly(), ct.ContratoHandler.Criar)
	ct2.PUT("/:id", adminOnly(), ct.ContratoHandler.Atualizar)
	ct2.DELETE("/:id", adminOnly(), ct.ContratoHandler.Excluir)
}

func registrarIntegracao(api *gin.RouterGroup, cfg *config.Config, ct *container.Container) {
	g := api.Group("/integracao", middlewares.ChaveAPIIntegracao(cfg.IntegracaoAPIKey))
	// Chamados (mapeiam para ordens de serviço). Sync completo: criar/atualizar
	// (idempotente por referencia_externa), ler o board e mover de coluna.
	g.POST("/ordens-servico", ct.IntegracaoHandler.Criar)
	g.GET("/ordens-servico", ct.IntegracaoHandler.Listar)
	g.GET("/ordens-servico/:id", ct.IntegracaoHandler.BuscarPorID)
	g.PATCH("/ordens-servico/:id/status", ct.IntegracaoHandler.DefinirStatus)
}

func registrarCategoriasChamado(g *gin.RouterGroup, ct *container.Container) {
	c := g.Group("/categorias-chamado")
	c.GET("", ct.CategoriaChamadoHandler.Listar)
	c.POST("", adminOnly(), ct.CategoriaChamadoHandler.Criar)
	c.PUT("/:id", adminOnly(), ct.CategoriaChamadoHandler.Atualizar)
	c.DELETE("/:id", adminOnly(), ct.CategoriaChamadoHandler.Excluir)
}

func registrarMeusChamados(g *gin.RouterGroup, ct *container.Container) {
	// Portal do solicitante. Ownership é garantido nos handlers (escopo ao
	// próprio usuário). Qualquer perfil autenticado pode usar.
	m := g.Group("/meus-chamados")
	m.GET("", ct.MeusChamadosHandler.Listar)
	m.POST("", ct.MeusChamadosHandler.Abrir)
	m.GET("/:id", ct.MeusChamadosHandler.BuscarPorID)
	m.GET("/:id/mensagens", ct.MeusChamadosHandler.ListarMensagens)
	m.POST("/:id/mensagens", ct.MeusChamadosHandler.EnviarMensagem)
	m.GET("/:id/mensagens/:msgId/anexo", ct.MeusChamadosHandler.BaixarAnexo)
}

func registrarReservas(g *gin.RouterGroup, ct *container.Container) {
	// Operador e administrador podem gerir reservas; exclusão só administrador.
	res := g.Group("/reservas")
	res.GET("", ct.ReservaHandler.Listar)
	res.GET("/:id", ct.ReservaHandler.BuscarPorID)
	res.POST("", ct.ReservaHandler.Criar)
	res.PUT("/:id", ct.ReservaHandler.Atualizar)
	res.PATCH("/:id/status", ct.ReservaHandler.DefinirStatus)
	res.DELETE("/:id", adminOnly(), ct.ReservaHandler.Excluir)
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
