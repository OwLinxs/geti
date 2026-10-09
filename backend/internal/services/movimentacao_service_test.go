package services_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupDB cria um banco SQLite em memória migrado, isolado por teste.
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("falha ao abrir banco: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Usuario{}, &models.Categoria{}, &models.Setor{},
		&models.Servidor{}, &models.Item{}, &models.Movimentacao{},
		&models.TermoResponsabilidade{},
	); err != nil {
		t.Fatalf("falha na migração: %v", err)
	}
	// Limpa para garantir isolamento (cache=shared compartilha a memória).
	db.Exec("DELETE FROM movimentacoes")
	db.Exec("DELETE FROM itens")
	db.Exec("DELETE FROM categorias")
	db.Exec("DELETE FROM setores")
	db.Exec("DELETE FROM servidores")
	db.Exec("DELETE FROM usuarios")
	return db
}

type fixtures struct {
	movSvc  *services.MovimentacaoService
	itemSvc *services.ItemService
	db      *gorm.DB
	usuario *models.Usuario
	catCons *models.Categoria
	catPatr *models.Categoria
}

func novoAmbiente(t *testing.T) *fixtures {
	t.Helper()
	db := setupDB(t)

	itemRepo := repositories.NewItemRepository(db)
	catRepo := repositories.NewCategoriaRepository(db)
	setorRepo := repositories.NewSetorRepository(db)
	servRepo := repositories.NewServidorRepository(db)
	movRepo := repositories.NewMovimentacaoRepository(db)
	termoRepo := repositories.NewTermoRepository(db)
	usrRepo := repositories.NewUsuarioRepository(db)

	itemSvc := services.NewItemService(itemRepo, catRepo, setorRepo, servRepo, movRepo, termoRepo)
	movSvc := services.NewMovimentacaoService(movRepo, itemRepo, setorRepo, servRepo)
	usrSvc := services.NewUsuarioService(usrRepo)
	catSvc := services.NewCategoriaService(catRepo)

	usuario, err := usrSvc.Criar(services.EntradaUsuario{
		Nome: "Teste", Email: "teste@sige.local.br", Senha: "12345678",
		Perfil: models.PerfilAdministrador, Ativo: true,
	})
	if err != nil {
		t.Fatalf("criar usuário: %v", err)
	}
	catCons, err := catSvc.Criar(services.EntradaCategoria{Nome: "Consumo", Consumivel: true})
	if err != nil {
		t.Fatalf("criar categoria consumo: %v", err)
	}
	catPatr, err := catSvc.Criar(services.EntradaCategoria{Nome: "Patrimônio", Consumivel: false})
	if err != nil {
		t.Fatalf("criar categoria patrimônio: %v", err)
	}

	return &fixtures{movSvc: movSvc, itemSvc: itemSvc, db: db, usuario: usuario, catCons: catCons, catPatr: catPatr}
}

func (f *fixtures) novoConsumivel(t *testing.T, qtd, minimo int) *models.Item {
	t.Helper()
	item, err := f.itemSvc.Criar(services.EntradaItem{
		Descricao: "Item Consumível", CategoriaID: f.catCons.ID,
		Quantidade: qtd, EstoqueMinimo: minimo, EstadoConservacao: models.EstadoNovo,
	})
	if err != nil {
		t.Fatalf("criar item consumível: %v", err)
	}
	return item
}

// --- Regra: ENTRADA soma ao estoque ---
func TestEntradaAumentaEstoque(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 5, 2)

	res, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovEntradaCompra, Quantidade: 10,
		RegistradoPorID: f.usuario.ID,
	})
	if err != nil {
		t.Fatalf("entrada falhou: %v", err)
	}
	if res.Item.Quantidade != 15 {
		t.Errorf("esperado saldo 15, obtido %d", res.Item.Quantidade)
	}
	if res.Movimentacao.SaldoResultante != 15 {
		t.Errorf("saldo resultante esperado 15, obtido %d", res.Movimentacao.SaldoResultante)
	}
	if res.AlertaEstoque {
		t.Errorf("não deveria haver alerta com 15 > mínimo 2")
	}
}

// --- Regra: SAÍDA subtrai do estoque ---
func TestSaidaReduzEstoque(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 10, 2)

	res, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovSaidaTransferencia, Quantidade: 4,
		SetorDestinoID: nil, RegistradoPorID: f.usuario.ID,
	})
	// Transferência exige setor de destino — usamos descarte? Não: testamos
	// empréstimo que não exige destino.
	if err == nil {
		t.Fatalf("transferência sem destino deveria falhar")
	}

	res, err = f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovSaidaEmprestimo, Quantidade: 4,
		RegistradoPorID: f.usuario.ID,
	})
	if err != nil {
		t.Fatalf("saída falhou: %v", err)
	}
	if res.Item.Quantidade != 6 {
		t.Errorf("esperado saldo 6, obtido %d", res.Item.Quantidade)
	}
}

