import * as React from "react";
import { Trash2, LogOut, LogIn, PackagePlus } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { itensApi, reservasApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import { rotuloStatusReserva, varianteStatusReserva } from "@/lib/rotulos";
import { dateInputParaISO, formatarData } from "@/lib/format";
import type {
  EquipamentoReservavel,
  Item,
  ReservaPayload,
  Servidor,
} from "@/types";

const SEM = "0";

export default function Reservas() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();
  const { servidores } = useReferencias();

  const [equipamentos, setEquipamentos] = React.useState<
    EquipamentoReservavel[]
  >([]);
  const [carregando, setCarregando] = React.useState(true);
  const [processando, setProcessando] = React.useState<number | null>(null);

  // Diálogo de alocação (item fixo).
  const [alocarItem, setAlocarItem] = React.useState<Item | null>(null);
  // Diálogo "adicionar equipamento ao pool".
  const [adicionarAberto, setAdicionarAberto] = React.useState(false);
  // Remoção do pool.
  const [removerPool, setRemoverPool] = React.useState<Item | null>(null);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    reservasApi
      .equipamentos()
      .then(setEquipamentos)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar equipamentos",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  async function voltarAoDepartamento(eq: EquipamentoReservavel) {
    if (!eq.reserva_ativa) return;
    setProcessando(eq.item.id);
    try {
      await reservasApi.definirStatus(eq.reserva_ativa.id, "devolvida");
      toast({ titulo: "Equipamento de volta ao departamento.", variant: "success" });
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível registrar a devolução",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProcessando(null);
    }
  }

  async function confirmarRemoverPool() {
    if (!removerPool) return;
    setProcessando(removerPool.id);
    try {
      await itensApi.definirReservavel(removerPool.id, false);
      toast({ titulo: "Equipamento removido do pool.", variant: "success" });
      setRemoverPool(null);
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível remover",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProcessando(null);
    }
  }

  const noDepartamento = equipamentos.filter((e) => !e.reserva_ativa).length;
  const alocados = equipamentos.length - noDepartamento;

  return (
    <div>
      <PageHeader
        titulo="Reservas de Equipamento"
        descricao="Pool de equipamentos disponíveis para alocação temporária."
        acao={
          <Button onClick={() => setAdicionarAberto(true)}>
            <PackagePlus className="h-4 w-4" /> Adicionar equipamento
          </Button>
        }
      />

      <div className="mb-4 flex gap-3 text-sm">
        <Badge variant="success">No departamento: {noDepartamento}</Badge>
        <Badge variant="warning">Alocados: {alocados}</Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Equipamentos</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : equipamentos.length === 0 ? (
            <EstadoVazio
              titulo="Nenhum equipamento no pool"
              descricao="Clique em 'Adicionar equipamento' para incluir computadores disponíveis para reserva."
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Equipamento</TableHead>
                  <TableHead>Patrimônio</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead className="text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {equipamentos.map((eq) => {
                  const r = eq.reserva_ativa;
                  return (
                    <TableRow key={eq.item.id}>
                      <TableCell className="font-medium">
                        {eq.item.descricao}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {eq.item.numero_patrimonio || "—"}
                      </TableCell>
                      <TableCell>
                        {r ? (
                          <div className="flex flex-col">
                            <Badge
                              variant={varianteStatusReserva(r.status)}
                              className="w-fit"
                            >
                              {rotuloStatusReserva(r.status)}
                            </Badge>
                            <span className="mt-1 text-xs text-muted-foreground">
                              {r.local_destino ? `Em ${r.local_destino}` : "Alocado"}
                              {" · "}
                              {formatarData(r.data_inicio)} — {formatarData(r.data_fim)}
                              {r.solicitante?.nome && ` · ${r.solicitante.nome}`}
                            </span>
                          </div>
                        ) : (
                          <Badge variant="success">No departamento</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          {processando === eq.item.id ? (
                            <Spinner className="h-4 w-4" />
                          ) : r ? (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => voltarAoDepartamento(eq)}
                            >
                              <LogIn className="h-4 w-4" /> Voltar ao
                              departamento
                            </Button>
                          ) : (
                            <>
                              <Button
                                size="sm"
                                onClick={() => setAlocarItem(eq.item)}
                              >
                                <LogOut className="h-4 w-4" /> Alocar
                              </Button>
                              {ehAdministrador && (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  aria-label="Remover do pool"
                                  title="Remover do pool de reservas"
                                  onClick={() => setRemoverPool(eq.item)}
                                >
                                  <Trash2 className="h-4 w-4 text-destructive" />
                                </Button>
                              )}
                            </>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <AlocarDialog
        item={alocarItem}
        servidores={servidores}
        onFechar={() => setAlocarItem(null)}
        onAlocado={() => {
          setAlocarItem(null);
          carregar();
        }}
      />

      <AdicionarEquipamentoDialog
        aberto={adicionarAberto}
        jaNoPool={equipamentos.map((e) => e.item.id)}
        onFechar={() => setAdicionarAberto(false)}
        onAdicionado={() => {
          setAdicionarAberto(false);
          carregar();
        }}
      />

      <ConfirmDialog
        aberto={!!removerPool}
        titulo="Remover do pool de reservas"
        descricao={`Remover "${removerPool?.descricao}" do pool? O equipamento continua no inventário.`}
        textoConfirmar="Remover"
        destrutivo
        processando={processando === removerPool?.id}
        onConfirmar={confirmarRemoverPool}
        onCancelar={() => setRemoverPool(null)}
      />
    </div>
  );
}

// AlocarDialog cria uma reserva "em uso" para um equipamento (sai do
// departamento para um destino).
function AlocarDialog({
  item,
  servidores,
  onFechar,
  onAlocado,
}: {
  item: Item | null;
  servidores: Servidor[];
  onFechar: () => void;
  onAlocado: () => void;
}) {
  const { toast } = useToast();
  const [local, setLocal] = React.useState("");
  const [inicio, setInicio] = React.useState("");
  const [fim, setFim] = React.useState("");
  const [solicitanteId, setSolicitanteId] = React.useState(SEM);
  const [finalidade, setFinalidade] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (item) {
      const hoje = new Date().toISOString().slice(0, 10);
      setLocal("");
      setInicio(hoje);
      setFim(hoje);
      setSolicitanteId(SEM);
      setFinalidade("");
      setErros({});
    }
  }, [item]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    if (!item) return;
    const novos: Record<string, string> = {};
    if (!local.trim()) novos.local_destino = "Informe para onde vai.";
    if (!inicio) novos.data_inicio = "Informe o início.";
    if (!fim) novos.data_fim = "Informe o término.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});
    setSalvando(true);
    try {
      const payload: ReservaPayload = {
        item_id: item.id,
        local_destino: local.trim(),
        data_inicio: dateInputParaISO(inicio),
        data_fim: dateInputParaISO(fim),
        solicitante_id: solicitanteId !== SEM ? Number(solicitanteId) : null,
        finalidade: finalidade.trim(),
        status: "em_uso",
      };
      await reservasApi.criar(payload);
      toast({ titulo: "Equipamento alocado.", variant: "success" });
      onAlocado();
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível alocar",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={!!item} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Alocar equipamento</DialogTitle>
        </DialogHeader>
        {item && (
          <div className="mb-2 rounded-md border border-border bg-muted/40 px-3 py-2 text-sm">
            <span className="font-medium">{item.descricao}</span>
            {item.numero_patrimonio ? ` — ${item.numero_patrimonio}` : ""}
          </div>
        )}
        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField
            label="Para onde vai (local/destino)"
            htmlFor="local"
            obrigatorio
            erro={erros.local_destino}
          >
            <Input
              id="local"
              value={local}
              onChange={(e) => setLocal(e.target.value)}
              placeholder="Ex.: Sala 3 / Secretaria de Saúde"
            />
          </FormField>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Início" htmlFor="di" obrigatorio erro={erros.data_inicio}>
              <Input
                id="di"
                type="date"
                value={inicio}
                onChange={(e) => setInicio(e.target.value)}
              />
            </FormField>
            <FormField label="Previsão de retorno" htmlFor="df" obrigatorio erro={erros.data_fim}>
              <Input
                id="df"
                type="date"
                value={fim}
                onChange={(e) => setFim(e.target.value)}
              />
            </FormField>
          </div>
          <FormField label="Responsável (opcional)">
            <Select value={solicitanteId} onValueChange={setSolicitanteId}>
              <SelectTrigger>
                <SelectValue placeholder="Opcional" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={SEM}>Não informado</SelectItem>
                {servidores.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.nome} ({s.matricula})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
          <FormField label="Finalidade (opcional)" htmlFor="fin">
            <Textarea
              id="fin"
              value={finalidade}
              onChange={(e) => setFinalidade(e.target.value)}
            />
          </FormField>
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Alocar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// AdicionarEquipamentoDialog marca um item do inventário como reservável.
function AdicionarEquipamentoDialog({
  aberto,
  jaNoPool,
  onFechar,
  onAdicionado,
}: {
  aberto: boolean;
  jaNoPool: number[];
  onFechar: () => void;
  onAdicionado: () => void;
}) {
  const { toast } = useToast();
  const [itens, setItens] = React.useState<Item[]>([]);
  const [itemId, setItemId] = React.useState("");
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (!aberto) return;
    setItemId("");
    itensApi
      .listar({ tamanho: 500, baixado: false })
      .then((r) =>
        setItens(
          r.dados.filter(
            (i) =>
              !i.reservavel &&
              !i.categoria?.consumivel &&
              !jaNoPool.includes(i.id)
          )
        )
      )
      .catch(() => setItens([]));
  }, [aberto, jaNoPool]);

  async function salvar() {
    if (!itemId) return;
    setSalvando(true);
    try {
      await itensApi.definirReservavel(Number(itemId), true);
      toast({ titulo: "Equipamento adicionado ao pool.", variant: "success" });
      onAdicionado();
    } catch (err) {
      toast({
        titulo: "Não foi possível adicionar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Adicionar equipamento ao pool</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Escolha um equipamento patrimoniado para disponibilizar para reserva.
          </p>
          <FormField label="Equipamento">
            <Select value={itemId} onValueChange={setItemId}>
              <SelectTrigger>
                <SelectValue placeholder="Selecione um equipamento" />
              </SelectTrigger>
              <SelectContent>
                {itens.map((i) => (
                  <SelectItem key={i.id} value={String(i.id)}>
                    {i.descricao}
                    {i.numero_patrimonio ? ` — ${i.numero_patrimonio}` : ""}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
          {itens.length === 0 && (
            <p className="text-xs text-muted-foreground">
              Nenhum equipamento disponível. Cadastre itens patrimoniados ou
              marque "Disponível para reserva" no cadastro do item.
            </p>
          )}
        </div>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onFechar}>
            Cancelar
          </Button>
          <Button onClick={salvar} disabled={salvando || !itemId}>
            {salvando ? <Spinner className="h-4 w-4" /> : "Adicionar"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
