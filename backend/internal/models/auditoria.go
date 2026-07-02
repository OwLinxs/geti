package models

import "time"

// RegistroAuditoria é a trilha de auditoria do sistema: registra QUEM fez O QUÊ
// e QUANDO, para operações sensíveis (login, cadastros/edições/exclusões,
// movimentações, redefinição de senha, etc.).
//
// Patrimônio público exige rastreabilidade. Este registro é imutável (apenas
// inserção — nunca editado ou removido pela aplicação) e guarda snapshots do
// nome do usuário para permanecer legível mesmo que a conta seja alterada.
//
// LGPD: guardamos apenas o mínimo para auditoria (id/nome do usuário, ação,
// recurso e IP de origem). Nenhum dado sensível é coletado.
type RegistroAuditoria struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	CriadoEm time.Time `gorm:"autoCreateTime;index" json:"criado_em"`

	// Autor da ação. UsuarioID é nulo em eventos sem sessão (ex.: login falho).
	UsuarioID    *uint  `gorm:"index" json:"usuario_id,omitempty"`
	UsuarioNome  string `gorm:"size:120;index" json:"usuario_nome"`
	UsuarioEmail string `gorm:"size:160" json:"usuario_email,omitempty"`

	// Ação legível em português (ex.: "criou", "excluiu", "login").
	Acao string `gorm:"size:40;not null;index" json:"acao"`
	// Recurso afetado (ex.: "item", "usuario", "movimentacao"), vazio p/ login.
	Recurso   string `gorm:"size:40;index" json:"recurso,omitempty"`
	RecursoID *uint  `gorm:"index" json:"recurso_id,omitempty"`

	// Contexto técnico da requisição.
	Metodo  string `gorm:"size:10" json:"metodo,omitempty"`
	Caminho string `gorm:"size:200" json:"caminho,omitempty"`
	Status  int    `gorm:"index" json:"status"`
	IP      string `gorm:"size:60" json:"ip,omitempty"`

	// Descrição curta e amigável (ex.: "login bem-sucedido", "excluiu item #12").
	Detalhe string `gorm:"size:255" json:"detalhe,omitempty"`
}

func (RegistroAuditoria) TableName() string { return "registros_auditoria" }
