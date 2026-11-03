package models

// PermissaoUsuario concede a um usuário específico uma ação privilegiada que,
// por padrão, só o administrador possui. É um override por usuário sobre a
// baseline do perfil: o administrador tem tudo (super-admin, não editável) e o
// operador parte sem nenhuma dessas ações, recebendo-as individualmente aqui.
//
// Cada linha representa uma concessão (a ausência da linha = não concedido).
type PermissaoUsuario struct {
	Base
	UsuarioID uint   `gorm:"not null;uniqueIndex:idx_usuario_permissao" json:"usuario_id"`
	Chave     string `gorm:"size:60;not null;uniqueIndex:idx_usuario_permissao" json:"chave"`
}

func (PermissaoUsuario) TableName() string { return "permissoes_usuario" }

// Permissao descreve uma ação privilegiada do catálogo (exibida na tela de
// gestão de permissões por usuário).
type Permissao struct {
	Chave     string `json:"chave"`
	Rotulo    string `json:"rotulo"`
	Grupo     string `json:"grupo"`
	Descricao string `json:"descricao"`
}

// Catálogo fixo de permissões delegáveis. Cada chave corresponde às rotas que,
// historicamente, eram exclusivas de administrador (adminOnly). Conceder a
// chave a um operador libera exatamente aquelas operações para ele.
const (
	PermUsuariosGerenciar     = "usuarios.gerenciar"
	PermAuditoriaVer          = "auditoria.ver"
	PermConfigGerenciar       = "config.gerenciar"
	PermCategoriasGerenciar   = "categorias.gerenciar"
	PermSetoresGerenciar      = "setores.gerenciar"
	PermServidoresExcluir     = "servidores.excluir"
	PermItensExcluir          = "itens.excluir"
	PermFornecedoresGerenciar = "fornecedores.gerenciar"
	PermContratosGerenciar    = "contratos.gerenciar"
	PermReservasExcluir       = "reservas.excluir"
	PermChamadosExcluir       = "chamados.excluir"
	PermChamadosCategorias    = "chamados.categorias"
	PermChamadosRespostas     = "chamados.respostas"
	PermConhecimentoExcluir   = "conhecimento.excluir"
)

// catalogoPermissoes é a lista completa, na ordem de exibição.
var catalogoPermissoes = []Permissao{
	// --- Administração ---
	{PermUsuariosGerenciar, "Gerenciar usuários", "Administração", "Criar, editar, resetar senha e ativar/desativar usuários do sistema."},
	{PermAuditoriaVer, "Ver trilha de auditoria", "Administração", "Consultar o histórico de operações registradas no sistema."},
	{PermConfigGerenciar, "Configurações de chamados", "Administração", "Alterar as configurações do módulo de chamados (SLA, auto-cadastro, etc.)."},
	// --- Inventário / Patrimônio ---
	{PermCategoriasGerenciar, "Gerenciar categorias", "Inventário", "Criar, editar e remover categorias de itens/patrimônio."},
	{PermSetoresGerenciar, "Gerenciar setores", "Inventário", "Criar, editar e remover setores."},
	{PermServidoresExcluir, "Excluir servidores", "Inventário", "Remover cadastros de servidores (responsáveis por equipamentos)."},
	{PermItensExcluir, "Excluir itens", "Inventário", "Excluir itens/patrimônio sem histórico (correção de cadastro)."},
	{PermFornecedoresGerenciar, "Gerenciar fornecedores", "Inventário", "Criar, editar e remover fornecedores."},
	{PermContratosGerenciar, "Gerenciar contratos", "Inventário", "Criar, editar e remover contratos."},
	{PermReservasExcluir, "Excluir reservas", "Inventário", "Remover reservas de equipamentos."},
	// --- Chamados / Manutenção ---
	{PermChamadosExcluir, "Excluir chamados", "Chamados", "Excluir ordens de serviço abertas por engano."},
	{PermChamadosCategorias, "Gerenciar categorias de chamado", "Chamados", "Criar, editar e remover categorias de chamado (com checklist)."},
	{PermChamadosRespostas, "Gerenciar respostas rápidas", "Chamados", "Criar, editar e remover as respostas rápidas do chat."},
	{PermConhecimentoExcluir, "Excluir artigos da base", "Chamados", "Excluir artigos da base de conhecimento."},
}

// CatalogoPermissoes devolve a lista de permissões delegáveis.
func CatalogoPermissoes() []Permissao {
	out := make([]Permissao, len(catalogoPermissoes))
	copy(out, catalogoPermissoes)
	return out
}

// PermissaoValida informa se a chave pertence ao catálogo.
func PermissaoValida(chave string) bool {
	for _, p := range catalogoPermissoes {
		if p.Chave == chave {
			return true
		}
	}
	return false
}
