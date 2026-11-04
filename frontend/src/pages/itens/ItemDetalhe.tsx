import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  ArrowLeft,
  Pencil,
  ArrowLeftRight,
  FileText,
  AlertTriangle,
  PlusCircle,
  Wrench,
} from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EstadoBadge } from "@/components/EstadoBadge";
import { HistoricoTabela } from "@/components/HistoricoTabela";
import { ItemForm } from "./ItemForm";
import { MovimentacaoForm } from "@/components/MovimentacaoForm";
import { TermoForm } from "@/components/TermoForm";
import { CarregandoTela } from "@/components/ui/spinner";
import { itensApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useReferencias } from "@/hooks/useReferencias";
import { formatarData, formatarDataHora, formatarMoeda } from "@/lib/format";
import { codigoInterno } from "@/lib/patrimonio";
import { cn } from "@/lib/utils";
import QRCode from "qrcode";
import { QrCode, Printer } from "lucide-react";
import type {
  EventoTimeline,
  Item,
  Movimentacao,
  TipoMovimentacao,
} from "@/types";

export default function ItemDetalhe() {
  const { id } = useParams<{ id: string }>();
  const itemId = Number(id);
  const navigate = useNavigate();
  const { toast } = useToast();
  const { categorias, setores, servidores } = useReferencias();

  const [item, setItem] = React.useState<Item | null>(null);
  const [historico, setHistorico] = React.useState<Movimentacao[]>([]);
  const [timeline, setTimeline] = React.useState<EventoTimeline[]>([]);
  const [abaHist, setAbaHist] = React.useState<"timeline" | "mov">("timeline");
  const [carregando, setCarregando] = React.useState(true);

  const [editAberto, setEditAberto] = React.useState(false);
  const [movAberto, setMovAberto] = React.useState(false);
  const [movTipoInicial, setMovTipoInicial] = React.useState<
    TipoMovimentacao | undefined
  >(undefined);
  const [termoAberto, setTermoAberto] = React.useState(false);

  // Abre a movimentação, opcionalmente já com um tipo pré-selecionado (ex.:
  // "Adicionar estoque" abre com entrada por compra).
  function abrirMovimentacao(tipo?: TipoMovimentacao) {
    setMovTipoInicial(tipo);
    setMovAberto(true);
  }

  const carregar = React.useCallback(() => {
    setCarregando(true);
    // Item é o dado crítico: se carregar, a tela renderiza.
    itensApi
      .buscarPorId(itemId)
      .then((it) => setItem(it))
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o item",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
    // Histórico e linha do tempo são secundários: falha neles não derruba a
    // tela (ex.: backend antigo sem o endpoint /timeline).
    itensApi
      .historico(itemId)
      .then((h) => setHistorico(h ?? []))
      .catch(() => setHistorico([]));
    itensApi
      .timeline(itemId)
      .then((tl) => setTimeline(tl ?? []))
      .catch(() => setTimeline([]));
  }, [itemId, toast]);

  React.useEffect(carregar, [carregar]);

  if (carregando) return <CarregandoTela texto="Carregando item..." />;
  if (!item)
    return (
      <div className="py-12 text-center text-muted-foreground">
        Item não encontrado.
      </div>
    );

  const alerta =
    item.estoque_minimo > 0 && item.quantidade < item.estoque_minimo;
  const bemPermanente = !item.categoria?.consumivel;

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate(item.categoria?.consumivel ? "/itens" : "/patrimonio")}
      >
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Button>

      <PageHeader
        titulo={item.descricao}
        descricao={`${item.categoria?.nome ?? "Item"} · ${codigoInterno(item.id)}`}
        acao={
          <>
            {!item.baixado && !bemPermanente && (
              <Button onClick={() => abrirMovimentacao("entrada_compra")}>
                <PlusCircle className="h-4 w-4" /> Adicionar estoque
              </Button>
            )}
            <Button variant="outline" onClick={() => abrirMovimentacao()}>
              <ArrowLeftRight className="h-4 w-4" /> Movimentar
            </Button>
            {bemPermanente && <Button variant="outline" onClick={() => setTermoAberto(true)}>
              <FileText className="h-4 w-4" /> Emitir termo
            </Button>}
            <Button variant="outline" onClick={() => setEditAberto(true)}>
              <Pencil className="h-4 w-4" /> Editar
            </Button>
          </>
        }
      />

      {item.baixado && (
        <div className="mb-4 flex items-center gap-2 rounded-md border border-border bg-muted/50 px-4 py-3 text-sm">
          <Badge variant="muted">Item baixado</Badge>
          <span className="text-muted-foreground">
            {item.motivo_baixa
              ? `Motivo: ${item.motivo_baixa}`
              : "Baixa patrimonial registrada."}
            {item.data_baixa && ` · ${formatarData(item.data_baixa)}`}
          </span>
        </div>
      )}

      {alerta && (
        <div className="mb-4 flex items-center gap-2 rounded-md border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <AlertTriangle className="h-4 w-4" />
          Estoque abaixo do mínimo ({item.quantidade} de {item.estoque_minimo}).
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base">Dados do item</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <Campo rotulo="Código interno">{codigoInterno(item.id)}</Campo>
            {item.lote_codigo && <Campo rotulo="Lote de recebimento">{item.lote_codigo}</Campo>}
            <Campo rotulo="Estado">
              <EstadoBadge estado={item.estado_conservacao} />
            </Campo>
            <Campo rotulo="Quantidade">
              <span className="flex items-center gap-2">
                {item.quantidade}
                {!item.baixado && !bemPermanente && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    aria-label="Adicionar ao estoque"
                    title="Adicionar ao estoque (registrar entrada)"
                    onClick={() => abrirMovimentacao("entrada_compra")}
                  >
                    <PlusCircle className="h-4 w-4" />
                  </Button>
                )}
              </span>
            </Campo>
            {!bemPermanente && <Campo rotulo="Estoque mínimo">{item.estoque_minimo}</Campo>}
            {bemPermanente && <Campo rotulo="Nº patrimônio">
              {item.numero_patrimonio ?? "Aguardando tombamento"}
            </Campo>}
            <Campo rotulo="Nº série">{item.numero_serie ?? "—"}</Campo>
            <Campo rotulo="Marca / modelo">
              {[item.marca, item.modelo].filter(Boolean).join(" · ") || "—"}
            </Campo>
            <Campo rotulo="Departamento">{item.setor?.nome ?? "Local pendente"}</Campo>
            <Campo rotulo="Responsável">
              {item.responsavel
                ? `${item.responsavel.nome} (${item.responsavel.matricula})`
                : "—"}
            </Campo>
            <Campo rotulo="Aquisição">
              {formatarData(item.data_aquisicao)}
            </Campo>
            <Campo rotulo="Valor">{formatarMoeda(item.valor)}</Campo>
            {bemPermanente && (
              <Campo rotulo="Garantia">
                <Garantia ate={item.garantia_ate} />
              </Campo>
            )}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between gap-2 space-y-0">
            <CardTitle className="text-base">
              {abaHist === "timeline"
                ? "Linha do tempo"
                : "Histórico de movimentações"}
            </CardTitle>
            <div className="flex gap-1">
              <AbaBotao
                ativo={abaHist === "timeline"}
                onClick={() => setAbaHist("timeline")}
              >
                Linha do tempo
              </AbaBotao>
              <AbaBotao
                ativo={abaHist === "mov"}
                onClick={() => setAbaHist("mov")}
              >
                Movimentações
              </AbaBotao>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {abaHist === "mov" ? (
              <HistoricoTabela movimentacoes={historico} />
            ) : (
              <LinhaDoTempo eventos={timeline} onAbrirOS={(id) => navigate(`/manutencao/${id}`)} />
            )}
          </CardContent>
        </Card>
      </div>

      {bemPermanente && <EtiquetaQR item={item} />}

      <ItemForm
        aberto={editAberto}
        item={item}
        categorias={categorias.filter((c) => c.consumivel === item.categoria?.consumivel)}
        setores={setores}
        servidores={servidores}
        onFechar={() => setEditAberto(false)}
        onSalvo={() => {
          setEditAberto(false);
          carregar();
        }}
        onRegistrarEntrada={bemPermanente ? undefined : () => {
          setEditAberto(false);
          abrirMovimentacao("entrada_compra");
        }}
      />

      <MovimentacaoForm
        aberto={movAberto}
        itemFixo={item}
        setores={setores}
        servidores={servidores}
        tipoInicial={movTipoInicial}
        onFechar={() => setMovAberto(false)}
        onRegistrado={() => {
          setMovAberto(false);
          carregar();
        }}
      />

      <TermoForm
        aberto={termoAberto}
        itemFixo={item}
        servidores={servidores}
        onFechar={() => setTermoAberto(false)}
        onEmitido={() => setTermoAberto(false)}
      />
    </div>
  );
}

