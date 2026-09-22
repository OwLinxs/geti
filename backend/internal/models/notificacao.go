package models

// Tipos de notificação in-app.
const (
	NotifChamadoNovo = "chamado_novo"
	NotifMensagem    = "mensagem"
	NotifStatus      = "status"
)

// Notificacao é um aviso in-app destinado a um usuário (sino do sistema).
type Notificacao struct {
	Base

	UsuarioID uint   `gorm:"not null;index" json:"usuario_id"`
	Titulo    string `gorm:"size:150;not null" json:"titulo"`
	Mensagem  string `gorm:"size:500" json:"mensagem,omitempty"`
	Tipo      string `gorm:"size:30;index" json:"tipo"`
	// RecursoID referencia o chamado/OS relacionado (para link do sino).
	RecursoID *uint `gorm:"index" json:"recurso_id,omitempty"`
	Lida      bool  `gorm:"not null;default:false;index" json:"lida"`
}

func (Notificacao) TableName() string { return "notificacoes" }