// --- Regra crítica: SAÍDA maior que estoque deve ser bloqueada ---
func TestSaidaMaiorQueEstoqueBloqueada(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 3, 1)

	_, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovSaidaEmprestimo, Quantidade: 5,
		RegistradoPorID: f.usuario.ID,
	})
	if !errors.Is(err, services.ErrEstoqueInsuficiente) {
		t.Fatalf("esperado ErrEstoqueInsuficiente, obtido %v", err)
	}

	// Estoque deve permanecer intacto.
	atual, _ := f.itemSvc.BuscarPorID(item.ID)
	if atual.Quantidade != 3 {
		t.Errorf("estoque deveria permanecer 3, obtido %d", atual.Quantidade)
	}
}

// --- Regra: ALERTA quando estoque fica abaixo do mínimo ---
func TestAlertaEstoqueBaixo(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 5, 4)

	res, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovSaidaEmprestimo, Quantidade: 2,
		RegistradoPorID: f.usuario.ID,
	})
	if err != nil {
		t.Fatalf("saída falhou: %v", err)
	}
	// Saldo 3 < mínimo 4 -> alerta.
	if !res.AlertaEstoque {
		t.Errorf("esperado alerta de estoque baixo (saldo 3 < mínimo 4)")
	}

	abaixo, err := f.itemSvc.ListarAbaixoDoMinimo()
	if err != nil {
		t.Fatalf("listar abaixo do mínimo: %v", err)
	}
	if len(abaixo) != 1 {
		t.Errorf("esperado 1 item em alerta, obtido %d", len(abaixo))
	}
}

// --- Regra: BAIXA patrimonial marca item, não remove ---
func TestBaixaPatrimonial(t *testing.T) {
	f := novoAmbiente(t)
	patNum := "PMFB-999"
	item, err := f.itemSvc.Criar(services.EntradaItem{
		Descricao: "Notebook Velho", CategoriaID: f.catPatr.ID,
		NumeroPatrimonio: &patNum, EstadoConservacao: models.EstadoRegular,
	})
	if err != nil {
		t.Fatalf("criar patrimoniado: %v", err)
	}

	res, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovSaidaDescarte, Quantidade: 1,
		MotivoBaixa: "Equipamento obsoleto e sem reparo", RegistradoPorID: f.usuario.ID,
	})
	if err != nil {
		t.Fatalf("baixa falhou: %v", err)
	}
	if !res.Item.Baixado {
		t.Errorf("item deveria estar marcado como baixado")
	}
	if res.Item.EstadoConservacao != models.EstadoInservivel {
		t.Errorf("estado deveria ser inservível, obtido %s", res.Item.EstadoConservacao)
	}

	// Item permanece consultável (rastreabilidade total).
	persistido, err := f.itemSvc.BuscarPorID(item.ID)
	if err != nil {
		t.Fatalf("item baixado deveria continuar visível: %v", err)
	}
	if !persistido.Baixado || persistido.MotivoBaixa == "" {
		t.Errorf("flag/motivo de baixa não persistidos")
	}

	// Nova movimentação em item baixado deve ser bloqueada.
	_, err = f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovEntradaDevolucao, Quantidade: 1,
		RegistradoPorID: f.usuario.ID,
	})
	if !errors.Is(err, services.ErrItemBaixado) {
		t.Errorf("esperado ErrItemBaixado, obtido %v", err)
	}
}

// --- Regra: exclusão de item sem histórico (correção de cadastro) ---
func TestExcluirItemSemHistorico(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 5, 1)

	if err := f.itemSvc.Excluir(item.ID); err != nil {
		t.Fatalf("exclusão de item sem histórico deveria funcionar: %v", err)
	}

	// Após soft delete, o item não deve mais ser encontrado.
	if _, err := f.itemSvc.BuscarPorID(item.ID); !errors.Is(err, services.ErrNaoEncontrado) {
		t.Errorf("esperado ErrNaoEncontrado após exclusão, obtido %v", err)
	}
}

// --- Regra: exclusão bloqueada quando o item tem histórico ---
func TestExcluirItemComHistoricoBloqueada(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 5, 1)

	// Gera histórico com uma entrada.
	if _, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovEntradaCompra, Quantidade: 2,
		RegistradoPorID: f.usuario.ID,
	}); err != nil {
		t.Fatalf("entrada falhou: %v", err)
	}

	err := f.itemSvc.Excluir(item.ID)
	if !errors.Is(err, services.ErrItemComHistorico) {
		t.Fatalf("esperado ErrItemComHistorico, obtido %v", err)
	}

	// Item deve continuar existindo.
	if _, err := f.itemSvc.BuscarPorID(item.ID); err != nil {
		t.Errorf("item com histórico não deveria ter sido removido: %v", err)
	}
}