// Garantia exibe a data de garantia com destaque quando está vencida ou perto.
function Garantia({ ate }: { ate?: string | null }) {
  if (!ate) return <>—</>;
  const dias = Math.ceil(
    (new Date(ate).getTime() - Date.now()) / (1000 * 60 * 60 * 24)
  );
  const cor =
    dias < 0 ? "text-destructive" : dias <= 30 ? "text-amber-600" : "";
  const sufixo =
    dias < 0 ? " (vencida)" : dias <= 30 ? ` (${dias}d)` : "";
  return (
    <span className={cor}>
      {formatarData(ate)}
      {sufixo}
    </span>
  );
}

function AbaBotao({
  ativo,
  onClick,
  children,
}: {
  ativo: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-md border px-2.5 py-1 text-xs font-medium transition-colors",
        ativo
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border bg-card text-muted-foreground hover:bg-accent"
      )}
    >
      {children}
    </button>
  );
}

// LinhaDoTempo renderiza os eventos consolidados (movimentação, OS, termo) numa
// trilha vertical, do mais recente ao mais antigo.
function LinhaDoTempo({
  eventos,
  onAbrirOS,
}: {
  eventos: EventoTimeline[];
  onAbrirOS: (id: number) => void;
}) {
  if (!eventos || eventos.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        Nenhum evento registrado para este item.
      </p>
    );
  }

  const estilo: Record<
    EventoTimeline["tipo"],
    { icone: React.ElementType; cor: string; rotulo: string }
  > = {
    movimentacao: {
      icone: ArrowLeftRight,
      cor: "bg-sky-100 text-sky-700",
      rotulo: "Movimentação",
    },
    os: { icone: Wrench, cor: "bg-amber-100 text-amber-700", rotulo: "Manutenção" },
    termo: {
      icone: FileText,
      cor: "bg-violet-100 text-violet-700",
      rotulo: "Termo",
    },
  };

  return (
    <ol className="space-y-0 p-4">
      {eventos.map((e, i) => {
        const s = estilo[e.tipo];
        const Icone = s.icone;
        const clicavel = e.tipo === "os";
        return (
          <li key={`${e.tipo}-${e.ref_id}-${i}`} className="flex gap-3">
            {/* Trilho + marcador. */}
            <div className="flex flex-col items-center">
              <span
                className={cn(
                  "flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
                  s.cor
                )}
              >
                <Icone className="h-4 w-4" />
              </span>
              {i < eventos.length - 1 && (
                <span className="my-1 w-px flex-1 bg-border" />
              )}
            </div>
            {/* Conteúdo. */}
            <div
              className={cn(
                "flex-1 pb-5",
                clicavel && "cursor-pointer"
              )}
              onClick={clicavel ? () => onAbrirOS(e.ref_id) : undefined}
            >
              <div className="flex flex-wrap items-center gap-x-2">
                <span className="text-sm font-medium text-foreground">
                  {e.titulo}
                </span>
                <span className="text-[11px] text-muted-foreground">
                  {s.rotulo}
                </span>
              </div>
              {e.detalhe && (
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {e.detalhe}
                </p>
              )}
              <p className="mt-0.5 text-xs text-muted-foreground">
                {formatarDataHora(e.data)}
                {clicavel && " · abrir OS"}
              </p>
            </div>
          </li>
        );
      })}
    </ol>
  );
}

