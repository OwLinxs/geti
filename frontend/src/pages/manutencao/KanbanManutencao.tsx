import * as React from "react";
import { useNavigate } from "react-router-dom";
import { List, MessageCircle } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CarregandoTela } from "@/components/ui/spinner";
import { ordensServicoApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import {
  rotuloPrioridadeOS,
  rotuloStatusOS,
  variantePrioridadeOS,
} from "@/lib/rotulos";
import { cn } from "@/lib/utils";
import type { OrdemServico, StatusOS } from "@/types";

// Colunas do board (exclui "cancelada" — chamados encerrados saem do fluxo).
const COLUNAS: StatusOS[] = [
  "aberta",
  "em_andamento",
  "aguardando_peca",
  "concluida",
];

export default function KanbanManutencao() {
  const { toast } = useToast();
  const navigate = useNavigate();

  const [lista, setLista] = React.useState<OrdemServico[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [arrastando, setArrastando] = React.useState<number | null>(null);
  const [sobreColuna, setSobreColuna] = React.useState<StatusOS | null>(null);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    ordensServicoApi
      .listar({ tamanho: 200 })
      .then((r) => setLista(r.dados))
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o kanban",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  const porColuna = React.useMemo(() => {
    const mapa: Record<StatusOS, OrdemServico[]> = {
      aberta: [],
      em_andamento: [],
      aguardando_peca: [],
      concluida: [],
      cancelada: [],
    };
    for (const os of lista) mapa[os.status]?.push(os);
    return mapa;
  }, [lista]);

  async function soltarNaColuna(coluna: StatusOS) {
    const id = arrastando;
    setArrastando(null);
    setSobreColuna(null);
    if (!id) return;
    const os = lista.find((o) => o.id === id);
    if (!os || os.status === coluna) return;

    const anterior = os.status;
    // Atualização otimista.
    setLista((prev) =>
      prev.map((o) => (o.id === id ? { ...o, status: coluna } : o))
    );
    try {
      await ordensServicoApi.definirStatus(id, coluna);
    } catch (err) {
      // Reverte em caso de erro.
      setLista((prev) =>
        prev.map((o) => (o.id === id ? { ...o, status: anterior } : o))
      );
      toast({
        titulo: "Não foi possível mover o chamado",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Kanban de Manutenção"
        descricao="Arraste os chamados entre as colunas para mudar o status."
        acao={
          <Button variant="outline" onClick={() => navigate("/manutencao")}>
            <List className="h-4 w-4" /> Ver lista
          </Button>
        }
      />

      {carregando ? (
        <CarregandoTela />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {COLUNAS.map((coluna) => (
            <div
              key={coluna}
              onDragOver={(e) => {
                e.preventDefault();
                setSobreColuna(coluna);
              }}
              onDragLeave={() =>
                setSobreColuna((c) => (c === coluna ? null : c))
              }
              onDrop={() => soltarNaColuna(coluna)}
              className={cn(
                "flex flex-col rounded-lg border bg-muted/30 p-2 transition-colors",
                sobreColuna === coluna
                  ? "border-primary bg-primary/5"
                  : "border-border"
              )}
            >
              <div className="mb-2 flex items-center justify-between px-1">
                <span className="text-sm font-semibold">
                  {rotuloStatusOS(coluna)}
                </span>
                <Badge variant="muted">{porColuna[coluna].length}</Badge>
              </div>

              <div className="flex min-h-24 flex-col gap-2">
                {porColuna[coluna].map((os) => (
                  <article
                    key={os.id}
                    draggable
                    onDragStart={() => setArrastando(os.id)}
                    onDragEnd={() => setArrastando(null)}
                    onClick={() => navigate(`/manutencao/${os.id}`)}
                    className={cn(
                      "cursor-grab rounded-md border border-border bg-card p-3 shadow-sm transition-opacity active:cursor-grabbing",
                      arrastando === os.id && "opacity-50"
                    )}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-medium text-muted-foreground">
                        {os.numero}
                      </span>
                      <Badge variant={variantePrioridadeOS(os.prioridade)}>
                        {rotuloPrioridadeOS(os.prioridade)}
                      </Badge>
                    </div>
                    <p className="mt-1 line-clamp-2 text-sm font-medium">
                      {os.equipamento_snapshot}
                    </p>
                    {os.defeito_relatado && (
                      <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                        {os.defeito_relatado}
                      </p>
                    )}
                    <div className="mt-2 flex flex-wrap items-center gap-1.5">
                      {os.origem && os.origem !== "interno" && (
                        <Badge variant="secondary" className="gap-1">
                          <MessageCircle className="h-3 w-3" />
                          {os.origem}
                        </Badge>
                      )}
                      {(os.solicitante?.nome ||
                        os.solicitante_nome_snapshot) && (
                        <span className="text-xs text-muted-foreground">
                          {os.solicitante?.nome ??
                            os.solicitante_nome_snapshot}
                        </span>
                      )}
                    </div>
                  </article>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
