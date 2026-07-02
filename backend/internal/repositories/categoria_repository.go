package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type CategoriaRepository interface {
	Criar(c *models.Categoria) error
	BuscarPorID(id uint) (*models.Categoria, error)
	Listar() ([]models.Categoria, error)
	Atualizar(c *models.Categoria) error
	Remover(id uint) error
	ContarItens(categoriaID uint) (int64, error)
}

type categoriaRepository struct {
	db *gorm.DB
}

func NewCategoriaRepository(db *gorm.DB) CategoriaRepository {
	return &categoriaRepository{db: db}
}

func (r *categoriaRepository) Criar(c *models.Categoria) error {
	return r.db.Create(c).Error
}

func (r *categoriaRepository) BuscarPorID(id uint) (*models.Categoria, error) {
	var c models.Categoria
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoriaRepository) Listar() ([]models.Categoria, error) {
	var cs []models.Categoria
	if err := r.db.Order("nome ASC").Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

func (r *categoriaRepository) Atualizar(c *models.Categoria) error {
	return r.db.Save(c).Error
}

func (r *categoriaRepository) Remover(id uint) error {
	return r.db.Delete(&models.Categoria{}, id).Error
}

func (r *categoriaRepository) ContarItens(categoriaID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Item{}).Where("categoria_id = ?", categoriaID).Count(&n).Error
	return n, err
}
