package repositories

import (
	"errors"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroReserva concentra os filtros de consulta de reservas.
type FiltroReserva struct {
	ItemID  *uint
	Status  string
	Pagina  int
	Tamanho int
}

type ReservaRepository interface {
	Criar(r *models.Reserva) error
	Atualizar(r *models.Reserva) error
	BuscarPorID(id uint) (*models.Reserva, error)
	Listar(f FiltroReserva) ([]models.Reserva, int64, error)
	Remover(id uint) error
	// ExisteConflito informa se há reserva ativa do mesmo item cujo período se
	// sobrepõe a [inicio, fim]. ignorarID exclui a própria reserva (edição).
	ExisteConflito(itemID uint, inicio, fim time.Time, ignorarID uint) (bool, error)
}

type reservaRepository struct {
	db *gorm.DB
}

func NewReservaRepository(db *gorm.DB) ReservaRepository {
	return &reservaRepository{db: db}
}

func (r *reservaRepository) preloads(q *gorm.DB) *gorm.DB {
	return q.Preload("Item").Preload("Solicitante").Preload("AprovadoPor")
}

func (r *reservaRepository) Criar(res *models.Reserva) error {
	return r.db.Create(res).Error
}

func (r *reservaRepository) Atualizar(res *models.Reserva) error {
	return r.db.Save(res).Error
}

func (r *reservaRepository) BuscarPorID(id uint) (*models.Reserva, error) {
	var res models.Reserva
	if err := r.preloads(r.db).First(&res, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &res, nil
}

func (r *reservaRepository) Listar(f FiltroReserva) ([]models.Reserva, int64, error) {
	q := r.db.Model(&models.Reserva{})

	if f.ItemID != nil {
		q = q.Where("item_id = ?", *f.ItemID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Tamanho <= 0 {
		f.Tamanho = 20
	}
	if f.Pagina <= 0 {
		f.Pagina = 1
	}
	offset := (f.Pagina - 1) * f.Tamanho

	var lista []models.Reserva
	err := r.preloads(q).
		Order("data_inicio DESC, id DESC").
		Limit(f.Tamanho).Offset(offset).
		Find(&lista).Error
	if err != nil {
		return nil, 0, err
	}
	return lista, total, nil
}

func (r *reservaRepository) Remover(id uint) error {
	return r.db.Delete(&models.Reserva{}, id).Error
}

func (r *reservaRepository) ExisteConflito(itemID uint, inicio, fim time.Time, ignorarID uint) (bool, error) {
	// Sobreposição: reserva existente começa antes do fim da nova E termina
	// depois do início da nova. Considera apenas reservas ativas.
	q := r.db.Model(&models.Reserva{}).
		Where("item_id = ?", itemID).
		Where("status IN ?", []models.StatusReserva{models.ReservaReservada, models.ReservaEmUso}).
		Where("data_inicio <= ? AND data_fim >= ?", fim, inicio)
	if ignorarID != 0 {
		q = q.Where("id <> ?", ignorarID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
