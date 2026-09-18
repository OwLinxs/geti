package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type CategoriaChamadoRepository interface {
	Criar(c *models.CategoriaChamado) error
	Atualizar(c *models.CategoriaChamado) error
	BuscarPorID(id uint) (*models.CategoriaChamado, error)
	Listar(apenasAtivas bool) ([]models.CategoriaChamado, error)
	Remover(id uint) error
	Contar() (int64, error)
	ContarChamados(categoriaID uint) (int64, error)
}

type categoriaChamadoRepository struct {
	db *gorm.DB
}

func NewCategoriaChamadoRepository(db *gorm.DB) CategoriaChamadoRepository {
	return &categoriaChamadoRepository{db: db}
}

func (r *categoriaChamadoRepository) Criar(c *models.CategoriaChamado) error {
	return r.db.Create(c).Error
}

func (r *categoriaChamadoRepository) Atualizar(c *models.CategoriaChamado) error {
	return r.db.Save(c).Error
}

func (r *categoriaChamadoRepository) BuscarPorID(id uint) (*models.CategoriaChamado, error) {
	var c models.CategoriaChamado
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoriaChamadoRepository) Listar(apenasAtivas bool) ([]models.CategoriaChamado, error) {
	q := r.db.Model(&models.CategoriaChamado{})
	if apenasAtivas {
		q = q.Where("ativo = ?", true)
	}
	var lista []models.CategoriaChamado
	err := q.Order("ordem ASC, nome ASC").Find(&lista).Error
	return lista, err
}

func (r *categoriaChamadoRepository) Remover(id uint) error {
	return r.db.Delete(&models.CategoriaChamado{}, id).Error
}

func (r *categoriaChamadoRepository) Contar() (int64, error) {
	var n int64
	err := r.db.Model(&models.CategoriaChamado{}).Count(&n).Error
	return n, err
}

func (r *categoriaChamadoRepository) ContarChamados(categoriaID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.OrdemServico{}).
		Where("categoria_chamado_id = ?", categoriaID).Count(&n).Error
	return n, err
}
