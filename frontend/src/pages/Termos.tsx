import * as React from "react";
import { Plus, Download, FileText } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
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
import { TermoForm } from "@/components/TermoForm";
import { CarregandoTela } from "@/components/ui/spinner";
import { itensApi, termosApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useReferencias } from "@/hooks/useReferencias";
import { baixarBlob, formatarData } from "@/lib/format";
import type { Item, TermoResponsabilidade } from "@/types";

export default function Termos() {
  const { toast } = useToast();
  const { servidores } = useReferencias();
  const [lista, setLista] = React.useState<TermoResponsabilidade[]>([]);
  const [itens, setItens] = React.useState<Item[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [formAberto, setFormAberto] = React.useState(false);
  const [baixandoId, setBaixandoId] = React.useState<number | null>(null);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    termosApi
      .listar()
      .then(setLista)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar termos",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  // Itens não baixados para o seletor (apenas patrimoniados fazem sentido,
  // mas o backend valida; aqui carregamos todos para simplicidade).
  React.useEffect(() => {
    itensApi
      .listar({ tamanho: 500 })
      .then((r) => setItens(r.dados.filter((i) => !i.baixado)))
      .catch(() => {
        /* silencioso */
      });
  }, []);

  async function baixarPdf(t: TermoResponsabilidade) {
    setBaixandoId(t.id);
    try {
      const blob = await termosApi.baixarPdf(t.id);
      baixarBlob(blob, `termo-${t.numero}.pdf`);
    } catch (err) {
      toast({
        titulo: "Erro ao baixar PDF",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setBaixandoId(null);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Termos de responsabilidade"
        descricao="Recibos de entrega de equipamentos aos servidores."
        acao={
          <Button onClick={() => setFormAberto(true)}>
            <Plus className="h-4 w-4" /> Emitir termo
          </Button>
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio
              titulo="Nenhum termo emitido"
              descricao="Emita o primeiro termo de responsabilidade."
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Número</TableHead>
                  <TableHead>Item</TableHead>
                  <TableHead>Patrimônio</TableHead>
                  <TableHead>Servidor</TableHead>
                  <TableHead>Emissão</TableHead>
                  <TableHead className="w-24 text-right">PDF</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((t) => (
                  <TableRow key={t.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <FileText className="h-4 w-4 text-primary" />
                        {t.numero}
                      </div>
                    </TableCell>
                    <TableCell>{t.item_descricao_snapshot}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {t.patrimonio_snapshot || "—"}
                    </TableCell>
                    <TableCell>
                      {t.servidor_nome_snapshot}
                      <span className="text-muted-foreground">
                        {" "}
                        ({t.servidor_matricula_snapshot})
                      </span>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatarData(t.data_emissao)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => baixarPdf(t)}
                        disabled={baixandoId === t.id}
                        aria-label="Baixar PDF"
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <TermoForm
        aberto={formAberto}
        itens={itens}
        servidores={servidores}
        onFechar={() => setFormAberto(false)}
        onEmitido={() => {
          setFormAberto(false);
          carregar();
        }}
      />
    </div>
  );
}
