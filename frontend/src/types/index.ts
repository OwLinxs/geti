// Tipos TypeScript espelhando os DTOs/JSON do backend Go (SIGE-TI).
// As tags `json` do backend ditam os nomes em snake_case usados aqui.

export type Perfil = "administrador" | "operador" | "solicitante";

export type EstadoConservacao = "novo" | "bom" | "regular" | "inservivel";

export type TipoMovimentacao =
  | "entrada_compra"
  | "entrada_doacao"
  | "entrada_devolucao"
  | "saida_emprestimo"
  | "saida_transferencia"
  | "saida_descarte";

// Campos de auditoria comuns (models.Base).
export interface Base {
  id: number;
  criado_em: string;
  atualizado_em: string;
}

export interface Usuario extends Base {
  nome: string;
  email: string;
  perfil: Perfil;
  ativo: boolean;
}

export interface Categoria extends Base {
  nome: string;
  descricao: string;
  consumivel: boolean;
}

export interface Setor extends Base {
  nome: string;
  sigla: string;
  localizacao: string;
  // Hierarquia: unidade superior (null = topo/Secretaria).
  pai_id?: number | null;
  pai?: Setor | null;
}

export interface Servidor extends Base {
  nome: string;
  matricula: string;
  setor_id?: number | null;
  setor?: Setor | null;
  ativo: boolean;
}

export interface Item extends Base {
  descricao: string;
  categoria_id: number;
  categoria?: Categoria | null;
  numero_patrimonio?: string | null;
  numero_serie?: string | null;
  marca: string;
  modelo: string;
  estado_conservacao: EstadoConservacao;
  quantidade: number;
  estoque_minimo: number;
  setor_id?: number | null;
  setor?: Setor | null;
  responsavel_id?: number | null;
  responsavel?: Servidor | null;
  data_aquisicao?: string | null;
  valor?: number | null;
  baixado: boolean;
  data_baixa?: string | null;
  motivo_baixa?: string;
}

export interface Movimentacao extends Base {
  item_id: number;
  item?: Item | null;
  tipo: TipoMovimentacao;
  quantidade: number;
  saldo_resultante: number;
  setor_origem_id?: number | null;
  setor_origem?: Setor | null;
  setor_destino_id?: number | null;
  setor_destino?: Setor | null;
  servidor_id?: number | null;
  servidor?: Servidor | null;
  registrado_por_id: number;
  registrado_por?: Usuario | null;
  origem_descricao?: string;
  destino_descricao?: string;
  observacao?: string;
  data_evento: string;
}

export interface TermoResponsabilidade extends Base {
  numero: string;
  item_id: number;
  item?: Item | null;
  servidor_id: number;
  servidor?: Servidor | null;
  movimentacao_id?: number | null;
  emitido_por_id: number;
  emitido_por?: Usuario | null;
  item_descricao_snapshot: string;
  patrimonio_snapshot?: string;
  servidor_nome_snapshot: string;
  servidor_matricula_snapshot: string;
  observacao?: string;
  data_emissao: string;
}

// ===== Manutenção / Ordem de Serviço =====

export type StatusOS =
  | "aberta"
  | "em_andamento"
  | "aguardando_peca"
  | "concluida"
  | "cancelada";

export type PrioridadeOS = "baixa" | "normal" | "alta";

export interface OrdemServicoPasso extends Base {
  ordem_servico_id: number;
  ordem: number;
  descricao: string;
  concluido: boolean;
  observacao?: string;
}

export interface OrdemServico extends Base {
  numero: string;
  origem?: string;
  referencia_externa?: string;
  assunto?: string;
  solicitante_contato?: string;
  item_id?: number;
  item?: Item | null;
  equipamento_descricao?: string;
  equipamento_identificacao?: string;
  equipamento_snapshot: string;
  patrimonio_snapshot?: string;
  setor_id?: number | null;
  setor?: Setor | null;
  solicitante_id?: number | null;
  solicitante?: Servidor | null;
  solicitante_nome_snapshot?: string;
  defeito_relatado: string;
  diagnostico?: string;
  solucao_aplicada?: string;
  prioridade: PrioridadeOS;
  status: StatusOS;
  tecnico_id?: number | null;
  tecnico?: Usuario | null;
  aberto_por_id: number;
  aberto_por?: Usuario | null;
  data_abertura: string;
  data_conclusao?: string | null;
  passos?: OrdemServicoPasso[];
}

// ===== Conversa do chamado (mensagens) =====

export type DirecaoMensagem = "saida" | "entrada" | "interna";
export type AutorTipoMensagem = "tecnico" | "servidor" | "sistema";

export interface MensagemChamado extends Base {
  ordem_servico_id: number;
  direcao: DirecaoMensagem;
  autor_tipo: AutorTipoMensagem;
  autor_id?: number | null;
  autor?: Usuario | null;
  autor_nome: string;
  texto: string;
  interna: boolean;
  status: string;
  anexo_nome?: string;
  anexo_tipo?: string;
  anexo_tamanho?: number;
  enviada_em: string;
}

// ===== Base de conhecimento =====

export interface ArtigoConhecimento extends Base {
  titulo: string;
  categoria?: string;
  conteudo: string;
  autor_id?: number | null;
  autor?: Usuario | null;
  publicado: boolean;
  visualizacoes: number;
}

