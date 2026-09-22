package repositories

import (
	"errors"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// FiltroConhecimento concentra os filtros de consulta da base de conhecimento.
type FiltroConhecimento struct {
	Q         string // busca em título/conteúdo
	Categoria string
	Publicado *bool
	Pagina    int
	Tamanho   int
}

type ConhecimentoRepository interface {
	Criar(a *models.ArtigoConhecimento) error
	Atualizar(a *models.ArtigoConhecimento) error
	BuscarPorID(id uint) (*models.ArtigoConhecimento, error)
	Listar(f FiltroConhecimento) ([]models.ArtigoConhecimento, int64, error)
	Remover(id uint) error
	IncrementarVisualizacoes(id uint) error
}

type conhecimentoRepository struct {
	db *gorm.DB
}

func NewConhecimentoRepository(db *gorm.DB) ConhecimentoRepository {
	return &conhecimentoRepository{db: db}
}

func (r *conhecimentoRepository) Criar(a *models.ArtigoConhecimento) error {
	return r.db.Create(a).Error
}

func (r *conhecimentoRepository) Atualizar(a *models.ArtigoConhecimento) error {
	return r.db.Save(a).Error
}

func (r *conhecimentoRepository) BuscarPorID(id uint) (*models.ArtigoConhecimento, error) {
	var a models.ArtigoConhecimento
	if err := r.db.Preload("Autor").First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &a, nil
}

func (r *conhecimentoRepository) Listar(f FiltroConhecimento) ([]models.ArtigoConhecimento, int64, error) {
	q := r.db.Model(&models.ArtigoConhecimento{})

	if termo := strings.TrimSpace(f.Q); termo != "" {
		like := "%" + strings.ToLower(termo) + "%"
		q = q.Where("LOWER(titulo) LIKE ? OR LOWER(conteudo) LIKE ?", like, like)
	}
	if f.Categoria != "" {
		q = q.Where("categoria = ?", f.Categoria)
	}
	if f.Publicado != nil {
		q = q.Where("publicado = ?", *f.Publicado)
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

	var lista []models.ArtigoConhecimento
	err := q.Preload("Autor").
		Order("atualizado_em DESC, id DESC").
		Limit(f.Tamanho).Offset(offset).
		Find(&lista).Error
	if err != nil {
		return nil, 0, err
	}
	return lista, total, nil
}

func (r *conhecimentoRepository) Remover(id uint) error {
	return r.db.Delete(&models.ArtigoConhecimento{}, id).Error
}

func (r *conhecimentoRepository) IncrementarVisualizacoes(id uint) error {
	return r.db.Model(&models.ArtigoConhecimento{}).
		Where("id = ?", id).
		UpdateColumn("visualizacoes", gorm.Expr("visualizacoes + 1")).Error
}
