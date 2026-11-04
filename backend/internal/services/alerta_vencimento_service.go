package services

import (
	"fmt"
	"log"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// AlertaVencimentoService varre garantias de itens e contratos próximos do
// vencimento e avisa a equipe (uma vez por registro, via flag de controle).
type AlertaVencimentoService struct {
	itemRepo     repositories.ItemRepository
	contratoRepo repositories.ContratoRepository
	notifSvc     *NotificacaoService
	dias         int
}

func NewAlertaVencimentoService(
	itemRepo repositories.ItemRepository,
	contratoRepo repositories.ContratoRepository,
	notifSvc *NotificacaoService,
	dias int,
) *AlertaVencimentoService {
	return &AlertaVencimentoService{
		itemRepo:     itemRepo,
		contratoRepo: contratoRepo,
		notifSvc:     notifSvc,
		dias:         dias,
	}
}

// Verificar roda a varredura de vencimentos. Chamado periodicamente pelo
// agendador. Best-effort: erros são logados, não interrompem o restante.
func (s *AlertaVencimentoService) Verificar() {
	if s.dias <= 0 {
		return
	}
	limite := time.Now().UTC().Add(time.Duration(s.dias) * 24 * time.Hour)
	s.verificarGarantias(limite)
	s.verificarContratos(limite)
}

func (s *AlertaVencimentoService) verificarGarantias(limite time.Time) {
	itens, err := s.itemRepo.ListarGarantiaVencendo(limite)
	if err != nil {
		log.Printf("[venc] falha ao listar garantias: %v", err)
		return
	}
	for i := range itens {
		it := itens[i]
		if it.GarantiaAte == nil {
			continue
		}
		msg := fmt.Sprintf("%s · garantia até %s.",
			it.Descricao, it.GarantiaAte.Format("02/01/2006"))
		s.notifSvc.NotificarEquipe("Garantia vencendo", msg, models.NotifVencimento, it.ID)

		it.AlertaGarantiaEnviado = true
		if err := s.itemRepo.Atualizar(&it); err != nil {
			log.Printf("[venc] falha ao marcar item %d: %v", it.ID, err)
		}
	}
}

func (s *AlertaVencimentoService) verificarContratos(limite time.Time) {
	contratos, err := s.contratoRepo.ListarVencendoNaoNotificados(limite)
	if err != nil {
		log.Printf("[venc] falha ao listar contratos: %v", err)
		return
	}
	for i := range contratos {
		ct := contratos[i]
		if ct.DataFim == nil {
			continue
		}
		fornecedor := ""
		if ct.Fornecedor != nil {
			fornecedor = " · " + ct.Fornecedor.Nome
		}
		msg := fmt.Sprintf("Contrato %s%s vence em %s.",
			ct.Numero, fornecedor, ct.DataFim.Format("02/01/2006"))
		s.notifSvc.NotificarEquipe("Contrato vencendo", msg, models.NotifVencimento, ct.ID)

		ct.AlertaVencimentoEnviado = true
		if err := s.contratoRepo.Atualizar(&ct); err != nil {
			log.Printf("[venc] falha ao marcar contrato %d: %v", ct.ID, err)
		}
	}
}
