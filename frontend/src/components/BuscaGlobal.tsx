import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Search, Loader2, Wrench, Package, Monitor } from "lucide-react";
import { ordensServicoApi, itensApi } from "@/services/api";
import { cn } from "@/lib/utils";
import type { Item, OrdemServico } from "@/types";

interface Resultado {
  chave: string;
  icone: React.ReactNode;
  titulo: string;
  sub: string;
  rota: string;
}

// BuscaGlobal: paleta de comando (Ctrl/Cmd+K) para pular direto a chamados e
// itens por número, descrição ou solicitante. Montada no AppShell (equipe).
export function BuscaGlobal() {
  const navigate = useNavigate();
  const [aberto, setAberto] = React.useState(false);
  const [termo, setTermo] = React.useState("");
  const [carregando, setCarregando] = React.useState(false);
  const [resultados, setResultados] = React.useState<Resultado[]>([]);
  const [ativo, setAtivo] = React.useState(0);
  const inputRef = React.useRef<HTMLInputElement | null>(null);

  // Atalho global Ctrl/Cmd+K.
  React.useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setAberto((v) => !v);
      }
      if (e.key === "Escape") setAberto(false);
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  React.useEffect(() => {
    if (aberto) setTimeout(() => inputRef.current?.focus(), 30);
    else {
      setTermo("");
      setResultados([]);
      setAtivo(0);
    }
  }, [aberto]);

  // Busca com debounce.
  React.useEffect(() => {
    const q = termo.trim();
    if (q.length < 2) {
      setResultados([]);
      return;
    }
    setCarregando(true);
    const t = setTimeout(async () => {
      try {
        const [chamados, itens] = await Promise.all([
          ordensServicoApi.listar({ q, tamanho: 6 }).then((r) => r.dados).catch(() => [] as OrdemServico[]),
          itensApi.listar({ q, tamanho: 6 }).then((r) => r.dados).catch(() => [] as Item[]),
        ]);
        const res: Resultado[] = [];
        for (const os of chamados) {
          res.push({
            chave: `os-${os.id}`,
            icone: <Wrench className="h-4 w-4 text-muted-foreground" />,
            titulo: `${os.numero} — ${os.assunto || os.equipamento_snapshot || "Chamado"}`,
            sub:
              os.solicitante?.nome ||
              os.solicitante_nome_snapshot ||
              os.solicitante_contato ||
              "Chamado",
            rota: `/manutencao/${os.id}`,
          });
        }
        for (const it of itens) {
          const consumivel = it.categoria?.consumivel;
          res.push({
            chave: `item-${it.id}`,
            icone: consumivel ? (
              <Package className="h-4 w-4 text-muted-foreground" />
            ) : (
              <Monitor className="h-4 w-4 text-muted-foreground" />
            ),
            titulo: it.descricao,
            sub:
              (it.numero_patrimonio ? `Patrimônio ${it.numero_patrimonio} · ` : "") +
              (it.categoria?.nome ?? ""),
            rota: `${consumivel ? "/itens" : "/patrimonio"}/${it.id}`,
          });
        }
        setResultados(res);
        setAtivo(0);
      } finally {
        setCarregando(false);
      }
    }, 250);
    return () => clearTimeout(t);
  }, [termo]);

  function abrirResultado(r: Resultado) {
    setAberto(false);
    navigate(r.rota);
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setAtivo((i) => Math.min(i + 1, resultados.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setAtivo((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && resultados[ativo]) {
      e.preventDefault();
      abrirResultado(resultados[ativo]);
    }
  }

  if (!aberto) {
    // Gatilho discreto no header (também abre por Ctrl+K).
    return (
      <button
        type="button"
        onClick={() => setAberto(true)}
        className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 text-sm text-muted-foreground hover:bg-accent"
        aria-label="Buscar (Ctrl+K)"
      >
        <Search className="h-4 w-4" />
        <span className="hidden sm:inline">Buscar…</span>
        <kbd className="hidden rounded bg-muted px-1 text-[10px] sm:inline">Ctrl K</kbd>
      </button>
    );
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 p-4 pt-24"
      onClick={() => setAberto(false)}
    >
      <div
        className="w-full max-w-lg overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 border-b border-border px-3">
          <Search className="h-4 w-4 text-muted-foreground" />
          <input
            ref={inputRef}
            value={termo}
            onChange={(e) => setTermo(e.target.value)}
            onKeyDown={onKeyDown}
            placeholder="Buscar chamado, item, patrimônio, solicitante…"
            className="h-12 flex-1 bg-transparent text-sm outline-none"
          />
          {carregando && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
        </div>
        <div className="max-h-80 overflow-y-auto">
          {termo.trim().length < 2 ? (
            <p className="px-4 py-6 text-center text-sm text-muted-foreground">
              Digite ao menos 2 caracteres.
            </p>
          ) : resultados.length === 0 && !carregando ? (
            <p className="px-4 py-6 text-center text-sm text-muted-foreground">
              Nada encontrado.
            </p>
          ) : (
            resultados.map((r, i) => (
              <button
                key={r.chave}
                type="button"
                onClick={() => abrirResultado(r)}
                onMouseEnter={() => setAtivo(i)}
                className={cn(
                  "flex w-full items-center gap-3 px-4 py-2.5 text-left",
                  i === ativo ? "bg-accent" : "hover:bg-accent/50"
                )}
              >
                {r.icone}
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">{r.titulo}</span>
                  <span className="block truncate text-xs text-muted-foreground">{r.sub}</span>
                </span>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
