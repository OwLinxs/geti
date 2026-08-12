package services

import (
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// ContratoService gerencia contratos com fornecedores.
type ContratoService struct {
	repo     repositories.ContratoRepository
	fornRepo repositories.FornecedorRepository
}

func NewContratoService(
	repo repositories.ContratoRepository,
	fornRepo repositories.FornecedorRepository,
) *ContratoService {
	return &ContratoService{repo: repo, fornRepo: fornRepo}
}

// EntradaContrato reúne os dados de criação/edição de um contrato.
type EntradaContrato struct {
	Numero       string
	FornecedorID uint
	Objeto       string
	DataInicio   *time.Time
	DataFim      *time.Time
	Valor        *float64
	Status       string
	Observacao   string
}

func (s *ContratoService) validar(in EntradaContrato) (models.StatusContrato, error) {
	ev := NovoErroValidacao()

	if strings.TrimSpace(in.Numero) == "" {
		ev.Add("numero", "Informe o número do contrato.")
	}
	if strings.TrimSpace(in.Objeto) == "" {
		ev.Add("objeto", "Informe o objeto do contrato.")
	}
	if in.FornecedorID == 0 {
		ev.Add("fornecedor_id", "Selecione o fornecedor.")
	} else if _, err := s.fornRepo.BuscarPorID(in.FornecedorID); err != nil {
		ev.Add("fornecedor_id", "Fornecedor não encontrado.")
	}
	if in.DataInicio != nil && in.DataFim != nil && in.DataFim.Before(*in.DataInicio) {
		ev.Add("data_fim", "A data de término não pode ser anterior ao início.")
	}
	if in.Valor != nil && *in.Valor < 0 {
		ev.Add("valor", "O valor não pode ser negativo.")
	}

	status := models.StatusContrato(strings.TrimSpace(in.Status))
	if status == "" {
		status = models.ContratoVigente
	}
	if !models.StatusContratoValido(status) {
		ev.Add("status", "Status inválido.")
	}

	if ev.TemErros() {
		return "", ev
	}
	return status, nil
}

func aplicarContrato(c *models.Contrato, in EntradaContrato, status models.StatusContrato) {
	c.Numero = strings.TrimSpace(in.Numero)
	c.FornecedorID = in.FornecedorID
	c.Objeto = strings.TrimSpace(in.Objeto)
	c.DataInicio = in.DataInicio
	c.DataFim = in.DataFim
	c.Valor = in.Valor
	c.Status = status
	c.Observacao = strings.TrimSpace(in.Observacao)
}

func (s *ContratoService) Criar(in EntradaContrato) (*models.Contrato, error) {
	status, err := s.validar(in)
	if err != nil {
		return nil, err
	}
	c := &models.Contrato{}
	aplicarContrato(c, in, status)
	if err := s.repo.Criar(c); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(c.ID)
}

func (s *ContratoService) Atualizar(id uint, in EntradaContrato) (*models.Contrato, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	status, err := s.validar(in)
	if err != nil {
		return nil, err
	}
	aplicarContrato(c, in, status)
	c.Fornecedor = nil // evita upsert da associação no Save
	if err := s.repo.Atualizar(c); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

func (s *ContratoService) BuscarPorID(id uint) (*models.Contrato, error) {
	c, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return c, nil
}

func (s *ContratoService) Listar(f repositories.FiltroContrato) ([]models.Contrato, int64, error) {
	return s.repo.Listar(f)
}

func (s *ContratoService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	return s.repo.Remover(id)
}
