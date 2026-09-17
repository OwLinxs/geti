import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConversaChamado } from "@/components/ConversaChamado";
import { CarregandoTela } from "@/components/ui/spinner";
import { portalApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { formatarDataHora } from "@/lib/format";
import {
  rotuloPrioridadeOS,
  rotuloStatusOS,
  variantePrioridadeOS,
  varianteStatusOS,
} from "@/lib/rotulos";
import type { OrdemServico } from "@/types";

export default function PortalChamadoDetalhe() {
  const { id } = useParams<{ id: string }>();
  const chamadoId = Number(id);
  const navigate = useNavigate();
  const { toast } = useToast();

  const [os, setOs] = React.useState<OrdemServico | null>(null);
  const [carregando, setCarregando] = React.useState(true);

  React.useEffect(() => {
    setCarregando(true);
    portalApi
      .buscarPorId(chamadoId)
      .then(setOs)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o chamado",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [chamadoId, toast]);

  if (carregando) return <CarregandoTela texto="Carregando chamado..." />;
  if (!os)
    return (
      <div className="py-12 text-center text-muted-foreground">
        Chamado não encontrado.
      </div>
    );

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

      <Card className="mb-4">
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs font-medium text-muted-foreground">
              {os.numero}
            </span>
            <Badge variant={varianteStatusOS(os.status)}>
              {rotuloStatusOS(os.status)}
            </Badge>
            <Badge variant={variantePrioridadeOS(os.prioridade)}>
              {rotuloPrioridadeOS(os.prioridade)}
            </Badge>
          </div>
          <CardTitle className="text-lg">
            {os.assunto || os.equipamento_snapshot}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p className="whitespace-pre-wrap text-foreground">
            {os.defeito_relatado}
          </p>
          <p className="text-xs text-muted-foreground">
            Aberto em {formatarDataHora(os.data_abertura)}
            {os.tecnico?.nome && ` · Técnico: ${os.tecnico.nome}`}
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Conversa</CardTitle>
        </CardHeader>
        <CardContent>
          <ConversaChamado osId={os.id} modo="portal" />
        </CardContent>
      </Card>
    </div>
  );
}
