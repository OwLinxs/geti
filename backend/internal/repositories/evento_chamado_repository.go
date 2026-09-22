package repositories

import (
	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type EventoChamadoRepository interface {
	Criar(e *models.EventoChamado) error
	ListarPorOS(osID uint) ([]models.EventoChamado, error)
}

type eventoChamadoRepository struct {
	db *gorm.DB
}

func NewEventoChamadoRepository(db *gorm.DB) EventoChamadoRepository {
	return &eventoChamadoRepository{db: db}
}

func (r *eventoChamadoRepository) Criar(e *models.EventoChamado) error {
	return r.db.Create(e).Error
}

func (r *eventoChamadoRepository) ListarPorOS(osID uint) ([]models.EventoChamado, error) {
	var lista []models.EventoChamado
	err := r.db.Where("ordem_servico_id = ?", osID).
		Order("criado_em ASC, id ASC").Find(&lista).Error
	return lista, err
}
