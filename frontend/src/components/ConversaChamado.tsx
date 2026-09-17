import * as React from "react";
import { Paperclip, Send, Lock, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Spinner } from "@/components/ui/spinner";
import { chamadoMensagensApi, type BaseChamado } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { formatarDataHora } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { MensagemChamado } from "@/types";

interface Props {
  osId: number;
  // "equipe" = técnico (área interna, pode nota interna); "portal" = solicitante.
  modo: "equipe" | "portal";
}

export function ConversaChamado({ osId, modo }: Props) {
  const { toast } = useToast();
  const base: BaseChamado = modo === "equipe" ? "ordens-servico" : "meus-chamados";
  const meuTipo = modo === "equipe" ? "tecnico" : "servidor";

  const [msgs, setMsgs] = React.useState<MensagemChamado[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [texto, setTexto] = React.useState("");
  const [interna, setInterna] = React.useState(false);
  const [arquivo, setArquivo] = React.useState<File | null>(null);
  const [enviando, setEnviando] = React.useState(false);
  // Cache de object URLs de anexos de imagem (id da msg -> url).
  const [imagens, setImagens] = React.useState<Record<number, string>>({});
  const fimRef = React.useRef<HTMLDivElement | null>(null);

  const carregar = React.useCallback(
    (rolar = false) => {
      chamadoMensagensApi
        .listar(base, osId)
        .then((lista) => {
          setMsgs(lista);
          if (rolar)
            setTimeout(
              () => fimRef.current?.scrollIntoView({ behavior: "smooth" }),
              50
            );
        })
        .catch((err) =>
          toast({
            titulo: "Erro ao carregar a conversa",
            descricao: mensagemErro(err),
            variant: "destructive",
          })
        )
        .finally(() => setCarregando(false));
    },
    [base, osId, toast]
  );

  React.useEffect(() => {
    carregar(true);
    // Polling leve para receber respostas do outro lado.
    const t = setInterval(() => carregar(false), 15000);
    return () => clearInterval(t);
  }, [carregar]);

  // Carrega miniaturas de anexos de imagem.
  React.useEffect(() => {
    let ativo = true;
    for (const m of msgs) {
      if (
        m.anexo_tipo?.startsWith("image/") &&
        !imagens[m.id]
      ) {
        chamadoMensagensApi
          .baixarAnexo(base, osId, m.id)
          .then((blob) => {
            if (!ativo) return;
            const url = URL.createObjectURL(blob);
            setImagens((prev) => ({ ...prev, [m.id]: url }));
          })
          .catch(() => undefined);
      }
    }
    return () => {
      ativo = false;
    };
  }, [msgs, base, osId, imagens]);

  async function enviar(e: React.FormEvent) {
    e.preventDefault();
    if (!texto.trim() && !arquivo) return;
    setEnviando(true);
    try {
      await chamadoMensagensApi.enviar(base, osId, {
        texto: texto.trim(),
        interna,
        arquivo,
      });
      setTexto("");
      setArquivo(null);
      setInterna(false);
      carregar(true);
    } catch (err) {
      toast({
        titulo: "Não foi possível enviar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setEnviando(false);
    }
  }

  async function abrirAnexo(m: MensagemChamado) {
    // Abre em nova aba (HTTP-safe): busca o blob autenticado e exibe.
    const aba = window.open("", "_blank");
    try {
      const blob = await chamadoMensagensApi.baixarAnexo(base, osId, m.id);
      const url = URL.createObjectURL(blob);
      if (aba) aba.location.href = url;
      setTimeout(() => URL.revokeObjectURL(url), 60000);
    } catch (err) {
      if (aba) aba.close();
      toast({
        titulo: "Não foi possível abrir o anexo",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    }
  }

  return (
    <div className="flex h-[28rem] flex-col rounded-lg border border-border">
      <div className="flex-1 space-y-3 overflow-y-auto p-3">
        {carregando ? (
          <div className="flex h-full items-center justify-center">
            <Spinner className="h-5 w-5" />
          </div>
        ) : msgs.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">
            Nenhuma mensagem ainda. Inicie a conversa abaixo.
          </p>
        ) : (
          msgs.map((m) => {
            const meu = m.autor_tipo === meuTipo;
            return (
              <div
                key={m.id}
                className={cn("flex", meu ? "justify-end" : "justify-start")}
              >
                <div
                  className={cn(
                    "max-w-[80%] rounded-lg px-3 py-2 text-sm",
                    m.interna
                      ? "border border-amber-300 bg-amber-50 text-amber-900"
                      : meu
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-foreground"
                  )}
                >
                  <div className="mb-0.5 flex items-center gap-1 text-[11px] opacity-80">
                    {m.interna && <Lock className="h-3 w-3" />}
                    <span className="font-medium">
                      {m.autor_nome || (m.autor_tipo === "servidor" ? "Solicitante" : "Equipe")}
                    </span>
                    {m.interna && <span>· nota interna</span>}
                  </div>
                  {m.texto && <p className="whitespace-pre-wrap">{m.texto}</p>}
                  {m.anexo_nome && (
                    <div className="mt-1">
                      {imagens[m.id] ? (
                        <img
                          src={imagens[m.id]}
                          alt={m.anexo_nome}
                          className="max-h-48 cursor-pointer rounded-md"
                          onClick={() => abrirAnexo(m)}
                        />
                      ) : (
                        <button
                          type="button"
                          onClick={() => abrirAnexo(m)}
                          className="flex items-center gap-1 rounded-md bg-black/10 px-2 py-1 text-xs underline"
                        >
                          <Download className="h-3 w-3" /> {m.anexo_nome}
                        </button>
                      )}
                    </div>
                  )}
                  <div className="mt-0.5 text-right text-[10px] opacity-70">
                    {formatarDataHora(m.enviada_em)}
                  </div>
                </div>
              </div>
            );
          })
        )}
        <div ref={fimRef} />
      </div>

      <form onSubmit={enviar} className="border-t border-border p-2">
        {arquivo && (
          <div className="mb-2 flex items-center justify-between rounded-md bg-muted px-2 py-1 text-xs">
            <span className="truncate">📎 {arquivo.name}</span>
            <button
              type="button"
              className="text-destructive"
              onClick={() => setArquivo(null)}
            >
              remover
            </button>
          </div>
        )}
        <div className="flex items-end gap-2">
          <label className="cursor-pointer rounded-md border border-border p-2 hover:bg-accent">
            <Paperclip className="h-4 w-4" />
            <input
              type="file"
              className="hidden"
              onChange={(e) => setArquivo(e.target.files?.[0] ?? null)}
            />
          </label>
          <Textarea
            value={texto}
            onChange={(e) => setTexto(e.target.value)}
            placeholder="Escreva uma mensagem…"
            rows={1}
            className="min-h-10 flex-1 resize-none"
          />
          <Button type="submit" disabled={enviando}>
            {enviando ? <Spinner className="h-4 w-4" /> : <Send className="h-4 w-4" />}
          </Button>
        </div>
        {modo === "equipe" && (
          <label className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
            <input
              type="checkbox"
              checked={interna}
              onChange={(e) => setInterna(e.target.checked)}
            />
            Nota interna (não visível ao solicitante)
          </label>
        )}
      </form>
    </div>
  );
}
