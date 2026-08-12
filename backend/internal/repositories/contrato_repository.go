package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroContrato concentra os filtros de consulta de contratos.
type FiltroContrato struct {
	Q            string
	Status       string
	FornecedorID *uint
	// VencendoAteDias, quando > 0, filtra contratos vigentes cuja data de
	// término cai entre hoje e hoje+N dias (alerta de vencimento).
	VencendoAteDias *int
	Pagina          int
	Tamanho         int
}

type ContratoRepository interface {
	Criar(c *models.Contrato) error
	Atualizar(c *models.Contrato) error
	BuscarPorID(id uint) (*models.Contrato, error)
	Listar(f FiltroContrato) ([]models.Contrato, int64, error)
	Remover(id uint) error
}

type contratoRepository struct {
	db *gorm.DB
}

func NewContratoRepository(db *gorm.DB) ContratoRepository {
	return &contratoRepository{db: db}
}

func (r *contratoRepository) Criar(c *models.Contrato) error {
	return r.db.Create(c).Error
}

func (r *contratoRepository) Atualizar(c *models.Contrato) error {
	return r.db.Save(c).Error
}

func (r *contratoRepository) BuscarPorID(id uint) (*models.Contrato, error) {
	var c models.Contrato
	if err := r.db.Preload("Fornecedor").First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &c, nil
}

func (r *contratoRepository) Listar(f FiltroContrato) ([]models.Contrato, int64, error) {
	q := r.db.Model(&models.Contrato{})

	if termo := strings.TrimSpace(f.Q); termo != "" {
		like := "%" + termo + "%"
		q = q.Where("numero LIKE ? OR objeto LIKE ?", like, like)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.FornecedorID != nil {
		q = q.Where("fornecedor_id = ?", *f.FornecedorID)
	}
	if f.VencendoAteDias != nil && *f.VencendoAteDias > 0 {
		agora := time.Now().UTC()
		limite := agora.Add(time.Duration(*f.VencendoAteDias) * 24 * time.Hour)
		q = q.Where("status = ? AND data_fim IS NOT NULL AND data_fim >= ? AND data_fim <= ?",
			models.ContratoVigente, agora, limite)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Tamanho <= 0 {
		f.Tamanho = 20
	}
	if f.Pagina <= 0 {
		f.Pagina = 1
	}
	offset := (f.Pagina - 1) * f.Tamanho

	var lista []models.Contrato
	err := q.Preload("Fornecedor").
		Order("data_fim IS NULL, data_fim ASC, id DESC").
		Limit(f.Tamanho).Offset(offset).
		Find(&lista).Error
	if err != nil {
		return nil, 0, err
	}
	return lista, total, nil
}

func (r *contratoRepository) Remover(id uint) error {
	return r.db.Delete(&models.Contrato{}, id).Error
}
