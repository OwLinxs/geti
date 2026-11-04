package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// CampoCategoriaRepository gerencia as definições de campos customizados por
// categoria e os valores (atributos) desses campos nos itens.
type CampoCategoriaRepository interface {
	ListarPorCategoria(categoriaID uint) ([]models.CampoCategoria, error)
	BuscarPorID(id uint) (*models.CampoCategoria, error)
	Criar(c *models.CampoCategoria) error
	Atualizar(c *models.CampoCategoria) error
	Remover(id uint) error
	// SubstituirAtributosDoItem apaga os atributos atuais do item e grava os
	// novos, em uma transação.
	SubstituirAtributosDoItem(itemID uint, atributos []models.ItemAtributo) error
}

type campoCategoriaRepository struct {
	db *gorm.DB
}

func NewCampoCategoriaRepository(db *gorm.DB) CampoCategoriaRepository {
	return &campoCategoriaRepository{db: db}
}

func (r *campoCategoriaRepository) ListarPorCategoria(categoriaID uint) ([]models.CampoCategoria, error) {
	var cs []models.CampoCategoria
	err := r.db.Where("categoria_id = ?", categoriaID).
		Order("ordem ASC, id ASC").
		Find(&cs).Error
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func (r *campoCategoriaRepository) BuscarPorID(id uint) (*models.CampoCategoria, error) {
	var c models.CampoCategoria
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &c, nil
}

func (r *campoCategoriaRepository) Criar(c *models.CampoCategoria) error {
	return r.db.Create(c).Error
}

func (r *campoCategoriaRepository) Atualizar(c *models.CampoCategoria) error {
	return r.db.Save(c).Error
}

func (r *campoCategoriaRepository) Remover(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove os valores desse campo em todos os itens e, depois, o campo.
		if err := tx.Where("campo_id = ?", id).Delete(&models.ItemAtributo{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.CampoCategoria{}, id).Error
	})
}

func (r *campoCategoriaRepository) SubstituirAtributosDoItem(itemID uint, atributos []models.ItemAtributo) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("item_id = ?", itemID).Delete(&models.ItemAtributo{}).Error; err != nil {
			return err
		}
		if len(atributos) == 0 {
			return nil
		}
		for i := range atributos {
			atributos[i].ItemID = itemID
		}
		return tx.Create(&atributos).Error
	})
}
