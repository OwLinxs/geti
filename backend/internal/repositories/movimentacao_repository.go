package repositories

import (
	"errors"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroMovimentacao agrupa critérios de consulta ao histórico.
type FiltroMovimentacao struct {
	ItemID    *uint
	Tipo      *models.TipoMovimentacao
	DataInicio *time.Time
	DataFim    *time.Time
	Pagina        int
	TamanhoPagina int
}

type MovimentacaoRepository interface {
	// CriarComTx insere a movimentação dentro de uma transação (atomicidade
	// com o recálculo de estoque do item).
	CriarComTx(tx *gorm.DB, m *models.Movimentacao) error
	BuscarPorID(id uint) (*models.Movimentacao, error)
	Listar(f FiltroMovimentacao) ([]models.Movimentacao, int64, error)
	ListarPorItem(itemID uint) ([]models.Movimentacao, error)
	// ContarPorItem informa quantas movimentações o item possui (histórico).
	ContarPorItem(itemID uint) (int64, error)
	// DB expõe o handle para abertura de transações no serviço.
	DB() *gorm.DB
}

type movimentacaoRepository struct {
	db *gorm.DB
}

func NewMovimentacaoRepository(db *gorm.DB) MovimentacaoRepository {
	return &movimentacaoRepository{db: db}
}

func (r *movimentacaoRepository) DB() *gorm.DB { return r.db }

func (r *movimentacaoRepository) CriarComTx(tx *gorm.DB, m *models.Movimentacao) error {
	return tx.Create(m).Error
}

func (r *movimentacaoRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Item").
		Preload("SetorOrigem").
		Preload("SetorDestino").
		Preload("Servidor").
		Preload("RegistradoPor")
}

func (r *movimentacaoRepository) BuscarPorID(id uint) (*models.Movimentacao, error) {
	var m models.Movimentacao
	if err := r.preloads(r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &m, nil
}

func (r *movimentacaoRepository) Listar(f FiltroMovimentacao) ([]models.Movimentacao, int64, error) {
	q := r.db.Model(&models.Movimentacao{})

	if f.ItemID != nil {
		q = q.Where("item_id = ?", *f.ItemID)
	}
	if f.Tipo != nil {
		q = q.Where("tipo = ?", *f.Tipo)
	}
	if f.DataInicio != nil {
		q = q.Where("data_evento >= ?", *f.DataInicio)
	}
	if f.DataFim != nil {
		q = q.Where("data_evento <= ?", *f.DataFim)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = r.preloads(q).Order("data_evento DESC, id DESC")

	if f.TamanhoPagina > 0 {
		offset := (f.Pagina - 1) * f.TamanhoPagina
		if offset < 0 {
			offset = 0
		}
		q = q.Limit(f.TamanhoPagina).Offset(offset)
	}

	var movs []models.Movimentacao
	if err := q.Find(&movs).Error; err != nil {
		return nil, 0, err
	}
	return movs, total, nil
}

func (r *movimentacaoRepository) ContarPorItem(itemID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Movimentacao{}).Where("item_id = ?", itemID).Count(&n).Error
	return n, err
}

func (r *movimentacaoRepository) ListarPorItem(itemID uint) ([]models.Movimentacao, error) {
	var movs []models.Movimentacao
	err := r.preloads(r.db).
		Where("item_id = ?", itemID).
		Order("data_evento DESC, id DESC").
		Find(&movs).Error
	return movs, err
}
