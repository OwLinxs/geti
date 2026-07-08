import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  ArrowLeft,
  Pencil,
  ArrowLeftRight,
  FileText,
  AlertTriangle,
  PlusCircle,
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
import { formatarData, formatarMoeda } from "@/lib/format";
import type { Item, Movimentacao, TipoMovimentacao } from "@/types";

export default function ItemDetalhe() {
  const { id } = useParams<{ id: string }>();
  const itemId = Number(id);
  const navigate = useNavigate();
  const { toast } = useToast();
  const { categorias, setores, servidores } = useReferencias();

  const [item, setItem] = React.useState<Item | null>(null);
  const [historico, setHistorico] = React.useState<Movimentacao[]>([]);
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
    Promise.all([itensApi.buscarPorId(itemId), itensApi.historico(itemId)])
      .then(([it, hist]) => {
        setItem(it);
        setHistorico(hist);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o item",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
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

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate("/itens")}
      >
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Button>

      <PageHeader
        titulo={item.descricao}
        descricao={item.categoria?.nome}
        acao={
          <>
            {!item.baixado && (
              <Button onClick={() => abrirMovimentacao("entrada_compra")}>
                <PlusCircle className="h-4 w-4" /> Adicionar estoque
              </Button>
            )}
            <Button variant="outline" onClick={() => abrirMovimentacao()}>
              <ArrowLeftRight className="h-4 w-4" /> Movimentar
            </Button>
            <Button variant="outline" onClick={() => setTermoAberto(true)}>
              <FileText className="h-4 w-4" /> Emitir termo
            </Button>
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
            <Campo rotulo="Estado">
              <EstadoBadge estado={item.estado_conservacao} />
            </Campo>
            <Campo rotulo="Quantidade">
              <span className="flex items-center gap-2">
                {item.quantidade}
                {!item.baixado && (
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
            <Campo rotulo="Estoque mínimo">{item.estoque_minimo}</Campo>
            <Campo rotulo="Nº patrimônio">
              {item.numero_patrimonio ?? "—"}
            </Campo>
            <Campo rotulo="Nº série">{item.numero_serie ?? "—"}</Campo>
            <Campo rotulo="Marca / modelo">
              {[item.marca, item.modelo].filter(Boolean).join(" · ") || "—"}
            </Campo>
            <Campo rotulo="Departamento">{item.setor?.nome ?? "—"}</Campo>
            <Campo rotulo="Responsável">
              {item.responsavel
                ? `${item.responsavel.nome} (${item.responsavel.matricula})`
                : "—"}
            </Campo>
            <Campo rotulo="Aquisição">
              {formatarData(item.data_aquisicao)}
            </Campo>
            <Campo rotulo="Valor">{formatarMoeda(item.valor)}</Campo>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">
              Histórico de movimentações
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <HistoricoTabela movimentacoes={historico} />
          </CardContent>
        </Card>
      </div>

      <ItemForm
        aberto={editAberto}
        item={item}
        categorias={categorias}
        setores={setores}
        servidores={servidores}
        onFechar={() => setEditAberto(false)}
        onSalvo={() => {
          setEditAberto(false);
          carregar();
        }}
        onRegistrarEntrada={() => {
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
