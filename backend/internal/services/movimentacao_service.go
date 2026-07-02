package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"gorm.io/gorm"
)

// MovimentacaoService centraliza TODAS as regras críticas de estoque:
//   - não permitir saída maior que o estoque disponível;
//   - recalcular o saldo do item após cada movimentação;
//   - registrar histórico imutável (com saldo resultante);
//   - sinalizar alerta quando o estoque ficar abaixo do mínimo;
//   - baixa patrimonial: registra movimentação + marca o item como baixado
//     (estado inservível), mantendo-o visível (rastreabilidade total).
//
// Toda a operação roda em transação para garantir atomicidade entre o
// registro da movimentação e o recálculo do saldo do item.
type MovimentacaoService struct {
	repo      repositories.MovimentacaoRepository
	itemRepo  repositories.ItemRepository
	setorRepo repositories.SetorRepository
	servRepo  repositories.ServidorRepository
}

func NewMovimentacaoService(
	repo repositories.MovimentacaoRepository,
	itemRepo repositories.ItemRepository,
	setorRepo repositories.SetorRepository,
	servRepo repositories.ServidorRepository,
) *MovimentacaoService {
	return &MovimentacaoService{repo: repo, itemRepo: itemRepo, setorRepo: setorRepo, servRepo: servRepo}
}

// EntradaMovimentacao representa os dados para registrar uma movimentação.
type EntradaMovimentacao struct {
	ItemID          uint
	Tipo            models.TipoMovimentacao
	Quantidade      int
	SetorOrigemID   *uint
	SetorDestinoID  *uint
	ServidorID      *uint
	Observacao      string
	MotivoBaixa     string // usado quando Tipo == saida_descarte
	RegistradoPorID uint
	DataEvento      *time.Time
}

// ResultadoMovimentacao devolve a movimentação criada, o item atualizado e
// se o item entrou em alerta de estoque baixo após a operação.
type ResultadoMovimentacao struct {
	Movimentacao  *models.Movimentacao
	Item          *models.Item
	AlertaEstoque bool
}

func (s *MovimentacaoService) Registrar(in EntradaMovimentacao) (*ResultadoMovimentacao, error) {
	if err := s.validarEntrada(in); err != nil {
		return nil, err
	}

	var resultado ResultadoMovimentacao

	// Transação: leitura do item, validação de saldo, escrita do histórico e
	// atualização do saldo do item — tudo atômico.
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item models.Item
		// Bloqueio pessimista quando suportado (no SQLite é no-op, mas a
		// limitação a 1 conexão de escrita já serializa as gravações).
		if err := tx.First(&item, in.ItemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNaoEncontrado
			}
			return err
		}

		// Item já baixado não aceita novas movimentações.
		if item.Baixado {
			return ErrItemBaixado
		}

		novoSaldo, err := s.calcularNovoSaldo(&item, in.Tipo, in.Quantidade)
		if err != nil {
			return err
		}

		dataEvento := time.Now().UTC()
		if in.DataEvento != nil {
			dataEvento = in.DataEvento.UTC()
		}

		mov := &models.Movimentacao{
			ItemID:           item.ID,
			Tipo:             in.Tipo,
			Quantidade:       in.Quantidade,
			SaldoResultante:  novoSaldo,
			SetorOrigemID:    in.SetorOrigemID,
			SetorDestinoID:   in.SetorDestinoID,
			ServidorID:       in.ServidorID,
			RegistradoPorID:  in.RegistradoPorID,
			OrigemDescricao:  s.descreverSetor(tx, in.SetorOrigemID),
			DestinoDescricao: s.descreverSetor(tx, in.SetorDestinoID),
			Observacao:       strings.TrimSpace(in.Observacao),
			DataEvento:       dataEvento,
		}
		if err := s.repo.CriarComTx(tx, mov); err != nil {
			return err
		}

		// Recalcula o saldo e aplica efeitos colaterais conforme o tipo.
		item.Quantidade = novoSaldo

		// Transferência: atualiza setor/responsável atuais do item.
		if in.Tipo == models.MovSaidaTransferencia {
			if in.SetorDestinoID != nil {
				item.SetorID = in.SetorDestinoID
			}
			item.ResponsavelID = in.ServidorID
		}
		if in.Tipo == models.MovSaidaEmprestimo && in.ServidorID != nil {
			item.ResponsavelID = in.ServidorID
		}
		if in.Tipo == models.MovEntradaDevolucao {
			// Devolução retorna ao estoque: limpa responsável.
			item.ResponsavelID = nil
		}

		// Baixa patrimonial: marca o item, NÃO remove (rastreabilidade total).
		if in.Tipo.EhBaixa() {
			item.Baixado = true
			item.EstadoConservacao = models.EstadoInservivel
			item.DataBaixa = &dataEvento
			item.MotivoBaixa = strings.TrimSpace(in.MotivoBaixa)
		}

		if err := s.itemRepo.AtualizarComTx(tx, &item); err != nil {
			return err
		}

		resultado.Movimentacao = mov
		resultado.Item = &item
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Alerta de estoque baixo (só faz sentido para consumíveis / mínimo > 0).
	resultado.AlertaEstoque = resultado.Item.EstoqueAbaixoDoMinimo()

	// Recarrega com preloads para a resposta da API.
	movCompleto, _ := s.recarregarMovimentacao(resultado.Movimentacao.ID)
	if movCompleto != nil {
		resultado.Movimentacao = movCompleto
	}
	itemCompleto, _ := s.itemRepo.BuscarPorID(resultado.Item.ID)
	if itemCompleto != nil {
		resultado.Item = itemCompleto
	}
	return &resultado, nil
}

