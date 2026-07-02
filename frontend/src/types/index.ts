// Tipos TypeScript espelhando os DTOs/JSON do backend Go (SIGE-TI).
// As tags `json` do backend ditam os nomes em snake_case usados aqui.

export type Perfil = "administrador" | "operador";

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
