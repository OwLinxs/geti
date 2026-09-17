package models

import "time"

// DirecaoMensagem indica o sentido da mensagem no chamado.
type DirecaoMensagem string

const (
	MsgSaida   DirecaoMensagem = "saida"   // do sistema/técnico para o servidor
	MsgEntrada DirecaoMensagem = "entrada" // do servidor para o sistema (via WhatsApp, futuro)
	MsgInterna DirecaoMensagem = "interna" // nota interna da equipe (não vai ao servidor)
)

// AutorTipo classifica quem escreveu a mensagem.
type AutorTipo string

const (
	AutorTecnico  AutorTipo = "tecnico"
	AutorServidor AutorTipo = "servidor"
	AutorSistema  AutorTipo = "sistema"
)

// StatusMensagem acompanha a entrega (relevante quando integrado ao WhatsApp).
type StatusMensagem string

const (
	MsgEnviada  StatusMensagem = "enviado"
	MsgEntregue StatusMensagem = "entregue"
	MsgLida     StatusMensagem = "lido"
	MsgRecebida StatusMensagem = "recebido"
	MsgFalhou   StatusMensagem = "falhou"
)

// MensagemChamado é uma mensagem da conversa vinculada a uma ordem de serviço
// (chamado). Modelada para suportar, no futuro, o transporte via WhatsApp
// (IdExterno, Status, Direcao) sem alterar o restante do módulo.
type MensagemChamado struct {
	Base

	OrdemServicoID uint            `gorm:"not null;index" json:"ordem_servico_id"`
	Direcao        DirecaoMensagem `gorm:"size:10;not null;default:saida;index" json:"direcao"`
	AutorTipo      AutorTipo       `gorm:"size:10;not null;default:tecnico" json:"autor_tipo"`

	// Autor interno (usuário) quando a mensagem parte da equipe.
	AutorID   *uint    `gorm:"index" json:"autor_id,omitempty"`
	Autor     *Usuario `gorm:"foreignKey:AutorID" json:"autor,omitempty"`
	AutorNome string   `gorm:"size:120" json:"autor_nome"`

	Texto   string         `gorm:"type:text" json:"texto"`
	Interna bool           `gorm:"not null;default:false;index" json:"interna"`
	Status  StatusMensagem `gorm:"size:12;not null;default:enviado" json:"status"`

	// Correlação com a plataforma externa (id da mensagem no WhatsApp/mensageiro).
	IdExterno string `gorm:"size:120;index" json:"id_externo,omitempty"`

	// Anexo opcional (foto do equipamento, documento). Guardado em disco.
	AnexoNome    string `gorm:"size:200" json:"anexo_nome,omitempty"`
	AnexoTipo    string `gorm:"size:100" json:"anexo_tipo,omitempty"`
	AnexoCaminho string `gorm:"size:400" json:"-"`
	AnexoTamanho int64  `gorm:"default:0" json:"anexo_tamanho,omitempty"`

	EnviadaEm time.Time `gorm:"not null;index" json:"enviada_em"`
}

func (MensagemChamado) TableName() string { return "mensagens_chamado" }

// TemAnexo informa se a mensagem carrega um arquivo.
func (m *MensagemChamado) TemAnexo() bool { return m.AnexoCaminho != "" }