// calcularNovoSaldo aplica a regra central: entradas somam, saídas subtraem,
// e saída nunca pode exceder o estoque disponível.
func (s *MovimentacaoService) calcularNovoSaldo(item *models.Item, tipo models.TipoMovimentacao, qtd int) (int, error) {
	if tipo.EhEntrada() {
		return item.Quantidade + qtd, nil
	}
	// Saída
	if qtd > item.Quantidade {
		return 0, fmt.Errorf("%w: disponível %d, solicitado %d", ErrEstoqueInsuficiente, item.Quantidade, qtd)
	}
	return item.Quantidade - qtd, nil
}

func (s *MovimentacaoService) validarEntrada(in EntradaMovimentacao) error {
	ev := NovoErroValidacao()

	if in.ItemID == 0 {
		ev.Add("item_id", "Informe o item da movimentação.")
	}
	if !models.TipoMovimentacaoValido(in.Tipo) {
		ev.Add("tipo", "Tipo de movimentação inválido.")
	}
	if in.Quantidade <= 0 {
		ev.Add("quantidade", "A quantidade deve ser maior que zero.")
	}
	if in.RegistradoPorID == 0 {
		ev.Add("registrado_por", "Movimentação sem usuário responsável.")
	}

	// Validações específicas por tipo.
	if in.Tipo == models.MovSaidaTransferencia && in.SetorDestinoID == nil {
		ev.Add("setor_destino_id", "Transferência exige o setor de destino.")
	}
	if in.Tipo.EhBaixa() && strings.TrimSpace(in.MotivoBaixa) == "" {
		ev.Add("motivo_baixa", "Informe o motivo da baixa patrimonial.")
	}

	// Existência de FKs opcionais.
	if in.SetorOrigemID != nil {
		if _, err := s.setorRepo.BuscarPorID(*in.SetorOrigemID); err != nil {
			ev.Add("setor_origem_id", "Setor de origem não existe.")
		}
	}
	if in.SetorDestinoID != nil {
		if _, err := s.setorRepo.BuscarPorID(*in.SetorDestinoID); err != nil {
			ev.Add("setor_destino_id", "Setor de destino não existe.")
		}
	}
	if in.ServidorID != nil {
		if _, err := s.servRepo.BuscarPorID(*in.ServidorID); err != nil {
			ev.Add("servidor_id", "Servidor informado não existe.")
		}
	}

	if ev.TemErros() {
		return ev
	}
	return nil
}

func (s *MovimentacaoService) descreverSetor(tx *gorm.DB, setorID *uint) string {
	if setorID == nil {
		return ""
	}
	var st models.Setor
	if err := tx.First(&st, *setorID).Error; err != nil {
		return ""
	}
	if st.Sigla != "" {
		return fmt.Sprintf("%s (%s)", st.Nome, st.Sigla)
	}
	return st.Nome
}

func (s *MovimentacaoService) recarregarMovimentacao(id uint) (*models.Movimentacao, error) {
	return s.repo.BuscarPorID(id)
}

// Listar consulta o histórico com filtros.
func (s *MovimentacaoService) Listar(f repositories.FiltroMovimentacao) ([]models.Movimentacao, int64, error) {
	return s.repo.Listar(f)
}

// HistoricoPorItem devolve o histórico completo de um item.
func (s *MovimentacaoService) HistoricoPorItem(itemID uint) ([]models.Movimentacao, error) {
	if _, err := s.itemRepo.BuscarPorID(itemID); err != nil {
		return nil, traduzErroRepo(err)
	}
	return s.repo.ListarPorItem(itemID)
}
