package models

import "gorm.io/gorm"

// RespostaRapida é um texto-modelo que a equipe insere no chat do chamado com
// um clique. Configurável pela administração.
type RespostaRapida struct {
	Base

	Titulo   string `gorm:"size:120;not null" json:"titulo"`
	Conteudo string `gorm:"type:text;not null" json:"conteudo"`
	Ordem    int    `gorm:"not null;default:0;index" json:"ordem"`
	Ativo    bool   `gorm:"not null;index" json:"ativo"`

	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RespostaRapida) TableName() string { return "respostas_rapidas" }
