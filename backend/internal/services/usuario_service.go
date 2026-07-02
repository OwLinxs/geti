package services

import (
	"errors"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// UsuarioService concentra as regras de negócio de usuários do sistema.
type UsuarioService struct {
	repo repositories.UsuarioRepository
}

func NewUsuarioService(repo repositories.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

// EntradaUsuario representa os dados de criação/edição de um usuário.
type EntradaUsuario struct {
	Nome   string
	Email  string
	Senha  string
	Perfil models.Perfil
	Ativo  bool
}

func (s *UsuarioService) Criar(in EntradaUsuario) (*models.Usuario, error) {
	ev := NovoErroValidacao()
	in.Nome = strings.TrimSpace(in.Nome)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	if in.Nome == "" {
		ev.Add("nome", "Informe o nome do usuário.")
	}
	if !emailValido(in.Email) {
		ev.Add("email", "Informe um e-mail válido.")
	}
	if len(in.Senha) < 6 {
		ev.Add("senha", "A senha deve ter ao menos 6 caracteres.")
	}
	if !models.PerfilValido(in.Perfil) {
		ev.Add("perfil", "Perfil inválido. Use 'administrador' ou 'operador'.")
	}
	if ev.TemErros() {
		return nil, ev
	}

	if _, err := s.repo.BuscarPorEmail(in.Email); err == nil {
		ev.Add("email", "Já existe um usuário com este e-mail.")
		return nil, ev
	} else if !errors.Is(err, repositories.ErrNaoEncontrado) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.Usuario{
		Nome:      in.Nome,
		Email:     in.Email,
		SenhaHash: string(hash),
		Perfil:    in.Perfil,
		Ativo:     in.Ativo,
	}
	if err := s.repo.Criar(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UsuarioService) Listar() ([]models.Usuario, error) {
	return s.repo.Listar()
}

// RedefinirSenha troca a senha de um usuário (uso administrativo). Gera novo
// hash bcrypt. Não revela se o usuário existe além do erro padrão.
func (s *UsuarioService) RedefinirSenha(id uint, novaSenha string) (*models.Usuario, error) {
	if len(novaSenha) < 6 {
		ev := NovoErroValidacao()
		ev.Add("senha", "A nova senha deve ter ao menos 6 caracteres.")
		return nil, ev
	}
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.SenhaHash = string(hash)
	if err := s.repo.Atualizar(u); err != nil {
		return nil, err
	}
	return u, nil
}

// DefinirAtivo ativa ou desativa um usuário (não há exclusão física: usuário
// inativo não consegue autenticar). Impede desativar o último administrador
// ativo, para não trancar o acesso administrativo ao sistema.
func (s *UsuarioService) DefinirAtivo(id uint, ativo bool) (*models.Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if u.Ativo == ativo {
		return u, nil // sem mudança
	}
	// Ao desativar um administrador, garante que reste ao menos um admin ativo.
	if !ativo && u.Perfil == models.PerfilAdministrador {
		n, err := s.repo.ContarAdministradores()
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			ev := NovoErroValidacao()
			ev.Add("ativo", "Não é possível desativar o último administrador ativo do sistema.")
			return nil, ev
		}
	}
	u.Ativo = ativo
	if err := s.repo.Atualizar(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UsuarioService) BuscarPorID(id uint) (*models.Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return u, nil
}

// Autenticar valida credenciais e devolve o usuário se válido e ativo.
func (s *UsuarioService) Autenticar(email, senha string) (*models.Usuario, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.repo.BuscarPorEmail(email)
	if err != nil {
		if errors.Is(err, repositories.ErrNaoEncontrado) {
			return nil, ErrCredenciaisInvalidas
		}
		return nil, err
	}
	if !u.Ativo {
		return nil, ErrCredenciaisInvalidas
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(senha)); err != nil {
		return nil, ErrCredenciaisInvalidas
	}
	return u, nil
}
