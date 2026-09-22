package services_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ambChamado struct {
	db        *gorm.DB
	osSvc     *services.OrdemServicoService
	catSvc    *services.CategoriaChamadoService
	notifRepo repositories.NotificacaoRepository
	osRepo    repositories.OrdemServicoRepository
	adminID   uint
}

func novoAmbChamado(t *testing.T) *ambChamado {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("abrir db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Usuario{}, &models.Categoria{}, &models.Setor{},
		&models.Servidor{}, &models.Item{}, &models.CategoriaChamado{},
		&models.OrdemServico{}, &models.OrdemServicoPasso{},
		&models.MensagemChamado{}, &models.Configuracao{},
		&models.Notificacao{}, &models.EventoChamado{},
	); err != nil {
		t.Fatalf("migrar: %v", err)
	}

	usuarioRepo := repositories.NewUsuarioRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	setorRepo := repositories.NewSetorRepository(db)
	servRepo := repositories.NewServidorRepository(db)
	catChamadoRepo := repositories.NewCategoriaChamadoRepository(db)
	osRepo := repositories.NewOrdemServicoRepository(db)
	configRepo := repositories.NewConfiguracaoRepository(db)
	notifRepo := repositories.NewNotificacaoRepository(db)
	eventoRepo := repositories.NewEventoChamadoRepository(db)

	configSvc := services.NewConfiguracaoService(configRepo)
	notifSvc := services.NewNotificacaoService(notifRepo, usuarioRepo, configSvc)
	catSvc := services.NewCategoriaChamadoService(catChamadoRepo)
	osSvc := services.NewOrdemServicoService(osRepo, itemRepo, setorRepo, servRepo, usuarioRepo, catChamadoRepo, &config.Config{})
	osSvc.SetNotificador(notifSvc)
	osSvc.SetConfig(configSvc)
	osSvc.SetEventos(eventoRepo)

	admin := &models.Usuario{
		Nome: "Admin", Email: "admin@t.local", SenhaHash: "x",
		Perfil: models.PerfilAdministrador, Ativo: true,
	}
	if err := usuarioRepo.Criar(admin); err != nil {
		t.Fatalf("criar admin: %v", err)
	}

	return &ambChamado{db: db, osSvc: osSvc, catSvc: catSvc, notifRepo: notifRepo, osRepo: osRepo, adminID: admin.ID}
}

func (a *ambChamado) abrir(t *testing.T, prioridade string) *models.OrdemServico {
	t.Helper()
	os, err := a.osSvc.Criar(services.EntradaOS{
		Assunto:         "Teste",
		DefeitoRelatado: "não funciona",
		Prioridade:      prioridade,
		AbertoPorID:     a.adminID,
	})
	if err != nil {
		t.Fatalf("abrir chamado: %v", err)
	}
	return os
}

func TestConfigChamadosPadrao(t *testing.T) {
	c := services.ConfigChamadosPadrao()
	if !c.NotificarEquipeNovoChamado || c.EmailAtivo {
		t.Fatalf("defaults inesperados: %+v", c)
	}
	resp, resol := c.HorasSLA("alta")
	if resp != 2 || resol != 8 {
		t.Fatalf("HorasSLA(alta) = %d,%d; quero 2,8", resp, resol)
	}
	if r, _ := c.HorasSLA("desconhecida"); r != c.SLARespostaNormalH {
		t.Fatalf("prioridade desconhecida deve cair no normal")
	}
}

func TestCriarDefinePrazosSLA(t *testing.T) {
	a := novoAmbChamado(t)
	os := a.abrir(t, "alta")
	if os.PrazoRespostaEm == nil || os.PrazoResolucaoEm == nil {
		t.Fatal("prazos de SLA não foram definidos")
	}
	h := os.PrazoRespostaEm.Sub(os.DataAbertura).Hours()
	if h < 1.9 || h > 2.1 {
		t.Fatalf("prazo de resposta = %.1fh; quero ~2h", h)
	}
	if os.Origem != "interno" {
		t.Fatalf("origem = %q; quero interno", os.Origem)
	}
}

