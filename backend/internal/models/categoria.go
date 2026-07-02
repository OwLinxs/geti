package models

// Categoria classifica os itens (computador, monitor, impressora, periférico,
// rede, material de consumo, etc.). O campo Consumivel indica se itens desta
// categoria são materiais de consumo (controlados por quantidade/estoque
// mínimo) em vez de bens patrimoniados individualmente.
type Categoria struct {
	Base
	Nome       string `gorm:"size:80;uniqueIndex;not null" json:"nome"`
	Descricao  string `gorm:"size:255" json:"descricao"`
	Consumivel bool   `gorm:"not null;default:false" json:"consumivel"`
}

func (Categoria) TableName() string { return "categorias" }
