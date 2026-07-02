package repositories

import (
	"errors"

	"github.com/pmfb/sige-ti/internal/models"
	"gorm.io/gorm"
)

// UsuarioRepository abstrai o acesso a dados de usuários.
type UsuarioRepository interface {
	Criar(u *models.Usuario) error
	BuscarPorID(id uint) (*models.Usuario, error)
	BuscarPorEmail(email string) (*models.Usuario, error)
	Listar() ([]models.Usuario, error)
	Atualizar(u *models.Usuario) error
	ContarAdministradores() (int64, error)
}

type usuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) Criar(u *models.Usuario) error {
	return r.db.Create(u).Error
}

func (r *usuarioRepository) BuscarPorID(id uint) (*models.Usuario, error) {
	var u models.Usuario
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &u, nil
}

func (r *usuarioRepository) BuscarPorEmail(email string) (*models.Usuario, error) {
	var u models.Usuario
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return &u, nil
}

func (r *usuarioRepository) Listar() ([]models.Usuario, error) {
	var us []models.Usuario
	if err := r.db.Order("nome ASC").Find(&us).Error; err != nil {
		return nil, err
	}
	return us, nil
}

func (r *usuarioRepository) Atualizar(u *models.Usuario) error {
	return r.db.Save(u).Error
}

func (r *usuarioRepository) ContarAdministradores() (int64, error) {
	var n int64
	err := r.db.Model(&models.Usuario{}).
		Where("perfil = ? AND ativo = ?", models.PerfilAdministrador, true).
		Count(&n).Error
	return n, err
}
