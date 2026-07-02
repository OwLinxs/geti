package repositories

import (
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroAuditoria concentra os filtros de consulta da trilha de auditoria.
type FiltroAuditoria struct {
	UsuarioID *uint
	Recurso   string
	Acao      string
	De        *time.Time
	Ate       *time.Time
	Pagina    int
	Tamanho   int
}

type AuditoriaRepository interface {
	Registrar(r *models.RegistroAuditoria) error
	Listar(f FiltroAuditoria) ([]models.RegistroAuditoria, int64, error)
}

type auditoriaRepository struct {
	db *gorm.DB
}

func NewAuditoriaRepository(db *gorm.DB) AuditoriaRepository {
	return &auditoriaRepository{db: db}
}

func (r *auditoriaRepository) Registrar(reg *models.RegistroAuditoria) error {
	return r.db.Create(reg).Error
}

func (r *auditoriaRepository) Listar(f FiltroAuditoria) ([]models.RegistroAuditoria, int64, error) {
	q := r.db.Model(&models.RegistroAuditoria{})

	if f.UsuarioID != nil {
		q = q.Where("usuario_id = ?", *f.UsuarioID)
	}
	if f.Recurso != "" {
		q = q.Where("recurso = ?", f.Recurso)
	}
	if f.Acao != "" {
		q = q.Where("acao = ?", f.Acao)
	}
	if f.De != nil {
		q = q.Where("criado_em >= ?", *f.De)
	}
	if f.Ate != nil {
		q = q.Where("criado_em <= ?", *f.Ate)
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

	var regs []models.RegistroAuditoria
	if err := q.Order("criado_em DESC").Limit(f.Tamanho).Offset(offset).Find(&regs).Error; err != nil {
		return nil, 0, err
	}
	return regs, total, nil
}
