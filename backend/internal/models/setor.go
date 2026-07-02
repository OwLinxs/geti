package models

// Setor representa uma unidade organizacional da Prefeitura (Secretaria,
// Departamento, Divisão, Seção...). É uma ÁRVORE: cada unidade pode ter uma
// unidade-pai (PaiID) e várias filhas. Unidades de topo (Secretarias) têm
// PaiID nulo.
//
// Itens e servidores continuam vinculados a uma unidade (setor_id) — a
// hierarquia é usada para organização e relatórios agregados.
type Setor struct {
	Base
	Nome        string `gorm:"size:120;uniqueIndex;not null" json:"nome"`
	Sigla       string `gorm:"size:20;index" json:"sigla"`
	Localizacao string `gorm:"size:160" json:"localizacao"` // prédio/sala, opcional

	// Hierarquia (auto-relação). PaiID nulo = unidade de topo (Secretaria).
	PaiID  *uint   `gorm:"index" json:"pai_id,omitempty"`
	Pai    *Setor  `gorm:"foreignKey:PaiID" json:"pai,omitempty"`
	Filhos []Setor `gorm:"foreignKey:PaiID" json:"filhos,omitempty"`
}

func (Setor) TableName() string { return "setores" }
