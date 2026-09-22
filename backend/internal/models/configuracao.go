package models

import "time"

// Configuracao guarda parâmetros configuráveis do sistema como pares
// chave/valor (valor em JSON), editáveis pela administração.
type Configuracao struct {
	Chave        string    `gorm:"size:60;primaryKey" json:"chave"`
	Valor        string    `gorm:"type:text" json:"valor"`
	AtualizadoEm time.Time `gorm:"autoUpdateTime" json:"atualizado_em"`
}

func (Configuracao) TableName() string { return "configuracoes" }
