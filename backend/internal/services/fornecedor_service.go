package services

import (
	"fmt"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// FornecedorService gerencia o cadastro de fornecedores.
type FornecedorService struct {
	repo repositories.FornecedorRepository
}

func NewFornecedorService(repo repositories.FornecedorRepository) *FornecedorService {
	return &FornecedorService{repo: repo}
}

// EntradaFornecedor reúne os dados de criação/edição de um fornecedor.
type EntradaFornecedor struct {
	Nome       string
	CNPJ       string
	Email      string
	Telefone   string
	Endereco   string
	Observacao string
}

func (s *FornecedorService) validar(in EntradaFornecedor) error {
	ev := NovoErroValidacao()
	if strings.TrimSpace(in.Nome) == "" {
		ev.Add("nome", "Informe o nome do fornecedor.")
	}
	if ev.TemErros() {
		return ev
	}
	return nil
}

func aplicarFornecedor(f *models.Fornecedor, in EntradaFornecedor) {
	f.Nome = strings.TrimSpace(in.Nome)
	f.CNPJ = strings.TrimSpace(in.CNPJ)
	f.Email = strings.TrimSpace(in.Email)
	f.Telefone = strings.TrimSpace(in.Telefone)
	f.Endereco = strings.TrimSpace(in.Endereco)
	f.Observacao = strings.TrimSpace(in.Observacao)
}

func (s *FornecedorService) Criar(in EntradaFornecedor) (*models.Fornecedor, error) {
	if err := s.validar(in); err != nil {
		return nil, err
	}
	f := &models.Fornecedor{}
	aplicarFornecedor(f, in)
	if err := s.repo.Criar(f); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(f.ID)
}

func (s *FornecedorService) Atualizar(id uint, in EntradaFornecedor) (*models.Fornecedor, error) {
	f, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	aplicarFornecedor(f, in)
	if err := s.repo.Atualizar(f); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

func (s *FornecedorService) BuscarPorID(id uint) (*models.Fornecedor, error) {
	f, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return f, nil
}

func (s *FornecedorService) Listar(q string) ([]models.Fornecedor, error) {
	return s.repo.Listar(q)
}

// Excluir remove um fornecedor, recusando quando há contratos vinculados
// (preserva a integridade do histórico contratual).
func (s *FornecedorService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	n, err := s.repo.ContarContratos(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: o fornecedor possui contratos vinculados", ErrRegraNegocio)
	}
	return s.repo.Remover(id)
}
