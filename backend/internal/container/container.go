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
	UsuarioService      *services.UsuarioService
	AuthService         *services.AuthService
	CategoriaService    *services.CategoriaService
	SetorService        *services.SetorService
	ServidorService     *services.ServidorService
	ItemService         *services.ItemService
	MovimentacaoService *services.MovimentacaoService
	TermoService        *services.TermoService
	RelatorioService    *services.RelatorioService
	AuditoriaService    *services.AuditoriaService

	// Handlers.
	HealthHandler       *handlers.HealthHandler
	AuthHandler         *handlers.AuthHandler
	UsuarioHandler      *handlers.UsuarioHandler
	CategoriaHandler    *handlers.CategoriaHandler
	SetorHandler        *handlers.SetorHandler
	ServidorHandler     *handlers.ServidorHandler
	ItemHandler         *handlers.ItemHandler
	MovimentacaoHandler *handlers.MovimentacaoHandler
	TermoHandler        *handlers.TermoHandler
	RelatorioHandler    *handlers.RelatorioHandler
	AuditoriaHandler    *handlers.AuditoriaHandler
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

	// Handlers
	return &Container{
		Config:              cfg,
		DB:                  db,
		UsuarioService:      usuarioSvc,
		AuthService:         authSvc,
		CategoriaService:    categoriaSvc,
		SetorService:        setorSvc,
		ServidorService:     servidorSvc,
		ItemService:         itemSvc,
		MovimentacaoService: movSvc,
		TermoService:        termoSvc,
		RelatorioService:    relatorioSvc,
		AuditoriaService:    auditoriaSvc,

		HealthHandler:       handlers.NewHealthHandler(db),
		AuthHandler:         handlers.NewAuthHandler(authSvc, usuarioSvc),
		UsuarioHandler:      handlers.NewUsuarioHandler(usuarioSvc),
		CategoriaHandler:    handlers.NewCategoriaHandler(categoriaSvc),
		SetorHandler:        handlers.NewSetorHandler(setorSvc),
		ServidorHandler:     handlers.NewServidorHandler(servidorSvc),
		ItemHandler:         handlers.NewItemHandler(itemSvc, movSvc),
		MovimentacaoHandler: handlers.NewMovimentacaoHandler(movSvc),
		TermoHandler:        handlers.NewTermoHandler(termoSvc),
		RelatorioHandler:    handlers.NewRelatorioHandler(relatorioSvc),
		AuditoriaHandler:    handlers.NewAuditoriaHandler(auditoriaSvc),
	}
}
