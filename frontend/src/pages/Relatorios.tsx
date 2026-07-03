import * as React from "react";
import { FileText, FileSpreadsheet } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { ordenarComoArvore, prefixoIndentacao } from "@/lib/setores";
import { EstadoBadge } from "@/components/EstadoBadge";
import { EstadoVazio } from "@/components/EstadoVazio";
import { HistoricoTabela } from "@/components/HistoricoTabela";
import { CarregandoTela } from "@/components/ui/spinner";
import { relatoriosApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useReferencias } from "@/hooks/useReferencias";
import { baixarBlob, formatarData } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Item, Movimentacao } from "@/types";

type TipoRelatorio =
  | "inventario"
  | "itens-por-setor"
  | "itens-por-responsavel"
  | "estoque-baixo"
  | "movimentacoes";

const ABAS: { id: TipoRelatorio; rotulo: string }[] = [
  { id: "inventario", rotulo: "Inventário geral" },
  { id: "itens-por-setor", rotulo: "Por departamento" },
  { id: "itens-por-responsavel", rotulo: "Por responsável" },
  { id: "estoque-baixo", rotulo: "Estoque baixo" },
  { id: "movimentacoes", rotulo: "Movimentações" },
];

const TODOS = "todos";

export default function Relatorios() {
  const { toast } = useToast();
  const { setores, servidores } = useReferencias();
  const [aba, setAba] = React.useState<TipoRelatorio>("inventario");

  const [itens, setItens] = React.useState<Item[]>([]);
  const [movs, setMovs] = React.useState<Movimentacao[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [baixando, setBaixando] = React.useState(false);

  // Parâmetros por relatório.
  const [setorId, setSetorId] = React.useState<string>(TODOS);
  const [responsavelId, setResponsavelId] = React.useState<string>(TODOS);
  const [dataInicio, setDataInicio] = React.useState("");
  const [dataFim, setDataFim] = React.useState("");

  const ehMovimentacoes = aba === "movimentacoes";

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const onErro = (err: unknown) =>
      toast({
        titulo: "Erro ao carregar relatório",
        descricao: mensagemErro(err),
        variant: "destructive",
      });

    if (aba === "inventario") {
      relatoriosApi.inventario().then(setItens).catch(onErro).finally(fim);
    } else if (aba === "itens-por-setor") {
      relatoriosApi
        .itensPorSetor(setorId !== TODOS ? Number(setorId) : undefined)
        .then(setItens)
        .catch(onErro)
        .finally(fim);
    } else if (aba === "itens-por-responsavel") {
      relatoriosApi
        .itensPorResponsavel(
          responsavelId !== TODOS ? Number(responsavelId) : undefined
        )
        .then(setItens)
        .catch(onErro)
        .finally(fim);
    } else if (aba === "estoque-baixo") {
      relatoriosApi.estoqueBaixo().then(setItens).catch(onErro).finally(fim);
    } else {
      relatoriosApi
        .movimentacoes({
          data_inicio: dataInicio || undefined,
          data_fim: dataFim || undefined,
        })
        .then(setMovs)
        .catch(onErro)
        .finally(fim);
    }
    function fim() {
      setCarregando(false);
    }
  }, [aba, setorId, responsavelId, dataInicio, dataFim, toast]);

  React.useEffect(carregar, [carregar]);

  async function exportar(formato: "csv" | "pdf") {
    setBaixando(true);
    try {
      const params: Record<string, unknown> = {};
      if (aba === "itens-por-setor" && setorId !== TODOS)
        params.setor_id = Number(setorId);
      if (aba === "itens-por-responsavel" && responsavelId !== TODOS)
        params.responsavel_id = Number(responsavelId);
      if (ehMovimentacoes) {
        if (dataInicio) params.data_inicio = dataInicio;
        if (dataFim) params.data_fim = dataFim;
      }
      const blob = await relatoriosApi.baixar(aba, formato, params);
      baixarBlob(blob, `${aba}.${formato}`);
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

  // PDF está disponível apenas para relatórios de itens (backend).
  const permitePdf = !ehMovimentacoes;

  return (
    <div>
      <PageHeader
        titulo="Relatórios"
        descricao="Consultas e exportações para CSV e PDF."
        acao={
          <>
            <Button
              variant="outline"
              onClick={() => exportar("csv")}
              disabled={baixando}
            >
              <FileSpreadsheet className="h-4 w-4" /> CSV
            </Button>
            {permitePdf && (
              <Button
                variant="outline"
                onClick={() => exportar("pdf")}
                disabled={baixando}
              >
                <FileText className="h-4 w-4" /> PDF
              </Button>
            )}
          </>
        }
      />

      <div className="mb-4 flex flex-wrap gap-2">
        {ABAS.map((a) => (
          <button
            key={a.id}
            onClick={() => setAba(a.id)}
            className={cn(
              "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
              aba === a.id
                ? "border-primary bg-primary text-primary-foreground"
                : "border-border bg-card text-foreground hover:bg-accent"
            )}
          >
            {a.rotulo}
          </button>
        ))}
      </div>

      {/* Filtros contextuais por relatório. */}
      {(aba === "itens-por-setor" ||
        aba === "itens-por-responsavel" ||
        ehMovimentacoes) && (
        <Card className="mb-4">
          <CardContent className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
            {aba === "itens-por-setor" && (
              <FormField label="Departamento" className="sm:w-64">
                <Select value={setorId} onValueChange={setSetorId}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={TODOS}>Todos os departamentos</SelectItem>
                    {ordenarComoArvore(setores).map(({ setor: s, nivel }) => (
                      <SelectItem key={s.id} value={String(s.id)}>
                        {prefixoIndentacao(nivel)}
                        {s.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}
            {aba === "itens-por-responsavel" && (
              <FormField label="Responsável" className="sm:w-64">
                <Select value={responsavelId} onValueChange={setResponsavelId}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={TODOS}>Todos os responsáveis</SelectItem>
                    {servidores.map((s) => (
                      <SelectItem key={s.id} value={String(s.id)}>
                        {s.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}
            {ehMovimentacoes && (
              <>
                <FormField label="De" className="sm:w-40">
                  <input
                    type="date"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={dataInicio}
                    onChange={(e) => setDataInicio(e.target.value)}
                  />
                </FormField>
                <FormField label="Até" className="sm:w-40">
                  <input
                    type="date"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={dataFim}
                    onChange={(e) => setDataFim(e.target.value)}
                  />
                </FormField>
              </>
            )}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            {ABAS.find((a) => a.id === aba)?.rotulo}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : ehMovimentacoes ? (
            <HistoricoTabela movimentacoes={movs} mostrarItem />
          ) : itens.length === 0 ? (
            <EstadoVazio titulo="Sem resultados para este relatório" />
          ) : (
            <TabelaItens itens={itens} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function TabelaItens({ itens }: { itens: Item[] }) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Descrição</TableHead>
          <TableHead>Categoria</TableHead>
          <TableHead>Patrimônio</TableHead>
          <TableHead>Departamento</TableHead>
          <TableHead>Responsável</TableHead>
          <TableHead className="text-center">Qtd.</TableHead>
          <TableHead>Estado</TableHead>
          <TableHead>Aquisição</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {itens.map((i) => (
          <TableRow key={i.id}>
            <TableCell className="font-medium">{i.descricao}</TableCell>
            <TableCell className="text-muted-foreground">
              {i.categoria?.nome ?? "—"}
            </TableCell>
            <TableCell>{i.numero_patrimonio ?? "—"}</TableCell>
            <TableCell className="text-muted-foreground">
              {i.setor?.nome ?? "—"}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {i.responsavel?.nome ?? "—"}
            </TableCell>
            <TableCell className="text-center">{i.quantidade}</TableCell>
            <TableCell>
              <EstadoBadge estado={i.estado_conservacao} />
            </TableCell>
            <TableCell className="text-muted-foreground">
              {formatarData(i.data_aquisicao)}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
