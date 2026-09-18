package services

import (
	"fmt"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// CategoriaChamadoService gerencia as categorias de chamado (configuráveis).
type CategoriaChamadoService struct {
	repo repositories.CategoriaChamadoRepository
}

func NewCategoriaChamadoService(repo repositories.CategoriaChamadoRepository) *CategoriaChamadoService {
	return &CategoriaChamadoService{repo: repo}
}

type EntradaCategoriaChamado struct {
	Nome      string
	Descricao string
	Ordem     int
	Ativo     bool
}

func (s *CategoriaChamadoService) validar(in EntradaCategoriaChamado) error {
	ev := NovoErroValidacao()
	if strings.TrimSpace(in.Nome) == "" {
		ev.Add("nome", "Informe o nome da categoria.")
	}
	if ev.TemErros() {
		return ev
	}
	return nil
}

func (s *CategoriaChamadoService) Criar(in EntradaCategoriaChamado) (*models.CategoriaChamado, error) {
	if err := s.validar(in); err != nil {
		return nil, err
	}
	c := &models.CategoriaChamado{
		Nome:      strings.TrimSpace(in.Nome),
		Descricao: strings.TrimSpace(in.Descricao),
		Ordem:     in.Ordem,
		Ativo:     in.Ativo,
	}
	if err := s.repo.Criar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoriaChamadoService) Atualizar(id uint, in EntradaCategoriaChamado) (*models.CategoriaChamado, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	c.Nome = strings.TrimSpace(in.Nome)
	c.Descricao = strings.TrimSpace(in.Descricao)
	c.Ordem = in.Ordem
	c.Ativo = in.Ativo
	if err := s.repo.Atualizar(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoriaChamadoService) Listar(apenasAtivas bool) ([]models.CategoriaChamado, error) {
	return s.repo.Listar(apenasAtivas)
}

// Excluir recusa a remoção quando há chamados vinculados (mantém integridade).
func (s *CategoriaChamadoService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	n, err := s.repo.ContarChamados(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: a categoria possui chamados vinculados (desative-a em vez de excluir)", ErrRegraNegocio)
	}
	return s.repo.Remover(id)
}

// SemearPadrao cria as categorias iniciais se ainda não houver nenhuma.
func (s *CategoriaChamadoService) SemearPadrao() error {
	n, err := s.repo.Contar()
	if err != nil || n > 0 {
		return err
	}
	padrao := []string{"Rede", "E-mail", "Impressora", "Hardware", "Software", "Sistema", "Outros"}
	for i, nome := range padrao {
		_ = s.repo.Criar(&models.CategoriaChamado{Nome: nome, Ordem: i + 1, Ativo: true})
	}
	return nil
}
