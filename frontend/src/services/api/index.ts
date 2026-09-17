import { api } from "./client";
import type {
  Categoria,
  CategoriaPayload,
  Item,
  ItemPayload,
  ArtigoConhecimento,
  ArtigoConhecimentoPayload,
  MensagemChamado,
  Contrato,
  ContratoPayload,
  Fornecedor,
  FornecedorPayload,
  Reserva,
  ReservaPayload,
  Perfil,
  Movimentacao,
  MovimentacaoPayload,
  OrdemServico,
  OrdemServicoPayload,
  PassoPayload,
  RegistroAuditoria,
  RespostaPaginada,
  ResultadoImportacao,
  ResultadoLogin,
  ResultadoMovimentacao,
  Servidor,
  ServidorPayload,
  Setor,
  SetorPayload,
  TermoPayload,
  TermoResponsabilidade,
  Usuario,
  UsuarioPayload,
} from "@/types";

// ===== Autenticação =====
export const authApi = {
  login: (email: string, senha: string) =>
    api.post<ResultadoLogin>("/auth/login", { email, senha }).then((r) => r.data),
  registrar: (nome: string, email: string, senha: string) =>
    api
      .post<ResultadoLogin>("/auth/registrar", { nome, email, senha })
      .then((r) => r.data),
  euMesmo: () => api.get<Usuario>("/auth/eu").then((r) => r.data),
};

// ===== Portal do solicitante (meus chamados) =====
export interface AbrirChamadoPayload {
  assunto: string;
  defeito_relatado: string;
  equipamento_descricao?: string;
  prioridade?: string;
}

export const portalApi = {
  listar: (f: { status?: string; pagina?: number; tamanho?: number } = {}) =>
    api
      .get<RespostaPaginada<OrdemServico>>("/meus-chamados", { params: limpar(f) })
      .then((r) => r.data),
  abrir: (p: AbrirChamadoPayload) =>
    api.post<OrdemServico>("/meus-chamados", p).then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<OrdemServico>(`/meus-chamados/${id}`).then((r) => r.data),
};

// ===== Conversa do chamado (equipe usa "ordens-servico"; solicitante "meus-chamados") =====
export type BaseChamado = "ordens-servico" | "meus-chamados";

export const chamadoMensagensApi = {
  listar: (base: BaseChamado, osId: number) =>
    api
      .get<MensagemChamado[]>(`/${base}/${osId}/mensagens`)
      .then((r) => r.data),
  enviar: (
    base: BaseChamado,
    osId: number,
    dados: { texto: string; interna?: boolean; arquivo?: File | null }
  ) => {
    const form = new FormData();
    form.append("texto", dados.texto);
    if (dados.interna) form.append("interna", "true");
    if (dados.arquivo) form.append("arquivo", dados.arquivo);
    return api
      .post<MensagemChamado>(`/${base}/${osId}/mensagens`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      })
      .then((r) => r.data);
  },
  baixarAnexo: (base: BaseChamado, osId: number, msgId: number) =>
    api
      .get(`/${base}/${osId}/mensagens/${msgId}/anexo`, { responseType: "blob" })
      .then((r) => r.data as Blob),
};

// ===== Categorias =====
export const categoriasApi = {
  listar: () => api.get<Categoria[]>("/categorias").then((r) => r.data),
  criar: (p: CategoriaPayload) =>
    api.post<Categoria>("/categorias", p).then((r) => r.data),
  atualizar: (id: number, p: CategoriaPayload) =>
    api.put<Categoria>(`/categorias/${id}`, p).then((r) => r.data),
  remover: (id: number) => api.delete(`/categorias/${id}`).then(() => undefined),
};

// ===== Setores =====
export const setoresApi = {
  listar: () => api.get<Setor[]>("/setores").then((r) => r.data),
  criar: (p: SetorPayload) => api.post<Setor>("/setores", p).then((r) => r.data),
  atualizar: (id: number, p: SetorPayload) =>
    api.put<Setor>(`/setores/${id}`, p).then((r) => r.data),
  remover: (id: number) => api.delete(`/setores/${id}`).then(() => undefined),
};

