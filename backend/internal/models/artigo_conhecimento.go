package models

import "gorm.io/gorm"

// ArtigoConhecimento é um artigo da base de conhecimento (KB) do setor de T.I.:
// procedimentos, soluções recorrentes e orientações reutilizáveis pela equipe.
type ArtigoConhecimento struct {
	Base

	Titulo    string `gorm:"size:200;not null;index" json:"titulo"`
	Categoria string `gorm:"size:80;index" json:"categoria,omitempty"`
	Conteudo  string `gorm:"type:text;not null" json:"conteudo"`

	AutorID *uint    `gorm:"index" json:"autor_id,omitempty"`
	Autor   *Usuario `gorm:"foreignKey:AutorID" json:"autor,omitempty"`

	Publicado     bool `gorm:"not null;default:true;index" json:"publicado"`
	Visualizacoes int  `gorm:"not null;default:0" json:"visualizacoes"`

	// Soft delete para correção/remoção sem perder rastreabilidade.
	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ArtigoConhecimento) TableName() string { return "artigos_conhecimento" }
