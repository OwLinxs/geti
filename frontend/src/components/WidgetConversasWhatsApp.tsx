import * as React from "react";
import { MessageCircle, X, Minus, ChevronLeft, Circle } from "lucide-react";
import { ordensServicoApi } from "@/services/api";
import { ConversaChamado } from "@/components/ConversaChamado";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";
import { formatarDataHora } from "@/lib/format";
import type { OrdemServico } from "@/types";

// Conversas de WhatsApp abertas (chamados origem="whatsapp" não encerrados).
const STATUS_ATIVOS = new Set(["aberta", "em_andamento", "aguardando_peca"]);

function dentroJanela(os: OrdemServico): boolean {
  if (!os.ultima_msg_solicitante_em) return false;
  const t = new Date(os.ultima_msg_solicitante_em).getTime();
  return Date.now() - t < 24 * 60 * 60 * 1000;
}

function nomeContato(os: OrdemServico): string {
  return (
    os.solicitante_nome_snapshot ||
    os.solicitante?.nome ||
    os.solicitante_contato ||
    "Contato"
  );
}

// WidgetConversasWhatsApp: janela flutuante minimizável (estilo suporte) para a
// equipe responder chamados vindos do WhatsApp sem sair da tela atual.
export function WidgetConversasWhatsApp() {
  const [aberto, setAberto] = React.useState(false);
  const [lista, setLista] = React.useState<OrdemServico[]>([]);
  const [carregando, setCarregando] = React.useState(false);
  const [selecionada, setSelecionada] = React.useState<OrdemServico | null>(null);

  const carregar = React.useCallback(() => {
    ordensServicoApi
      .listar({ tamanho: 100 })
      .then((r) => {
        const wpp = r.dados.filter(
          (o) => o.origem === "whatsapp" && STATUS_ATIVOS.has(o.status)
        );
        // Ordena por atividade recente do solicitante (janela aberta primeiro).
        wpp.sort((a, b) => {
          const ta = a.ultima_msg_solicitante_em
            ? new Date(a.ultima_msg_solicitante_em).getTime()
            : 0;
          const tb = b.ultima_msg_solicitante_em
            ? new Date(b.ultima_msg_solicitante_em).getTime()
            : 0;
          return tb - ta;
        });
        setLista(wpp);
      })
      .catch(() => undefined)
      .finally(() => setCarregando(false));
  }, []);

  // Poll leve sempre (para o badge), mesmo com o widget fechado.
  React.useEffect(() => {
    setCarregando(true);
    carregar();
    const t = setInterval(carregar, 20000);
    return () => clearInterval(t);
  }, [carregar]);

  const ativas = lista.filter(dentroJanela).length;

  // ---- Launcher (bolha) ----
  if (!aberto) {
    return (
      <button
        type="button"
        onClick={() => setAberto(true)}
        aria-label="Abrir conversas de WhatsApp"
        className="fixed bottom-4 right-4 z-40 flex h-14 w-14 items-center justify-center rounded-full bg-emerald-600 text-white shadow-lg transition hover:bg-emerald-700"
      >
        <MessageCircle className="h-6 w-6" />
        {ativas > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1 text-[11px] font-bold text-white">
            {ativas}
          </span>
        )}
      </button>
    );
  }

  // ---- Painel ----
  return (
    <div className="fixed bottom-4 right-4 z-40 flex h-[32rem] max-h-[calc(100vh-2rem)] w-[22rem] max-w-[calc(100vw-2rem)] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-2xl">
      {/* Cabeçalho */}
      <div className="flex items-center justify-between bg-emerald-600 px-3 py-2 text-white">
        <div className="flex items-center gap-2">
          {selecionada && (
            <button
              type="button"
              onClick={() => setSelecionada(null)}
              className="rounded p-0.5 hover:bg-white/20"
              aria-label="Voltar à lista"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
          )}
          <MessageCircle className="h-4 w-4" />
          <span className="text-sm font-semibold">
            {selecionada
              ? `${selecionada.numero} · ${nomeContato(selecionada)}`
              : "Conversas WhatsApp"}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setAberto(false)}
            className="rounded p-0.5 hover:bg-white/20"
            aria-label="Minimizar"
          >
            <Minus className="h-4 w-4" />
          </button>
          <button
            type="button"
            onClick={() => {
              setAberto(false);
              setSelecionada(null);
            }}
            className="rounded p-0.5 hover:bg-white/20"
            aria-label="Fechar"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Corpo */}
      {selecionada ? (
        <div className="flex-1 overflow-hidden p-2">
          {!dentroJanela(selecionada) && (
            <div className="mb-2 rounded-md bg-amber-50 px-2 py-1 text-[11px] text-amber-800">
              🟡 Fora da janela de 24h — resposta livre não será entregue (use
              template).
            </div>
          )}
          <ConversaChamado osId={selecionada.id} modo="equipe" />
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto">
          {carregando && lista.length === 0 ? (
            <div className="flex h-full items-center justify-center">
              <Spinner className="h-5 w-5" />
            </div>
          ) : lista.length === 0 ? (
            <p className="p-6 text-center text-sm text-muted-foreground">
              Nenhuma conversa de WhatsApp em aberto.
            </p>
          ) : (
            lista.map((os) => {
              const janela = dentroJanela(os);
              return (
                <button
                  key={os.id}
                  type="button"
                  onClick={() => setSelecionada(os)}
                  className="flex w-full items-start gap-2 border-b border-border/60 px-3 py-2 text-left hover:bg-accent"
                >
                  <Circle
                    className={cn(
                      "mt-1 h-2.5 w-2.5 shrink-0",
                      janela
                        ? "fill-emerald-500 text-emerald-500"
                        : "fill-muted-foreground/40 text-muted-foreground/40"
                    )}
                  />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate text-sm font-medium">
                        {nomeContato(os)}
                      </span>
                      <span className="shrink-0 text-[10px] text-muted-foreground">
                        {os.numero}
                      </span>
                    </div>
                    <p className="truncate text-xs text-muted-foreground">
                      {os.assunto || os.equipamento_snapshot || "Chamado WhatsApp"}
                    </p>
                    {os.ultima_msg_solicitante_em && (
                      <p className="text-[10px] text-muted-foreground/80">
                        {janela ? "Janela aberta · " : ""}
                        {formatarDataHora(os.ultima_msg_solicitante_em)}
                      </p>
                    )}
                  </div>
                </button>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}
