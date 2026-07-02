package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type TermoRepository interface {
	Criar(t *models.TermoResponsabilidade) error
	CriarComTx(tx *gorm.DB, t *models.TermoResponsabilidade) error
	BuscarPorID(id uint) (*models.TermoResponsabilidade, error)
	Listar() ([]models.TermoResponsabilidade, error)
	// ContarPorItem informa quantos termos foram emitidos para o item.
	ContarPorItem(itemID uint) (int64, error)
	// ProximoNumero gera o próximo número sequencial no formato TR-AAAA-NNNN.
	ProximoNumero(tx *gorm.DB, ano int) (string, error)
	// DB expõe o handle para abertura de transações no serviço.
	DB() *gorm.DB
}

type termoRepository struct {
	db *gorm.DB
}

func NewTermoRepository(db *gorm.DB) TermoRepository {
	return &termoRepository{db: db}
}

func (r *termoRepository) DB() *gorm.DB { return r.db }

func (r *termoRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Item").
		Preload("Servidor").
		Preload("EmitidoPor").
		Preload("Movimentacao")
}

func (r *termoRepository) Criar(t *models.TermoResponsabilidade) error {
	return r.db.Create(t).Error
}

func (r *termoRepository) CriarComTx(tx *gorm.DB, t *models.TermoResponsabilidade) error {
	return tx.Create(t).Error
}

func (r *termoRepository) BuscarPorID(id uint) (*models.TermoResponsabilidade, error) {
	var t models.TermoResponsabilidade
	if err := r.preloads(r.db).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &t, nil
}

func (r *termoRepository) Listar() ([]models.TermoResponsabilidade, error) {
	var ts []models.TermoResponsabilidade
	err := r.preloads(r.db).Order("data_emissao DESC, id DESC").Find(&ts).Error
	return ts, err
}

func (r *termoRepository) ContarPorItem(itemID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.TermoResponsabilidade{}).Where("item_id = ?", itemID).Count(&n).Error
	return n, err
}

func (r *termoRepository) ProximoNumero(tx *gorm.DB, ano int) (string, error) {
	var count int64
	inicio := time.Date(ano, 1, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(ano+1, 1, 1, 0, 0, 0, 0, time.UTC)
	err := tx.Model(&models.TermoResponsabilidade{}).
		Where("data_emissao >= ? AND data_emissao < ?", inicio, fim).
		Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TR-%04d-%04d", ano, count+1), nil
}
