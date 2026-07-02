import { api } from "./client";
import type {
  Categoria,
  CategoriaPayload,
  Item,
  ItemPayload,
  Movimentacao,
  MovimentacaoPayload,
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
  euMesmo: () => api.get<Usuario>("/auth/eu").then((r) => r.data),
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

// ===== Usuários (admin) =====
export const usuariosApi = {
  listar: () => api.get<Usuario[]>("/usuarios").then((r) => r.data),
  criar: (p: UsuarioPayload) =>
    api.post<Usuario>("/usuarios", p).then((r) => r.data),
  redefinirSenha: (id: number, senha: string) =>
    api.patch<Usuario>(`/usuarios/${id}/senha`, { senha }).then((r) => r.data),
  definirAtivo: (id: number, ativo: boolean) =>
    api.patch<Usuario>(`/usuarios/${id}/ativo`, { ativo }).then((r) => r.data),
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
