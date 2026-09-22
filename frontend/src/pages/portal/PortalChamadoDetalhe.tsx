import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Star, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ConversaChamado } from "@/components/ConversaChamado";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
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
  const [nota, setNota] = React.useState(0);
  const [comentario, setComentario] = React.useState("");
  const [motivo, setMotivo] = React.useState("");
  const [acao, setAcao] = React.useState(false);

  async function avaliar() {
    if (!os || nota < 1) return;
    setAcao(true);
    try {
      const at = await portalApi.avaliar(os.id, nota, comentario.trim());
      setOs(at);
      toast({ titulo: "Obrigado pela avaliação!", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível avaliar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setAcao(false);
    }
  }

  async function reabrir() {
    if (!os) return;
    setAcao(true);
    try {
      const at = await portalApi.reabrir(os.id, motivo.trim());
      setOs(at);
      setMotivo("");
      toast({ titulo: "Chamado reaberto.", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível reabrir",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setAcao(false);
    }
  }

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
            {os.categoria_chamado?.nome && (
              <Badge variant="outline">{os.categoria_chamado.nome}</Badge>
            )}
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

      {os.status === "concluida" && (
        <Card className="mb-4">
          <CardHeader>
            <CardTitle className="text-base">Atendimento concluído</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {os.avaliacao_nota ? (
              <div>
                <p className="text-sm text-muted-foreground">Sua avaliação:</p>
                <div className="mt-1 flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((n) => (
                    <Star
                      key={n}
                      className={
                        n <= (os.avaliacao_nota ?? 0)
                          ? "h-5 w-5 fill-amber-400 text-amber-400"
                          : "h-5 w-5 text-muted-foreground"
                      }
                    />
                  ))}
                </div>
                {os.avaliacao_comentario && (
                  <p className="mt-2 text-sm">{os.avaliacao_comentario}</p>
                )}
              </div>
            ) : (
              <div>
                <p className="text-sm font-medium">Avalie o atendimento</p>
                <div className="mt-1 flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((n) => (
                    <button
                      key={n}
                      type="button"
                      onClick={() => setNota(n)}
                      aria-label={`${n} estrelas`}
                    >
                      <Star
                        className={
                          n <= nota
                            ? "h-7 w-7 fill-amber-400 text-amber-400"
                            : "h-7 w-7 text-muted-foreground hover:text-amber-400"
                        }
                      />
                    </button>
                  ))}
                </div>
                <Textarea
                  className="mt-2"
                  value={comentario}
                  onChange={(e) => setComentario(e.target.value)}
                  placeholder="Comentário (opcional)"
                />
                <Button
                  className="mt-2"
                  onClick={avaliar}
                  disabled={acao || nota < 1}
                >
                  {acao ? <Spinner className="h-4 w-4" /> : "Enviar avaliação"}
                </Button>
              </div>
            )}

            <div className="border-t border-border pt-3">
              <p className="text-sm text-muted-foreground">
                Não resolveu? Reabra o chamado.
              </p>
              <div className="mt-2 flex flex-col gap-2 sm:flex-row">
                <Input
                  value={motivo}
                  onChange={(e) => setMotivo(e.target.value)}
                  placeholder="Motivo da reabertura (opcional)"
                  className="flex-1"
                />
                <Button variant="outline" onClick={reabrir} disabled={acao}>
                  <RotateCcw className="h-4 w-4" /> Reabrir chamado
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

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
