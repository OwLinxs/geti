package repositories

import (
	"errors"
	"strconv"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroItem agrupa os critérios de busca/filtragem de itens.
type FiltroItem struct {
	Texto           string // busca por descrição, patrimônio, série, marca, modelo
	CategoriaID     *uint
	SetorID         *uint
	ResponsavelID   *uint
	Estado          *models.EstadoConservacao
	SomenteBaixados *bool // nil = todos; true = só baixados; false = só ativos
	AbaixoDoMinimo  bool  // true = só consumíveis em alerta de estoque
	Reservavel      *bool // nil = todos; true = só reserváveis
	Consumivel      *bool // nil = todos; true = materiais de consumo; false = patrimônio
	LoteCodigo      string
	SemPatrimonio   bool
	SemSetor        bool
	Pagina          int
	TamanhoPagina   int
}

type ItemRepository interface {
	Criar(i *models.Item) error
	CriarLote(itens []models.Item) error
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

func (r *itemRepository) CriarLote(itens []models.Item) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(&itens, 50).Error
	})
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
		busca := strings.ToUpper(strings.TrimSpace(f.Texto))
		id, err := strconv.ParseUint(strings.TrimPrefix(busca, "TI-"), 10, 64)
		if strings.HasPrefix(busca, "TI-") && err == nil && id > 0 {
			q = q.Where("id = ?", id)
		} else {
			// LOWER(...) para busca case-insensitive portável (SQLite e PostgreSQL).
			like := "%" + strings.ToLower(f.Texto) + "%"
			q = q.Where(
				"LOWER(descricao) LIKE ? OR LOWER(numero_patrimonio) LIKE ? OR LOWER(numero_serie) LIKE ? OR LOWER(marca) LIKE ? OR LOWER(modelo) LIKE ?",
				like, like, like, like, like,
			)
		}
	}
	if f.LoteCodigo != "" {
		q = q.Where("lote_codigo = ?", f.LoteCodigo)
	}
	if f.SemPatrimonio {
		q = q.Where("numero_patrimonio IS NULL AND categoria_id IN (SELECT id FROM categorias WHERE consumivel = ?)", false)
	}
	if f.SemSetor {
		q = q.Where("setor_id IS NULL AND categoria_id IN (SELECT id FROM categorias WHERE consumivel = ?)", false)
	}
	if f.CategoriaID != nil {
		q = q.Where("categoria_id = ?", *f.CategoriaID)
	}
	if f.Consumivel != nil {
		q = q.Where("categoria_id IN (SELECT id FROM categorias WHERE consumivel = ?)", *f.Consumivel)
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
	if f.Reservavel != nil {
		q = q.Where("reservavel = ?", *f.Reservavel)
	}
	if f.SomenteBaixados != nil {
		q = q.Where("baixado = ?", *f.SomenteBaixados)
	}
	if f.AbaixoDoMinimo {
		q = q.Where("estoque_minimo > 0 AND quantidade < estoque_minimo AND categoria_id IN (SELECT id FROM categorias WHERE consumivel = ?)", true)
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
		Where("baixado = ? AND estoque_minimo > 0 AND quantidade < estoque_minimo AND categoria_id IN (SELECT id FROM categorias WHERE consumivel = ?)", false, true).
		Order("descricao ASC").
		Find(&itens).Error
	return itens, err
}
