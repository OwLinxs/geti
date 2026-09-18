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
	AuthService             *services.AuthService
	CategoriaService        *services.CategoriaService
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
	ConhecimentoService     *services.ConhecimentoService
	FornecedorService       *services.FornecedorService
	ContratoService         *services.ContratoService
	ReservaService          *services.ReservaService

	// Handlers.
	HealthHandler           *handlers.HealthHandler
	AuthHandler             *handlers.AuthHandler
	UsuarioHandler          *handlers.UsuarioHandler
	CategoriaHandler        *handlers.CategoriaHandler
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
	ConhecimentoHandler     *handlers.ConhecimentoHandler
	FornecedorHandler       *handlers.FornecedorHandler
	ContratoHandler         *handlers.ContratoHandler
	ReservaHandler          *handlers.ReservaHandler
	IntegracaoHandler       *handlers.IntegracaoHandler
}

func New(cfg *config.Config, db *gorm.DB) *Container {
	// Repositories
	usuarioRepo := repositories.NewUsuarioRepository(db)
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
	conhecimentoRepo := repositories.NewConhecimentoRepository(db)
	fornecedorRepo := repositories.NewFornecedorRepository(db)
	contratoRepo := repositories.NewContratoRepository(db)
	reservaRepo := repositories.NewReservaRepository(db)

	// Services
	usuarioSvc := services.NewUsuarioService(usuarioRepo)
	authSvc := services.NewAuthService(usuarioSvc, cfg.JWTSecret, cfg.JWTExpiresIn)
	categoriaSvc := services.NewCategoriaService(categoriaRepo)
	setorSvc := services.NewSetorService(setorRepo)
	servidorSvc := services.NewServidorService(servidorRepo, setorRepo)
	itemSvc := services.NewItemService(itemRepo, categoriaRepo, setorRepo, servidorRepo, movRepo, termoRepo)
	movSvc := services.NewMovimentacaoService(movRepo, itemRepo, setorRepo, servidorRepo)
	termoSvc := services.NewTermoService(termoRepo, itemRepo, servidorRepo, cfg)
	relatorioSvc := services.NewRelatorioService(itemRepo, movRepo, cfg)
	auditoriaSvc := services.NewAuditoriaService(auditoriaRepo)
	categoriaChamadoSvc := services.NewCategoriaChamadoService(categoriaChamadoRepo)
	ordemServicoSvc := services.NewOrdemServicoService(ordemServicoRepo, itemRepo, setorRepo, servidorRepo, usuarioRepo, categoriaChamadoRepo, cfg)
	mensagemSvc := services.NewMensagemService(mensagemRepo, ordemServicoRepo, cfg)
	conhecimentoSvc := services.NewConhecimentoService(conhecimentoRepo)
	fornecedorSvc := services.NewFornecedorService(fornecedorRepo)
	contratoSvc := services.NewContratoService(contratoRepo, fornecedorRepo)
	reservaSvc := services.NewReservaService(reservaRepo, itemRepo, servidorRepo)

	// Handlers
	return &Container{
		Config:                  cfg,
		DB:                      db,
		UsuarioService:          usuarioSvc,
		AuthService:             authSvc,
		CategoriaService:        categoriaSvc,
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
		ConhecimentoService:     conhecimentoSvc,
		FornecedorService:       fornecedorSvc,
		ContratoService:         contratoSvc,
		ReservaService:          reservaSvc,

		HealthHandler:           handlers.NewHealthHandler(db),
		AuthHandler:             handlers.NewAuthHandler(authSvc, usuarioSvc),
		UsuarioHandler:          handlers.NewUsuarioHandler(usuarioSvc),
		CategoriaHandler:        handlers.NewCategoriaHandler(categoriaSvc),
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
		ConhecimentoHandler:     handlers.NewConhecimentoHandler(conhecimentoSvc),
		FornecedorHandler:       handlers.NewFornecedorHandler(fornecedorSvc),
		ContratoHandler:         handlers.NewContratoHandler(contratoSvc),
		ReservaHandler:          handlers.NewReservaHandler(reservaSvc),
		IntegracaoHandler:       handlers.NewIntegracaoHandler(ordemServicoSvc),
	}
}
