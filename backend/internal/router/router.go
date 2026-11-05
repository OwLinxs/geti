package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/container"
	"github.com/pmfb/sige-ti/internal/middlewares"
	"github.com/pmfb/sige-ti/internal/models"
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

	// ZapGov (gateway de WhatsApp): webhook de eventos + listagem por telefone.
	// Ficam fora de /api/v1 para casar exatamente com o contrato do gateway.
	r.GET("/v1/chamados",
		middlewares.ChaveAPIIntegracao(cfg.IntegracaoAPIKey),
		ct.ZapGovHandler.ListarChamados,
	)
	r.POST("/integracao/zapgov/eventos",
		middlewares.WebhookZapGov(cfg.IntegracaoAPIKey, cfg.IntegracaoHMACSecret),
		ct.ZapGovHandler.Eventos,
	)
	// Mídia pública assinada dos anexos (o ZapGov busca o arquivo que o técnico
	// enviou). Sem login; protegida por assinatura HMAC no parâmetro ?sig=.
	r.GET("/midia/:osId/:msgId", ct.MidiaHandler.Servir)

	// Rotas autenticadas.
	auth := api.Group("")
	auth.Use(middlewares.Autenticacao(ct.AuthService))
	auth.Use(middlewares.RateLimitAutenticado(cfg.APIRateLimite, cfg.APIRateJanela))
	{
		// Disponível para qualquer perfil autenticado (inclusive solicitante).
		auth.GET("/auth/eu", ct.AuthHandler.EuMesmo)

		// Portal do solicitante (usuário final): abre e acompanha os PRÓPRIOS
		// chamados + conversa. Acesso escopado ao próprio usuário nos handlers.
		registrarMeusChamados(auth, ct)

		// Base de conhecimento no portal (só artigos publicados).
		pc := auth.Group("/portal/conhecimento")
		pc.GET("", ct.ConhecimentoHandler.ListarPortal)
		pc.GET("/:id", ct.ConhecimentoHandler.VisualizarPortal)

		// Categorias de chamado: leitura para qualquer perfil (o solicitante
		// precisa escolher ao abrir); escrita só administrador (configuração).
		registrarCategoriasChamado(auth, ct)

		// Notificações in-app: cada usuário vê as suas (qualquer perfil).
		registrarNotificacoes(auth, ct)

		// Configurações do módulo de chamados (somente administrador).
		registrarConfiguracoes(auth, ct)

		// Área da equipe de T.I. (administrador/operador). Solicitantes são
		// barrados aqui.
		equipe := auth.Group("", middlewares.SomenteEquipe())
		registrarPermissoes(equipe, ct)
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
		registrarRespostasRapidas(equipe, ct)
	}

	return r
}

// adminOnly é atalho para o middleware de perfil administrador.
func adminOnly() gin.HandlerFunc { return middlewares.SomenteAdministrador() }

// exigir é atalho para o middleware de permissão por usuário. O administrador
// sempre passa; os demais perfis dependem de concessão explícita.
func exigir(ct *container.Container, chave string) gin.HandlerFunc {
	return middlewares.ExigirPermissao(ct.PermissaoService, chave)
}

func registrarPermissoes(g *gin.RouterGroup, ct *container.Container) {
	// Catálogo de permissões delegáveis (consultado pela tela de gestão). Só
	// administradores enxergam/gerenciam permissões.
	g.GET("/permissoes/catalogo", adminOnly(), ct.PermissaoHandler.Catalogo)
}

