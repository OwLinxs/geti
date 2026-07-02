package services

import (
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

type SetorService struct {
	repo repositories.SetorRepository
}

func NewSetorService(repo repositories.SetorRepository) *SetorService {
	return &SetorService{repo: repo}
}

type EntradaSetor struct {
	Nome        string
	Sigla       string
	Localizacao string
}

func (s *SetorService) Criar(in EntradaSetor) (*models.Setor, error) {
	if err := validarSetor(in); err != nil {
		return nil, err
	}
	st := &models.Setor{
		Nome:        strings.TrimSpace(in.Nome),
		Sigla:       strings.TrimSpace(in.Sigla),
		Localizacao: strings.TrimSpace(in.Localizacao),
	}
	if err := s.repo.Criar(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *SetorService) Listar() ([]models.Setor, error) { return s.repo.Listar() }

func (s *SetorService) BuscarPorID(id uint) (*models.Setor, error) {
	st, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return st, nil
}

func (s *SetorService) Atualizar(id uint, in EntradaSetor) (*models.Setor, error) {
	st, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := validarSetor(in); err != nil {
		return nil, err
	}
	st.Nome = strings.TrimSpace(in.Nome)
	st.Sigla = strings.TrimSpace(in.Sigla)
	st.Localizacao = strings.TrimSpace(in.Localizacao)
	if err := s.repo.Atualizar(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *SetorService) Remover(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	n, err := s.repo.ContarItens(id)
	if err != nil {
		return err
	}
	if n > 0 {
		ev := NovoErroValidacao()
		ev.Add("setor", "Não é possível remover: existem itens alocados neste setor.")
		return ev
	}
	return s.repo.Remover(id)
}

func validarSetor(in EntradaSetor) error {
	if strings.TrimSpace(in.Nome) == "" {
		ev := NovoErroValidacao()
		ev.Add("nome", "Informe o nome do setor.")
		return ev
	}
	return nil
}
