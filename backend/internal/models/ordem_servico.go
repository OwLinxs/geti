package models

import (
	"time"

	"gorm.io/gorm"
)

// StatusOS descreve a fase da ordem de serviço na fila de manutenção.
type StatusOS string

const (
	OSAberta         StatusOS = "aberta"
	OSEmAndamento    StatusOS = "em_andamento"
	OSAguardandoPeca StatusOS = "aguardando_peca"
	OSConcluida      StatusOS = "concluida"
	OSCancelada      StatusOS = "cancelada"
)

// StatusOSValido valida o status informado.
func StatusOSValido(s StatusOS) bool {
	switch s {
	case OSAberta, OSEmAndamento, OSAguardandoPeca, OSConcluida, OSCancelada:
		return true
	}
	return false
}

// PrioridadeOS classifica a urgência do atendimento.
type PrioridadeOS string

const (
	PrioridadeBaixa  PrioridadeOS = "baixa"
	PrioridadeNormal PrioridadeOS = "normal"
	PrioridadeAlta   PrioridadeOS = "alta"
)

// PrioridadeOSValida valida a prioridade informada.
func PrioridadeOSValida(p PrioridadeOS) bool {
	switch p {
	case PrioridadeBaixa, PrioridadeNormal, PrioridadeAlta:
		return true
	}
	return false
}

// PassosPadraoManutencao é o template inicial semeado em toda nova OS. É
// editável por OS (adicionar/remover/marcar) e serve de checklist no documento
// impresso. Mantido como constante para simplicidade; pode virar configurável
// (tabela de modelos) sem alterar o restante do módulo.
var PassosPadraoManutencao = []string{
	"Registrar backup dos dados do usuário (quando aplicável)",
	"Limpeza física interna e externa do equipamento",
	"Verificar e reassentar conexões (memória, cabos, armazenamento)",
	"Teste de memória RAM",
	"Teste de disco / unidade de armazenamento (SMART)",
	"Verificação e remoção de malware",
	"Formatação e reinstalação do sistema (quando necessário)",
	"Instalação de atualizações e drivers",
	"Instalação dos softwares padrão do órgão",
	"Restaurar backup dos dados do usuário (quando aplicável)",
	"Teste final de funcionamento",
}

