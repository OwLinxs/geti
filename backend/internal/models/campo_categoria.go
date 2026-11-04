package models

// TipoCampo define o tipo de dado de um campo customizado de categoria.
type TipoCampo string

const (
	CampoTexto    TipoCampo = "texto"
	CampoNumero   TipoCampo = "numero"
	CampoLista    TipoCampo = "lista"
	CampoData     TipoCampo = "data"
	CampoBooleano TipoCampo = "booleano"
)

// TipoCampoValido valida o tipo informado.
func TipoCampoValido(t TipoCampo) bool {
	switch t {
	case CampoTexto, CampoNumero, CampoLista, CampoData, CampoBooleano:
		return true
	}
	return false
}

// CampoCategoria é a definição de um campo customizado (atributo) aplicável aos
// itens de uma categoria. Ex.: categoria "Desktop" com os campos Processador,
// Memória (número) e Tipo de disco (lista).
type CampoCategoria struct {
	Base
	// Índice composto único (categoria_id, chave): a mesma chave não se repete
	// dentro de uma categoria.
	CategoriaID uint      `gorm:"not null;index:idx_campo_cat_chave,unique,priority:1" json:"categoria_id"`
	Chave       string    `gorm:"size:80;not null;index:idx_campo_cat_chave,unique,priority:2" json:"chave"`
	Nome        string    `gorm:"size:80;not null" json:"nome"`
	Tipo        TipoCampo `gorm:"size:20;not null;default:texto" json:"tipo"`
	// Opcoes guarda as opções do tipo "lista", uma por linha.
	Opcoes      string `gorm:"size:1000" json:"opcoes,omitempty"`
	Obrigatorio bool   `gorm:"not null;default:false" json:"obrigatorio"`
	Ordem       int    `gorm:"not null;default:0" json:"ordem"`
}

func (CampoCategoria) TableName() string { return "campos_categoria" }

// ItemAtributo guarda o valor de um campo customizado para um item específico.
// O valor é armazenado sempre como texto; a conversão (número, data, booleano)
// é responsabilidade da camada de apresentação, conforme o Tipo do campo.
type ItemAtributo struct {
	Base
	ItemID  uint            `gorm:"not null;index:idx_item_campo,unique,priority:1" json:"item_id"`
	CampoID uint            `gorm:"not null;index:idx_item_campo,unique,priority:2" json:"campo_id"`
	Campo   *CampoCategoria `gorm:"foreignKey:CampoID" json:"campo,omitempty"`
	Valor   string          `gorm:"size:1000" json:"valor"`
}

func (ItemAtributo) TableName() string { return "item_atributos" }