// ===== Fornecedores e contratos =====

export interface Fornecedor extends Base {
  nome: string;
  cnpj?: string;
  email?: string;
  telefone?: string;
  endereco?: string;
  observacao?: string;
}

export type StatusContrato = "vigente" | "encerrado" | "cancelado";

export interface Contrato extends Base {
  numero: string;
  fornecedor_id: number;
  fornecedor?: Fornecedor | null;
  objeto: string;
  data_inicio?: string | null;
  data_fim?: string | null;
  valor?: number | null;
  status: StatusContrato;
  observacao?: string;
}

// ===== Reservas de equipamento =====

export type StatusReserva = "reservada" | "em_uso" | "devolvida" | "cancelada";

export interface Reserva extends Base {
  item_id: number;
  item?: Item | null;
  solicitante_id?: number | null;
  solicitante?: Servidor | null;
  data_inicio: string;
  data_fim: string;
  finalidade?: string;
  status: StatusReserva;
  aprovado_por_id?: number | null;
  aprovado_por?: Usuario | null;
  observacao?: string;
}

// ===== Respostas =====

export interface ResultadoLogin {
  token: string;
  expira_em: string;
  usuario: Usuario;
}

// Resposta paginada (itens e movimentações).
export interface RespostaPaginada<T> {
  dados: T[];
  total: number;
  pagina: number;
  tamanho: number;
}

// Registro da trilha de auditoria (admin).
export interface RegistroAuditoria {
  id: number;
  criado_em: string;
  usuario_id?: number | null;
  usuario_nome: string;
  usuario_email?: string;
  acao: string;
  recurso?: string;
  recurso_id?: number | null;
  metodo?: string;
  caminho?: string;
  status: number;
  ip?: string;
  detalhe?: string;
}

// Erro de uma linha na importação em massa de itens.
export interface ErroLinhaImportacao {
  linha: number;
  descricao: string;
  mensagem: string;
}

// Resultado da importação (ou simulação) de itens via CSV.
export interface ResultadoImportacao {
  validacao: boolean;
  total: number;
  importados: number;
  validas: number;
  erros: ErroLinhaImportacao[];
}

// Resposta do registro de movimentação.
export interface ResultadoMovimentacao {
  movimentacao: Movimentacao;
  item: Item;
  alerta_estoque: boolean;
}

// Erro de validação por campo (HTTP 422).
export interface ErroValidacao {
  erro: string;
  campos?: Record<string, string>;
}

// ===== Payloads de entrada =====

export interface ItemPayload {
  descricao: string;
  categoria_id: number;
  numero_patrimonio?: string | null;
  numero_serie?: string | null;
  marca: string;
  modelo: string;
  estado_conservacao: EstadoConservacao;
  quantidade: number;
  estoque_minimo: number;
  setor_id?: number | null;
  responsavel_id?: number | null;
  data_aquisicao?: string | null;
  valor?: number | null;
}

export interface MovimentacaoPayload {
  item_id: number;
  tipo: TipoMovimentacao;
  quantidade: number;
  setor_origem_id?: number | null;
  setor_destino_id?: number | null;
  servidor_id?: number | null;
  observacao?: string;
  motivo_baixa?: string;
  data_evento?: string | null;
}

export interface CategoriaPayload {
  nome: string;
  descricao: string;
  consumivel: boolean;
}

export interface SetorPayload {
  nome: string;
  sigla: string;
  localizacao: string;
  pai_id?: number | null;
}

export interface ServidorPayload {
  nome: string;
  matricula: string;
  setor_id?: number | null;
  ativo?: boolean;
}

export interface UsuarioPayload {
  nome: string;
  email: string;
  senha: string;
  perfil: Perfil;
  ativo?: boolean;
}

export interface TermoPayload {
  item_id: number;
  servidor_id: number;
  movimentacao_id?: number | null;
  observacao?: string;
}

export interface OrdemServicoPayload {
  item_id?: number;
  equipamento_descricao?: string;
  equipamento_identificacao?: string;
  setor_id?: number | null;
  solicitante_id?: number | null;
  defeito_relatado: string;
  diagnostico?: string;
  solucao_aplicada?: string;
  prioridade: PrioridadeOS;
  tecnico_id?: number | null;
}

export interface PassoPayload {
  descricao: string;
  concluido: boolean;
  observacao?: string;
}

export interface ArtigoConhecimentoPayload {
  titulo: string;
  categoria?: string;
  conteudo: string;
  publicado?: boolean;
}

export interface FornecedorPayload {
  nome: string;
  cnpj?: string;
  email?: string;
  telefone?: string;
  endereco?: string;
  observacao?: string;
}

export interface ContratoPayload {
  numero: string;
  fornecedor_id: number;
  objeto: string;
  data_inicio?: string | null;
  data_fim?: string | null;
  valor?: number | null;
  status?: StatusContrato;
  observacao?: string;
}

export interface ReservaPayload {
  item_id: number;
  solicitante_id?: number | null;
  data_inicio?: string | null;
  data_fim?: string | null;
  finalidade?: string;
  status?: StatusReserva;
  observacao?: string;
}
