import * as React from "react";
import { useNavigate } from "react-router-dom";
import { List } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CarregandoTela } from "@/components/ui/spinner";
import { ordensServicoApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { rotuloStatusOS, rotuloPrioridadeOS } from "@/lib/rotulos";
import type { ContagemRotulo, DashboardChamados as Dados } from "@/types";

export default function DashboardChamados() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const [d, setD] = React.useState<Dados | null>(null);
  const [carregando, setCarregando] = React.useState(true);

  React.useEffect(() => {
    ordensServicoApi
      .dashboard()
      .then(setD)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o dashboard",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  return (
    <div>
      <PageHeader
        titulo="Dashboard de Chamados"
        descricao="Visão geral do atendimento e SLA."
        acao={
          <Button variant="outline" onClick={() => navigate("/manutencao")}>
            <List className="h-4 w-4" /> Ver lista
          </Button>
        }
      />

      {carregando ? (
        <CarregandoTela />
      ) : !d ? null : (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            <Kpi rotulo="Abertos" valor={d.abertos_total} />
            <Kpi rotulo="Sem técnico" valor={d.sem_tecnico} alerta={d.sem_tecnico > 0} />
            <Kpi
              rotulo="Resposta atrasada"
              valor={d.resposta_atrasada}
              alerta={d.resposta_atrasada > 0}
            />
            <Kpi
              rotulo="Resolução atrasada"
              valor={d.resolucao_atrasada}
              alerta={d.resolucao_atrasada > 0}
            />
            <Kpi rotulo="Concluídos (30d)" valor={d.concluidos_ultimos_30} />
            <Kpi
              rotulo="Tempo médio resolução"
              valor={
                d.tempo_medio_resolucao_horas > 0
                  ? `${d.tempo_medio_resolucao_horas.toFixed(1)} h`
                  : "—"
              }
            />
            <Kpi
              rotulo="Nota média"
              valor={
                d.total_avaliacoes > 0
                  ? `${d.nota_media.toFixed(1)} ★`
                  : "—"
              }
              sub={
                d.total_avaliacoes > 0
                  ? `${d.total_avaliacoes} avaliação(ões)`
                  : "sem avaliações"
              }
            />
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <Barras
              titulo="Por status"
              itens={d.por_status.map((x) => ({
                rotulo: rotuloStatusOS(x.rotulo as never),
                total: x.total,
              }))}
            />
            <Barras
              titulo="Por prioridade (abertos)"
              itens={d.por_prioridade.map((x) => ({
                rotulo: rotuloPrioridadeOS(x.rotulo as never),
                total: x.total,
              }))}
            />
            <Barras titulo="Por categoria (abertos)" itens={d.por_categoria} />
          </div>
        </div>
      )}
    </div>
  );
}

function Kpi({
  rotulo,
  valor,
  sub,
  alerta,
}: {
  rotulo: string;
  valor: number | string;
  sub?: string;
  alerta?: boolean;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <p className="text-xs text-muted-foreground">{rotulo}</p>
        <p
          className={
            alerta
              ? "text-2xl font-semibold text-destructive"
              : "text-2xl font-semibold text-foreground"
          }
        >
          {valor}
        </p>
        {sub && <p className="text-[11px] text-muted-foreground">{sub}</p>}
      </CardContent>
    </Card>
  );
}

function Barras({
  titulo,
  itens,
}: {
  titulo: string;
  itens: ContagemRotulo[];
}) {
  const max = Math.max(1, ...itens.map((i) => i.total));
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{titulo}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {itens.length === 0 ? (
          <p className="text-sm text-muted-foreground">Sem dados.</p>
        ) : (
          itens.map((i) => (
            <div key={i.rotulo}>
              <div className="mb-0.5 flex justify-between text-xs">
                <span className="text-foreground">{i.rotulo}</span>
                <span className="text-muted-foreground">{i.total}</span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-primary"
                  style={{ width: `${(i.total / max) * 100}%` }}
                />
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}