func TestAvaliarSomenteConcluido(t *testing.T) {
	a := novoAmbChamado(t)
	os := a.abrir(t, "normal")

	if _, err := a.osSvc.Avaliar(os.ID, 5, "ok"); err == nil {
		t.Fatal("avaliar chamado não concluído deveria falhar")
	}
	if _, err := a.osSvc.DefinirStatus(os.ID, models.OSConcluida); err != nil {
		t.Fatalf("concluir: %v", err)
	}
	if _, err := a.osSvc.Avaliar(os.ID, 9, ""); err == nil {
		t.Fatal("nota 9 deveria falhar")
	}
	av, err := a.osSvc.Avaliar(os.ID, 4, "bom")
	if err != nil {
		t.Fatalf("avaliar: %v", err)
	}
	if av.AvaliacaoNota == nil || *av.AvaliacaoNota != 4 {
		t.Fatalf("nota não gravada: %+v", av.AvaliacaoNota)
	}
}

func TestReabrirSomenteConcluido(t *testing.T) {
	a := novoAmbChamado(t)
	os := a.abrir(t, "normal")
	if _, err := a.osSvc.Reabrir(os.ID, "x"); err == nil {
		t.Fatal("reabrir chamado aberto deveria falhar")
	}
	_, _ = a.osSvc.DefinirStatus(os.ID, models.OSConcluida)
	re, err := a.osSvc.Reabrir(os.ID, "voltou")
	if err != nil {
		t.Fatalf("reabrir: %v", err)
	}
	if re.Status != models.OSAberta || re.DataConclusao != nil {
		t.Fatalf("reabertura não voltou para aberta: %+v", re.Status)
	}
}

func TestAtribuirTecnico(t *testing.T) {
	a := novoAmbChamado(t)
	os := a.abrir(t, "normal")
	up, err := a.osSvc.AtribuirTecnico(os.ID, a.adminID)
	if err != nil {
		t.Fatalf("atribuir: %v", err)
	}
	if up.TecnicoID == nil || *up.TecnicoID != a.adminID {
		t.Fatal("técnico não atribuído")
	}
	// técnico inexistente
	if _, err := a.osSvc.AtribuirTecnico(os.ID, 99999); err == nil {
		t.Fatal("técnico inexistente deveria falhar")
	}
}

func TestMonitorSLANotificaUmaVez(t *testing.T) {
	a := novoAmbChamado(t)
	os := a.abrir(t, "normal")
	// Força o prazo de resposta para o passado.
	passado := time.Now().UTC().Add(-2 * time.Hour)
	if err := a.db.Model(&models.OrdemServico{}).Where("id = ?", os.ID).
		Update("prazo_resposta_em", passado).Error; err != nil {
		t.Fatalf("forçar prazo: %v", err)
	}

	a.osSvc.VerificarSLA()
	n1, _ := a.notifRepo.ContarNaoLidas(a.adminID)
	if n1 == 0 {
		t.Fatal("monitor de SLA não gerou notificação")
	}
	// Segunda passada não deve duplicar (flag de notificado).
	a.osSvc.VerificarSLA()
	n2, _ := a.notifRepo.ContarNaoLidas(a.adminID)
	if n2 != n1 {
		t.Fatalf("SLA notificou de novo: %d -> %d", n1, n2)
	}
}

func TestExcluirCategoriaComChamado(t *testing.T) {
	a := novoAmbChamado(t)
	cat, err := a.catSvc.Criar(services.EntradaCategoriaChamado{Nome: "Rede", Ativo: true})
	if err != nil {
		t.Fatalf("criar categoria: %v", err)
	}
	cid := cat.ID
	if _, err := a.osSvc.Criar(services.EntradaOS{
		Assunto: "x", DefeitoRelatado: "y", AbertoPorID: a.adminID,
		CategoriaChamadoID: &cid,
	}); err != nil {
		t.Fatalf("abrir com categoria: %v", err)
	}
	if err := a.catSvc.Excluir(cid); err == nil {
		t.Fatal("excluir categoria com chamado deveria falhar")
	}
}
