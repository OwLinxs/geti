package models

import "gorm.io/gorm"

// Fornecedor representa uma empresa/prestador de onde vêm equipamentos,
// serviços ou contratos do setor de T.I.
type Fornecedor struct {
	Base

	Nome       string `gorm:"size:200;not null;index" json:"nome"`
	CNPJ       string `gorm:"size:20;index" json:"cnpj,omitempty"`
	Email      string `gorm:"size:120" json:"email,omitempty"`
	Telefone   string `gorm:"size:40" json:"telefone,omitempty"`
	Endereco   string `gorm:"size:255" json:"endereco,omitempty"`
	Observacao string `gorm:"size:500" json:"observacao,omitempty"`

	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Fornecedor) TableName() string { return "fornecedores" }
