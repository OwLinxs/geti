import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { CarregandoTela } from "@/components/ui/spinner";
import { portalApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { ArtigoConhecimento } from "@/types";

export default function PortalArtigo() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [artigo, setArtigo] = React.useState<ArtigoConhecimento | null>(null);
  const [carregando, setCarregando] = React.useState(true);

  React.useEffect(() => {
    portalApi
      .ajudaArtigo(Number(id))
      .then(setArtigo)
      .catch((err) =>
        toast({
          titulo: "Erro ao abrir o artigo",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [id, toast]);

  if (carregando) return <CarregandoTela texto="Carregando…" />;
  if (!artigo)
    return (
      <div className="py-12 text-center text-muted-foreground">
        Artigo não encontrado.
      </div>
    );

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate("/ajuda")}
      >
        <ArrowLeft className="h-4 w-4" /> Central de Ajuda
      </Button>

      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h1 className="text-xl font-semibold text-foreground">
          {artigo.titulo}
        </h1>
        {artigo.categoria && <Badge variant="secondary">{artigo.categoria}</Badge>}
      </div>

      <Card>
        <CardContent className="p-6">
          <div className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
            {artigo.conteudo}
          </div>
        </CardContent>
      </Card>

      <div className="mt-4 rounded-md border border-border bg-muted/40 p-4 text-sm">
        Não resolveu?{" "}
        <button
          className="font-medium text-primary hover:underline"
          onClick={() => navigate("/")}
        >
          Abrir um chamado
        </button>
        .
      </div>
    </div>
  );
}