// OrdemServico representa uma ordem de serviço de manutenção de equipamento.
// O equipamento pode ser um item do inventário (ItemID) OU uma máquina externa
// (não cadastrada) descrita em texto livre.
//
// Guardamos snapshots do equipamento e do solicitante no momento da abertura
// para que o documento permaneça fiel mesmo que cadastros mudem depois.
//
// Regra de negócio: a OS NÃO altera o estoque/quantidade do item — manutenção
// não é movimentação. Apenas registra o atendimento e gera o documento.
type OrdemServico struct {
	Base

	Numero string `gorm:"size:40;uniqueIndex;not null" json:"numero"` // ex.: OS-2026-0001

	// Origem do chamado: "interno" (aberto no sistema) ou "whatsapp"/externo
	// (criado pela API de integração).
	Origem string `gorm:"size:20;not null;default:interno;index" json:"origem"`
	// ReferenciaExterna guarda o id do card na plataforma externa, garantindo
	// idempotência do sync (criar/atualizar sem duplicar).
	ReferenciaExterna string `gorm:"size:100;index" json:"referencia_externa,omitempty"`

	// Assunto: título curto do chamado (usado quando aberto pelo solicitante,
	// sem equipamento vinculado).
	Assunto string `gorm:"size:150" json:"assunto,omitempty"`

	// Categoria do chamado (rede, e-mail, impressora...), configurável.
	CategoriaChamadoID *uint             `gorm:"index" json:"categoria_chamado_id,omitempty"`
	CategoriaChamado   *CategoriaChamado `gorm:"foreignKey:CategoriaChamadoID" json:"categoria_chamado,omitempty"`

	// Equipamento: do inventário (opcional) ou externo (descrito em texto).
	// Ponteiro para gravar NULL (e não 0) quando for máquina externa, evitando
	// violar a foreign key para itens.
	ItemID                   *uint  `gorm:"index" json:"item_id,omitempty"`
	Item                     *Item  `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	EquipamentoDescricao     string `gorm:"size:200" json:"equipamento_descricao,omitempty"`
	EquipamentoIdentificacao string `gorm:"size:80" json:"equipamento_identificacao,omitempty"` // série/tag/plaqueta
	EquipamentoSnapshot      string `gorm:"size:200" json:"equipamento_snapshot"`
	PatrimonioSnapshot       string `gorm:"size:60" json:"patrimonio_snapshot,omitempty"`

	SetorID *uint  `gorm:"index" json:"setor_id,omitempty"`
	Setor   *Setor `gorm:"foreignKey:SetorID" json:"setor,omitempty"`

	SolicitanteID           *uint     `gorm:"index" json:"solicitante_id,omitempty"`
	Solicitante             *Servidor `gorm:"foreignKey:SolicitanteID" json:"solicitante,omitempty"`
	SolicitanteNomeSnapshot string    `gorm:"size:120" json:"solicitante_nome_snapshot,omitempty"`
	// Contato do solicitante externo (telefone/WhatsApp), quando não é servidor.
	SolicitanteContato string `gorm:"size:80" json:"solicitante_contato,omitempty"`

	DefeitoRelatado string `gorm:"size:1000;not null" json:"defeito_relatado"`
	Diagnostico     string `gorm:"size:1000" json:"diagnostico,omitempty"`
	SolucaoAplicada string `gorm:"size:1000" json:"solucao_aplicada,omitempty"`

	Prioridade PrioridadeOS `gorm:"size:20;not null;default:normal;index" json:"prioridade"`
	Status     StatusOS     `gorm:"size:20;not null;default:aberta;index" json:"status"`

	TecnicoID *uint    `gorm:"index" json:"tecnico_id,omitempty"`
	Tecnico   *Usuario `gorm:"foreignKey:TecnicoID" json:"tecnico,omitempty"`

	// Autor da abertura. Chamados externos (integração) usam o administrador do
	// sistema como autor, preservando a restrição NOT NULL e a foreign key.
	AbertoPorID uint     `gorm:"not null;index" json:"aberto_por_id"`
	AbertoPor   *Usuario `gorm:"foreignKey:AbertoPorID" json:"aberto_por,omitempty"`

	DataAbertura  time.Time  `gorm:"not null;index" json:"data_abertura"`
	DataConclusao *time.Time `json:"data_conclusao,omitempty"`

	// SLA (prazos calculados na abertura a partir da configuração; nil = sem
	// SLA). PrimeiraRespostaEm marca quando a equipe respondeu pela 1ª vez.
	PrazoRespostaEm    *time.Time `gorm:"index" json:"prazo_resposta_em,omitempty"`
	PrazoResolucaoEm   *time.Time `gorm:"index" json:"prazo_resolucao_em,omitempty"`
	PrimeiraRespostaEm *time.Time `json:"primeira_resposta_em,omitempty"`
	// Flags para o monitor de SLA não notificar o mesmo estouro repetidamente.
	SLARespostaNotificada  bool `gorm:"not null;default:false" json:"-"`
	SLAResolucaoNotificada bool `gorm:"not null;default:false" json:"-"`

	// Avaliação do solicitante ao encerrar (nota 1–5 + comentário).
	AvaliacaoNota       *int       `json:"avaliacao_nota,omitempty"`
	AvaliacaoComentario string     `gorm:"size:1000" json:"avaliacao_comentario,omitempty"`
	AvaliadoEm          *time.Time `json:"avaliado_em,omitempty"`

	Passos []OrdemServicoPasso `gorm:"foreignKey:OrdemServicoID;constraint:OnDelete:CASCADE" json:"passos,omitempty"`

	// Soft delete para correção de cadastro aberto por engano (admin), somente
	// quando a OS ainda não teve andamento (status aberta/cancelada).
	DeletadoEm gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OrdemServico) TableName() string { return "ordens_servico" }

// OrdemServicoPasso é um item do checklist de manutenção de uma OS.
type OrdemServicoPasso struct {
	Base

	OrdemServicoID uint   `gorm:"not null;index" json:"ordem_servico_id"`
	Ordem          int    `gorm:"not null;default:0" json:"ordem"`
	Descricao      string `gorm:"size:300;not null" json:"descricao"`
	Concluido      bool   `gorm:"not null;default:false" json:"concluido"`
	Observacao     string `gorm:"size:300" json:"observacao,omitempty"`
}

func (OrdemServicoPasso) TableName() string { return "ordens_servico_passos" }