// ===== Servidores =====
export const servidoresApi = {
  listar: () => api.get<Servidor[]>("/servidores").then((r) => r.data),
  criar: (p: ServidorPayload) =>
    api.post<Servidor>("/servidores", p).then((r) => r.data),
  atualizar: (id: number, p: ServidorPayload) =>
    api.put<Servidor>(`/servidores/${id}`, p).then((r) => r.data),
  remover: (id: number) => api.delete(`/servidores/${id}`).then(() => undefined),
};

// ===== Itens =====
export interface FiltroItens {
  q?: string;
  categoria_id?: number;
  setor_id?: number;
  responsavel_id?: number;
  estado?: string;
  baixado?: boolean;
  abaixo_minimo?: boolean;
  pagina?: number;
  tamanho?: number;
}

export const itensApi = {
  listar: (f: FiltroItens = {}) =>
    api
      .get<RespostaPaginada<Item>>("/itens", { params: limpar(f) })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<Item>(`/itens/${id}`).then((r) => r.data),
  criar: (p: ItemPayload) => api.post<Item>("/itens", p).then((r) => r.data),
  atualizar: (id: number, p: ItemPayload) =>
    api.put<Item>(`/itens/${id}`, p).then((r) => r.data),
  historico: (id: number) =>
    api.get<Movimentacao[]>(`/itens/${id}/historico`).then((r) => r.data),
  alertasEstoqueBaixo: () =>
    api.get<Item[]>("/itens/alertas/estoque-baixo").then((r) => r.data),
  excluir: (id: number) => api.delete(`/itens/${id}`).then(() => undefined),

  // Importação em massa via CSV. validar=true faz apenas simulação (dry-run).
  importar: (arquivo: File, validar: boolean) => {
    const form = new FormData();
    form.append("arquivo", arquivo);
    return api
      .post<ResultadoImportacao>("/itens/importar", form, {
        params: validar ? { validar: "true" } : {},
        headers: { "Content-Type": "multipart/form-data" },
      })
      .then((r) => r.data);
  },
  baixarModeloCsv: () =>
    api
      .get("/itens/modelo-csv", { responseType: "blob" })
      .then((r) => r.data as Blob),
};

// ===== Movimentações =====
export interface FiltroMovimentacoes {
  item_id?: number;
  tipo?: string;
  data_inicio?: string;
  data_fim?: string;
  pagina?: number;
  tamanho?: number;
}

export const movimentacoesApi = {
  listar: (f: FiltroMovimentacoes = {}) =>
    api
      .get<RespostaPaginada<Movimentacao>>("/movimentacoes", { params: limpar(f) })
      .then((r) => r.data),
  registrar: (p: MovimentacaoPayload) =>
    api.post<ResultadoMovimentacao>("/movimentacoes", p).then((r) => r.data),
};

// ===== Termos de responsabilidade =====
export const termosApi = {
  listar: () =>
    api.get<TermoResponsabilidade[]>("/termos").then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<TermoResponsabilidade>(`/termos/${id}`).then((r) => r.data),
  emitir: (p: TermoPayload) =>
    api.post<TermoResponsabilidade>("/termos", p).then((r) => r.data),
  pdfUrl: (id: number) => `/termos/${id}/pdf`,
  baixarPdf: (id: number) =>
    api
      .get(`/termos/${id}/pdf`, { responseType: "blob" })
      .then((r) => r.data as Blob),
};

// ===== Manutenção / Ordens de serviço =====
export interface FiltroOrdensServico {
  status?: string;
  tecnico_id?: number;
  de?: string;
  ate?: string;
  pagina?: number;
  tamanho?: number;
}

