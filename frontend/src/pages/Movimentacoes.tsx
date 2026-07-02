import * as React from "react";
import { Plus, Download, X } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { HistoricoTabela } from "@/components/HistoricoTabela";
import { MovimentacaoForm } from "@/components/MovimentacaoForm";
import { CarregandoTela } from "@/components/ui/spinner";
import {
  movimentacoesApi,
  relatoriosApi,
  itensApi,
  type FiltroMovimentacoes,
} from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useReferencias } from "@/hooks/useReferencias";
import { baixarBlob } from "@/lib/format";
import { TIPOS_MOVIMENTACAO } from "@/lib/rotulos";
import type { Item, Movimentacao } from "@/types";

const TODOS = "todos";

export default function Movimentacoes() {
  const { toast } = useToast();
  const { setores, servidores, adicionarServidor } = useReferencias();

  const [movs, setMovs] = React.useState<Movimentacao[]>([]);
  const [total, setTotal] = React.useState(0);
  const [carregando, setCarregando] = React.useState(true);
  const [itens, setItens] = React.useState<Item[]>([]);
  const [baixando, setBaixando] = React.useState(false);
  const [formAberto, setFormAberto] = React.useState(false);

  // Filtros.
  const [tipo, setTipo] = React.useState<string>(TODOS);
  const [dataInicio, setDataInicio] = React.useState("");
  const [dataFim, setDataFim] = React.useState("");

  // Carrega itens (para o seletor do formulário) uma vez.
  React.useEffect(() => {
    itensApi
      .listar({ tamanho: 500 })
      .then((r) => setItens(r.dados))
      .catch(() => {
        /* silencioso: o formulário ainda funciona sem a lista completa */
      });
  }, []);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const filtro: FiltroMovimentacoes = {
      tipo: tipo !== TODOS ? tipo : undefined,
      data_inicio: dataInicio || undefined,
      data_fim: dataFim || undefined,
      tamanho: 100,
    };
    movimentacoesApi
      .listar(filtro)
      .then((r) => {
        setMovs(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar movimentações",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [tipo, dataInicio, dataFim, toast]);

  React.useEffect(carregar, [carregar]);

  const temFiltro = tipo !== TODOS || dataInicio || dataFim;

  async function exportarCSV() {
    setBaixando(true);
    try {
      const blob = await relatoriosApi.baixar("movimentacoes", "csv", {
        tipo: tipo !== TODOS ? tipo : undefined,
        data_inicio: dataInicio || undefined,
        data_fim: dataFim || undefined,
      });
      baixarBlob(blob, "movimentacoes.csv");
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
        titulo="Movimentações"
        descricao="Entradas, saídas e baixas de estoque com histórico completo."
        acao={
          <>
            <Button
              variant="outline"
              onClick={exportarCSV}
              disabled={baixando}
            >
              <Download className="h-4 w-4" /> Exportar CSV
            </Button>
            <Button onClick={() => setFormAberto(true)}>
              <Plus className="h-4 w-4" /> Nova movimentação
            </Button>
          </>
        }
      />

      <Card className="mb-4">
        <CardContent className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
          <FormField label="Tipo" className="sm:w-52">
            <Select
              value={tipo}
              onValueChange={(v) => setTipo(v)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={TODOS}>Todos os tipos</SelectItem>
                {TIPOS_MOVIMENTACAO.map((t) => (
                  <SelectItem key={t.valor} value={t.valor}>
                    {t.rotulo}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <FormField label="De" className="sm:w-40">
            <Input
              type="date"
              value={dataInicio}
              onChange={(e) => setDataInicio(e.target.value)}
            />
          </FormField>

          <FormField label="Até" className="sm:w-40">
            <Input
              type="date"
              value={dataFim}
              onChange={(e) => setDataFim(e.target.value)}
            />
          </FormField>

          {temFiltro && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setTipo(TODOS);
                setDataInicio("");
                setDataFim("");
              }}
            >
              <X className="h-4 w-4" /> Limpar
            </Button>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : (
            <HistoricoTabela movimentacoes={movs} mostrarItem />
          )}
        </CardContent>
      </Card>

      {total > 0 && (
        <p className="mt-3 text-sm text-muted-foreground">
          Exibindo {movs.length} de {total} movimentações.
        </p>
      )}

      <MovimentacaoForm
        aberto={formAberto}
        itens={itens}
        setores={setores}
        servidores={servidores}
        onServidorCriado={adicionarServidor}
        onFechar={() => setFormAberto(false)}
        onRegistrado={() => {
          setFormAberto(false);
          carregar();
        }}
      />
    </div>
  );
}
