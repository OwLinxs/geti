package models

import "time"

// Base concentra campos comuns de auditoria a todas as entidades.
// Não usamos DeletedAt (soft delete) deliberadamente em entidades de
// patrimônio: a baixa é registrada via movimentação + flag, preservando
// rastreabilidade total (decisão de negócio do SIGE-TI).
type Base struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CriadoEm  time.Time `gorm:"autoCreateTime" json:"criado_em"`
	AtualizadoEm time.Time `gorm:"autoUpdateTime" json:"atualizado_em"`
}