// --- Validação: quantidade inválida ---
func TestValidacaoQuantidadeInvalida(t *testing.T) {
	f := novoAmbiente(t)
	item := f.novoConsumivel(t, 5, 1)

	_, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: item.ID, Tipo: models.MovEntradaCompra, Quantidade: 0,
		RegistradoPorID: f.usuario.ID,
	})
	if !errors.Is(err, services.ErrValidacao) {
		t.Fatalf("esperado ErrValidacao para quantidade 0, obtido %v", err)
	}
}

func TestLotePatrimonialSemTombamentoETransferencia(t *testing.T) {
	f := novoAmbiente(t)
	lote, err := f.itemSvc.CriarLote(services.EntradaItem{
		Descricao: "Monitor 24 polegadas", CategoriaID: f.catPatr.ID,
		Marca: "LG", Modelo: "24", EstadoConservacao: models.EstadoNovo,
	}, 48)
	if err != nil {
		t.Fatalf("criar lote: %v", err)
	}
	if lote.Quantidade != 48 || len(lote.IDs) != 48 || lote.IDs[0] == 0 || lote.LoteCodigo == "" {
		t.Fatalf("resultado de lote inesperado: %+v", lote)
	}
	var total int64
	if err := f.db.Model(&models.Item{}).Where("lote_codigo = ? AND numero_patrimonio IS NULL AND setor_id IS NULL AND quantidade = 1", lote.LoteCodigo).Count(&total).Error; err != nil || total != 48 {
		t.Fatalf("esperadas 48 unidades pendentes, obtidas %d (erro: %v)", total, err)
	}

	setor := &models.Setor{Nome: "Secretaria de Saúde"}
	if err := f.db.Create(setor).Error; err != nil {
		t.Fatal(err)
	}
	res, err := f.movSvc.Registrar(services.EntradaMovimentacao{
		ItemID: lote.IDs[0], Tipo: models.MovSaidaTransferencia,
		Quantidade: 1, SetorDestinoID: &setor.ID, RegistradoPorID: f.usuario.ID,
	})
	if err != nil {
		t.Fatalf("transferir bem: %v", err)
	}
	if res.Item.Quantidade != 1 || res.Movimentacao.SaldoResultante != 1 || res.Item.SetorID == nil || *res.Item.SetorID != setor.ID {
		t.Fatalf("transferência consumiu unidade ou não atualizou destino: %+v", res.Item)
	}
	_, err = f.itemSvc.Atualizar(lote.IDs[0], services.EntradaItem{
		Descricao: "Monitor 24 polegadas", CategoriaID: f.catPatr.ID,
		EstadoConservacao: models.EstadoNovo, Quantidade: 1,
	})
	if !errors.Is(err, services.ErrValidacao) {
		t.Fatalf("edição direta do destino deveria ser recusada: %v", err)
	}
	patrimonio := "PM-2026-001"
	_, err = f.itemSvc.Atualizar(lote.IDs[0], services.EntradaItem{
		Descricao: "Monitor 24 polegadas", CategoriaID: f.catPatr.ID,
		NumeroPatrimonio: &patrimonio, Marca: "LG", Modelo: "24",
		EstadoConservacao: models.EstadoNovo, Quantidade: 1, SetorID: &setor.ID,
	})
	if err != nil {
		t.Fatalf("tombar bem após distribuição: %v", err)
	}
	var restantes int64
	if err := f.db.Model(&models.Item{}).Where("lote_codigo = ? AND numero_patrimonio IS NULL", lote.LoteCodigo).Count(&restantes).Error; err != nil || restantes != 47 {
		t.Fatalf("esperadas 47 unidades ainda sem patrimônio, obtidas %d (erro: %v)", restantes, err)
	}
	lista, total, err := repositories.NewItemRepository(f.db).Listar(repositories.FiltroItem{
		Texto: "TI-" + fmt.Sprintf("%06d", lote.IDs[0]),
	})
	if err != nil || total != 1 || len(lista) != 1 || lista[0].ID != lote.IDs[0] {
		t.Fatalf("busca por código interno falhou: total=%d erro=%v", total, err)
	}
}

func TestLoteRecusaConsumoEQuantidadeExcessiva(t *testing.T) {
	f := novoAmbiente(t)
	_, err := f.itemSvc.CriarLote(services.EntradaItem{Descricao: "Toner", CategoriaID: f.catCons.ID}, 48)
	if !errors.Is(err, services.ErrValidacao) {
		t.Fatalf("lote individual de consumível deveria falhar: %v", err)
	}
	_, err = f.itemSvc.Criar(services.EntradaItem{Descricao: "Monitor", CategoriaID: f.catPatr.ID, Quantidade: 48})
	if !errors.Is(err, services.ErrValidacao) {
		t.Fatalf("bem permanente com quantidade 48 deveria falhar: %v", err)
	}
}
