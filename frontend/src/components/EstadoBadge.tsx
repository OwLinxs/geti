import { Badge } from "@/components/ui/badge";
import { rotuloEstado } from "@/lib/rotulos";
import type { EstadoConservacao } from "@/types";

const VARIANTE: Record<
  EstadoConservacao,
  "success" | "secondary" | "warning" | "destructive"
> = {
  novo: "success",
  bom: "secondary",
  regular: "warning",
  inservivel: "destructive",
};

export function EstadoBadge({ estado }: { estado: EstadoConservacao }) {
  return <Badge variant={VARIANTE[estado]}>{rotuloEstado(estado)}</Badge>;
}
