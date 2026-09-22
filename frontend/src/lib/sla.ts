import type { OrdemServico } from "@/types";

type VarianteBadge =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "success"
  | "warning"
  | "muted";

export interface EstadoSLA {
  codigo: "no_prazo" | "resposta_atrasada" | "resolucao_atrasada";
  rotulo: string;
  variante: VarianteBadge;
}

// estadoSLA avalia o SLA de um chamado (a partir dos prazos gravados + agora).
// Retorna null quando não há SLA aplicável (chamado fechado ou sem prazos).
export function estadoSLA(os: OrdemServico): EstadoSLA | null {
  if (os.status === "concluida" || os.status === "cancelada") return null;

  const agora = Date.now();
  const prazoResol = os.prazo_resolucao_em
    ? new Date(os.prazo_resolucao_em).getTime()
    : null;
  const prazoResp = os.prazo_resposta_em
    ? new Date(os.prazo_resposta_em).getTime()
    : null;

  if (prazoResol && agora > prazoResol) {
    return {
      codigo: "resolucao_atrasada",
      rotulo: "Resolução atrasada",
      variante: "destructive",
    };
  }
  if (!os.primeira_resposta_em && prazoResp && agora > prazoResp) {
    return {
      codigo: "resposta_atrasada",
      rotulo: "Resposta atrasada",
      variante: "warning",
    };
  }
  if (prazoResol || prazoResp) {
    return { codigo: "no_prazo", rotulo: "No prazo", variante: "success" };
  }
  return null;
}

// slaAtrasado indica se há qualquer atraso (para contadores/filtros).
export function slaAtrasado(os: OrdemServico): boolean {
  const e = estadoSLA(os);
  return e?.codigo === "resposta_atrasada" || e?.codigo === "resolucao_atrasada";
}
