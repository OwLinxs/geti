package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type ConfiguracaoRepository interface {
	Obter(chave string) (string, error)
	Salvar(chave, valor string) error
}

type configuracaoRepository struct {
	db *gorm.DB
}

func NewConfiguracaoRepository(db *gorm.DB) ConfiguracaoRepository {
	return &configuracaoRepository{db: db}
}

func (r *configuracaoRepository) Obter(chave string) (string, error) {
	var c models.Configuracao
	if err := r.db.First(&c, "chave = ?", chave).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNaoEncontrado
		}
		return "", err
	}
	return c.Valor, nil
}

func (r *configuracaoRepository) Salvar(chave, valor string) error {
	c := models.Configuracao{Chave: chave, Valor: valor}
	// Upsert por chave.
	return r.db.Save(&c).Error
}
