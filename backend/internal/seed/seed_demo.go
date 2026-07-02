package seed

import (
	"log"
	"time"

	"github.com/pmfb/sige-ti/internal/container"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
	"github.com/pmfb/sige-ti/internal/services"
)

// ExecutarDemo insere dados fictícios para desenvolvimento/testes: operador,
// servidores, itens patrimoniados e de consumo, movimentações (incluindo uma
// que dispara alerta de estoque) e um termo de responsabilidade.
//
// É idempotente em relação a si mesmo: se já houver itens, não recria nada.
// Reaproveita as categorias e setores criados pelo seed base, buscando-os por
// nome.
func ExecutarDemo(ct *container.Container) error {
	// Idempotência: se já existir algum item, assume que o demo já rodou.
	if _, total, err := ct.ItemService.Listar(buscarTudo()); err != nil {
		return err
	} else if total > 0 {
		return nil
	}

	log.Println("[seed] populando dados DEMO (itens e movimentações fictícios)...")

	// Administrador já existe (criado no seed base): localiza para auditoria.
	admin, err := primeiroAdmin(ct)
	if err != nil {
		return err
	}

	// Operador de exemplo (só em demo).
	if _, err := ct.UsuarioService.Criar(services.EntradaUsuario{
		Nome: "Operador de Estoque", Email: "operador@sige-ti.local",
		Senha: "operador123", Perfil: models.PerfilOperador, Ativo: true,
	}); err != nil {
		return err
	}

	// Localiza categorias e setores criados pelo seed base.
	catNotebook := idCategoria(ct, "Notebook")
	catMonitor := idCategoria(ct, "Monitor")
	catTeclado := idCategoria(ct, "Teclado")
	catToner := idCategoria(ct, "Toner/Cartucho")
	catCabo := idCategoria(ct, "Cabo de rede")

	setorTI := idSetor(ct, "Departamento de T.I.")
	setorFinancas := idSetor(ct, "Secretaria de Finanças")
	setorSaude := idSetor(ct, "Secretaria de Saúde")

	// Servidores (LGPD: apenas nome e matrícula).
	servAna, _ := ct.ServidorService.Criar(services.EntradaServidor{Nome: "Ana Paula Souza", Matricula: "1001", SetorID: setorFinancas, Ativo: true})
	servCarlos, _ := ct.ServidorService.Criar(services.EntradaServidor{Nome: "Carlos Eduardo Lima", Matricula: "1002", SetorID: setorSaude, Ativo: true})
	_, _ = ct.ServidorService.Criar(services.EntradaServidor{Nome: "Mariana Alves", Matricula: "1003", SetorID: setorTI, Ativo: true})

	pat := func(s string) *string { return &s }
	dataAq := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	valor := func(v float64) *float64 { return &v }

	// Itens patrimoniados.
	item1, _ := ct.ItemService.Criar(services.EntradaItem{
		Descricao: "Notebook Dell Latitude 3420", CategoriaID: deref(catNotebook),
		NumeroPatrimonio: pat("PMFB-000123"), NumeroSerie: pat("DLL-AB12CD"),
		Marca: "Dell", Modelo: "Latitude 3420", EstadoConservacao: models.EstadoBom,
		SetorID: setorTI, DataAquisicao: &dataAq, Valor: valor(4200.00),
	})
	_, _ = ct.ItemService.Criar(services.EntradaItem{
		Descricao: "Monitor LG 24 polegadas", CategoriaID: deref(catMonitor),
		NumeroPatrimonio: pat("PMFB-000124"), Marca: "LG", Modelo: "24MK430H",
		EstadoConservacao: models.EstadoNovo, SetorID: setorTI, Valor: valor(750.00),
	})
	_, _ = ct.ItemService.Criar(services.EntradaItem{
		Descricao: "Teclado USB ABNT2", CategoriaID: deref(catTeclado),
		Quantidade: 8, EstoqueMinimo: 3, Marca: "Multilaser",
		EstadoConservacao: models.EstadoNovo, SetorID: setorTI,
	})

	// Itens de consumo (com estoque mínimo).
	itemToner, _ := ct.ItemService.Criar(services.EntradaItem{
		Descricao: "Toner HP 85A (CE285A)", CategoriaID: deref(catToner),
		Quantidade: 10, EstoqueMinimo: 4, EstadoConservacao: models.EstadoNovo,
		SetorID: setorTI,
	})
	_, _ = ct.ItemService.Criar(services.EntradaItem{
		Descricao: "Cabo de rede UTP Cat6 (caixa 305m)", CategoriaID: deref(catCabo),
		Quantidade: 3, EstoqueMinimo: 2, EstadoConservacao: models.EstadoNovo,
		SetorID: setorTI,
	})

	// Movimentações de exemplo (respeitam as regras de negócio).
	// Entrada de toner por compra.
	_, _ = ct.MovimentacaoService.Registrar(services.EntradaMovimentacao{
		ItemID: itemToner.ID, Tipo: models.MovEntradaCompra, Quantidade: 5,
		Observacao: "Compra via pregão 12/2024", RegistradoPorID: admin.ID,
	})
	// Saída/empréstimo do notebook para Ana (Finanças).
	_, _ = ct.MovimentacaoService.Registrar(services.EntradaMovimentacao{
		ItemID: item1.ID, Tipo: models.MovSaidaEmprestimo, Quantidade: 1,
		SetorOrigemID: setorTI, SetorDestinoID: setorFinancas,
		ServidorID: &servAna.ID, Observacao: "Empréstimo para trabalho remoto",
		RegistradoPorID: admin.ID,
	})
	// Saída grande de toner para deixar abaixo do mínimo e disparar alerta.
	_, _ = ct.MovimentacaoService.Registrar(services.EntradaMovimentacao{
		ItemID: itemToner.ID, Tipo: models.MovSaidaTransferencia, Quantidade: 12,
		SetorOrigemID: setorTI, SetorDestinoID: setorSaude,
		ServidorID: &servCarlos.ID, Observacao: "Reposição de impressoras da Saúde",
		RegistradoPorID: admin.ID,
	})

	// Termo de responsabilidade de exemplo (notebook para Ana).
	_, _ = ct.TermoService.Emitir(services.EntradaTermo{
		ItemID: item1.ID, ServidorID: servAna.ID,
		Observacao: "Entrega acompanhada de carregador e mochila.",
		EmitidoPorID: admin.ID,
	})

	log.Println("[seed] demo concluído.")
	return nil
}

// ---- Helpers de localização por nome -------------------------------------

func idCategoria(ct *container.Container, nome string) *uint {
	cats, err := ct.CategoriaService.Listar()
	if err != nil {
		return nil
	}
	for i := range cats {
		if cats[i].Nome == nome {
			id := cats[i].ID
			return &id
		}
	}
	return nil
}

func idSetor(ct *container.Container, nome string) *uint {
	setores, err := ct.SetorService.Listar()
	if err != nil {
		return nil
	}
	for i := range setores {
		if setores[i].Nome == nome {
			id := setores[i].ID
			return &id
		}
	}
	return nil
}

func primeiroAdmin(ct *container.Container) (*models.Usuario, error) {
	usuarios, err := ct.UsuarioService.Listar()
	if err != nil {
		return nil, err
	}
	for i := range usuarios {
		if usuarios[i].Perfil == models.PerfilAdministrador {
			return &usuarios[i], nil
		}
	}
	if len(usuarios) > 0 {
		return &usuarios[0], nil
	}
	return nil, nil
}

func deref(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}

func buscarTudo() repositories.FiltroItem {
	return repositories.FiltroItem{Pagina: 1, TamanhoPagina: 1}
}
