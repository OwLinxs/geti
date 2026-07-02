package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type SetorRepository interface {
	Criar(s *models.Setor) error
	BuscarPorID(id uint) (*models.Setor, error)
	Listar() ([]models.Setor, error)
	Atualizar(s *models.Setor) error
	Remover(id uint) error
	ContarItens(setorID uint) (int64, error)
}

type setorRepository struct {
	db *gorm.DB
}

func NewSetorRepository(db *gorm.DB) SetorRepository {
	return &setorRepository{db: db}
}

func (r *setorRepository) Criar(s *models.Setor) error {
	return r.db.Create(s).Error
}

func (r *setorRepository) BuscarPorID(id uint) (*models.Setor, error) {
	var s models.Setor
	if err := r.db.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &s, nil
}

func (r *setorRepository) Listar() ([]models.Setor, error) {
	var ss []models.Setor
	if err := r.db.Order("nome ASC").Find(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil
}

func (r *setorRepository) Atualizar(s *models.Setor) error {
	return r.db.Save(s).Error
}

func (r *setorRepository) Remover(id uint) error {
	return r.db.Delete(&models.Setor{}, id).Error
}

func (r *setorRepository) ContarItens(setorID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Item{}).Where("setor_id = ?", setorID).Count(&n).Error
	return n, err
}
