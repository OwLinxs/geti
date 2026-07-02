import { TrendingDown, TrendingUp } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EstadoVazio } from "@/components/EstadoVazio";
import { formatarDataHora } from "@/lib/format";
import { ehEntrada, rotuloTipoMov } from "@/lib/rotulos";
import type { Movimentacao } from "@/types";

interface HistoricoTabelaProps {
  movimentacoes: Movimentacao[];
  // Mostra a coluna do item (na tela geral de movimentações).
  mostrarItem?: boolean;
}

export function HistoricoTabela({
  movimentacoes,
  mostrarItem = false,
}: HistoricoTabelaProps) {
  if (movimentacoes.length === 0) {
    return <EstadoVazio titulo="Nenhuma movimentação registrada" />;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Data</TableHead>
          {mostrarItem && <TableHead>Item</TableHead>}
          <TableHead>Tipo</TableHead>
          <TableHead className="text-center">Qtd.</TableHead>
          <TableHead className="text-center">Saldo</TableHead>
          <TableHead>Origem / Destino</TableHead>
          <TableHead>Responsável</TableHead>
          <TableHead>Registrado por</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {movimentacoes.map((m) => {
          const entrada = ehEntrada(m.tipo);
          return (
            <TableRow key={m.id}>
              <TableCell className="whitespace-nowrap text-muted-foreground">
                {formatarDataHora(m.data_evento)}
              </TableCell>
              {mostrarItem && (
                <TableCell className="font-medium">
                  {m.item?.descricao ?? `Item #${m.item_id}`}
                </TableCell>
              )}
              <TableCell>
                <span
                  className={
                    entrada
                      ? "inline-flex items-center gap-1 text-emerald-700"
                      : "inline-flex items-center gap-1 text-rose-700"
                  }
                >
                  {entrada ? (
                    <TrendingUp className="h-3.5 w-3.5" />
                  ) : (
                    <TrendingDown className="h-3.5 w-3.5" />
                  )}
                  {rotuloTipoMov(m.tipo)}
                </span>
              </TableCell>
              <TableCell className="text-center font-medium">
                {entrada ? "+" : "−"}
                {m.quantidade}
              </TableCell>
              <TableCell className="text-center text-muted-foreground">
                {m.saldo_resultante}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {origemDestino(m)}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {m.servidor?.nome ?? "—"}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {m.registrado_por?.nome ?? "—"}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}

function origemDestino(m: Movimentacao): string {
  const origem = m.setor_origem?.nome ?? m.origem_descricao;
  const destino = m.setor_destino?.nome ?? m.destino_descricao;
  if (origem && destino) return `${origem} → ${destino}`;
  if (destino) return `→ ${destino}`;
  if (origem) return `${origem} →`;
  return "—";
}
