package models

// Setor representa a localização organizacional (departamento, secretaria,
// sala) onde um item pode estar alocado.
type Setor struct {
	Base
	Nome      string `gorm:"size:120;uniqueIndex;not null" json:"nome"`
	Sigla     string `gorm:"size:20;index" json:"sigla"`
	Localizacao string `gorm:"size:160" json:"localizacao"` // prédio/sala, opcional
}

func (Setor) TableName() string { return "setores" }
