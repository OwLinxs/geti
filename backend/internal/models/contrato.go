package models

import (
	"time"

	"gorm.io/gorm"
)

// StatusContrato descreve a situação do contrato.
type StatusContrato string

const (
	ContratoVigente   StatusContrato = "vigente"
	ContratoEncerrado StatusContrato = "encerrado"
	ContratoCancelado StatusContrato = "cancelado"
)

// StatusContratoValido valida o status informado.
func StatusContratoValido(s StatusContrato) bool {
	switch s {
	case ContratoVigente, ContratoEncerrado, ContratoCancelado:
		return true
	}
	return false
}

// Contrato registra um contrato/serviço com um fornecedor, com controle de
// vigência (data de término) para alerta de vencimento.
type Contrato struct {
	Base

	Numero string `gorm:"size:60;not null;index" json:"numero"`

	FornecedorID uint        `gorm:"not null;index" json:"fornecedor_id"`
	Fornecedor   *Fornecedor `gorm:"foreignKey:FornecedorID" json:"fornecedor,omitempty"`

	Objeto     string         `gorm:"size:500;not null" json:"objeto"`
	DataInicio *time.Time     `json:"data_inicio,omitempty"`
	DataFim    *time.Time     `gorm:"index" json:"data_fim,omitempty"` // vencimento
	Valor      *float64       `json:"valor,omitempty"`
	Status     StatusContrato `gorm:"size:20;not null;default:vigente;index" json:"status"`
	Observacao string         `gorm:"size:1000" json:"observacao,omitempty"`

	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Contrato) TableName() string { return "contratos" }
