import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, Check } from "lucide-react";
import { notificacoesApi } from "@/services/api";
import { useAuth } from "@/contexts/AuthContext";
import { formatarDataHora } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Notificacao } from "@/types";

export function SinoNotificacoes() {
  const navigate = useNavigate();
  const { ehSolicitante } = useAuth();
  const [aberto, setAberto] = React.useState(false);
  const [naoLidas, setNaoLidas] = React.useState(0);
  const [lista, setLista] = React.useState<Notificacao[]>([]);
  const [carregando, setCarregando] = React.useState(false);
  const ref = React.useRef<HTMLDivElement | null>(null);

  const atualizarContador = React.useCallback(() => {
    notificacoesApi
      .contarNaoLidas()
      .then(setNaoLidas)
      .catch(() => undefined);
  }, []);

  // Polling do contador.
  React.useEffect(() => {
    atualizarContador();
    const t = setInterval(atualizarContador, 30000);
    return () => clearInterval(t);
  }, [atualizarContador]);

  // Fecha ao clicar fora.
  React.useEffect(() => {
    function onClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setAberto(false);
      }
    }
    if (aberto) document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, [aberto]);

  function abrir() {
    const novo = !aberto;
    setAberto(novo);
    if (novo) {
      setCarregando(true);
      notificacoesApi
        .listar(false, 20)
        .then((r) => {
          setLista(r.dados);
          setNaoLidas(r.nao_lidas);
        })
        .catch(() => undefined)
        .finally(() => setCarregando(false));
    }
  }

  async function abrirNotificacao(n: Notificacao) {
    setAberto(false);
    if (!n.lida) {
      try {
        await notificacoesApi.marcarLida(n.id);
        atualizarContador();
      } catch {
        /* best-effort */
      }
    }
    if (n.recurso_id) {
      navigate(
        ehSolicitante ? `/chamado/${n.recurso_id}` : `/manutencao/${n.recurso_id}`
      );
    }
  }

  async function marcarTodas() {
    try {
      await notificacoesApi.marcarTodasLidas();
      setLista((prev) => prev.map((n) => ({ ...n, lida: true })));
      setNaoLidas(0);
    } catch {
      /* best-effort */
    }
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={abrir}
        className="relative flex h-9 w-9 items-center justify-center rounded-md hover:bg-accent"
        aria-label="Notificações"
      >
        <Bell className="h-5 w-5" />
        {naoLidas > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-semibold text-destructive-foreground">
            {naoLidas > 9 ? "9+" : naoLidas}
          </span>
        )}
      </button>

      {aberto && (
        <div className="absolute right-0 z-50 mt-2 w-80 rounded-lg border border-border bg-card shadow-lg">
          <div className="flex items-center justify-between border-b border-border px-3 py-2">
            <span className="text-sm font-semibold">Notificações</span>
            {naoLidas > 0 && (
              <button
                onClick={marcarTodas}
                className="flex items-center gap-1 text-xs text-primary hover:underline"
              >
                <Check className="h-3 w-3" /> Marcar todas
              </button>
            )}
          </div>
          <div className="max-h-96 overflow-y-auto">
            {carregando ? (
              <p className="py-6 text-center text-sm text-muted-foreground">
                Carregando…
              </p>
            ) : lista.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">
                Nenhuma notificação.
              </p>
            ) : (
              lista.map((n) => (
                <button
                  key={n.id}
                  onClick={() => abrirNotificacao(n)}
                  className={cn(
                    "flex w-full flex-col items-start gap-0.5 border-b border-border/60 px-3 py-2 text-left last:border-0 hover:bg-accent",
                    !n.lida && "bg-primary/5"
                  )}
                >
                  <span className="flex w-full items-center gap-2">
                    {!n.lida && (
                      <span className="h-2 w-2 shrink-0 rounded-full bg-primary" />
                    )}
                    <span className="text-sm font-medium">{n.titulo}</span>
                  </span>
                  {n.mensagem && (
                    <span className="text-xs text-muted-foreground">
                      {n.mensagem}
                    </span>
                  )}
                  <span className="text-[10px] text-muted-foreground">
                    {formatarDataHora(n.criado_em)}
                  </span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
