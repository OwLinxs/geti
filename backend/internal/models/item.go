package models

import (
	"time"

	"gorm.io/gorm"
)

// EstadoConservacao descreve o estado físico do item.
type EstadoConservacao string

const (
	EstadoNovo       EstadoConservacao = "novo"
	EstadoBom        EstadoConservacao = "bom"
	EstadoRegular    EstadoConservacao = "regular"
	EstadoInservivel EstadoConservacao = "inservivel"
)

// EstadoConservacaoValido valida o estado informado.
func EstadoConservacaoValido(e EstadoConservacao) bool {
	switch e {
	case EstadoNovo, EstadoBom, EstadoRegular, EstadoInservivel:
		return true
	}
	return false
}

// Item representa tanto bens patrimoniados (com nº de patrimônio, controle
// unitário) quanto materiais de consumo (controle por quantidade/estoque
// mínimo). A distinção vem da Categoria.Consumivel e da presença de
// NumeroPatrimonio.
//
// Regras de estoque:
//   - Consumível: Quantidade reflete o saldo atual; EstoqueMinimo dispara
//     alerta quando Quantidade < EstoqueMinimo.
//   - Patrimoniado: Quantidade normalmente é 1; o controle é unitário.
//
// Baixa patrimonial (decisão de negócio): NÃO é soft delete. O item
// permanece visível, com Baixado=true, EstadoConservacao=inservivel e
// a movimentação de baixa registrada no histórico (rastreabilidade total).
type Item struct {
	Base

	Descricao        string `gorm:"size:200;not null;index" json:"descricao"`
	CategoriaID      uint   `gorm:"not null;index" json:"categoria_id"`
	Categoria        *Categoria `gorm:"foreignKey:CategoriaID" json:"categoria,omitempty"`

	// Identificação patrimonial (opcionais para consumíveis).
	NumeroPatrimonio *string `gorm:"size:60;uniqueIndex" json:"numero_patrimonio,omitempty"`
	NumeroSerie      *string `gorm:"size:80;index" json:"numero_serie,omitempty"`
	Marca            string  `gorm:"size:80" json:"marca"`
	Modelo           string  `gorm:"size:80" json:"modelo"`

	EstadoConservacao EstadoConservacao `gorm:"size:20;not null;default:bom;index" json:"estado_conservacao"`

	// Controle de estoque (relevante para consumíveis; patrimoniados = 1).
	Quantidade    int `gorm:"not null;default:0" json:"quantidade"`
	EstoqueMinimo int `gorm:"not null;default:0" json:"estoque_minimo"`

	// Localização e responsabilidade atuais.
	SetorID            *uint     `gorm:"index" json:"setor_id,omitempty"`
	Setor              *Setor    `gorm:"foreignKey:SetorID" json:"setor,omitempty"`
	ResponsavelID      *uint     `gorm:"index" json:"responsavel_id,omitempty"`
	Responsavel        *Servidor `gorm:"foreignKey:ResponsavelID" json:"responsavel,omitempty"`

	// Dados de aquisição (valor opcional).
	DataAquisicao *time.Time `json:"data_aquisicao,omitempty"`
	Valor         *float64   `json:"valor,omitempty"`

	// Flag de baixa patrimonial (mantém item visível, rastreável).
	Baixado    bool       `gorm:"not null;default:false;index" json:"baixado"`
	DataBaixa  *time.Time `json:"data_baixa,omitempty"`
	MotivoBaixa string    `gorm:"size:255" json:"motivo_baixa,omitempty"`

	// Soft delete EXCLUSIVO para correção de cadastro errado (admin), e somente
	// quando o item NÃO tem histórico (movimentações/termos). Não confundir com
	// baixa patrimonial: a baixa preserva o item visível e rastreável; este
	// soft delete remove da lista um registro criado por engano. GORM filtra
	// automaticamente registros com DeletedAt preenchido.
	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Item) TableName() string { return "itens" }

// EstoqueAbaixoDoMinimo indica se o item (consumível) está em alerta.
func (i *Item) EstoqueAbaixoDoMinimo() bool {
	return i.EstoqueMinimo > 0 && i.Quantidade < i.EstoqueMinimo
}
