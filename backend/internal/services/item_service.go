package services

import (
	"errors"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

type ItemService struct {
	repo          repositories.ItemRepository
	categoriaRepo repositories.CategoriaRepository
	setorRepo     repositories.SetorRepository
	servidorRepo  repositories.ServidorRepository
	movRepo       repositories.MovimentacaoRepository
	termoRepo     repositories.TermoRepository
}

func NewItemService(
	repo repositories.ItemRepository,
	categoriaRepo repositories.CategoriaRepository,
	setorRepo repositories.SetorRepository,
	servidorRepo repositories.ServidorRepository,
	movRepo repositories.MovimentacaoRepository,
	termoRepo repositories.TermoRepository,
) *ItemService {
	return &ItemService{
		repo:          repo,
		categoriaRepo: categoriaRepo,
		setorRepo:     setorRepo,
		servidorRepo:  servidorRepo,
		movRepo:       movRepo,
		termoRepo:     termoRepo,
	}
}

type EntradaItem struct {
	Descricao         string
	CategoriaID       uint
	NumeroPatrimonio  *string
	NumeroSerie       *string
	Marca             string
	Modelo            string
	EstadoConservacao models.EstadoConservacao
	Quantidade        int
	EstoqueMinimo     int
	SetorID           *uint
	ResponsavelID     *uint
	DataAquisicao     *time.Time
	Valor             *float64
}

func (s *ItemService) Criar(in EntradaItem) (*models.Item, error) {
	categoria, err := s.validar(in)
	if err != nil {
		return nil, err
	}

	item := &models.Item{
		Descricao:         strings.TrimSpace(in.Descricao),
		CategoriaID:       in.CategoriaID,
		NumeroPatrimonio:  normalizarPtr(in.NumeroPatrimonio),
		NumeroSerie:       normalizarPtr(in.NumeroSerie),
		Marca:             strings.TrimSpace(in.Marca),
		Modelo:            strings.TrimSpace(in.Modelo),
		EstadoConservacao: in.EstadoConservacao,
		Quantidade:        in.Quantidade,
		EstoqueMinimo:     in.EstoqueMinimo,
		SetorID:           in.SetorID,
		ResponsavelID:     in.ResponsavelID,
		DataAquisicao:     in.DataAquisicao,
		Valor:             in.Valor,
	}
	// Patrimoniados têm controle unitário: padroniza quantidade mínima 1.
	if !categoria.Consumivel && item.Quantidade == 0 {
		item.Quantidade = 1
	}

	if err := s.repo.Criar(item); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(item.ID)
}

func (s *ItemService) Atualizar(id uint, in EntradaItem) (*models.Item, error) {
	item, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if item.Baixado {
		return nil, ErrItemBaixado
	}
	if _, err := s.validar(in); err != nil {
		return nil, err
	}

	item.Descricao = strings.TrimSpace(in.Descricao)
	item.CategoriaID = in.CategoriaID
	item.NumeroPatrimonio = normalizarPtr(in.NumeroPatrimonio)
	item.NumeroSerie = normalizarPtr(in.NumeroSerie)
	item.Marca = strings.TrimSpace(in.Marca)
	item.Modelo = strings.TrimSpace(in.Modelo)
	item.EstadoConservacao = in.EstadoConservacao
	item.EstoqueMinimo = in.EstoqueMinimo
	item.SetorID = in.SetorID
	item.ResponsavelID = in.ResponsavelID
	item.DataAquisicao = in.DataAquisicao
	item.Valor = in.Valor
	// Observação: a Quantidade NÃO é editada aqui — só muda via movimentações
	// (regra de negócio: estoque é sempre recalculado por movimentação).

	if err := s.repo.Atualizar(item); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(item.ID)
}

// Excluir remove (soft delete) um item cadastrado por engano. Uso restrito a
// administradores (controlado na rota). Por segurança e rastreabilidade, recusa
// a exclusão se o item já possui histórico — movimentações ou termos emitidos —
// orientando o uso da BAIXA patrimonial nesses casos.
func (s *ItemService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}

	nMov, err := s.movRepo.ContarPorItem(id)
	if err != nil {
		return err
	}
	nTermo, err := s.termoRepo.ContarPorItem(id)
	if err != nil {
		return err
	}
	if nMov > 0 || nTermo > 0 {
		return ErrItemComHistorico
	}

	return s.repo.Remover(id)
}

func (s *ItemService) BuscarPorID(id uint) (*models.Item, error) {
	item, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	return item, nil
}

func (s *ItemService) Listar(f repositories.FiltroItem) ([]models.Item, int64, error) {
	return s.repo.Listar(f)
}

func (s *ItemService) ListarAbaixoDoMinimo() ([]models.Item, error) {
	return s.repo.ListarAbaixoDoMinimo()
}

// validar checa consistência e existência de FKs; devolve a categoria
// carregada para decisões dependentes (consumível vs patrimoniado).
func (s *ItemService) validar(in EntradaItem) (*models.Categoria, error) {
	ev := NovoErroValidacao()

	if strings.TrimSpace(in.Descricao) == "" {
		ev.Add("descricao", "Informe a descrição do item.")
	}
	if in.EstadoConservacao == "" {
		in.EstadoConservacao = models.EstadoBom
	} else if !models.EstadoConservacaoValido(in.EstadoConservacao) {
		ev.Add("estado_conservacao", "Estado de conservação inválido.")
	}
	if in.Quantidade < 0 {
		ev.Add("quantidade", "A quantidade não pode ser negativa.")
	}
	if in.EstoqueMinimo < 0 {
		ev.Add("estoque_minimo", "O estoque mínimo não pode ser negativo.")
	}
	if in.Valor != nil && *in.Valor < 0 {
		ev.Add("valor", "O valor não pode ser negativo.")
	}

	var categoria *models.Categoria
	if in.CategoriaID == 0 {
		ev.Add("categoria_id", "Selecione uma categoria.")
	} else {
		c, err := s.categoriaRepo.BuscarPorID(in.CategoriaID)
		if err != nil {
			if errors.Is(err, repositories.ErrNaoEncontrado) {
				ev.Add("categoria_id", "Categoria informada não existe.")
			} else {
				return nil, err
			}
		} else {
			categoria = c
		}
	}

	if in.SetorID != nil {
		if _, err := s.setorRepo.BuscarPorID(*in.SetorID); err != nil {
			if errors.Is(err, repositories.ErrNaoEncontrado) {
				ev.Add("setor_id", "Setor informado não existe.")
			} else {
				return nil, err
			}
		}
	}
	if in.ResponsavelID != nil {
		if _, err := s.servidorRepo.BuscarPorID(*in.ResponsavelID); err != nil {
			if errors.Is(err, repositories.ErrNaoEncontrado) {
				ev.Add("responsavel_id", "Servidor responsável informado não existe.")
			} else {
				return nil, err
			}
		}
	}

	// Regra: item patrimoniado deve ter número de patrimônio.
	if categoria != nil && !categoria.Consumivel {
		if normalizarPtr(in.NumeroPatrimonio) == nil {
			ev.Add("numero_patrimonio", "Itens patrimoniados exigem número de patrimônio.")
		}
	}

	if ev.TemErros() {
		return nil, ev
	}
	return categoria, nil
}

func normalizarPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}
