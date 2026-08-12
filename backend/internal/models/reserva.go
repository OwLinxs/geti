package models

import (
	"time"

	"gorm.io/gorm"
)

// StatusReserva descreve a situação de uma reserva de equipamento.
type StatusReserva string

const (
	ReservaReservada StatusReserva = "reservada"
	ReservaEmUso     StatusReserva = "em_uso"
	ReservaDevolvida StatusReserva = "devolvida"
	ReservaCancelada StatusReserva = "cancelada"
)

// StatusReservaValido valida o status informado.
func StatusReservaValido(s StatusReserva) bool {
	switch s {
	case ReservaReservada, ReservaEmUso, ReservaDevolvida, ReservaCancelada:
		return true
	}
	return false
}

// ReservaAtiva indica se o status ocupa o equipamento (conta para conflito de
// datas). Devolvida/cancelada liberam o equipamento.
func (s StatusReserva) Ativa() bool {
	return s == ReservaReservada || s == ReservaEmUso
}

// Reserva registra o agendamento de uso de um equipamento patrimoniado por um
// período. Impede conflito de datas para o mesmo item (ver serviço).
type Reserva struct {
	Base

	ItemID uint  `gorm:"not null;index" json:"item_id"`
	Item   *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`

	SolicitanteID *uint     `gorm:"index" json:"solicitante_id,omitempty"`
	Solicitante   *Servidor `gorm:"foreignKey:SolicitanteID" json:"solicitante,omitempty"`

	DataInicio time.Time `gorm:"not null;index" json:"data_inicio"`
	DataFim    time.Time `gorm:"not null;index" json:"data_fim"`

	Finalidade string        `gorm:"size:500" json:"finalidade,omitempty"`
	Status     StatusReserva `gorm:"size:20;not null;default:reservada;index" json:"status"`

	AprovadoPorID *uint    `gorm:"index" json:"aprovado_por_id,omitempty"`
	AprovadoPor   *Usuario `gorm:"foreignKey:AprovadoPorID" json:"aprovado_por,omitempty"`

	Observacao string `gorm:"size:500" json:"observacao,omitempty"`

	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Reserva) TableName() string { return "reservas" }
