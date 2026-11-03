package repositories

import (
	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// PermissaoUsuarioRepository abstrai o acesso às concessões de permissão por
// usuário.
type PermissaoUsuarioRepository interface {
	// Concedidas devolve as chaves concedidas a um usuário.
	Concedidas(usuarioID uint) ([]string, error)
	// Existe informa se o usuário tem a chave concedida.
	Existe(usuarioID uint, chave string) (bool, error)
	// Substituir troca o conjunto de concessões do usuário pelo informado
	// (operação atômica: apaga as antigas e grava as novas).
	Substituir(usuarioID uint, chaves []string) error
}

type permissaoUsuarioRepository struct {
	db *gorm.DB
}

func NewPermissaoUsuarioRepository(db *gorm.DB) PermissaoUsuarioRepository {
	return &permissaoUsuarioRepository{db: db}
}

func (r *permissaoUsuarioRepository) Concedidas(usuarioID uint) ([]string, error) {
	var chaves []string
	err := r.db.Model(&models.PermissaoUsuario{}).
		Where("usuario_id = ?", usuarioID).
		Order("chave ASC").
		Pluck("chave", &chaves).Error
	return chaves, err
}

func (r *permissaoUsuarioRepository) Existe(usuarioID uint, chave string) (bool, error) {
	var n int64
	err := r.db.Model(&models.PermissaoUsuario{}).
		Where("usuario_id = ? AND chave = ?", usuarioID, chave).
		Count(&n).Error
	return n > 0, err
}

func (r *permissaoUsuarioRepository) Substituir(usuarioID uint, chaves []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("usuario_id = ?", usuarioID).
			Delete(&models.PermissaoUsuario{}).Error; err != nil {
			return err
		}
		if len(chaves) == 0 {
			return nil
		}
		linhas := make([]models.PermissaoUsuario, 0, len(chaves))
		for _, ch := range chaves {
			linhas = append(linhas, models.PermissaoUsuario{UsuarioID: usuarioID, Chave: ch})
		}
		return tx.Create(&linhas).Error
	})
}
