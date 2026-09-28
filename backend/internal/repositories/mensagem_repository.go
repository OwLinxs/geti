package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type MensagemRepository interface {
	Criar(m *models.MensagemChamado) error
	ListarPorOS(osID uint) ([]models.MensagemChamado, error)
	// ListarPublicasPorOS exclui notas internas (visão do solicitante).
	ListarPublicasPorOS(osID uint) ([]models.MensagemChamado, error)
	BuscarPorID(id uint) (*models.MensagemChamado, error)
	// ExistePorIdExterno indica se já há mensagem com esse id externo (dedup de
	// entregas repetidas do WhatsApp/SSE).
	ExistePorIdExterno(idExterno string) (bool, error)
	AtualizarStatusPorIdExterno(idExterno string, status models.StatusMensagem) error
	// AtualizarEnvioExterno grava o id externo (wa_message_id) e o status de
	// entrega de uma mensagem enviada ao WhatsApp.
	AtualizarEnvioExterno(id uint, idExterno string, status models.StatusMensagem) error
}

type mensagemRepository struct {
	db *gorm.DB
}

func NewMensagemRepository(db *gorm.DB) MensagemRepository {
	return &mensagemRepository{db: db}
}

func (r *mensagemRepository) Criar(m *models.MensagemChamado) error {
	return r.db.Create(m).Error
}

func (r *mensagemRepository) ListarPorOS(osID uint) ([]models.MensagemChamado, error) {
	var msgs []models.MensagemChamado
	err := r.db.Preload("Autor").
		Where("ordem_servico_id = ?", osID).
		Order("enviada_em ASC, id ASC").
		Find(&msgs).Error
	return msgs, err
}

func (r *mensagemRepository) ListarPublicasPorOS(osID uint) ([]models.MensagemChamado, error) {
	var msgs []models.MensagemChamado
	err := r.db.Preload("Autor").
		Where("ordem_servico_id = ? AND interna = ?", osID, false).
		Order("enviada_em ASC, id ASC").
		Find(&msgs).Error
	return msgs, err
}

func (r *mensagemRepository) BuscarPorID(id uint) (*models.MensagemChamado, error) {
	var m models.MensagemChamado
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &m, nil
}

func (r *mensagemRepository) AtualizarEnvioExterno(id uint, idExterno string, status models.StatusMensagem) error {
	campos := map[string]any{"status": status}
	if idExterno != "" {
		campos["id_externo"] = idExterno
	}
	return r.db.Model(&models.MensagemChamado{}).
		Where("id = ?", id).
		Updates(campos).Error
}

func (r *mensagemRepository) ExistePorIdExterno(idExterno string) (bool, error) {
	if idExterno == "" {
		return false, nil
	}
	var n int64
	err := r.db.Model(&models.MensagemChamado{}).
		Where("id_externo = ?", idExterno).
		Count(&n).Error
	return n > 0, err
}

func (r *mensagemRepository) AtualizarStatusPorIdExterno(idExterno string, status models.StatusMensagem) error {
	if idExterno == "" {
		return nil
	}
	return r.db.Model(&models.MensagemChamado{}).
		Where("id_externo = ?", idExterno).
		Update("status", status).Error
}
