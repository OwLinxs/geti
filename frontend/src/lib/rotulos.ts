import type {
  EstadoConservacao,
  Perfil,
  TipoMovimentacao,
} from "@/types";

// Rótulos legíveis em PT-BR para os enums do backend.

export const ESTADOS_CONSERVACAO: { valor: EstadoConservacao; rotulo: string }[] =
  [
    { valor: "novo", rotulo: "Novo" },
    { valor: "bom", rotulo: "Bom" },
    { valor: "regular", rotulo: "Regular" },
    { valor: "inservivel", rotulo: "Inservível" },
  ];

export function rotuloEstado(e: EstadoConservacao): string {
  return ESTADOS_CONSERVACAO.find((x) => x.valor === e)?.rotulo ?? e;
}

export const PERFIS: { valor: Perfil; rotulo: string }[] = [
  { valor: "administrador", rotulo: "Administrador" },
  { valor: "operador", rotulo: "Operador" },
];

export function rotuloPerfil(p: Perfil): string {
  return PERFIS.find((x) => x.valor === p)?.rotulo ?? p;
}

export const TIPOS_ENTRADA: { valor: TipoMovimentacao; rotulo: string }[] = [
  { valor: "entrada_compra", rotulo: "Entrada — Compra" },
  { valor: "entrada_doacao", rotulo: "Entrada — Doação" },
  { valor: "entrada_devolucao", rotulo: "Entrada — Devolução" },
];

export const TIPOS_SAIDA: { valor: TipoMovimentacao; rotulo: string }[] = [
  { valor: "saida_emprestimo", rotulo: "Saída — Empréstimo" },
  { valor: "saida_transferencia", rotulo: "Saída — Transferência" },
  { valor: "saida_descarte", rotulo: "Saída — Descarte / Baixa" },
];

export const TIPOS_MOVIMENTACAO = [...TIPOS_ENTRADA, ...TIPOS_SAIDA];

export function rotuloTipoMov(t: TipoMovimentacao): string {
  return TIPOS_MOVIMENTACAO.find((x) => x.valor === t)?.rotulo ?? t;
}

export function ehEntrada(t: TipoMovimentacao): boolean {
  return t.startsWith("entrada_");
}

export function ehBaixa(t: TipoMovimentacao): boolean {
  return t === "saida_descarte";
}
