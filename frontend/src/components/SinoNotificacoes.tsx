import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, Check, Volume2, VolumeX } from "lucide-react";
import { notificacoesApi } from "@/services/api";
import { useAuth } from "@/contexts/AuthContext";
import { formatarDataHora } from "@/lib/format";
import { cn } from "@/lib/utils";
import { tocarSomNotificacao } from "@/lib/som";
import type { Notificacao } from "@/types";

export function SinoNotificacoes() {
  const navigate = useNavigate();
  const { ehSolicitante } = useAuth();
  const [aberto, setAberto] = React.useState(false);
  const [naoLidas, setNaoLidas] = React.useState(0);
  const [lista, setLista] = React.useState<Notificacao[]>([]);
  const [carregando, setCarregando] = React.useState(false);
  const [silencioso, setSilencioso] = React.useState(() => {
    try {
      return localStorage.getItem("notif_som") === "off";
    } catch {
      return false;
    }
  });
  const ref = React.useRef<HTMLDivElement | null>(null);
  // Último contador conhecido: se o novo for maior, chegou algo -> toca som.
  const anteriorRef = React.useRef<number | null>(null);
  // Mantém o valor atual acessível dentro do callback de polling.
  const silenciosoRef = React.useRef(silencioso);
  silenciosoRef.current = silencioso;

  const atualizarContador = React.useCallback(() => {
    notificacoesApi
      .contarNaoLidas()
      .then((valor) => {
        setNaoLidas(valor);
        const anterior = anteriorRef.current;
        if (anterior !== null && valor > anterior && !silenciosoRef.current) {
          tocarSomNotificacao();
        }
        anteriorRef.current = valor;
      })
      .catch(() => undefined);
  }, []);

  function alternarSom() {
    setSilencioso((v) => {
      const novo = !v;
      try {
        localStorage.setItem("notif_som", novo ? "off" : "on");
      } catch {
        /* best-effort */
      }
      return novo;
    });
  }

  // Polling do contador.
  React.useEffect(() => {
    atualizarContador();
    const t = setInterval(atualizarContador, 10000);
    return () => clearInterval(t);
  }, [atualizarContador]);

  // Atualiza na hora quando outra tela marca notificações como lidas
  // (ex.: abrir a conversa de um chamado).
  React.useEffect(() => {
    window.addEventListener("notificacoes:atualizar", atualizarContador);
    return () =>
      window.removeEventListener("notificacoes:atualizar", atualizarContador);
  }, [atualizarContador]);

  // Reflete o nº de não lidas no título da aba (visível mesmo fora da tela).
  React.useEffect(() => {
    const base = document.title.replace(/^\(\d+\)\s*/, "");
    document.title = naoLidas > 0 ? `(${naoLidas}) ${base}` : base;
    return () => {
      document.title = document.title.replace(/^\(\d+\)\s*/, "");
    };
  }, [naoLidas]);

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
          anteriorRef.current = r.nao_lidas;
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
      anteriorRef.current = 0;
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
            <div className="flex items-center gap-3">
              {naoLidas > 0 && (
                <button
                  onClick={marcarTodas}
                  className="flex items-center gap-1 text-xs text-primary hover:underline"
                >
                  <Check className="h-3 w-3" /> Marcar todas
                </button>
              )}
              <button
                onClick={alternarSom}
                className="text-muted-foreground hover:text-foreground"
                title={silencioso ? "Ativar som" : "Silenciar som"}
                aria-label={silencioso ? "Ativar som" : "Silenciar som"}
              >
                {silencioso ? (
                  <VolumeX className="h-4 w-4" />
                ) : (
                  <Volume2 className="h-4 w-4" />
                )}
              </button>
            </div>
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
