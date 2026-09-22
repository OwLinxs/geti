package repositories

import (
	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type NotificacaoRepository interface {
	CriarVarias(ns []models.Notificacao) error
	ListarPorUsuario(usuarioID uint, apenasNaoLidas bool, limite int) ([]models.Notificacao, error)
	ContarNaoLidas(usuarioID uint) (int64, error)
	MarcarLida(id, usuarioID uint) error
	MarcarTodasLidas(usuarioID uint) error
}

type notificacaoRepository struct {
	db *gorm.DB
}

func NewNotificacaoRepository(db *gorm.DB) NotificacaoRepository {
	return &notificacaoRepository{db: db}
}

func (r *notificacaoRepository) CriarVarias(ns []models.Notificacao) error {
	if len(ns) == 0 {
		return nil
	}
	return r.db.Create(&ns).Error
}

func (r *notificacaoRepository) ListarPorUsuario(usuarioID uint, apenasNaoLidas bool, limite int) ([]models.Notificacao, error) {
	q := r.db.Where("usuario_id = ?", usuarioID)
	if apenasNaoLidas {
		q = q.Where("lida = ?", false)
	}
	if limite <= 0 {
		limite = 30
	}
	var ns []models.Notificacao
	err := q.Order("criado_em DESC, id DESC").Limit(limite).Find(&ns).Error
	return ns, err
}

func (r *notificacaoRepository) ContarNaoLidas(usuarioID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Notificacao{}).
		Where("usuario_id = ? AND lida = ?", usuarioID, false).
		Count(&n).Error
	return n, err
}

func (r *notificacaoRepository) MarcarLida(id, usuarioID uint) error {
	return r.db.Model(&models.Notificacao{}).
		Where("id = ? AND usuario_id = ?", id, usuarioID).
		Update("lida", true).Error
}

func (r *notificacaoRepository) MarcarTodasLidas(usuarioID uint) error {
	return r.db.Model(&models.Notificacao{}).
		Where("usuario_id = ? AND lida = ?", usuarioID, false).
		Update("lida", true).Error
}
