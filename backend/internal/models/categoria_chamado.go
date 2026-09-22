package models

import "gorm.io/gorm"

// CategoriaChamado classifica os chamados (rede, e-mail, impressora, etc.).
// É configurável pela administração (CRUD) e usada em filtros e relatórios.
// Distinta de Categoria (que classifica itens do inventário).
type CategoriaChamado struct {
	Base

	Nome      string `gorm:"size:80;uniqueIndex;not null" json:"nome"`
	Descricao string `gorm:"size:255" json:"descricao,omitempty"`
	Ordem     int    `gorm:"not null;default:0;index" json:"ordem"`
	Ativo     bool   `gorm:"not null;index" json:"ativo"`

	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoriaChamado) TableName() string { return "categorias_chamado" }