function Campo({
  rotulo,
  children,
}: {
  rotulo: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-2 last:border-0">
      <span className="text-muted-foreground">{rotulo}</span>
      <span className="text-right font-medium text-foreground">{children}</span>
    </div>
  );
}

// EtiquetaQR gera um QR Code que aponta para a página do bem. ColadA no
// equipamento, abre o item ao escanear com o celular. Permite imprimir a
// etiqueta (QR + código TI + patrimônio).
function EtiquetaQR({ item }: { item: Item }) {
  const [dataUrl, setDataUrl] = React.useState("");
  const url = `${window.location.origin}/patrimonio/${item.id}`;
  const codigo = codigoInterno(item.id);

  React.useEffect(() => {
    QRCode.toDataURL(url, { width: 220, margin: 1 })
      .then(setDataUrl)
      .catch(() => setDataUrl(""));
  }, [url]);

  function imprimir() {
    const w = window.open("", "_blank", "width=400,height=500");
    if (!w) return;
    w.document.write(`
      <html><head><title>Etiqueta ${codigo}</title>
      <style>
        body{font-family:system-ui,sans-serif;text-align:center;padding:16px;margin:0}
        img{width:200px;height:200px}
        .cod{font-family:monospace;font-size:14px;margin-top:8px}
        .desc{font-size:13px;color:#333;margin-top:4px}
        .pat{font-size:12px;color:#555;margin-top:2px}
      </style></head><body>
        <img src="${dataUrl}" alt="QR"/>
        <div class="cod">${codigo}</div>
        <div class="desc">${(item.descricao || "").replace(/</g, "&lt;")}</div>
        ${item.numero_patrimonio ? `<div class="pat">Patrimônio ${item.numero_patrimonio}</div>` : ""}
      </body></html>`);
    w.document.close();
    w.focus();
    setTimeout(() => w.print(), 300);
  }

  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <QrCode className="h-4 w-4" /> Etiqueta QR
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-3 sm:flex-row sm:items-center sm:gap-6">
        {dataUrl ? (
          <img src={dataUrl} alt="QR Code do item" className="h-36 w-36" />
        ) : (
          <div className="h-36 w-36 animate-pulse rounded bg-muted" />
        )}
        <div className="text-center sm:text-left">
          <p className="text-sm text-muted-foreground">
            Escaneie para abrir este bem no celular.
          </p>
          <p className="mt-1 font-mono text-xs">{codigo}</p>
          <Button
            variant="outline"
            size="sm"
            className="mt-3"
            onClick={imprimir}
            disabled={!dataUrl}
          >
            <Printer className="h-4 w-4" /> Imprimir etiqueta
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
