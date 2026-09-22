import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Search, ArrowLeft, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela } from "@/components/ui/spinner";
import { portalApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { ArtigoConhecimento } from "@/types";

export default function PortalAjuda() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const [lista, setLista] = React.useState<ArtigoConhecimento[]>([]);
  const [busca, setBusca] = React.useState("");
  const [buscaAtiva, setBuscaAtiva] = React.useState("");
  const [carregando, setCarregando] = React.useState(true);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    portalApi
      .ajudaListar(buscaAtiva || undefined)
      .then((r) => setLista(r.dados))
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar a central de ajuda",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [buscaAtiva, toast]);

  React.useEffect(carregar, [carregar]);

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate("/")}
      >
        <ArrowLeft className="h-4 w-4" /> Meus chamados
      </Button>

      <div className="mb-4">
        <h1 className="text-xl font-semibold text-foreground">
          Central de Ajuda
        </h1>
        <p className="text-sm text-muted-foreground">
          Procure uma solução antes de abrir um chamado.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          setBuscaAtiva(busca.trim());
        }}
        className="mb-4 flex gap-2"
      >
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            placeholder="Buscar ajuda… (ex.: e-mail, impressora)"
            className="pl-9"
          />
        </div>
        <Button type="submit" variant="outline">
          Buscar
        </Button>
      </form>

      {carregando ? (
        <CarregandoTela />
      ) : lista.length === 0 ? (
        <Card>
          <CardContent className="p-0">
            <EstadoVazio
              titulo="Nada encontrado"
              descricao="Não achou o que procura? Abra um chamado que a equipe ajuda."
            />
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {lista.map((a) => (
            <Card
              key={a.id}
              className="cursor-pointer transition-colors hover:bg-accent/40"
              onClick={() => navigate(`/ajuda/${a.id}`)}
            >
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <h3 className="font-medium text-foreground">{a.titulo}</h3>
                  {a.categoria && (
                    <Badge variant="secondary">{a.categoria}</Badge>
                  )}
                </div>
                <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                  {a.conteudo}
                </p>
                <p className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                  <Eye className="h-3 w-3" /> {a.visualizacoes}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
