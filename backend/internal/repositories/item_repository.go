package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroItem agrupa os critérios de busca/filtragem de itens.
type FiltroItem struct {
	Texto            string // busca por descrição, patrimônio, série, marca, modelo
	CategoriaID      *uint
	SetorID          *uint
	ResponsavelID    *uint
	Estado           *models.EstadoConservacao
	SomenteBaixados  *bool // nil = todos; true = só baixados; false = só ativos
	AbaixoDoMinimo   bool  // true = só consumíveis em alerta de estoque
	Pagina           int
	TamanhoPagina    int
}

type ItemRepository interface {
	Criar(i *models.Item) error
	BuscarPorID(id uint) (*models.Item, error)
	Listar(f FiltroItem) ([]models.Item, int64, error)
	Atualizar(i *models.Item) error
	// Remover faz soft delete (GORM): o registro fica oculto nas consultas.
	Remover(id uint) error
	ListarAbaixoDoMinimo() ([]models.Item, error)
	// AtualizarComTx permite ajustar o item dentro de uma transação externa
	// (usado pelo serviço de movimentação para recalcular estoque atomicamente).
	AtualizarComTx(tx *gorm.DB, i *models.Item) error
}

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.Preload("Categoria").Preload("Setor").Preload("Responsavel")
}

func (r *itemRepository) Criar(i *models.Item) error {
	return r.db.Create(i).Error
}

func (r *itemRepository) BuscarPorID(id uint) (*models.Item, error) {
	var i models.Item
	if err := r.preloads(r.db).First(&i, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &i, nil
}

func (r *itemRepository) Listar(f FiltroItem) ([]models.Item, int64, error) {
	q := r.db.Model(&models.Item{})

	if f.Texto != "" {
		like := "%" + f.Texto + "%"
		q = q.Where(
			"descricao LIKE ? OR numero_patrimonio LIKE ? OR numero_serie LIKE ? OR marca LIKE ? OR modelo LIKE ?",
			like, like, like, like, like,
		)
	}
	if f.CategoriaID != nil {
		q = q.Where("categoria_id = ?", *f.CategoriaID)
	}
	if f.SetorID != nil {
		q = q.Where("setor_id = ?", *f.SetorID)
	}
	if f.ResponsavelID != nil {
		q = q.Where("responsavel_id = ?", *f.ResponsavelID)
	}
	if f.Estado != nil {
		q = q.Where("estado_conservacao = ?", *f.Estado)
	}
	if f.SomenteBaixados != nil {
		q = q.Where("baixado = ?", *f.SomenteBaixados)
	}
	if f.AbaixoDoMinimo {
		q = q.Where("estoque_minimo > 0 AND quantidade < estoque_minimo")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = r.preloads(q).Order("descricao ASC")

	if f.TamanhoPagina > 0 {
		offset := (f.Pagina - 1) * f.TamanhoPagina
		if offset < 0 {
			offset = 0
		}
		q = q.Limit(f.TamanhoPagina).Offset(offset)
	}

	var itens []models.Item
	if err := q.Find(&itens).Error; err != nil {
		return nil, 0, err
	}
	return itens, total, nil
}

func (r *itemRepository) Atualizar(i *models.Item) error {
	return r.db.Save(i).Error
}

func (r *itemRepository) AtualizarComTx(tx *gorm.DB, i *models.Item) error {
	return tx.Save(i).Error
}

func (r *itemRepository) Remover(id uint) error {
	return r.db.Delete(&models.Item{}, id).Error
}

func (r *itemRepository) ListarAbaixoDoMinimo() ([]models.Item, error) {
	var itens []models.Item
	err := r.preloads(r.db).
		Where("baixado = ? AND estoque_minimo > 0 AND quantidade < estoque_minimo", false).
		Order("descricao ASC").
		Find(&itens).Error
	return itens, err
}
