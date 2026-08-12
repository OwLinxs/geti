package repositories

import (
	"errors"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

type FornecedorRepository interface {
	Criar(f *models.Fornecedor) error
	Atualizar(f *models.Fornecedor) error
	BuscarPorID(id uint) (*models.Fornecedor, error)
	Listar(q string) ([]models.Fornecedor, error)
	Remover(id uint) error
	ContarContratos(fornecedorID uint) (int64, error)
}

type fornecedorRepository struct {
	db *gorm.DB
}

func NewFornecedorRepository(db *gorm.DB) FornecedorRepository {
	return &fornecedorRepository{db: db}
}

func (r *fornecedorRepository) Criar(f *models.Fornecedor) error {
	return r.db.Create(f).Error
}

func (r *fornecedorRepository) Atualizar(f *models.Fornecedor) error {
	return r.db.Save(f).Error
}

func (r *fornecedorRepository) BuscarPorID(id uint) (*models.Fornecedor, error) {
	var f models.Fornecedor
	if err := r.db.First(&f, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &f, nil
}

func (r *fornecedorRepository) Listar(q string) ([]models.Fornecedor, error) {
	query := r.db.Model(&models.Fornecedor{})
	if termo := strings.TrimSpace(q); termo != "" {
		like := "%" + termo + "%"
		query = query.Where("nome LIKE ? OR cnpj LIKE ?", like, like)
	}
	var lista []models.Fornecedor
	err := query.Order("nome ASC").Find(&lista).Error
	return lista, err
}

func (r *fornecedorRepository) Remover(id uint) error {
	return r.db.Delete(&models.Fornecedor{}, id).Error
}

func (r *fornecedorRepository) ContarContratos(fornecedorID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Contrato{}).
		Where("fornecedor_id = ?", fornecedorID).Count(&n).Error
	return n, err
}
