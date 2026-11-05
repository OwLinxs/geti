// Package container faz a injeção de dependências simples do SIGE-TI,
// montando a cadeia repositories -> services -> handlers a partir de uma
// única conexão de banco e da configuração.
package container

import (
	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/handlers"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
	"gorm.io/gorm"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB

	// Services (expostos para uso em seed/testes).
	UsuarioService          *services.UsuarioService
	PermissaoService        *services.PermissaoService
	AuthService             *services.AuthService
	CategoriaService        *services.CategoriaService
	CampoCategoriaService   *services.CampoCategoriaService
	InventarioService       *services.InventarioService
	SetorService            *services.SetorService
	ServidorService         *services.ServidorService
	ItemService             *services.ItemService
	MovimentacaoService     *services.MovimentacaoService
	TermoService            *services.TermoService
	RelatorioService        *services.RelatorioService
	AuditoriaService        *services.AuditoriaService
	OrdemServicoService     *services.OrdemServicoService
	MensagemService         *services.MensagemService
	CategoriaChamadoService *services.CategoriaChamadoService
	ConfiguracaoService     *services.ConfiguracaoService
	NotificacaoService      *services.NotificacaoService
	AlertaVencimentoService *services.AlertaVencimentoService
	RespostaRapidaService   *services.RespostaRapidaService
	ConhecimentoService     *services.ConhecimentoService
	FornecedorService       *services.FornecedorService
	ContratoService         *services.ContratoService
	ReservaService          *services.ReservaService

	// Handlers.
	HealthHandler           *handlers.HealthHandler
	AuthHandler             *handlers.AuthHandler
	UsuarioHandler          *handlers.UsuarioHandler
	PermissaoHandler        *handlers.PermissaoHandler
	CategoriaHandler        *handlers.CategoriaHandler
	CampoCategoriaHandler   *handlers.CampoCategoriaHandler
	SetorHandler            *handlers.SetorHandler
	ServidorHandler         *handlers.ServidorHandler
	ItemHandler             *handlers.ItemHandler
	MovimentacaoHandler     *handlers.MovimentacaoHandler
	TermoHandler            *handlers.TermoHandler
	RelatorioHandler        *handlers.RelatorioHandler
	AuditoriaHandler        *handlers.AuditoriaHandler
	OrdemServicoHandler     *handlers.OrdemServicoHandler
	MensagemHandler         *handlers.MensagemHandler
	MeusChamadosHandler     *handlers.MeusChamadosHandler
	CategoriaChamadoHandler *handlers.CategoriaChamadoHandler
	ConfiguracaoHandler     *handlers.ConfiguracaoHandler
	NotificacaoHandler      *handlers.NotificacaoHandler
	RespostaRapidaHandler   *handlers.RespostaRapidaHandler
	ConhecimentoHandler     *handlers.ConhecimentoHandler
	FornecedorHandler       *handlers.FornecedorHandler
	ContratoHandler         *handlers.ContratoHandler
	ReservaHandler          *handlers.ReservaHandler
	IntegracaoHandler       *handlers.IntegracaoHandler
	InventarioHandler       *handlers.InventarioHandler
	ZapGovHandler           *handlers.ZapGovHandler
	MidiaHandler            *handlers.MidiaHandler
}

