package services

import (
	"errors"
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// InventarioService recebe as coletas automáticas de configuração das máquinas
// (script PowerShell nas estações) e faz o UPSERT do item no inventário.
//
// Chave natural: número de série (BIOS). Reexecutar a coleta atualiza o mesmo
// item em vez de duplicar. Campos controlados pelo operador (setor, responsável,
// patrimônio, estado) NUNCA são sobrescritos pela coleta.
type InventarioService struct {
	itemRepo      repositories.ItemRepository
	categoriaRepo repositories.CategoriaRepository
	campoRepo     repositories.CampoCategoriaRepository
}

func NewInventarioService(
	itemRepo repositories.ItemRepository,
	categoriaRepo repositories.CategoriaRepository,
	campoRepo repositories.CampoCategoriaRepository,
) *InventarioService {
	return &InventarioService{
		itemRepo:      itemRepo,
		categoriaRepo: categoriaRepo,
		campoRepo:     campoRepo,
	}
}

// EspecificacaoColeta é um par rótulo/valor de configuração coletada (ex.:
// "Processador" -> "Intel Core i5-10400").
type EspecificacaoColeta struct {
	Nome  string
	Valor string
}

// EntradaInventario são os dados enviados pelo coletor de uma máquina.
type EntradaInventario struct {
	Hostname       string
	NumeroSerie    string
	Fabricante     string
	Modelo         string
	Categoria      string // "Desktop", "Notebook"... (casada por nome; criada se faltar)
	Especificacoes []EspecificacaoColeta
}

// categoriaPadraoInventario é usada quando a coleta não informa a categoria.
const categoriaPadraoInventario = "Computadores"

// Coletar faz o upsert do item a partir da coleta. Devolve o item e se foi
// criado (true) ou atualizado (false).
func (s *InventarioService) Coletar(in EntradaInventario) (*models.Item, bool, error) {
	ev := NovoErroValidacao()
	serie := strings.TrimSpace(in.NumeroSerie)
	hostname := strings.TrimSpace(in.Hostname)
	if serie == "" && hostname == "" {
		ev.Add("numero_serie", "Informe o número de série ou o hostname da máquina.")
		return nil, false, ev
	}

	categoria, err := s.resolverCategoria(in.Categoria)
	if err != nil {
		return nil, false, err
	}

	descricao := hostname
	if descricao == "" {
		descricao = strings.TrimSpace(in.Fabricante + " " + in.Modelo)
	}
	if descricao == "" {
		descricao = "Computador"
	}

	// Procura item existente pela série (chave natural).
	var item *models.Item
	if serie != "" {
		achado, err := s.itemRepo.BuscarPorNumeroSerie(serie)
		if err != nil && !errors.Is(err, repositories.ErrNaoEncontrado) {
			return nil, false, err
		}
		item = achado
	}

	criado := false
	if item == nil {
		novo := &models.Item{
			Descricao:         descricao,
			CategoriaID:       categoria.ID,
			Marca:             strings.TrimSpace(in.Fabricante),
			Modelo:            strings.TrimSpace(in.Modelo),
			EstadoConservacao: models.EstadoBom,
			Quantidade:        1,
		}
		if serie != "" {
			novo.NumeroSerie = &serie
		}
		if err := s.itemRepo.Criar(novo); err != nil {
			return nil, false, err
		}
		item = novo
		criado = true
	} else {
		// Atualiza apenas o que a coleta conhece; preserva setor/responsável/
		// patrimônio/estado (controle do operador).
		item.Descricao = descricao
		item.Marca = strings.TrimSpace(in.Fabricante)
		item.Modelo = strings.TrimSpace(in.Modelo)
		if item.CategoriaID == 0 {
			item.CategoriaID = categoria.ID
		}
		if err := s.itemRepo.Atualizar(item); err != nil {
			return nil, false, err
		}
	}

	// Especificações viram atributos customizados da categoria do item. Upsert
	// por campo: não apaga atributos preenchidos à mão.
	for _, esp := range in.Especificacoes {
		nome := strings.TrimSpace(esp.Nome)
		valor := strings.TrimSpace(esp.Valor)
		if nome == "" || valor == "" {
			continue
		}
		campo, err := s.resolverCampo(item.CategoriaID, nome)
		if err != nil {
			continue // campo problemático não derruba a coleta inteira
		}
		if err := s.campoRepo.UpsertAtributo(item.ID, campo.ID, valor); err != nil {
			continue
		}
	}

	atualizado, err := s.itemRepo.BuscarPorID(item.ID)
	if err != nil {
		return nil, criado, err
	}
	return atualizado, criado, nil
}

// resolverCategoria casa a categoria pelo nome (ignora caixa/acentos); cria uma
// patrimoniada se não existir.
func (s *InventarioService) resolverCategoria(nome string) (*models.Categoria, error) {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		nome = categoriaPadraoInventario
	}
	lista, err := s.categoriaRepo.Listar()
	if err != nil {
		return nil, err
	}
	alvo := normalizarChave(nome)
	for i := range lista {
		if normalizarChave(lista[i].Nome) == alvo {
			return &lista[i], nil
		}
	}
	nova := &models.Categoria{Nome: nome, Consumivel: false}
	if err := s.categoriaRepo.Criar(nova); err != nil {
		return nil, err
	}
	return nova, nil
}

// resolverCampo casa o campo da categoria pela chave gerada do nome; cria um
// campo de texto se não existir.
func (s *InventarioService) resolverCampo(categoriaID uint, nome string) (*models.CampoCategoria, error) {
	campos, err := s.campoRepo.ListarPorCategoria(categoriaID)
	if err != nil {
		return nil, err
	}
	chave := gerarChaveCampo(nome)
	for i := range campos {
		if campos[i].Chave == chave {
			return &campos[i], nil
		}
	}
	novo := &models.CampoCategoria{
		CategoriaID: categoriaID,
		Chave:       chave,
		Nome:        nome,
		Tipo:        models.CampoTexto,
		Ordem:       len(campos),
	}
	if err := s.campoRepo.Criar(novo); err != nil {
		return nil, err
	}
	return novo, nil
}