func registrarUsuarios(g *gin.RouterGroup, ct *container.Container) {
	// Gestão de usuários: delegável via permissão "usuarios.gerenciar".
	u := g.Group("/usuarios", exigir(ct, models.PermUsuariosGerenciar))
	u.GET("", ct.UsuarioHandler.Listar)
	u.POST("", ct.UsuarioHandler.Criar)
	u.GET("/:id", ct.UsuarioHandler.BuscarPorID)
	// Gestão de contas: reset de senha e ativação/desativação (sem exclusão
	// física — usuário inativo não consegue logar).
	u.PATCH("/:id/senha", ct.UsuarioHandler.RedefinirSenha)
	u.PATCH("/:id/ativo", ct.UsuarioHandler.DefinirAtivo)
	u.PATCH("/:id/perfil", ct.UsuarioHandler.DefinirPerfil)

	// Permissões por usuário: atribuir/consultar é SEMPRE exclusivo de
	// administradores (adminOnly reforça, mesmo que a gestão de usuários tenha
	// sido delegada a um operador) — evita escalonamento de privilégio.
	u.GET("/:id/permissoes", adminOnly(), ct.PermissaoHandler.DoUsuario)
	u.PUT("/:id/permissoes", adminOnly(), ct.PermissaoHandler.Salvar)
}

func registrarCategorias(g *gin.RouterGroup, ct *container.Container) {
	c := g.Group("/categorias")
	// Leitura: ambos os perfis.
	c.GET("", ct.CategoriaHandler.Listar)
	c.GET("/:id", ct.CategoriaHandler.BuscarPorID)
	// Escrita: administrador.
	c.POST("", exigir(ct, models.PermCategoriasGerenciar), ct.CategoriaHandler.Criar)
	c.PUT("/:id", exigir(ct, models.PermCategoriasGerenciar), ct.CategoriaHandler.Atualizar)
	c.DELETE("/:id", exigir(ct, models.PermCategoriasGerenciar), ct.CategoriaHandler.Remover)

	// Campos customizados (atributos) da categoria.
	c.GET("/:id/campos", ct.CampoCategoriaHandler.Listar)
	c.POST("/:id/campos", exigir(ct, models.PermCategoriasGerenciar), ct.CampoCategoriaHandler.Criar)
	c.PUT("/:id/campos/:campoId", exigir(ct, models.PermCategoriasGerenciar), ct.CampoCategoriaHandler.Atualizar)
	c.DELETE("/:id/campos/:campoId", exigir(ct, models.PermCategoriasGerenciar), ct.CampoCategoriaHandler.Remover)
}

func registrarSetores(g *gin.RouterGroup, ct *container.Container) {
	s := g.Group("/setores")
	s.GET("", ct.SetorHandler.Listar)
	s.GET("/:id", ct.SetorHandler.BuscarPorID)
	s.POST("", exigir(ct, models.PermSetoresGerenciar), ct.SetorHandler.Criar)
	s.PUT("/:id", exigir(ct, models.PermSetoresGerenciar), ct.SetorHandler.Atualizar)
	s.DELETE("/:id", exigir(ct, models.PermSetoresGerenciar), ct.SetorHandler.Remover)
}

func registrarServidores(g *gin.RouterGroup, ct *container.Container) {
	s := g.Group("/servidores")
	s.GET("", ct.ServidorHandler.Listar)
	// Importação em massa via CSV (operador e administrador).
	s.GET("/modelo-csv", ct.ServidorHandler.ModeloCSV)
	s.POST("/importar", ct.ServidorHandler.Importar)
	s.GET("/:id", ct.ServidorHandler.BuscarPorID)
	// Operador pode cadastrar/editar servidores (fluxo operacional).
	s.POST("", ct.ServidorHandler.Criar)
	s.PUT("/:id", ct.ServidorHandler.Atualizar)
	s.DELETE("/:id", exigir(ct, models.PermServidoresExcluir), ct.ServidorHandler.Remover)
}

func registrarItens(g *gin.RouterGroup, ct *container.Container) {
	i := g.Group("/itens")
	// Leitura/consulta/busca/filtros e alertas: ambos os perfis.
	i.GET("", ct.ItemHandler.Listar)
	i.GET("/alertas/estoque-baixo", ct.ItemHandler.AlertasEstoqueBaixo)
	// Importação em massa via CSV (operador e administrador).
	i.GET("/modelo-csv", ct.ItemHandler.ModeloCSV)
	i.POST("/lote", ct.ItemHandler.CriarLote)
	i.POST("/importar", ct.ItemHandler.Importar)
	i.GET("/:id", ct.ItemHandler.BuscarPorID)
	i.GET("/:id/historico", ct.ItemHandler.Historico)
	i.GET("/:id/timeline", ct.ItemHandler.Timeline)
	// Cadastro/edição: operador e administrador.
	i.POST("", ct.ItemHandler.Criar)
	i.PUT("/:id", ct.ItemHandler.Atualizar)
	i.PATCH("/:id/reservavel", ct.ItemHandler.DefinirReservavel)
	// Exclusão (correção de cadastro errado): só administrador, e somente se o
	// item não tiver histórico (caso contrário, usar baixa patrimonial).
	i.DELETE("/:id", exigir(ct, models.PermItensExcluir), ct.ItemHandler.Excluir)
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
	rel.GET("/catalogo", ct.RelatorioHandler.Catalogo)
	rel.GET("", ct.RelatorioHandler.Gerar)
}

