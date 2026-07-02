package models

import "time"

// TermoResponsabilidade registra a emissão de um termo (recibo) de entrega de
// equipamento a um servidor. O PDF é gerado a partir destes dados + parâmetros
// institucionais (nome da Prefeitura, brasão), que vêm de variáveis de
// ambiente — sem hardcode.
//
// Guardamos snapshots dos nomes/identificadores no momento da emissão para que
// o termo permaneça fiel mesmo que cadastros sejam alterados depois.
type TermoResponsabilidade struct {
	Base

	Numero string `gorm:"size:40;uniqueIndex;not null" json:"numero"` // ex.: TR-2026-0001

	ItemID uint  `gorm:"not null;index" json:"item_id"`
	Item   *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`

	ServidorID uint      `gorm:"not null;index" json:"servidor_id"`
	Servidor   *Servidor `gorm:"foreignKey:ServidorID" json:"servidor,omitempty"`

	// Movimentação de saída que originou o termo (opcional, mas usual).
	MovimentacaoID *uint         `gorm:"index" json:"movimentacao_id,omitempty"`
	Movimentacao   *Movimentacao `gorm:"foreignKey:MovimentacaoID" json:"movimentacao,omitempty"`

	EmitidoPorID uint     `gorm:"not null;index" json:"emitido_por_id"`
	EmitidoPor   *Usuario `gorm:"foreignKey:EmitidoPorID" json:"emitido_por,omitempty"`

	// Snapshots para fidelidade do documento.
	ItemDescricaoSnapshot   string `gorm:"size:200" json:"item_descricao_snapshot"`
	PatrimonioSnapshot      string `gorm:"size:60" json:"patrimonio_snapshot,omitempty"`
	ServidorNomeSnapshot    string `gorm:"size:120" json:"servidor_nome_snapshot"`
	ServidorMatriculaSnapshot string `gorm:"size:40" json:"servidor_matricula_snapshot"`

	Observacao string    `gorm:"size:500" json:"observacao,omitempty"`
	DataEmissao time.Time `gorm:"not null;index" json:"data_emissao"`
}

func (TermoResponsabilidade) TableName() string { return "termos_responsabilidade" }
