import * as React from "react";
import { Link } from "react-router-dom";
import {
  Package,
  AlertTriangle,
  ArrowLeftRight,
  Archive,
  TrendingUp,
  TrendingDown,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/PageHeader";
import { CarregandoTela } from "@/components/ui/spinner";
import { EstadoVazio } from "@/components/EstadoVazio";
import { itensApi, movimentacoesApi, relatoriosApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { Item, Movimentacao } from "@/types";
import { formatarDataHora } from "@/lib/format";
import { ehEntrada, rotuloTipoMov } from "@/lib/rotulos";

export default function Dashboard() {
  const { toast } = useToast();
  const [carregando, setCarregando] = React.useState(true);
  const [totalItens, setTotalItens] = React.useState(0);
  const [inventario, setInventario] = React.useState<Item[]>([]);
  const [alertas, setAlertas] = React.useState<Item[]>([]);
  const [ultimasMovs, setUltimasMovs] = React.useState<Movimentacao[]>([]);

  React.useEffect(() => {
    Promise.all([
      itensApi.listar({ tamanho: 1, pagina: 1 }),
      relatoriosApi.inventario(),
      itensApi.alertasEstoqueBaixo(),
      movimentacoesApi.listar({ tamanho: 6, pagina: 1 }),
    ])
      .then(([pag, inv, al, movs]) => {
        setTotalItens(pag.total);
        setInventario(inv);
        setAlertas(al);
        setUltimasMovs(movs.dados);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o painel",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  if (carregando) return <CarregandoTela texto="Carregando painel..." />;

  const baixados = inventario.filter((i) => i.baixado).length;
  const valorTotal = inventario.reduce(
    (acc, i) => acc + (i.valor ?? 0) * (i.quantidade || 1),
    0
  );

  return (
    <div>
      <PageHeader
        titulo="Painel"
        descricao="Resumo geral do estoque e do patrimônio de T.I."
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <CardResumo
          titulo="Itens cadastrados"
          valor={totalItens}
          icone={Package}
          cor="text-primary"
        />
        <CardResumo
          titulo="Alertas de estoque"
          valor={alertas.length}
          icone={AlertTriangle}
          cor={alertas.length > 0 ? "text-amber-600" : "text-muted-foreground"}
          destaque={alertas.length > 0}
          link="/alertas"
        />
        <CardResumo
          titulo="Itens baixados"
          valor={baixados}
          icone={Archive}
          cor="text-muted-foreground"
        />
        <CardResumo
          titulo="Valor estimado"
          valor={valorTotal.toLocaleString("pt-BR", {
            style: "currency",
            currency: "BRL",
            maximumFractionDigits: 0,
          })}
          icone={TrendingUp}
          cor="text-emerald-600"
        />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle className="text-base">
              Alertas de estoque baixo
            </CardTitle>
            <Link
              to="/alertas"
              className="text-xs font-medium text-primary hover:underline"
            >
              Ver todos
            </Link>
          </CardHeader>
          <CardContent>
            {alertas.length === 0 ? (
              <EstadoVazio
                titulo="Nenhum alerta"
                descricao="Todos os consumíveis estão acima do mínimo."
              />
            ) : (
              <ul className="divide-y divide-border">
                {alertas.slice(0, 6).map((i) => (
                  <li
                    key={i.id}
                    className="flex items-center justify-between py-2.5"
                  >
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-foreground">
                        {i.descricao}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {i.categoria?.nome ?? "Sem categoria"}
                      </p>
                    </div>
                    <Badge variant="warning">
                      {i.quantidade} / mín. {i.estoque_minimo}
                    </Badge>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <CardTitle className="text-base">Movimentações recentes</CardTitle>
            <Link
              to="/movimentacoes"
              className="text-xs font-medium text-primary hover:underline"
            >
              Ver todas
            </Link>
          </CardHeader>
          <CardContent>
            {ultimasMovs.length === 0 ? (
              <EstadoVazio titulo="Sem movimentações" />
            ) : (
              <ul className="divide-y divide-border">
                {ultimasMovs.map((m) => {
                  const entrada = ehEntrada(m.tipo);
                  return (
                    <li
                      key={m.id}
                      className="flex items-center gap-3 py-2.5"
                    >
                      <div
                        className={
                          entrada
                            ? "rounded-full bg-emerald-100 p-1.5 text-emerald-700"
                            : "rounded-full bg-rose-100 p-1.5 text-rose-700"
                        }
                      >
                        {entrada ? (
                          <TrendingUp className="h-3.5 w-3.5" />
                        ) : (
                          <TrendingDown className="h-3.5 w-3.5" />
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-foreground">
                          {m.item?.descricao ?? `Item #${m.item_id}`}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {rotuloTipoMov(m.tipo)} · {m.quantidade} un.
                        </p>
                      </div>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {formatarDataHora(m.data_evento)}
                      </span>
                    </li>
                  );
                })}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="mt-6 flex flex-wrap gap-3">
        <AcaoRapida to="/itens" icone={Package} rotulo="Cadastrar item" />
        <AcaoRapida
          to="/movimentacoes"
          icone={ArrowLeftRight}
          rotulo="Registrar movimentação"
        />
        <AcaoRapida to="/termos" icone={Archive} rotulo="Emitir termo" />
      </div>
    </div>
  );
}

function CardResumo({
  titulo,
  valor,
  icone: Icone,
  cor,
  destaque,
  link,
}: {
  titulo: string;
  valor: React.ReactNode;
  icone: React.ComponentType<{ className?: string }>;
  cor: string;
  destaque?: boolean;
  link?: string;
}) {
  const conteudo = (
    <Card className={destaque ? "border-amber-300" : undefined}>
      <CardContent className="flex items-center justify-between p-5">
        <div>
          <p className="text-sm text-muted-foreground">{titulo}</p>
          <p className="mt-1 text-2xl font-semibold text-foreground">{valor}</p>
        </div>
        <Icone className={`h-8 w-8 ${cor}`} />
      </CardContent>
    </Card>
  );
  return link ? <Link to={link}>{conteudo}</Link> : conteudo;
}

function AcaoRapida({
  to,
  icone: Icone,
  rotulo,
}: {
  to: string;
  icone: React.ComponentType<{ className?: string }>;
  rotulo: string;
}) {
  return (
    <Link
      to={to}
      className="inline-flex items-center gap-2 rounded-md border border-border bg-card px-4 py-2 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-accent"
    >
      <Icone className="h-4 w-4 text-primary" />
      {rotulo}
    </Link>
  );
}