export const ordensServicoApi = {
  listar: (f: FiltroOrdensServico = {}) =>
    api
      .get<RespostaPaginada<OrdemServico>>("/ordens-servico", {
        params: limpar(f),
      })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<OrdemServico>(`/ordens-servico/${id}`).then((r) => r.data),
  criar: (p: OrdemServicoPayload) =>
    api.post<OrdemServico>("/ordens-servico", p).then((r) => r.data),
  atualizar: (id: number, p: OrdemServicoPayload) =>
    api.put<OrdemServico>(`/ordens-servico/${id}`, p).then((r) => r.data),
  definirStatus: (id: number, status: string) =>
    api
      .patch<OrdemServico>(`/ordens-servico/${id}/status`, { status })
      .then((r) => r.data),
  salvarPassos: (id: number, passos: PassoPayload[]) =>
    api
      .put<OrdemServico>(`/ordens-servico/${id}/passos`, { passos })
      .then((r) => r.data),
  baixarDocumento: (id: number, passos_extra: string[] = []) =>
    api
      .post(
        `/ordens-servico/${id}/documento`,
        { passos_extra },
        { responseType: "blob" }
      )
      .then((r) => r.data as Blob),
  excluir: (id: number) =>
    api.delete(`/ordens-servico/${id}`).then(() => undefined),
};

// ===== Base de conhecimento =====
export interface FiltroConhecimento {
  q?: string;
  categoria?: string;
  publicado?: boolean;
  pagina?: number;
  tamanho?: number;
}

export const conhecimentoApi = {
  listar: (f: FiltroConhecimento = {}) =>
    api
      .get<RespostaPaginada<ArtigoConhecimento>>("/conhecimento", {
        params: limpar(f),
      })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<ArtigoConhecimento>(`/conhecimento/${id}`).then((r) => r.data),
  criar: (p: ArtigoConhecimentoPayload) =>
    api.post<ArtigoConhecimento>("/conhecimento", p).then((r) => r.data),
  atualizar: (id: number, p: ArtigoConhecimentoPayload) =>
    api.put<ArtigoConhecimento>(`/conhecimento/${id}`, p).then((r) => r.data),
  excluir: (id: number) =>
    api.delete(`/conhecimento/${id}`).then(() => undefined),
};

// ===== Fornecedores =====
export const fornecedoresApi = {
  listar: (q?: string) =>
    api
      .get<Fornecedor[]>("/fornecedores", { params: limpar({ q }) })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<Fornecedor>(`/fornecedores/${id}`).then((r) => r.data),
  criar: (p: FornecedorPayload) =>
    api.post<Fornecedor>("/fornecedores", p).then((r) => r.data),
  atualizar: (id: number, p: FornecedorPayload) =>
    api.put<Fornecedor>(`/fornecedores/${id}`, p).then((r) => r.data),
  excluir: (id: number) =>
    api.delete(`/fornecedores/${id}`).then(() => undefined),
};

// ===== Contratos =====
export interface FiltroContratos {
  q?: string;
  status?: string;
  fornecedor_id?: number;
  vencendo?: number; // dias
  pagina?: number;
  tamanho?: number;
}

export const contratosApi = {
  listar: (f: FiltroContratos = {}) =>
    api
      .get<RespostaPaginada<Contrato>>("/contratos", { params: limpar(f) })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<Contrato>(`/contratos/${id}`).then((r) => r.data),
  criar: (p: ContratoPayload) =>
    api.post<Contrato>("/contratos", p).then((r) => r.data),
  atualizar: (id: number, p: ContratoPayload) =>
    api.put<Contrato>(`/contratos/${id}`, p).then((r) => r.data),
  excluir: (id: number) => api.delete(`/contratos/${id}`).then(() => undefined),
};

// ===== Reservas =====
export interface FiltroReservas {
  item_id?: number;
  status?: string;
  pagina?: number;
  tamanho?: number;
}

