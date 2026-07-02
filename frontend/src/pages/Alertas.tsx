import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AlertTriangle, Download } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela } from "@/components/ui/spinner";
import { itensApi, relatoriosApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { baixarBlob } from "@/lib/format";
import type { Item } from "@/types";

export default function Alertas() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const [itens, setItens] = React.useState<Item[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [baixando, setBaixando] = React.useState(false);

  React.useEffect(() => {
    itensApi
      .alertasEstoqueBaixo()
      .then(setItens)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar alertas",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  async function exportar(formato: "csv" | "pdf") {
    setBaixando(true);
    try {
      const blob = await relatoriosApi.baixar("estoque-baixo", formato);
      baixarBlob(blob, `estoque-baixo.${formato}`);
    } catch (err) {
      toast({
        titulo: "Erro ao exportar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setBaixando(false);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Alertas de estoque baixo"
        descricao="Materiais de consumo com quantidade abaixo do estoque mínimo."
        acao={
          itens.length > 0 && (
            <>
              <Button
                variant="outline"
                onClick={() => exportar("csv")}
                disabled={baixando}
              >
                <Download className="h-4 w-4" /> CSV
              </Button>
              <Button
                variant="outline"
                onClick={() => exportar("pdf")}
                disabled={baixando}
              >
                <Download className="h-4 w-4" /> PDF
              </Button>
            </>
          )
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : itens.length === 0 ? (
            <EstadoVazio
              titulo="Nenhum alerta no momento"
              descricao="Todos os materiais de consumo estão acima do mínimo."
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Item</TableHead>
                  <TableHead>Categoria</TableHead>
                  <TableHead>Setor</TableHead>
                  <TableHead className="text-center">Quantidade</TableHead>
                  <TableHead className="text-center">Mínimo</TableHead>
                  <TableHead className="text-center">Faltam</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {itens.map((i) => (
                  <TableRow
                    key={i.id}
                    className="cursor-pointer"
                    onClick={() => navigate(`/itens/${i.id}`)}
                  >
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <AlertTriangle className="h-4 w-4 text-amber-600" />
                        {i.descricao}
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {i.categoria?.nome ?? "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {i.setor?.nome ?? "—"}
                    </TableCell>
                    <TableCell className="text-center font-medium text-amber-700">
                      {i.quantidade}
                    </TableCell>
                    <TableCell className="text-center">
                      {i.estoque_minimo}
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge variant="warning">
                        {Math.max(0, i.estoque_minimo - i.quantidade)}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
