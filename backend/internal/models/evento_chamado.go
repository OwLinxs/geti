package models

import "time"

// EventoChamado é um registro da linha do tempo de um chamado/OS (abertura,
// mudanças de status, atribuição, reabertura, avaliação).
type EventoChamado struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrdemServicoID uint      `gorm:"not null;index" json:"ordem_servico_id"`
	Tipo           string    `gorm:"size:30;index" json:"tipo"`
	Descricao      string    `gorm:"size:300" json:"descricao"`
	AutorNome      string    `gorm:"size:120" json:"autor_nome,omitempty"`
	CriadoEm       time.Time `gorm:"autoCreateTime;index" json:"criado_em"`
}

func (EventoChamado) TableName() string { return "eventos_chamado" }
