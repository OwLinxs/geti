package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type ServidorRepository interface {
	Criar(s *models.Servidor) error
	BuscarPorID(id uint) (*models.Servidor, error)
	BuscarPorMatricula(matricula string) (*models.Servidor, error)
	Listar() ([]models.Servidor, error)
	Atualizar(s *models.Servidor) error
	Remover(id uint) error
	ContarItens(servidorID uint) (int64, error)
}

type servidorRepository struct {
	db *gorm.DB
}

func NewServidorRepository(db *gorm.DB) ServidorRepository {
	return &servidorRepository{db: db}
}

func (r *servidorRepository) Criar(s *models.Servidor) error {
	return r.db.Create(s).Error
}

func (r *servidorRepository) BuscarPorID(id uint) (*models.Servidor, error) {
	var s models.Servidor
	if err := r.db.Preload("Setor").First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &s, nil
}

func (r *servidorRepository) BuscarPorMatricula(matricula string) (*models.Servidor, error) {
	var s models.Servidor
	if err := r.db.Where("matricula = ?", matricula).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &s, nil
}

func (r *servidorRepository) Listar() ([]models.Servidor, error) {
	var ss []models.Servidor
	if err := r.db.Preload("Setor").Order("nome ASC").Find(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil
}

func (r *servidorRepository) Atualizar(s *models.Servidor) error {
	return r.db.Save(s).Error
}

func (r *servidorRepository) Remover(id uint) error {
	return r.db.Delete(&models.Servidor{}, id).Error
}

func (r *servidorRepository) ContarItens(servidorID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Item{}).Where("responsavel_id = ?", servidorID).Count(&n).Error
	return n, err
}
