package services

import (
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"gorm.io/gorm"
)

// TermoService gera termos de responsabilidade. Os parâmetros institucionais
// (nome da Prefeitura, logo, cabeçalho) vêm da configuração via env.
type TermoService struct {
	repo     repositories.TermoRepository
	itemRepo repositories.ItemRepository
	servRepo repositories.ServidorRepository
	cfg      *config.Config
}

func NewTermoService(
	repo repositories.TermoRepository,
	itemRepo repositories.ItemRepository,
	servRepo repositories.ServidorRepository,
	cfg *config.Config,
) *TermoService {
	return &TermoService{repo: repo, itemRepo: itemRepo, servRepo: servRepo, cfg: cfg}
}

type EntradaTermo struct {
	ItemID         uint
	ServidorID     uint
	MovimentacaoID *uint
	Observacao     string
	EmitidoPorID   uint
}

// Emitir cria o registro do termo (com numeração sequencial e snapshots).
func (s *TermoService) Emitir(in EntradaTermo) (*models.TermoResponsabilidade, error) {
	ev := NovoErroValidacao()
	if in.ItemID == 0 {
		ev.Add("item_id", "Informe o item do termo.")
	}
	if in.ServidorID == 0 {
		ev.Add("servidor_id", "Informe o servidor responsável.")
	}
	if in.EmitidoPorID == 0 {
		ev.Add("emitido_por", "Termo sem usuário emissor.")
	}
	if ev.TemErros() {
		return nil, ev
	}

	item, err := s.itemRepo.BuscarPorID(in.ItemID)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	servidor, err := s.servRepo.BuscarPorID(in.ServidorID)
	if err != nil {
		return nil, traduzErroRepo(err)
	}

	// Numeração sequencial + criação em transação para evitar números duplicados.
	var termo *models.TermoResponsabilidade
	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		numero, err := s.repo.ProximoNumero(tx, time.Now().UTC().Year())
		if err != nil {
			return err
		}
		termo = s.montar(numero, item, servidor, in)
		return s.repo.CriarComTx(tx, termo)
	})
	if err != nil {
		return nil, err
	}
	return termo, nil
}

func (s *TermoService) montar(numero string, item *models.Item, servidor *models.Servidor, in EntradaTermo) *models.TermoResponsabilidade {
	patrimonio := ""
	if item.NumeroPatrimonio != nil {
		patrimonio = *item.NumeroPatrimonio
	}
	return &models.TermoResponsabilidade{
		Numero:                    numero,
		ItemID:                    item.ID,
		ServidorID:                servidor.ID,
		MovimentacaoID:            in.MovimentacaoID,
		EmitidoPorID:              in.EmitidoPorID,
		ItemDescricaoSnapshot:     item.Descricao,
		PatrimonioSnapshot:        patrimonio,
		ServidorNomeSnapshot:      servidor.Nome,
		ServidorMatriculaSnapshot: servidor.Matricula,
		Observacao:                strings.TrimSpace(in.Observacao),
		DataEmissao:               time.Now().UTC(),
	}
}

func (s *TermoService) BuscarPorID(id uint) (*models.TermoResponsabilidade, error) {
	t, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return t, nil
}

func (s *TermoService) Listar() ([]models.TermoResponsabilidade, error) {
	return s.repo.Listar()
}

// GerarPDF produz o PDF do termo.
func (s *TermoService) GerarPDF(id uint) ([]byte, *models.TermoResponsabilidade, error) {
	t, err := s.BuscarPorID(id)
	if err != nil {
		return nil, nil, err
	}
	pdf, err := PDFTermoResponsabilidade(t, s.cfg)
	if err != nil {
		return nil, nil, err
	}
	return pdf, t, nil
}
