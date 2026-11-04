package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// Tipos de relatório suportados pelo construtor.
const (
	RelInventario     = "inventario"
	RelBaixados       = "baixados"
	RelEstoqueBaixo   = "estoque-baixo"
	RelTransferencias = "transferencias"
	RelEmprestimos    = "emprestimos"
	RelEntradas       = "entradas"
	RelMovimentacoes  = "movimentacoes"
)

// DefinicaoRelatorio descreve um tipo de relatório para o catálogo (usado pelo
// frontend para montar o construtor: quais filtros exibir por tipo).
type DefinicaoRelatorio struct {
	Tipo           string `json:"tipo"`
	Rotulo         string `json:"rotulo"`
	Descricao      string `json:"descricao"`
	EhEvento       bool   `json:"eh_evento"` // true = relatório de movimentações; false = de itens
	UsaPeriodo     bool   `json:"usa_periodo"`
	UsaSetor       bool   `json:"usa_setor"`       // filtro por secretaria/departamento
	UsaResponsavel bool   `json:"usa_responsavel"` // só itens
	UsaCategoria   bool   `json:"usa_categoria"`
	UsaEstado      bool   `json:"usa_estado"` // só itens
	PermitePDF     bool   `json:"permite_pdf"`
}

// catalogoRelatorios é a fonte única de verdade dos tipos e seus filtros.
var catalogoRelatorios = []DefinicaoRelatorio{
	{
		Tipo: RelInventario, Rotulo: "Inventário", Descricao: "Itens cadastrados, com filtros por local, categoria, estado e responsável.",
		EhEvento: false, UsaSetor: true, UsaResponsavel: true, UsaCategoria: true, UsaEstado: true, PermitePDF: true,
	},
	{
		Tipo: RelBaixados, Rotulo: "Baixados / descartados", Descricao: "Itens com baixa patrimonial, com motivo e data.",
		EhEvento: false, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
	{
		Tipo: RelEstoqueBaixo, Rotulo: "Estoque baixo", Descricao: "Materiais de consumo abaixo do estoque mínimo.",
		EhEvento: false, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
	{
		Tipo: RelTransferencias, Rotulo: "Transferências", Descricao: "Itens transferidos entre locais, por período.",
		EhEvento: true, UsaPeriodo: true, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
	{
		Tipo: RelEmprestimos, Rotulo: "Empréstimos", Descricao: "Equipamentos emprestados, por período.",
		EhEvento: true, UsaPeriodo: true, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
	{
		Tipo: RelEntradas, Rotulo: "Entradas / aquisições", Descricao: "Compras, doações e devoluções, por período.",
		EhEvento: true, UsaPeriodo: true, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
	{
		Tipo: RelMovimentacoes, Rotulo: "Movimentações (todas)", Descricao: "Histórico completo de movimentações, por período.",
		EhEvento: true, UsaPeriodo: true, UsaSetor: true, UsaCategoria: true, PermitePDF: true,
	},
}

func definicaoRelatorio(tipo string) (DefinicaoRelatorio, bool) {
	for _, d := range catalogoRelatorios {
		if d.Tipo == tipo {
			return d, true
		}
	}
	return DefinicaoRelatorio{}, false
}

// RelatorioService produz consultas agregadas para relatórios e exportações.
type RelatorioService struct {
	itemRepo  repositories.ItemRepository
	movRepo   repositories.MovimentacaoRepository
	setorRepo repositories.SetorRepository
	cfg       *config.Config
}

func NewRelatorioService(itemRepo repositories.ItemRepository, movRepo repositories.MovimentacaoRepository, setorRepo repositories.SetorRepository, cfg *config.Config) *RelatorioService {
	return &RelatorioService{itemRepo: itemRepo, movRepo: movRepo, setorRepo: setorRepo, cfg: cfg}
}

// Config expõe a configuração para geração de PDFs parametrizados.
func (s *RelatorioService) Config() *config.Config { return s.cfg }

// Catalogo devolve os tipos de relatório disponíveis e seus filtros.
func (s *RelatorioService) Catalogo() []DefinicaoRelatorio { return catalogoRelatorios }

// ParametrosRelatorio reúne os filtros do construtor de relatórios.
type ParametrosRelatorio struct {
	Tipo          string
	SecretariaID  *uint // filtro hierárquico: setor + descendentes
	SetorID       *uint // filtro de setor exato (departamento)
	ResponsavelID *uint
	CategoriaID   *uint
	Estado        *models.EstadoConservacao
	TipoMov       *models.TipoMovimentacao // só p/ "movimentacoes": filtra um tipo específico
	DataInicio    *time.Time
	DataFim       *time.Time
}

// ResultadoRelatorio carrega o resultado tipado (itens OU movimentações).
type ResultadoRelatorio struct {
	Tipo          string                `json:"tipo"`
	Titulo        string                `json:"titulo"`
	EhEvento      bool                  `json:"eh_evento"`
	Itens         []models.Item         `json:"itens,omitempty"`
	Movimentacoes []models.Movimentacao `json:"movimentacoes,omitempty"`
}

// Gerar executa o relatório solicitado aplicando os filtros compatíveis.
func (s *RelatorioService) Gerar(p ParametrosRelatorio) (*ResultadoRelatorio, error) {
	def, ok := definicaoRelatorio(p.Tipo)
	if !ok {
		return nil, fmt.Errorf("%w: tipo de relatório desconhecido", ErrValidacao)
	}

	setorIDs, err := s.resolverSetores(p)
	if err != nil {
		return nil, err
	}

	res := &ResultadoRelatorio{Tipo: def.Tipo, Titulo: def.Rotulo, EhEvento: def.EhEvento}

	if def.EhEvento {
		f := repositories.FiltroMovimentacao{
			DataInicio:  p.DataInicio,
			DataFim:     p.DataFim,
			CategoriaID: p.CategoriaID,
			SetorIDs:    setorIDs,
		}
		switch p.Tipo {
		case RelTransferencias:
			f.Tipos = []models.TipoMovimentacao{models.MovSaidaTransferencia}
		case RelEmprestimos:
			f.Tipos = []models.TipoMovimentacao{models.MovSaidaEmprestimo}
		case RelEntradas:
			f.Tipos = []models.TipoMovimentacao{
				models.MovEntradaCompra, models.MovEntradaDoacao, models.MovEntradaDevolucao,
			}
		case RelMovimentacoes:
			if p.TipoMov != nil {
				f.Tipo = p.TipoMov
			}
		}
		movs, _, err := s.movRepo.Listar(f)
		if err != nil {
			return nil, err
		}
		res.Movimentacoes = movs
		return res, nil
	}

	f := repositories.FiltroItem{
		SetorIDs:      setorIDs,
		ResponsavelID: p.ResponsavelID,
		CategoriaID:   p.CategoriaID,
		Estado:        p.Estado,
	}
	switch p.Tipo {
	case RelBaixados:
		baixados := true
		f.SomenteBaixados = &baixados
	case RelEstoqueBaixo:
		f.AbaixoDoMinimo = true
		ativos := false
		f.SomenteBaixados = &ativos
	}
	itens, _, err := s.itemRepo.Listar(f)
	if err != nil {
		return nil, err
	}
	res.Itens = itens
	return res, nil
}

// resolverSetores converte secretaria (hierárquica) + departamento (exato) em
// uma lista de IDs de setor. Retorna nil quando nenhum filtro de local foi dado.
func (s *RelatorioService) resolverSetores(p ParametrosRelatorio) ([]uint, error) {
	conjunto := map[uint]struct{}{}
	if p.SecretariaID != nil {
		ids, err := s.descendentes(*p.SecretariaID)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			conjunto[id] = struct{}{}
		}
	}
	if p.SetorID != nil {
		conjunto[*p.SetorID] = struct{}{}
	}
	if len(conjunto) == 0 {
		return nil, nil
	}
	out := make([]uint, 0, len(conjunto))
	for id := range conjunto {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// descendentes retorna o setor raiz e todos os seus descendentes (BFS na árvore
// de setores carregada em memória — portável entre SQLite e PostgreSQL).
func (s *RelatorioService) descendentes(raiz uint) ([]uint, error) {
	setores, err := s.setorRepo.Listar()
	if err != nil {
		return nil, err
	}
	filhos := map[uint][]uint{}
	for _, st := range setores {
		if st.PaiID != nil {
			filhos[*st.PaiID] = append(filhos[*st.PaiID], st.ID)
		}
	}
	var out []uint
	fila := []uint{raiz}
	for len(fila) > 0 {
		atual := fila[0]
		fila = fila[1:]
		out = append(out, atual)
		fila = append(fila, filhos[atual]...)
	}
	return out, nil
}
