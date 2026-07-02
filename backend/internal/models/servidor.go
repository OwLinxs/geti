package models

// Servidor é a pessoa (servidor público) responsável por equipamentos.
//
// LGPD: coletamos apenas nome e matrícula — o mínimo necessário para
// identificar o responsável e emitir o termo. NÃO armazenamos CPF, RG,
// endereço, telefone ou qualquer dado sensível.
type Servidor struct {
	Base
	Nome      string `gorm:"size:120;not null;index" json:"nome"`
	Matricula string `gorm:"size:40;uniqueIndex;not null" json:"matricula"`
	SetorID   *uint  `gorm:"index" json:"setor_id,omitempty"` // lotação atual (opcional)
	Setor     *Setor `gorm:"foreignKey:SetorID" json:"setor,omitempty"`
	Ativo     bool   `gorm:"not null;default:true" json:"ativo"`
}

func (Servidor) TableName() string { return "servidores" }
