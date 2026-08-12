import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, Eye } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { ArtigoForm } from "./ArtigoForm";
import { conhecimentoApi, type FiltroConhecimento } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { formatarData } from "@/lib/format";
import type { ArtigoConhecimento } from "@/types";

const TAMANHO = 20;

export default function Conhecimento() {
  const { toast } = useToast();
  const navigate = useNavigate();

  const [lista, setLista] = React.useState<ArtigoConhecimento[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [busca, setBusca] = React.useState("");
  const [buscaAtiva, setBuscaAtiva] = React.useState("");
  const [carregando, setCarregando] = React.useState(true);
  const [formAberto, setFormAberto] = React.useState(false);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const f: FiltroConhecimento = {
      pagina,
      tamanho: TAMANHO,
      q: buscaAtiva || undefined,
    };
    conhecimentoApi
      .listar(f)
      .then((r) => {
        setLista(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar a base de conhecimento",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [pagina, buscaAtiva, toast]);

  React.useEffect(carregar, [carregar]);

  function submeterBusca(e: React.FormEvent) {
    e.preventDefault();
    setPagina(1);
    setBuscaAtiva(busca.trim());
  }

  return (
    <div>
      <PageHeader
        titulo="Base de Conhecimento"
        descricao="Procedimentos, soluções e orientações da equipe de T.I."
        acao={
          <Button onClick={() => setFormAberto(true)}>
            <Plus className="h-4 w-4" /> Novo artigo
          </Button>
        }
      />

      <form onSubmit={submeterBusca} className="mb-4 flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            placeholder="Buscar por título ou conteúdo…"
            className="pl-9"
          />
        </div>
        <Button type="submit" variant="outline">
          Buscar
        </Button>
        {buscaAtiva && (
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              setBusca("");
              setBuscaAtiva("");
              setPagina(1);
            }}
          >
            Limpar
          </Button>
        )}
      </form>

      {carregando ? (
        <CarregandoTela />
      ) : lista.length === 0 ? (
        <Card>
          <CardContent className="p-0">
            <EstadoVazio
              titulo="Nenhum artigo encontrado"
              descricao={
                buscaAtiva
                  ? "Tente outro termo de busca."
                  : "Crie o primeiro artigo da base de conhecimento."
              }
            />
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {lista.map((a) => (
            <Card
              key={a.id}
              className="cursor-pointer transition-colors hover:bg-accent/40"
              onClick={() => navigate(`/conhecimento/${a.id}`)}
            >
              <CardContent className="flex items-start justify-between gap-4 p-4">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-medium text-foreground">{a.titulo}</h3>
                    {a.categoria && (
                      <Badge variant="secondary">{a.categoria}</Badge>
                    )}
                    {!a.publicado && <Badge variant="muted">Rascunho</Badge>}
                  </div>
                  <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                    {a.conteudo}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {a.autor?.nome ? `Por ${a.autor.nome} · ` : ""}
                    Atualizado em {formatarData(a.atualizado_em)}
                  </p>
                </div>
                <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                  <Eye className="h-3.5 w-3.5" /> {a.visualizacoes}
                </span>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {total > TAMANHO && (
        <div className="mt-4 flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {total} artigo(s) · página {pagina} de {totalPaginas}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={pagina <= 1 || carregando}
              onClick={() => setPagina((p) => p - 1)}
            >
              Anterior
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={pagina >= totalPaginas || carregando}
              onClick={() => setPagina((p) => p + 1)}
            >
              {carregando ? <Spinner className="h-4 w-4" /> : "Próxima"}
            </Button>
          </div>
        </div>
      )}

      <ArtigoForm
        aberto={formAberto}
        artigo={null}
        onFechar={() => setFormAberto(false)}
        onSalvo={() => {
          setFormAberto(false);
          setPagina(1);
          carregar();
        }}
      />
    </div>
  );
}
