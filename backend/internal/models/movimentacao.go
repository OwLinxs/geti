package models

import "time"

// TipoMovimentacao classifica a natureza da movimentação de estoque.
type TipoMovimentacao string

const (
	// Entradas
	MovEntradaCompra    TipoMovimentacao = "entrada_compra"
	MovEntradaDoacao    TipoMovimentacao = "entrada_doacao"
	MovEntradaDevolucao TipoMovimentacao = "entrada_devolucao"
	// Saídas
	MovSaidaEmprestimo    TipoMovimentacao = "saida_emprestimo"
	MovSaidaTransferencia TipoMovimentacao = "saida_transferencia"
	MovSaidaDescarte      TipoMovimentacao = "saida_descarte" // baixa patrimonial
)

// EhEntrada informa se o tipo aumenta o estoque.
func (t TipoMovimentacao) EhEntrada() bool {
	switch t {
	case MovEntradaCompra, MovEntradaDoacao, MovEntradaDevolucao:
		return true
	}
	return false
}

// EhSaida informa se o tipo reduz o estoque.
func (t TipoMovimentacao) EhSaida() bool {
	switch t {
	case MovSaidaEmprestimo, MovSaidaTransferencia, MovSaidaDescarte:
		return true
	}
	return false
}

// EhBaixa indica que a movimentação representa uma baixa patrimonial.
func (t TipoMovimentacao) EhBaixa() bool {
	return t == MovSaidaDescarte
}

// TipoMovimentacaoValido valida o tipo informado.
func TipoMovimentacaoValido(t TipoMovimentacao) bool {
	return t.EhEntrada() || t.EhSaida()
}

// Movimentacao registra um evento imutável de entrada/saída no histórico do
// item. Registros nunca são alterados nem removidos (auditoria/LGPD): para
// corrigir, registra-se uma nova movimentação compensatória.
//
// Origem/Destino são "snapshots" textuais do contexto no momento do evento,
// além das FKs estruturadas, para preservar o histórico mesmo se setores ou
// servidores forem renomeados/desativados posteriormente.
type Movimentacao struct {
	Base

	ItemID uint  `gorm:"not null;index" json:"item_id"`
	Item   *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`

	Tipo       TipoMovimentacao `gorm:"size:30;not null;index" json:"tipo"`
	Quantidade int              `gorm:"not null" json:"quantidade"`

	// Saldo do item após aplicar esta movimentação (auditoria).
	SaldoResultante int `gorm:"not null" json:"saldo_resultante"`

	// Contexto estruturado (FKs, opcionais conforme o tipo).
	SetorOrigemID   *uint     `gorm:"index" json:"setor_origem_id,omitempty"`
	SetorOrigem     *Setor    `gorm:"foreignKey:SetorOrigemID" json:"setor_origem,omitempty"`
	SetorDestinoID  *uint     `gorm:"index" json:"setor_destino_id,omitempty"`
	SetorDestino    *Setor    `gorm:"foreignKey:SetorDestinoID" json:"setor_destino,omitempty"`
	ServidorID      *uint     `gorm:"index" json:"servidor_id,omitempty"` // responsável envolvido
	Servidor        *Servidor `gorm:"foreignKey:ServidorID" json:"servidor,omitempty"`

	// Quem registrou a movimentação (usuário do sistema).
	RegistradoPorID uint     `gorm:"not null;index" json:"registrado_por_id"`
	RegistradoPor   *Usuario `gorm:"foreignKey:RegistradoPorID" json:"registrado_por,omitempty"`

	// Snapshots textuais para preservar histórico.
	OrigemDescricao  string `gorm:"size:160" json:"origem_descricao,omitempty"`
	DestinoDescricao string `gorm:"size:160" json:"destino_descricao,omitempty"`

	Observacao string    `gorm:"size:500" json:"observacao,omitempty"`
	DataEvento time.Time `gorm:"not null;index" json:"data_evento"`
}

func (Movimentacao) TableName() string { return "movimentacoes" }