func New(cfg *config.Config, db *gorm.DB) *Container {
	// Repositories
	usuarioRepo := repositories.NewUsuarioRepository(db)
	permissaoRepo := repositories.NewPermissaoUsuarioRepository(db)
	categoriaRepo := repositories.NewCategoriaRepository(db)
	setorRepo := repositories.NewSetorRepository(db)
	servidorRepo := repositories.NewServidorRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	movRepo := repositories.NewMovimentacaoRepository(db)
	termoRepo := repositories.NewTermoRepository(db)
	auditoriaRepo := repositories.NewAuditoriaRepository(db)
	ordemServicoRepo := repositories.NewOrdemServicoRepository(db)
	mensagemRepo := repositories.NewMensagemRepository(db)
	categoriaChamadoRepo := repositories.NewCategoriaChamadoRepository(db)
	configuracaoRepo := repositories.NewConfiguracaoRepository(db)
	notificacaoRepo := repositories.NewNotificacaoRepository(db)
	respostaRapidaRepo := repositories.NewRespostaRapidaRepository(db)
	eventoChamadoRepo := repositories.NewEventoChamadoRepository(db)
	conhecimentoRepo := repositories.NewConhecimentoRepository(db)
	fornecedorRepo := repositories.NewFornecedorRepository(db)
	contratoRepo := repositories.NewContratoRepository(db)
	reservaRepo := repositories.NewReservaRepository(db)

	// Services
	usuarioSvc := services.NewUsuarioService(usuarioRepo)
	permissaoSvc := services.NewPermissaoService(permissaoRepo, usuarioRepo)
	authSvc := services.NewAuthService(usuarioSvc, cfg.JWTSecret, cfg.JWTExpiresIn)
	categoriaSvc := services.NewCategoriaService(categoriaRepo)
	campoCategoriaRepo := repositories.NewCampoCategoriaRepository(db)
	campoCategoriaSvc := services.NewCampoCategoriaService(campoCategoriaRepo, categoriaRepo)
	inventarioSvc := services.NewInventarioService(itemRepo, categoriaRepo, campoCategoriaRepo)
	setorSvc := services.NewSetorService(setorRepo)
	servidorSvc := services.NewServidorService(servidorRepo, setorRepo)
	itemSvc := services.NewItemService(itemRepo, categoriaRepo, setorRepo, servidorRepo, movRepo, termoRepo)
	// Linha do tempo do item depende das OS (injeção opcional via setter).
	itemSvc.DefinirOrdemServicoRepo(ordemServicoRepo)
	// Campos customizados (atributos) da categoria (injeção opcional via setter).
	itemSvc.DefinirCampoRepo(campoCategoriaRepo)
	movSvc := services.NewMovimentacaoService(movRepo, itemRepo, setorRepo, servidorRepo)
	termoSvc := services.NewTermoService(termoRepo, itemRepo, servidorRepo, cfg)
	relatorioSvc := services.NewRelatorioService(itemRepo, movRepo, setorRepo, cfg)
	auditoriaSvc := services.NewAuditoriaService(auditoriaRepo)
	categoriaChamadoSvc := services.NewCategoriaChamadoService(categoriaChamadoRepo)
	configuracaoSvc := services.NewConfiguracaoService(configuracaoRepo)
	notificacaoSvc := services.NewNotificacaoService(notificacaoRepo, usuarioRepo, configuracaoSvc)
	respostaRapidaSvc := services.NewRespostaRapidaService(respostaRapidaRepo)
	ordemServicoSvc := services.NewOrdemServicoService(ordemServicoRepo, itemRepo, setorRepo, servidorRepo, usuarioRepo, categoriaChamadoRepo, cfg)
	mensagemSvc := services.NewMensagemService(mensagemRepo, ordemServicoRepo, cfg)
	conhecimentoSvc := services.NewConhecimentoService(conhecimentoRepo)

	// Cliente do ZapGov (gateway de WhatsApp). Inerte sem credenciais.
	zapGovCli := services.NewZapGovClient(cfg)

	// Liga o notificador aos serviços que emitem eventos de chamado.
	ordemServicoSvc.SetNotificador(notificacaoSvc)
	mensagemSvc.SetNotificador(notificacaoSvc)
	// Liga o envio ao WhatsApp (saída da equipe → contato).
	mensagemSvc.SetZapGov(zapGovCli)
	// Liga o mensageiro ao serviço de OS (aviso de encerramento no WhatsApp).
	ordemServicoSvc.SetMensageiro(mensagemSvc)
	// Liga a configuração ao serviço de OS (cálculo de prazos de SLA).
	ordemServicoSvc.SetConfig(configuracaoSvc)
	// Liga o log de eventos (linha do tempo do chamado).
	ordemServicoSvc.SetEventos(eventoChamadoRepo)
	// Liga a configuração ao auth (regras de auto-cadastro).
	authSvc.SetConfig(configuracaoSvc)
	fornecedorSvc := services.NewFornecedorService(fornecedorRepo)
	contratoSvc := services.NewContratoService(contratoRepo, fornecedorRepo)
	alertaVencimentoSvc := services.NewAlertaVencimentoService(itemRepo, contratoRepo, notificacaoSvc, cfg.AlertaVencimentoDias)
	reservaSvc := services.NewReservaService(reservaRepo, itemRepo, servidorRepo)

	// Handlers
	return &Container{
		Config:                  cfg,
		DB:                      db,
		UsuarioService:          usuarioSvc,
		PermissaoService:        permissaoSvc,
		AuthService:             authSvc,
		CategoriaService:        categoriaSvc,
		CampoCategoriaService:   campoCategoriaSvc,
		InventarioService:       inventarioSvc,
		SetorService:            setorSvc,
		ServidorService:         servidorSvc,
		ItemService:             itemSvc,
		MovimentacaoService:     movSvc,
		TermoService:            termoSvc,
		RelatorioService:        relatorioSvc,
		AuditoriaService:        auditoriaSvc,
		OrdemServicoService:     ordemServicoSvc,
		MensagemService:         mensagemSvc,
		CategoriaChamadoService: categoriaChamadoSvc,
		ConfiguracaoService:     configuracaoSvc,
		NotificacaoService:      notificacaoSvc,
		AlertaVencimentoService: alertaVencimentoSvc,
		RespostaRapidaService:   respostaRapidaSvc,
		ConhecimentoService:     conhecimentoSvc,
		FornecedorService:       fornecedorSvc,
		ContratoService:         contratoSvc,
		ReservaService:          reservaSvc,

		HealthHandler:           handlers.NewHealthHandler(db),
		AuthHandler:             handlers.NewAuthHandler(authSvc, usuarioSvc, permissaoSvc),
		UsuarioHandler:          handlers.NewUsuarioHandler(usuarioSvc),
		PermissaoHandler:        handlers.NewPermissaoHandler(permissaoSvc),
		CategoriaHandler:        handlers.NewCategoriaHandler(categoriaSvc),
		CampoCategoriaHandler:   handlers.NewCampoCategoriaHandler(campoCategoriaSvc),
		SetorHandler:            handlers.NewSetorHandler(setorSvc),
		ServidorHandler:         handlers.NewServidorHandler(servidorSvc),
		ItemHandler:             handlers.NewItemHandler(itemSvc, movSvc),
		MovimentacaoHandler:     handlers.NewMovimentacaoHandler(movSvc),
		TermoHandler:            handlers.NewTermoHandler(termoSvc),
		RelatorioHandler:        handlers.NewRelatorioHandler(relatorioSvc),
		AuditoriaHandler:        handlers.NewAuditoriaHandler(auditoriaSvc),
		OrdemServicoHandler:     handlers.NewOrdemServicoHandler(ordemServicoSvc),
		MensagemHandler:         handlers.NewMensagemHandler(mensagemSvc),
		MeusChamadosHandler:     handlers.NewMeusChamadosHandler(ordemServicoSvc, mensagemSvc),
		CategoriaChamadoHandler: handlers.NewCategoriaChamadoHandler(categoriaChamadoSvc),
		ConfiguracaoHandler:     handlers.NewConfiguracaoHandler(configuracaoSvc),
		NotificacaoHandler:      handlers.NewNotificacaoHandler(notificacaoSvc),
		RespostaRapidaHandler:   handlers.NewRespostaRapidaHandler(respostaRapidaSvc),
		ConhecimentoHandler:     handlers.NewConhecimentoHandler(conhecimentoSvc),
		FornecedorHandler:       handlers.NewFornecedorHandler(fornecedorSvc),
		ContratoHandler:         handlers.NewContratoHandler(contratoSvc),
		ReservaHandler:          handlers.NewReservaHandler(reservaSvc),
		IntegracaoHandler:       handlers.NewIntegracaoHandler(ordemServicoSvc),
		InventarioHandler:       handlers.NewInventarioHandler(inventarioSvc),
		ZapGovHandler:           handlers.NewZapGovHandler(ordemServicoSvc, mensagemSvc),
		MidiaHandler:            handlers.NewMidiaHandler(mensagemSvc, cfg.JWTSecret),
	}
}
