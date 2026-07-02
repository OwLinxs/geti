package services

import (
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// RelatorioService produz consultas agregadas para relatórios e exportações.
type RelatorioService struct {
	itemRepo repositories.ItemRepository
	movRepo  repositories.MovimentacaoRepository
	cfg      *config.Config
}

func NewRelatorioService(itemRepo repositories.ItemRepository, movRepo repositories.MovimentacaoRepository, cfg *config.Config) *RelatorioService {
	return &RelatorioService{itemRepo: itemRepo, movRepo: movRepo, cfg: cfg}
}

// Config expõe a configuração para geração de PDFs parametrizados.
func (s *RelatorioService) Config() *config.Config { return s.cfg }

// ItensPorSetor retorna itens filtrados por setor (todos se setorID == nil).
func (s *RelatorioService) ItensPorSetor(setorID *uint) ([]models.Item, error) {
	itens, _, err := s.itemRepo.Listar(repositories.FiltroItem{SetorID: setorID})
	return itens, err
}

// ItensPorResponsavel retorna itens sob responsabilidade de um servidor.
func (s *RelatorioService) ItensPorResponsavel(servidorID *uint) ([]models.Item, error) {
	itens, _, err := s.itemRepo.Listar(repositories.FiltroItem{ResponsavelID: servidorID})
	return itens, err
}

// EstoqueBaixo retorna consumíveis abaixo do mínimo.
func (s *RelatorioService) EstoqueBaixo() ([]models.Item, error) {
	return s.itemRepo.ListarAbaixoDoMinimo()
}

// MovimentacoesPorPeriodo retorna o histórico em um intervalo.
func (s *RelatorioService) MovimentacoesPorPeriodo(inicio, fim *time.Time, tipo *models.TipoMovimentacao) ([]models.Movimentacao, error) {
	f := repositories.FiltroMovimentacao{DataInicio: inicio, DataFim: fim, Tipo: tipo}
	movs, _, err := s.movRepo.Listar(f)
	return movs, err
}

// InventarioCompleto retorna todos os itens (para exportação geral).
func (s *RelatorioService) InventarioCompleto() ([]models.Item, error) {
	itens, _, err := s.itemRepo.Listar(repositories.FiltroItem{})
	return itens, err
}