func registrarAuditoria(g *gin.RouterGroup, ct *container.Container) {
	// Consulta da trilha de auditoria é exclusiva de administradores.
	a := g.Group("/auditoria", exigir(ct, models.PermAuditoriaVer))
	a.GET("", ct.AuditoriaHandler.Listar)
}

func registrarOrdensServico(g *gin.RouterGroup, ct *container.Container) {
	// Módulo de manutenção: operador e administrador podem operar; exclusão
	// (correção de OS aberta por engano) é exclusiva de administradores.
	o := g.Group("/ordens-servico")
	o.GET("", ct.OrdemServicoHandler.Listar)
	o.GET("/dashboard", ct.OrdemServicoHandler.Dashboard)
	o.GET("/exportar", ct.OrdemServicoHandler.Exportar)
	o.POST("", ct.OrdemServicoHandler.Criar)
	o.GET("/:id", ct.OrdemServicoHandler.BuscarPorID)
	o.GET("/:id/eventos", ct.OrdemServicoHandler.ListarEventos)
	o.GET("/:id/relacionadas", ct.OrdemServicoHandler.Relacionadas)
	o.POST("/:id/desdobrar", ct.OrdemServicoHandler.Desdobrar)
	o.PUT("/:id", ct.OrdemServicoHandler.Atualizar)
	o.PATCH("/:id/status", ct.OrdemServicoHandler.DefinirStatus)
	o.PATCH("/:id/atribuir", ct.OrdemServicoHandler.AtribuirTecnico)
	o.PATCH("/:id/classificacao", ct.OrdemServicoHandler.Classificar)
	o.PUT("/:id/passos", ct.OrdemServicoHandler.SalvarPassos)
	o.POST("/:id/documento", ct.OrdemServicoHandler.Documento)
	o.DELETE("/:id", exigir(ct, models.PermChamadosExcluir), ct.OrdemServicoHandler.Excluir)
	// Conversa do chamado (mensagens + anexos).
	o.GET("/:id/mensagens", ct.MensagemHandler.Listar)
	o.POST("/:id/mensagens", ct.MensagemHandler.Enviar)
	o.POST("/:id/template", ct.MensagemHandler.EnviarTemplate)
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
	k.DELETE("/:id", exigir(ct, models.PermConhecimentoExcluir), ct.ConhecimentoHandler.Excluir)
}

func registrarFornecedores(g *gin.RouterGroup, ct *container.Container) {
	// Leitura para qualquer perfil; escrita/exclusão para administradores.
	f := g.Group("/fornecedores")
	f.GET("", ct.FornecedorHandler.Listar)
	f.GET("/:id", ct.FornecedorHandler.BuscarPorID)
	f.POST("", exigir(ct, models.PermFornecedoresGerenciar), ct.FornecedorHandler.Criar)
	f.PUT("/:id", exigir(ct, models.PermFornecedoresGerenciar), ct.FornecedorHandler.Atualizar)
	f.DELETE("/:id", exigir(ct, models.PermFornecedoresGerenciar), ct.FornecedorHandler.Excluir)
}

