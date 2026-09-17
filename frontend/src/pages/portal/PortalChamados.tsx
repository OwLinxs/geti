import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { portalApi, type AbrirChamadoPayload } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { formatarData } from "@/lib/format";
import {
  PRIORIDADES_OS,
  rotuloStatusOS,
  varianteStatusOS,
} from "@/lib/rotulos";
import type { OrdemServico } from "@/types";

export default function PortalChamados() {
  const { toast } = useToast();
  const navigate = useNavigate();

  const [lista, setLista] = React.useState<OrdemServico[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [aberto, setAberto] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    portalApi
      .listar({ tamanho: 100 })
      .then((r) => setLista(r.dados))
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar seus chamados",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-foreground">
            Meus chamados
          </h1>
          <p className="text-sm text-muted-foreground">
            Acompanhe suas solicitações à equipe de T.I.
          </p>
        </div>
        <Button onClick={() => setAberto(true)}>
          <Plus className="h-4 w-4" /> Abrir chamado
        </Button>
      </div>

      {carregando ? (
        <CarregandoTela />
      ) : lista.length === 0 ? (
        <Card>
          <CardContent className="p-0">
            <EstadoVazio
              titulo="Você ainda não tem chamados"
              descricao="Clique em 'Abrir chamado' para solicitar suporte."
            />
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {lista.map((c) => (
            <Card
              key={c.id}
              className="cursor-pointer transition-colors hover:bg-accent/40"
              onClick={() => navigate(`/chamado/${c.id}`)}
            >
              <CardContent className="flex items-center justify-between gap-4 p-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-muted-foreground">
                      {c.numero}
                    </span>
                    <Badge variant={varianteStatusOS(c.status)}>
                      {rotuloStatusOS(c.status)}
                    </Badge>
                  </div>
                  <p className="mt-1 truncate font-medium">
                    {c.assunto || c.equipamento_snapshot}
                  </p>
                  <p className="truncate text-sm text-muted-foreground">
                    {c.defeito_relatado}
                  </p>
                </div>
                <span className="shrink-0 text-xs text-muted-foreground">
                  {formatarData(c.data_abertura)}
                </span>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <AbrirChamadoDialog
        aberto={aberto}
        onFechar={() => setAberto(false)}
        onCriado={(id) => {
          setAberto(false);
          navigate(`/chamado/${id}`);
        }}
      />
    </div>
  );
}

function AbrirChamadoDialog({
  aberto,
  onFechar,
  onCriado,
}: {
  aberto: boolean;
  onFechar: () => void;
  onCriado: (id: number) => void;
}) {
  const { toast } = useToast();
  const [assunto, setAssunto] = React.useState("");
  const [descricao, setDescricao] = React.useState("");
  const [local, setLocal] = React.useState("");
  const [prioridade, setPrioridade] = React.useState("normal");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [enviando, setEnviando] = React.useState(false);

  React.useEffect(() => {
    if (aberto) {
      setAssunto("");
      setDescricao("");
      setLocal("");
      setPrioridade("normal");
      setErros({});
    }
  }, [aberto]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!assunto.trim()) novos.assunto = "Informe um assunto.";
    if (!descricao.trim()) novos.defeito_relatado = "Descreva o problema.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});
    setEnviando(true);
    try {
      const payload: AbrirChamadoPayload = {
        assunto: assunto.trim(),
        defeito_relatado: descricao.trim(),
        equipamento_descricao: local.trim(),
        prioridade,
      };
      const os = await portalApi.abrir(payload);
      toast({ titulo: "Chamado aberto!", variant: "success" });
      onCriado(os.id);
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível abrir o chamado",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setEnviando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-h-[90vh] max-w-lg overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Abrir chamado</DialogTitle>
        </DialogHeader>
        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField label="Assunto" htmlFor="assunto" obrigatorio erro={erros.assunto}>
            <Input
              id="assunto"
              value={assunto}
              onChange={(e) => setAssunto(e.target.value)}
              placeholder="Ex.: Computador não liga"
            />
          </FormField>
          <FormField
            label="Descrição do problema"
            htmlFor="descricao"
            obrigatorio
            erro={erros.defeito_relatado}
          >
            <Textarea
              id="descricao"
              value={descricao}
              onChange={(e) => setDescricao(e.target.value)}
              placeholder="Conte o que está acontecendo, quando começou, etc."
              rows={5}
            />
          </FormField>
          <FormField label="Equipamento / local (opcional)" htmlFor="local">
            <Input
              id="local"
              value={local}
              onChange={(e) => setLocal(e.target.value)}
              placeholder="Ex.: PC da sala 3 / patrimônio 12345"
            />
          </FormField>
          <FormField label="Prioridade">
            <Select value={prioridade} onValueChange={setPrioridade}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PRIORIDADES_OS.map((p) => (
                  <SelectItem key={p.valor} value={p.valor}>
                    {p.rotulo}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={enviando}>
              {enviando ? <Spinner className="h-4 w-4" /> : "Abrir chamado"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
