package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// ReservaService gerencia reservas/agendamentos de equipamentos.
type ReservaService struct {
	repo     repositories.ReservaRepository
	itemRepo repositories.ItemRepository
	servRepo repositories.ServidorRepository
}

func NewReservaService(
	repo repositories.ReservaRepository,
	itemRepo repositories.ItemRepository,
	servRepo repositories.ServidorRepository,
) *ReservaService {
	return &ReservaService{repo: repo, itemRepo: itemRepo, servRepo: servRepo}
}

// EntradaReserva reúne os dados de criação/edição de uma reserva.
type EntradaReserva struct {
	ItemID        uint
	SolicitanteID *uint
	DataInicio    *time.Time
	DataFim       *time.Time
	Finalidade    string
	LocalDestino  string
	Status        string
	Observacao    string
	AprovadoPorID uint // usuário autenticado (na criação)
}

// validar checa os dados e devolve o status resolvido. ignorarID exclui a
// própria reserva da checagem de conflito (usado na edição).
func (s *ReservaService) validar(in EntradaReserva, ignorarID uint) (models.StatusReserva, error) {
	ev := NovoErroValidacao()

	if in.ItemID == 0 {
		ev.Add("item_id", "Selecione o equipamento.")
	} else {
		item, err := s.itemRepo.BuscarPorID(in.ItemID)
		if err != nil {
			ev.Add("item_id", "Equipamento não encontrado.")
		} else {
			if item.Baixado {
				ev.Add("item_id", "Equipamento baixado não pode ser reservado.")
			}
			if item.Categoria != nil && item.Categoria.Consumivel {
				ev.Add("item_id", "Materiais de consumo não são reserváveis; apenas equipamentos patrimoniados.")
			}
		}
	}

	if in.DataInicio == nil {
		ev.Add("data_inicio", "Informe a data de início.")
	}
	if in.DataFim == nil {
		ev.Add("data_fim", "Informe a data de término.")
	}
	if in.DataInicio != nil && in.DataFim != nil && in.DataFim.Before(*in.DataInicio) {
		ev.Add("data_fim", "A data de término não pode ser anterior ao início.")
	}

	if in.SolicitanteID != nil {
		if _, err := s.servRepo.BuscarPorID(*in.SolicitanteID); err != nil {
			ev.Add("solicitante_id", "Solicitante não encontrado.")
		}
	}

	status := models.StatusReserva(strings.TrimSpace(in.Status))
	if status == "" {
		status = models.ReservaReservada
	}
	if !models.StatusReservaValido(status) {
		ev.Add("status", "Status inválido.")
	}

	if ev.TemErros() {
		return "", ev
	}

	// Conflito de datas só importa quando a reserva ocupa o equipamento.
	if status.Ativa() {
		conflito, err := s.repo.ExisteConflito(in.ItemID, *in.DataInicio, *in.DataFim, ignorarID)
		if err != nil {
			return "", err
		}
		if conflito {
			return "", fmt.Errorf("%w: já existe uma reserva ativa deste equipamento no período informado", ErrRegraNegocio)
		}
	}

	return status, nil
}

func aplicarReserva(r *models.Reserva, in EntradaReserva, status models.StatusReserva) {
	r.ItemID = in.ItemID
	r.SolicitanteID = in.SolicitanteID
	r.DataInicio = *in.DataInicio
	r.DataFim = *in.DataFim
	r.Finalidade = strings.TrimSpace(in.Finalidade)
	r.LocalDestino = strings.TrimSpace(in.LocalDestino)
	r.Status = status
	r.Observacao = strings.TrimSpace(in.Observacao)
}

func (s *ReservaService) Criar(in EntradaReserva) (*models.Reserva, error) {
	status, err := s.validar(in, 0)
	if err != nil {
		return nil, err
	}
	r := &models.Reserva{}
	aplicarReserva(r, in, status)
	if in.AprovadoPorID != 0 {
		id := in.AprovadoPorID
		r.AprovadoPorID = &id
	}
	if err := s.repo.Criar(r); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(r.ID)
}

func (s *ReservaService) Atualizar(id uint, in EntradaReserva) (*models.Reserva, error) {
	r, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	status, err := s.validar(in, id)
	if err != nil {
		return nil, err
	}
	aplicarReserva(r, in, status)
	limparAssociacoesReserva(r)
	if err := s.repo.Atualizar(r); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

// DefinirStatus muda apenas o status (ex.: marcar "em uso", "devolvida",
// "cancelada"). Ao reativar, revalida conflito de datas.
func (s *ReservaService) DefinirStatus(id uint, novo models.StatusReserva) (*models.Reserva, error) {
	if !models.StatusReservaValido(novo) {
		ev := NovoErroValidacao()
		ev.Add("status", "Status inválido.")
		return nil, ev
	}
	r, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if novo.Ativa() {
		conflito, err := s.repo.ExisteConflito(r.ItemID, r.DataInicio, r.DataFim, r.ID)
		if err != nil {
			return nil, err
		}
		if conflito {
			return nil, fmt.Errorf("%w: já existe uma reserva ativa deste equipamento no período", ErrRegraNegocio)
		}
	}
	r.Status = novo
	limparAssociacoesReserva(r)
	if err := s.repo.Atualizar(r); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

func (s *ReservaService) BuscarPorID(id uint) (*models.Reserva, error) {
	r, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return r, nil
}

func (s *ReservaService) Listar(f repositories.FiltroReserva) ([]models.Reserva, int64, error) {
	return s.repo.Listar(f)
}

// EquipamentoReservavel representa um item do pool de reservas com seu estado
// atual (reserva ativa quando alocado; nil quando está no departamento).
type EquipamentoReservavel struct {
	Item         models.Item     `json:"item"`
	ReservaAtiva *models.Reserva `json:"reserva_ativa"`
}

// ListarEquipamentos devolve os itens marcados como reserváveis (não baixados)
// com a reserva ativa de cada um, para o painel de alocação.
func (s *ReservaService) ListarEquipamentos() ([]EquipamentoReservavel, error) {
	reservavel := true
	ativos := false
	itens, _, err := s.itemRepo.Listar(repositories.FiltroItem{
		Reservavel:      &reservavel,
		SomenteBaixados: &ativos,
		Pagina:          1,
		TamanhoPagina:   500,
	})
	if err != nil {
		return nil, err
	}
	out := make([]EquipamentoReservavel, 0, len(itens))
	for i := range itens {
		ativa, err := s.repo.BuscarAtivaPorItem(itens[i].ID)
		if err != nil {
			return nil, err
		}
		out = append(out, EquipamentoReservavel{Item: itens[i], ReservaAtiva: ativa})
	}
	return out, nil
}

func (s *ReservaService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	return s.repo.Remover(id)
}

func limparAssociacoesReserva(r *models.Reserva) {
	r.Item = nil
	r.Solicitante = nil
	r.AprovadoPor = nil
}