export const reservasApi = {
  listar: (f: FiltroReservas = {}) =>
    api
      .get<RespostaPaginada<Reserva>>("/reservas", { params: limpar(f) })
      .then((r) => r.data),
  buscarPorId: (id: number) =>
    api.get<Reserva>(`/reservas/${id}`).then((r) => r.data),
  criar: (p: ReservaPayload) =>
    api.post<Reserva>("/reservas", p).then((r) => r.data),
  atualizar: (id: number, p: ReservaPayload) =>
    api.put<Reserva>(`/reservas/${id}`, p).then((r) => r.data),
  definirStatus: (id: number, status: string) =>
    api
      .patch<Reserva>(`/reservas/${id}/status`, { status })
      .then((r) => r.data),
  excluir: (id: number) => api.delete(`/reservas/${id}`).then(() => undefined),
};

// ===== Usuários (admin) =====
export const usuariosApi = {
  listar: () => api.get<Usuario[]>("/usuarios").then((r) => r.data),
  criar: (p: UsuarioPayload) =>
    api.post<Usuario>("/usuarios", p).then((r) => r.data),
  redefinirSenha: (id: number, senha: string) =>
    api.patch<Usuario>(`/usuarios/${id}/senha`, { senha }).then((r) => r.data),
  definirAtivo: (id: number, ativo: boolean) =>
    api.patch<Usuario>(`/usuarios/${id}/ativo`, { ativo }).then((r) => r.data),
  definirPerfil: (id: number, perfil: Perfil) =>
    api.patch<Usuario>(`/usuarios/${id}/perfil`, { perfil }).then((r) => r.data),
};

// ===== Auditoria (admin) =====
export interface FiltroAuditoria {
  usuario_id?: number;
  recurso?: string;
  acao?: string;
  de?: string;
  ate?: string;
  pagina?: number;
  tamanho?: number;
}

export const auditoriaApi = {
  listar: (f: FiltroAuditoria = {}) =>
    api
      .get<RespostaPaginada<RegistroAuditoria>>("/auditoria", {
        params: limpar(f),
      })
      .then((r) => r.data),
};

// ===== Relatórios =====
export type FormatoRelatorio = "json" | "csv" | "pdf";

export interface ParamsMovimentacoesRelatorio {
  data_inicio?: string;
  data_fim?: string;
  tipo?: string;
}

export const relatoriosApi = {
  itensPorSetor: (setor_id?: number) =>
    api
      .get<Item[]>("/relatorios/itens-por-setor", { params: limpar({ setor_id }) })
      .then((r) => r.data),
  itensPorResponsavel: (responsavel_id?: number) =>
    api
      .get<Item[]>("/relatorios/itens-por-responsavel", {
        params: limpar({ responsavel_id }),
      })
      .then((r) => r.data),
  estoqueBaixo: () =>
    api.get<Item[]>("/relatorios/estoque-baixo").then((r) => r.data),
  inventario: () =>
    api.get<Item[]>("/relatorios/inventario").then((r) => r.data),
  movimentacoes: (p: ParamsMovimentacoesRelatorio) =>
    api
      .get<Movimentacao[]>("/relatorios/movimentacoes", { params: limpar(p) })
      .then((r) => r.data),

  // Download de arquivo (CSV/PDF) como Blob.
  baixar: (
    recurso:
      | "itens-por-setor"
      | "itens-por-responsavel"
      | "estoque-baixo"
      | "inventario"
      | "movimentacoes",
    formato: "csv" | "pdf",
    params: Record<string, unknown> = {}
  ) =>
    api
      .get(`/relatorios/${recurso}`, {
        params: limpar({ ...params, formato }),
        responseType: "blob",
      })
      .then((r) => r.data as Blob),
};

// Remove chaves com valor undefined/null/"" antes de enviar como query.
function limpar<T extends object>(obj: T): Partial<T> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(obj)) {
    if (v !== undefined && v !== null && v !== "") out[k] = v;
  }
  return out as Partial<T>;
}

export { api };
