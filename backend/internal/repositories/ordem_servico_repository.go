package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroOrdemServico concentra os filtros de consulta da fila de manutenção.
type FiltroOrdemServico struct {
	Status    string
	TecnicoID *uint
	De        *time.Time
	Ate       *time.Time
	Pagina    int
	Tamanho   int
}

type OrdemServicoRepository interface {
	CriarComTx(tx *gorm.DB, os *models.OrdemServico) error
	Atualizar(os *models.OrdemServico) error
	BuscarPorID(id uint) (*models.OrdemServico, error)
	Listar(f FiltroOrdemServico) ([]models.OrdemServico, int64, error)
	Remover(id uint) error
	// SubstituirPassos troca todos os passos de uma OS numa transação.
	SubstituirPassos(osID uint, passos []models.OrdemServicoPasso) error
	// ProximoNumero gera o próximo número sequencial no formato OS-AAAA-NNNN.
	ProximoNumero(tx *gorm.DB, ano int) (string, error)
	DB() *gorm.DB
}

type ordemServicoRepository struct {
	db *gorm.DB
}

func NewOrdemServicoRepository(db *gorm.DB) OrdemServicoRepository {
	return &ordemServicoRepository{db: db}
}

func (r *ordemServicoRepository) DB() *gorm.DB { return r.db }

func (r *ordemServicoRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Item").
		Preload("Setor").
		Preload("Solicitante").
		Preload("Tecnico").
		Preload("AbertoPor").
		Preload("Passos", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC, id ASC")
		})
}

func (r *ordemServicoRepository) CriarComTx(tx *gorm.DB, os *models.OrdemServico) error {
	return tx.Create(os).Error
}

func (r *ordemServicoRepository) Atualizar(os *models.OrdemServico) error {
	return r.db.Save(os).Error
}

func (r *ordemServicoRepository) BuscarPorID(id uint) (*models.OrdemServico, error) {
	var os models.OrdemServico
	if err := r.preloads(r.db).First(&os, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &os, nil
}

func (r *ordemServicoRepository) Listar(f FiltroOrdemServico) ([]models.OrdemServico, int64, error) {
	q := r.db.Model(&models.OrdemServico{})

	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.TecnicoID != nil {
		q = q.Where("tecnico_id = ?", *f.TecnicoID)
	}
	if f.De != nil {
		q = q.Where("data_abertura >= ?", *f.De)
	}
	if f.Ate != nil {
		q = q.Where("data_abertura <= ?", *f.Ate)
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

	var lista []models.OrdemServico
	err := r.preloads(q).
		Order("data_abertura DESC, id DESC").
		Limit(f.Tamanho).Offset(offset).
		Find(&lista).Error
	if err != nil {
		return nil, 0, err
	}
	return lista, total, nil
}

func (r *ordemServicoRepository) Remover(id uint) error {
	return r.db.Delete(&models.OrdemServico{}, id).Error
}

func (r *ordemServicoRepository) SubstituirPassos(osID uint, passos []models.OrdemServicoPasso) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ordem_servico_id = ?", osID).
			Delete(&models.OrdemServicoPasso{}).Error; err != nil {
			return err
		}
		if len(passos) == 0 {
			return nil
		}
		for i := range passos {
			passos[i].OrdemServicoID = osID
		}
		return tx.Create(&passos).Error
	})
}

func (r *ordemServicoRepository) ProximoNumero(tx *gorm.DB, ano int) (string, error) {
	var count int64
	inicio := time.Date(ano, 1, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(ano+1, 1, 1, 0, 0, 0, 0, time.UTC)
	// Conta inclusive registros soft-deletados para não reaproveitar números.
	err := tx.Unscoped().Model(&models.OrdemServico{}).
		Where("data_abertura >= ? AND data_abertura < ?", inicio, fim).
		Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("OS-%04d-%04d", ano, count+1), nil
}
