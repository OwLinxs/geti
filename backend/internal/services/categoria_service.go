package services

import (
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

type CategoriaService struct {
	repo repositories.CategoriaRepository
}

func NewCategoriaService(repo repositories.CategoriaRepository) *CategoriaService {
	return &CategoriaService{repo: repo}
}

type EntradaCategoria struct {
	Nome       string
	Descricao  string
	Consumivel bool
}

func (s *CategoriaService) Criar(in EntradaCategoria) (*models.Categoria, error) {
	in.Nome = strings.TrimSpace(in.Nome)
	if in.Nome == "" {
		ev := NovoErroValidacao()
		ev.Add("nome", "Informe o nome da categoria.")
		return nil, ev
	}
	c := &models.Categoria{Nome: in.Nome, Descricao: strings.TrimSpace(in.Descricao), Consumivel: in.Consumivel}
	if err := s.repo.Criar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoriaService) Listar() ([]models.Categoria, error) {
	return s.repo.Listar()
}

func (s *CategoriaService) BuscarPorID(id uint) (*models.Categoria, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return c, nil
}

func (s *CategoriaService) Atualizar(id uint, in EntradaCategoria) (*models.Categoria, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	in.Nome = strings.TrimSpace(in.Nome)
	if in.Nome == "" {
		ev := NovoErroValidacao()
		ev.Add("nome", "Informe o nome da categoria.")
		return nil, ev
	}
	c.Nome = in.Nome
	c.Descricao = strings.TrimSpace(in.Descricao)
	c.Consumivel = in.Consumivel
	if err := s.repo.Atualizar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoriaService) Remover(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	n, err := s.repo.ContarItens(id)
	if err != nil {
		return err
	}
	if n > 0 {
		ev := NovoErroValidacao()
		ev.Add("categoria", "Não é possível remover: existem itens vinculados a esta categoria.")
		return ev
	}
	return s.repo.Remover(id)
}
