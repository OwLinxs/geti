package services

import (
	"errors"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// ServidorService - lembrando: LGPD, apenas nome e matrícula.
type ServidorService struct {
	repo      repositories.ServidorRepository
	setorRepo repositories.SetorRepository
}

func NewServidorService(repo repositories.ServidorRepository, setorRepo repositories.SetorRepository) *ServidorService {
	return &ServidorService{repo: repo, setorRepo: setorRepo}
}

type EntradaServidor struct {
	Nome      string
	Matricula string
	SetorID   *uint
	Ativo     bool
}

func (s *ServidorService) Criar(in EntradaServidor) (*models.Servidor, error) {
	if err := s.validar(in, nil); err != nil {
		return nil, err
	}
	srv := &models.Servidor{
		Nome:      strings.TrimSpace(in.Nome),
		Matricula: strings.TrimSpace(in.Matricula),
		SetorID:   in.SetorID,
		Ativo:     in.Ativo,
	}
	if err := s.repo.Criar(srv); err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *ServidorService) Listar() ([]models.Servidor, error) { return s.repo.Listar() }

func (s *ServidorService) BuscarPorID(id uint) (*models.Servidor, error) {
	srv, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return srv, nil
}

func (s *ServidorService) Atualizar(id uint, in EntradaServidor) (*models.Servidor, error) {
	srv, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := s.validar(in, &id); err != nil {
		return nil, err
	}
	srv.Nome = strings.TrimSpace(in.Nome)
	srv.Matricula = strings.TrimSpace(in.Matricula)
	srv.SetorID = in.SetorID
	srv.Ativo = in.Ativo
	if err := s.repo.Atualizar(srv); err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *ServidorService) Remover(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	n, err := s.repo.ContarItens(id)
	if err != nil {
		return err
	}
	if n > 0 {
		ev := NovoErroValidacao()
		ev.Add("servidor", "Não é possível remover: existem itens sob responsabilidade deste servidor.")
		return ev
	}
	return s.repo.Remover(id)
}

func (s *ServidorService) validar(in EntradaServidor, idAtual *uint) error {
	ev := NovoErroValidacao()
	in.Nome = strings.TrimSpace(in.Nome)
	in.Matricula = strings.TrimSpace(in.Matricula)

	if in.Nome == "" {
		ev.Add("nome", "Informe o nome do servidor.")
	}
	if in.Matricula == "" {
		ev.Add("matricula", "Informe a matrícula do servidor.")
	}
	if in.SetorID != nil {
		if _, err := s.setorRepo.BuscarPorID(*in.SetorID); err != nil {
			if errors.Is(err, repositories.ErrNaoEncontrado) {
				ev.Add("setor_id", "Setor informado não existe.")
			} else {
				return err
			}
		}
	}
	if ev.TemErros() {
		return ev
	}

	// Unicidade de matrícula.
	if in.Matricula != "" {
		existente, err := s.repo.BuscarPorMatricula(in.Matricula)
		if err == nil && (idAtual == nil || existente.ID != *idAtual) {
			ev.Add("matricula", "Já existe um servidor com esta matrícula.")
			return ev
		} else if err != nil && !errors.Is(err, repositories.ErrNaoEncontrado) {
			return err
		}
	}
	return nil
}