func registrarContratos(g *gin.RouterGroup, ct *container.Container) {
	// Leitura para qualquer perfil; escrita/exclusão para administradores.
	ct2 := g.Group("/contratos")
	ct2.GET("", ct.ContratoHandler.Listar)
	ct2.GET("/:id", ct.ContratoHandler.BuscarPorID)
	ct2.POST("", exigir(ct, models.PermContratosGerenciar), ct.ContratoHandler.Criar)
	ct2.PUT("/:id", exigir(ct, models.PermContratosGerenciar), ct.ContratoHandler.Atualizar)
	ct2.DELETE("/:id", exigir(ct, models.PermContratosGerenciar), ct.ContratoHandler.Excluir)
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
	c.POST("", exigir(ct, models.PermChamadosCategorias), ct.CategoriaChamadoHandler.Criar)
	c.PUT("/:id", exigir(ct, models.PermChamadosCategorias), ct.CategoriaChamadoHandler.Atualizar)
	c.DELETE("/:id", exigir(ct, models.PermChamadosCategorias), ct.CategoriaChamadoHandler.Excluir)
}

func registrarNotificacoes(g *gin.RouterGroup, ct *container.Container) {
	n := g.Group("/notificacoes")
	n.GET("", ct.NotificacaoHandler.Listar)
	n.GET("/nao-lidas", ct.NotificacaoHandler.ContarNaoLidas)
	n.PATCH("/lidas", ct.NotificacaoHandler.MarcarTodasLidas)
	n.PATCH("/por-recurso", ct.NotificacaoHandler.MarcarLidasPorRecurso)
	n.PATCH("/:id/lida", ct.NotificacaoHandler.MarcarLida)
}

func registrarRespostasRapidas(g *gin.RouterGroup, ct *container.Container) {
	// Leitura para a equipe (usada no chat); escrita só administrador.
	r := g.Group("/respostas-rapidas")
	r.GET("", ct.RespostaRapidaHandler.Listar)
	r.POST("", exigir(ct, models.PermChamadosRespostas), ct.RespostaRapidaHandler.Criar)
	r.PUT("/:id", exigir(ct, models.PermChamadosRespostas), ct.RespostaRapidaHandler.Atualizar)
	r.DELETE("/:id", exigir(ct, models.PermChamadosRespostas), ct.RespostaRapidaHandler.Excluir)
}

func registrarConfiguracoes(g *gin.RouterGroup, ct *container.Container) {
	c := g.Group("/config", exigir(ct, models.PermConfigGerenciar))
	c.GET("/chamados", ct.ConfiguracaoHandler.ObterChamados)
	c.PUT("/chamados", ct.ConfiguracaoHandler.SalvarChamados)
}

func registrarMeusChamados(g *gin.RouterGroup, ct *container.Container) {
	// Portal do solicitante. Ownership é garantido nos handlers (escopo ao
	// próprio usuário). Qualquer perfil autenticado pode usar.
	m := g.Group("/meus-chamados")
	m.GET("", ct.MeusChamadosHandler.Listar)
	m.POST("", ct.MeusChamadosHandler.Abrir)
	m.GET("/:id", ct.MeusChamadosHandler.BuscarPorID)
	m.POST("/:id/avaliacao", ct.MeusChamadosHandler.Avaliar)
	m.POST("/:id/reabrir", ct.MeusChamadosHandler.Reabrir)
	m.GET("/:id/mensagens", ct.MeusChamadosHandler.ListarMensagens)
	m.POST("/:id/mensagens", ct.MeusChamadosHandler.EnviarMensagem)
	m.GET("/:id/mensagens/:msgId/anexo", ct.MeusChamadosHandler.BaixarAnexo)
}

func registrarReservas(g *gin.RouterGroup, ct *container.Container) {
	// Operador e administrador podem gerir reservas; exclusão só administrador.
	res := g.Group("/reservas")
	res.GET("", ct.ReservaHandler.Listar)
	res.GET("/equipamentos", ct.ReservaHandler.ListarEquipamentos)
	res.GET("/:id", ct.ReservaHandler.BuscarPorID)
	res.POST("", ct.ReservaHandler.Criar)
	res.PUT("/:id", ct.ReservaHandler.Atualizar)
	res.PATCH("/:id/status", ct.ReservaHandler.DefinirStatus)
	res.DELETE("/:id", exigir(ct, models.PermReservasExcluir), ct.ReservaHandler.Excluir)
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
