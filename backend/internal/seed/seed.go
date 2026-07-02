// Package seed popula o banco com dados iniciais. Há duas trilhas:
//
//   - Seed BASE (sempre): usuário administrador, taxonomia padrão de T.I.
//     (categorias patrimoniadas e de consumo) e setores/secretarias reais.
//     É o necessário para começar a usar o sistema em produção.
//   - Seed DEMO (apenas desenvolvimento, controlado por SEED_DEMO=true):
//     servidores, itens, movimentações e termo fictícios para testes.
//
// Tudo é idempotente: não duplica registros já existentes.
package seed

import (
	"log"

	"github.com/pmfb/sige-ti/internal/container"
	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/services"
)

// Executar roda o seed base e, se habilitado em configuração, o seed demo.
func Executar(ct *container.Container) error {
	criados, err := ExecutarBase(ct)
	if err != nil {
		return err
	}

	// Seed demo só roda em desenvolvimento (SEED_DEMO=true). A validação de
	// config já impede SEED_DEMO em produção.
	if ct.Config.SeedDemo {
		if err := ExecutarDemo(ct); err != nil {
			return err
		}
	}

	if criados {
		log.Printf("[seed] base concluída. Admin: %s", ct.Config.AdminEmail)
	}
	return nil
}

// ExecutarBase cria o administrador, a taxonomia de categorias e os setores
// padrão. Idempotente: se já houver usuário admin, não recria nada.
//
// Devolve true se efetivamente criou dados (primeira execução).
func ExecutarBase(ct *container.Container) (bool, error) {
	// Idempotência: se já existe algum usuário, assume que a base já rodou.
	usuarios, err := ct.UsuarioService.Listar()
	if err != nil {
		return false, err
	}
	if len(usuarios) > 0 {
		return false, nil
	}

	log.Println("[seed] populando dados base (admin, categorias, setores)...")

	// 1) Administrador inicial (credenciais vêm de configuração/ambiente).
	if _, err := ct.UsuarioService.Criar(services.EntradaUsuario{
		Nome:   ct.Config.AdminNome,
		Email:  ct.Config.AdminEmail,
		Senha:  ct.Config.AdminSenha,
		Perfil: models.PerfilAdministrador,
		Ativo:  true,
	}); err != nil {
		return false, err
	}

	// 2) Categorias patrimoniadas (controle unitário, exigem patrimônio).
	categoriasPatrimoniadas := []services.EntradaCategoria{
		{Nome: "Desktop", Descricao: "Computadores de mesa"},
		{Nome: "Notebook", Descricao: "Computadores portáteis"},
		{Nome: "Monitor", Descricao: "Monitores de vídeo"},
		{Nome: "Impressora", Descricao: "Impressoras e multifuncionais"},
		{Nome: "Scanner", Descricao: "Digitalizadores"},
		{Nome: "Nobreak", Descricao: "Fontes de energia ininterrupta (UPS)"},
		{Nome: "Switch", Descricao: "Comutadores de rede"},
		{Nome: "Roteador", Descricao: "Roteadores e access points"},
		{Nome: "Servidor", Descricao: "Servidores físicos"},
		{Nome: "Projetor", Descricao: "Projetores multimídia"},
		{Nome: "Telefone IP", Descricao: "Aparelhos de telefonia VoIP"},
	}
	for _, c := range categoriasPatrimoniadas {
		if _, err := ct.CategoriaService.Criar(c); err != nil {
			return false, err
		}
	}

	// 3) Categorias de consumo (controle por quantidade/estoque mínimo).
	categoriasConsumo := []services.EntradaCategoria{
		{Nome: "Toner/Cartucho", Descricao: "Suprimentos de impressão", Consumivel: true},
		{Nome: "Cabo de rede", Descricao: "Cabos e conectores de rede", Consumivel: true},
		{Nome: "Mouse", Descricao: "Mouses (consumo/reposição)", Consumivel: true},
		{Nome: "Teclado", Descricao: "Teclados (consumo/reposição)", Consumivel: true},
		{Nome: "Pen drive", Descricao: "Dispositivos de armazenamento USB", Consumivel: true},
		{Nome: "HD/SSD", Descricao: "Discos de armazenamento", Consumivel: true},
		{Nome: "Memória RAM", Descricao: "Módulos de memória", Consumivel: true},
		{Nome: "Fonte/Adaptador", Descricao: "Fontes e adaptadores de energia", Consumivel: true},
	}
	for _, c := range categoriasConsumo {
		if _, err := ct.CategoriaService.Criar(c); err != nil {
			return false, err
		}
	}

	// 4) Setores / secretarias padrão da Prefeitura.
	setores := []services.EntradaSetor{
		{Nome: "Gabinete do Prefeito", Sigla: "GAB"},
		{Nome: "Secretaria de Administração", Sigla: "SEAD"},
		{Nome: "Secretaria de Finanças", Sigla: "SEFIN"},
		{Nome: "Secretaria de Saúde", Sigla: "SMS"},
		{Nome: "Secretaria de Educação", Sigla: "SME"},
		{Nome: "Secretaria de Obras", Sigla: "SEOBR"},
		{Nome: "Secretaria de Assistência Social", Sigla: "SEAS"},
		{Nome: "Departamento de T.I.", Sigla: "DTI"},
		{Nome: "Procuradoria", Sigla: "PROC"},
		{Nome: "Recursos Humanos", Sigla: "RH"},
		{Nome: "Protocolo", Sigla: "PROT"},
	}
	for _, s := range setores {
		if _, err := ct.SetorService.Criar(s); err != nil {
			return false, err
		}
	}

	return true, nil
}
