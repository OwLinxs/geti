package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"gorm.io/gorm"
)

// OrdemServicoService orquestra o módulo de manutenção (ordens de serviço).
type OrdemServicoService struct {
	repo        repositories.OrdemServicoRepository
	itemRepo    repositories.ItemRepository
	setorRepo   repositories.SetorRepository
	servRepo    repositories.ServidorRepository
	usuarioRepo repositories.UsuarioRepository
	cfg         *config.Config
}

func NewOrdemServicoService(
	repo repositories.OrdemServicoRepository,
	itemRepo repositories.ItemRepository,
	setorRepo repositories.SetorRepository,
	servRepo repositories.ServidorRepository,
	usuarioRepo repositories.UsuarioRepository,
	cfg *config.Config,
) *OrdemServicoService {
	return &OrdemServicoService{
		repo:        repo,
		itemRepo:    itemRepo,
		setorRepo:   setorRepo,
		servRepo:    servRepo,
		usuarioRepo: usuarioRepo,
		cfg:         cfg,
	}
}

// EntradaOS reúne os dados de abertura/edição de uma ordem de serviço.
type EntradaOS struct {
	ItemID                   uint // 0 = máquina externa
	EquipamentoDescricao     string
	EquipamentoIdentificacao string
	SetorID                  *uint
	SolicitanteID            *uint
	DefeitoRelatado          string
	Diagnostico              string
	SolucaoAplicada          string
	Prioridade               string
	TecnicoID                *uint
	AbertoPorID              uint // usado apenas na criação
}

// dadosValidados guarda as entidades resolvidas durante a validação.
type dadosValidados struct {
	item            *models.Item
	solicitanteNome string
	prioridade      models.PrioridadeOS
}

func (s *OrdemServicoService) validar(in EntradaOS) (*dadosValidados, error) {
	ev := NovoErroValidacao()
	d := &dadosValidados{}

	if strings.TrimSpace(in.DefeitoRelatado) == "" {
		ev.Add("defeito_relatado", "Descreva o defeito relatado.")
	}

	// Equipamento: item do inventário OU descrição de máquina externa.
	if in.ItemID != 0 {
		item, err := s.itemRepo.BuscarPorID(in.ItemID)
		if err != nil {
			ev.Add("item_id", "Item do inventário não encontrado.")
		} else {
			d.item = item
		}
	} else if strings.TrimSpace(in.EquipamentoDescricao) == "" {
		ev.Add("equipamento", "Informe o item do inventário ou a descrição da máquina externa.")
	}

	if in.SetorID != nil {
		if _, err := s.setorRepo.BuscarPorID(*in.SetorID); err != nil {
			ev.Add("setor_id", "Departamento não encontrado.")
		}
	}
	if in.SolicitanteID != nil {
		serv, err := s.servRepo.BuscarPorID(*in.SolicitanteID)
		if err != nil {
			ev.Add("solicitante_id", "Servidor solicitante não encontrado.")
		} else {
			d.solicitanteNome = serv.Nome
		}
	}
	if in.TecnicoID != nil {
		if _, err := s.usuarioRepo.BuscarPorID(*in.TecnicoID); err != nil {
			ev.Add("tecnico_id", "Técnico responsável não encontrado.")
		}
	}

	prio := models.PrioridadeOS(strings.TrimSpace(in.Prioridade))
	if prio == "" {
		prio = models.PrioridadeNormal
	}
	if !models.PrioridadeOSValida(prio) {
		ev.Add("prioridade", "Prioridade inválida.")
	}
	d.prioridade = prio

	if ev.TemErros() {
		return nil, ev
	}
	return d, nil
}

// aplicarEquipamento preenche os campos de equipamento (e snapshots) na OS.
func aplicarEquipamento(os *models.OrdemServico, in EntradaOS, item *models.Item) {
	if in.ItemID != 0 {
		id := in.ItemID
		os.ItemID = &id
	} else {
		os.ItemID = nil
	}
	os.EquipamentoIdentificacao = strings.TrimSpace(in.EquipamentoIdentificacao)
	if item != nil {
		os.EquipamentoDescricao = ""
		os.EquipamentoSnapshot = item.Descricao
		os.PatrimonioSnapshot = derefStr(item.NumeroPatrimonio)
	} else {
		os.EquipamentoDescricao = strings.TrimSpace(in.EquipamentoDescricao)
		os.EquipamentoSnapshot = strings.TrimSpace(in.EquipamentoDescricao)
		os.PatrimonioSnapshot = ""
	}
}

// Criar abre uma nova OS, semeia os passos padrão e gera a numeração
// sequencial em transação (evita números duplicados).
func (s *OrdemServicoService) Criar(in EntradaOS) (*models.OrdemServico, error) {
	if in.AbertoPorID == 0 {
		return nil, ErrNaoAutorizado
	}
	d, err := s.validar(in)
	if err != nil {
		return nil, err
	}

	os := &models.OrdemServico{
		SetorID:                 in.SetorID,
		SolicitanteID:           in.SolicitanteID,
		SolicitanteNomeSnapshot: d.solicitanteNome,
		DefeitoRelatado:         strings.TrimSpace(in.DefeitoRelatado),
		Diagnostico:             strings.TrimSpace(in.Diagnostico),
		SolucaoAplicada:         strings.TrimSpace(in.SolucaoAplicada),
		Prioridade:              d.prioridade,
		Status:                  models.OSAberta,
		TecnicoID:               in.TecnicoID,
		AbertoPorID:             in.AbertoPorID,
		DataAbertura:            time.Now().UTC(),
		Passos:                  passosPadrao(),
	}
	aplicarEquipamento(os, in, d.item)

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		numero, err := s.repo.ProximoNumero(tx, time.Now().UTC().Year())
		if err != nil {
			return err
		}
		os.Numero = numero
		return s.repo.CriarComTx(tx, os)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(os.ID)
}

