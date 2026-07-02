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
	PaiID       *uint // unidade superior (nil = topo/Secretaria)
}

func (s *SetorService) Criar(in EntradaSetor) (*models.Setor, error) {
	if err := validarSetor(in); err != nil {
		return nil, err
	}
	if err := s.validarPaiExiste(in.PaiID); err != nil {
		return nil, err
	}
	st := &models.Setor{
		Nome:        strings.TrimSpace(in.Nome),
		Sigla:       strings.TrimSpace(in.Sigla),
		Localizacao: strings.TrimSpace(in.Localizacao),
		PaiID:       in.PaiID,
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
	if err := s.validarPaiExiste(in.PaiID); err != nil {
		return nil, err
	}
	// Impede ciclos: a nova unidade-pai não pode ser a própria unidade nem
	// uma de suas descendentes.
	if err := s.validarSemCiclo(id, in.PaiID); err != nil {
		return nil, err
	}

	st.Nome = strings.TrimSpace(in.Nome)
	st.Sigla = strings.TrimSpace(in.Sigla)
	st.Localizacao = strings.TrimSpace(in.Localizacao)
	st.PaiID = in.PaiID
	if err := s.repo.Atualizar(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *SetorService) Remover(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}

	nFilhos, err := s.repo.ContarFilhos(id)
	if err != nil {
		return err
	}
	if nFilhos > 0 {
		ev := NovoErroValidacao()
		ev.Add("setor", "Não é possível remover: esta unidade possui unidades subordinadas. Remova ou realoque as filhas primeiro.")
		return ev
	}

	n, err := s.repo.ContarItens(id)
	if err != nil {
		return err
	}
	if n > 0 {
		ev := NovoErroValidacao()
		ev.Add("setor", "Não é possível remover: existem itens alocados nesta unidade.")
		return ev
	}
	return s.repo.Remover(id)
}

// validarPaiExiste confirma que a unidade-pai informada existe.
func (s *SetorService) validarPaiExiste(paiID *uint) error {
	if paiID == nil {
		return nil
	}
	if _, err := s.repo.BuscarPorID(*paiID); err != nil {
		ev := NovoErroValidacao()
		ev.Add("pai_id", "Unidade superior informada não existe.")
		return ev
	}
	return nil
}

// validarSemCiclo garante que, ao mover uma unidade para baixo de outra, não
// se crie um ciclo (a unidade não pode ser sua própria ancestral). Caminha dos
// ancestrais do pai proposto até a raiz; se encontrar o próprio id, é ciclo.
func (s *SetorService) validarSemCiclo(id uint, paiID *uint) error {
	if paiID == nil {
		return nil
	}
	ev := NovoErroValidacao()
	if *paiID == id {
		ev.Add("pai_id", "Uma unidade não pode ser subordinada a si mesma.")
		return ev
	}
	atual := paiID
	// Limite de segurança contra dados inconsistentes.
	for i := 0; i < 1000 && atual != nil; i++ {
		if *atual == id {
			ev.Add("pai_id", "Vínculo inválido: geraria um ciclo na hierarquia.")
			return ev
		}
		no, err := s.repo.BuscarPorID(*atual)
		if err != nil {
			return traduzErroRepo(err)
		}
		atual = no.PaiID
	}
	return nil
}

func validarSetor(in EntradaSetor) error {
	if strings.TrimSpace(in.Nome) == "" {
		ev := NovoErroValidacao()
		ev.Add("nome", "Informe o nome da unidade.")
		return ev
	}
	return nil
}
