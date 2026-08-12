import type {
  EstadoConservacao,
  Perfil,
  PrioridadeOS,
  StatusContrato,
  StatusOS,
  StatusReserva,
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

// ===== Manutenção / Ordem de Serviço =====

type VarianteBadge =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "success"
  | "warning"
  | "muted";

export const STATUS_OS: {
  valor: StatusOS;
  rotulo: string;
  variante: VarianteBadge;
}[] = [
  { valor: "aberta", rotulo: "Aberta", variante: "secondary" },
  { valor: "em_andamento", rotulo: "Em andamento", variante: "default" },
  { valor: "aguardando_peca", rotulo: "Aguardando peça", variante: "warning" },
  { valor: "concluida", rotulo: "Concluída", variante: "success" },
  { valor: "cancelada", rotulo: "Cancelada", variante: "muted" },
];

export function rotuloStatusOS(s: StatusOS): string {
  return STATUS_OS.find((x) => x.valor === s)?.rotulo ?? s;
}

export function varianteStatusOS(s: StatusOS): VarianteBadge {
  return STATUS_OS.find((x) => x.valor === s)?.variante ?? "secondary";
}

export const PRIORIDADES_OS: {
  valor: PrioridadeOS;
  rotulo: string;
  variante: VarianteBadge;
}[] = [
  { valor: "baixa", rotulo: "Baixa", variante: "muted" },
  { valor: "normal", rotulo: "Normal", variante: "secondary" },
  { valor: "alta", rotulo: "Alta", variante: "destructive" },
];

export function rotuloPrioridadeOS(p: PrioridadeOS): string {
  return PRIORIDADES_OS.find((x) => x.valor === p)?.rotulo ?? p;
}

export function variantePrioridadeOS(p: PrioridadeOS): VarianteBadge {
  return PRIORIDADES_OS.find((x) => x.valor === p)?.variante ?? "secondary";
}

// ===== Contratos =====

export const STATUS_CONTRATO: {
  valor: StatusContrato;
  rotulo: string;
  variante: VarianteBadge;
}[] = [
  { valor: "vigente", rotulo: "Vigente", variante: "success" },
  { valor: "encerrado", rotulo: "Encerrado", variante: "muted" },
  { valor: "cancelado", rotulo: "Cancelado", variante: "destructive" },
];

export function rotuloStatusContrato(s: StatusContrato): string {
  return STATUS_CONTRATO.find((x) => x.valor === s)?.rotulo ?? s;
}

export function varianteStatusContrato(s: StatusContrato): VarianteBadge {
  return STATUS_CONTRATO.find((x) => x.valor === s)?.variante ?? "secondary";
}

// ===== Reservas =====

export const STATUS_RESERVA: {
  valor: StatusReserva;
  rotulo: string;
  variante: VarianteBadge;
}[] = [
  { valor: "reservada", rotulo: "Reservada", variante: "default" },
  { valor: "em_uso", rotulo: "Em uso", variante: "warning" },
  { valor: "devolvida", rotulo: "Devolvida", variante: "success" },
  { valor: "cancelada", rotulo: "Cancelada", variante: "muted" },
];

export function rotuloStatusReserva(s: StatusReserva): string {
  return STATUS_RESERVA.find((x) => x.valor === s)?.rotulo ?? s;
}

export function varianteStatusReserva(s: StatusReserva): VarianteBadge {
  return STATUS_RESERVA.find((x) => x.valor === s)?.variante ?? "secondary";
}