// passosPadrao materializa o template em registros de checklist.
func passosPadrao() []models.OrdemServicoPasso {
	passos := make([]models.OrdemServicoPasso, 0, len(models.PassosPadraoManutencao))
	for i, desc := range models.PassosPadraoManutencao {
		passos = append(passos, models.OrdemServicoPasso{
			Ordem:     i + 1,
			Descricao: desc,
		})
	}
	return passos
}

// Atualizar edita os dados da OS (não altera número, abertura, status nem os
// passos — estes têm endpoints próprios).
func (s *OrdemServicoService) Atualizar(id uint, in EntradaOS) (*models.OrdemServico, error) {
	os, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	d, err := s.validar(in)
	if err != nil {
		return nil, err
	}

	os.SetorID = in.SetorID
	os.SolicitanteID = in.SolicitanteID
	os.SolicitanteNomeSnapshot = d.solicitanteNome
	os.DefeitoRelatado = strings.TrimSpace(in.DefeitoRelatado)
	os.Diagnostico = strings.TrimSpace(in.Diagnostico)
	os.SolucaoAplicada = strings.TrimSpace(in.SolucaoAplicada)
	os.Prioridade = d.prioridade
	os.TecnicoID = in.TecnicoID
	aplicarEquipamento(os, in, d.item)

	limparAssociacoes(os)
	if err := s.repo.Atualizar(os); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

// DefinirStatus altera o status da OS. Ao concluir, carimba a data de
// conclusão; ao reabrir (sair de concluída), limpa-a.
func (s *OrdemServicoService) DefinirStatus(id uint, status models.StatusOS) (*models.OrdemServico, error) {
	if !models.StatusOSValido(status) {
		ev := NovoErroValidacao()
		ev.Add("status", "Status inválido.")
		return nil, ev
	}
	os, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}

	os.Status = status
	if status == models.OSConcluida {
		agora := time.Now().UTC()
		os.DataConclusao = &agora
	} else {
		os.DataConclusao = nil
	}

	limparAssociacoes(os)
	if err := s.repo.Atualizar(os); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

// PassoEntrada é um item do checklist recebido do cliente.
type PassoEntrada struct {
	Descricao  string
	Concluido  bool
	Observacao string
}

// SalvarPassos substitui todo o checklist da OS pelo conjunto informado.
func (s *OrdemServicoService) SalvarPassos(id uint, entradas []PassoEntrada) (*models.OrdemServico, error) {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return nil, traduzErroRepo(err)
	}

	passos := make([]models.OrdemServicoPasso, 0, len(entradas))
	ordem := 1
	for _, e := range entradas {
		desc := strings.TrimSpace(e.Descricao)
		if desc == "" {
			continue
		}
		passos = append(passos, models.OrdemServicoPasso{
			Ordem:      ordem,
			Descricao:  desc,
			Concluido:  e.Concluido,
			Observacao: strings.TrimSpace(e.Observacao),
		})
		ordem++
	}

	if err := s.repo.SubstituirPassos(id, passos); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

func (s *OrdemServicoService) BuscarPorID(id uint) (*models.OrdemServico, error) {
	os, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return os, nil
}

func (s *OrdemServicoService) Listar(f repositories.FiltroOrdemServico) ([]models.OrdemServico, int64, error) {
	return s.repo.Listar(f)
}

// Excluir remove uma OS aberta por engano (admin). Só permite quando a OS ainda
// não teve andamento (status aberta ou cancelada), preservando o histórico de
// atendimentos efetivos.
func (s *OrdemServicoService) Excluir(id uint) error {
	os, err := s.repo.BuscarPorID(id)
	if err != nil {
		return traduzErroRepo(err)
	}
	if os.Status != models.OSAberta && os.Status != models.OSCancelada {
		return fmt.Errorf("%w: não é possível excluir uma OS em andamento ou concluída; cancele-a antes", ErrRegraNegocio)
	}
	return s.repo.Remover(id)
}

// GerarPDF produz o documento da OS. passosExtra são acrescentados ao final do
// checklist apenas no documento (não são persistidos).
func (s *OrdemServicoService) GerarPDF(id uint, passosExtra []string) ([]byte, *models.OrdemServico, error) {
	os, err := s.BuscarPorID(id)
	if err != nil {
		return nil, nil, err
	}
	pdf, err := PDFOrdemServico(os, passosExtra, s.cfg)
	if err != nil {
		return nil, nil, err
	}
	return pdf, os, nil
}

// limparAssociacoes zera os ponteiros de associação antes de um Save, para que
// o GORM atualize apenas as colunas da própria OS (sem upsert em Item, Setor,
// Servidor, Usuário ou nos Passos).
func limparAssociacoes(os *models.OrdemServico) {
	os.Item = nil
	os.Setor = nil
	os.Solicitante = nil
	os.Tecnico = nil
	os.AbertoPor = nil
	os.Passos = nil
}
