package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type RespostaRapidaRepository interface {
	Criar(r *models.RespostaRapida) error
	Atualizar(r *models.RespostaRapida) error
	BuscarPorID(id uint) (*models.RespostaRapida, error)
	Listar(apenasAtivas bool) ([]models.RespostaRapida, error)
	Remover(id uint) error
	Contar() (int64, error)
}

type respostaRapidaRepository struct {
	db *gorm.DB
}

func NewRespostaRapidaRepository(db *gorm.DB) RespostaRapidaRepository {
	return &respostaRapidaRepository{db: db}
}

func (r *respostaRapidaRepository) Criar(x *models.RespostaRapida) error {
	return r.db.Create(x).Error
}

func (r *respostaRapidaRepository) Atualizar(x *models.RespostaRapida) error {
	return r.db.Save(x).Error
}

func (r *respostaRapidaRepository) BuscarPorID(id uint) (*models.RespostaRapida, error) {
	var x models.RespostaRapida
	if err := r.db.First(&x, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &x, nil
}

func (r *respostaRapidaRepository) Listar(apenasAtivas bool) ([]models.RespostaRapida, error) {
	q := r.db.Model(&models.RespostaRapida{})
	if apenasAtivas {
		q = q.Where("ativo = ?", true)
	}
	var lista []models.RespostaRapida
	err := q.Order("ordem ASC, titulo ASC").Find(&lista).Error
	return lista, err
}

func (r *respostaRapidaRepository) Remover(id uint) error {
	return r.db.Delete(&models.RespostaRapida{}, id).Error
}

func (r *respostaRapidaRepository) Contar() (int64, error) {
	var n int64
	err := r.db.Model(&models.RespostaRapida{}).Count(&n).Error
	return n, err
}
