package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	// osRepo é opcional (injetado via setter) para a linha do tempo do item.
	osRepo repositories.OrdemServicoRepository
	// campoRepo é opcional (injetado via setter) para os campos customizados.
	campoRepo repositories.CampoCategoriaRepository
}

// DefinirOrdemServicoRepo injeta o repositório de OS usado pela linha do tempo
// (dependência opcional, evita ciclo e mantém o construtor estável).
func (s *ItemService) DefinirOrdemServicoRepo(r repositories.OrdemServicoRepository) {
	s.osRepo = r
}

// DefinirCampoRepo injeta o repositório de campos customizados (atributos).
func (s *ItemService) DefinirCampoRepo(r repositories.CampoCategoriaRepository) {
	s.campoRepo = r
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
	GarantiaAte       *time.Time
	Reservavel        bool
	// CadastradoPorID é o usuário autenticado que registrou o item. Vem do
	// contexto da requisição (nunca do corpo), por segurança de autoria.
	CadastradoPorID *uint
	// Atributos: valores dos campos customizados da categoria. Quando nil, os
	// atributos do item não são alterados; quando não-nil (mesmo vazio),
	// substituem os atuais.
	Atributos []EntradaAtributo
}

// EntradaAtributo é o valor de um campo customizado enviado no cadastro/edição.
type EntradaAtributo struct {
	CampoID uint
	Valor   string
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
		GarantiaAte:       in.GarantiaAte,
		Reservavel:        in.Reservavel,
		CadastradoPorID:   in.CadastradoPorID,
	}
	// Bens permanentes têm controle unitário, inclusive antes do tombamento.
	if !categoria.Consumivel && item.Quantidade == 0 {
		item.Quantidade = 1
	}

	if err := s.repo.Criar(item); err != nil {
		return nil, err
	}
	if err := s.salvarAtributos(item.ID, in.Atributos); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(item.ID)
}

// salvarAtributos grava os valores dos campos customizados do item. Entradas
// com valor vazio são descartadas (não guarda atributo em branco). Não faz nada
// se o repositório de campos não foi injetado ou se atributos for nil.
func (s *ItemService) salvarAtributos(itemID uint, entradas []EntradaAtributo) error {
	if s.campoRepo == nil || entradas == nil {
		return nil
	}
	var atributos []models.ItemAtributo
	for _, e := range entradas {
		valor := strings.TrimSpace(e.Valor)
		if e.CampoID == 0 || valor == "" {
			continue
		}
		atributos = append(atributos, models.ItemAtributo{
			CampoID: e.CampoID,
			Valor:   valor,
		})
	}
	return s.campoRepo.SubstituirAtributosDoItem(itemID, atributos)
}

// ResultadoLote identifica um recebimento e as unidades criadas nele.
type ResultadoLote struct {
	LoteCodigo string `json:"lote_codigo"`
	Quantidade int    `json:"quantidade"`
	IDs        []uint `json:"ids"`
}

// CriarLote cadastra unidades permanentes individualmente em uma transação.
func (s *ItemService) CriarLote(in EntradaItem, unidades int) (*ResultadoLote, error) {
	ev := NovoErroValidacao()
	if unidades < 1 || unidades > 200 {
		ev.Add("unidades", "Informe de 1 a 200 unidades por lote.")
	}
	if in.NumeroPatrimonio != nil || in.NumeroSerie != nil {
		ev.Add("unidades", "Patrimônio e série devem ser preenchidos em cada unidade após o cadastro.")
	}
	if ev.TemErros() {
		return nil, ev
	}
	in.Quantidade = 1
	in.EstoqueMinimo = 0
	categoria, err := s.validar(in)
	if err != nil {
		return nil, err
	}
	if categoria.Consumivel {
		ev.Add("categoria_id", "O cadastro em lote individual é para bens permanentes.")
		return nil, ev
	}
	aleatorio := make([]byte, 8)
	if _, err := rand.Read(aleatorio); err != nil {
		return nil, err
	}
	lote := fmt.Sprintf("REC-%s-%s", time.Now().UTC().Format("20060102"), hex.EncodeToString(aleatorio))
	itens := make([]models.Item, unidades)
	for i := range itens {
		itens[i] = models.Item{
			Descricao: strings.TrimSpace(in.Descricao), CategoriaID: in.CategoriaID,
			LoteCodigo: lote, Marca: strings.TrimSpace(in.Marca), Modelo: strings.TrimSpace(in.Modelo),
			EstadoConservacao: in.EstadoConservacao, Quantidade: 1, SetorID: in.SetorID,
			DataAquisicao: in.DataAquisicao, Valor: in.Valor,
			CadastradoPorID: in.CadastradoPorID,
		}
	}
	if err := s.repo.CriarLote(itens); err != nil {
		return nil, err
	}
	ids := make([]uint, len(itens))
	for i := range itens {
		ids[i] = itens[i].ID
	}
	return &ResultadoLote{LoteCodigo: lote, Quantidade: unidades, IDs: ids}, nil
}

func (s *ItemService) Atualizar(id uint, in EntradaItem) (*models.Item, error) {
	item, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if item.Baixado {
		return nil, ErrItemBaixado
	}
	categoria, err := s.validar(in)
	if err != nil {
		return nil, err
	}
	if !categoria.Consumivel && (!mesmoID(item.SetorID, in.SetorID) || !mesmoID(item.ResponsavelID, in.ResponsavelID)) {
		ev := NovoErroValidacao()
		ev.Add("setor_id", "Para mudar o local ou responsável de um bem permanente, registre uma movimentação.")
		return nil, ev
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
	// Se a data de garantia mudou, rearma o alerta de vencimento.
	if !mesmaData(item.GarantiaAte, in.GarantiaAte) {
		item.GarantiaAte = in.GarantiaAte
		item.AlertaGarantiaEnviado = false
	}
	item.Reservavel = in.Reservavel
	// Observação: a Quantidade NÃO é editada aqui — só muda via movimentações
	// (regra de negócio: estoque é sempre recalculado por movimentação).

	if err := s.repo.Atualizar(item); err != nil {
		return nil, err
	}
	if err := s.salvarAtributos(item.ID, in.Atributos); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(item.ID)
}

func mesmoID(a, b *uint) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func mesmaData(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

// DefinirReservavel marca/desmarca o item como reservável (pool de reservas).
func (s *ItemService) DefinirReservavel(id uint, reservavel bool) (*models.Item, error) {
	item, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if item.Baixado {
		return nil, ErrItemBaixado
	}
	item.Reservavel = reservavel
	if err := s.repo.Atualizar(item); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
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

	// Bens permanentes entram antes do tombamento, sempre como unidades.
	if categoria != nil && !categoria.Consumivel && in.Quantidade > 1 {
		ev.Add("quantidade", "Cadastre cada bem permanente separadamente ou use o cadastro em lote.")
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
