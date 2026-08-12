import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
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
import {
  itensApi,
  reservasApi,
  type FiltroReservas,
} from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import {
  STATUS_RESERVA,
  rotuloStatusReserva,
  varianteStatusReserva,
} from "@/lib/rotulos";
import { dateInputParaISO, formatarData, isoParaDateInput } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Item, Reserva, ReservaPayload } from "@/types";

const TAMANHO = 20;
const TODOS = "todos";
const SEM = "0";

interface FormState {
  item_id: string;
  solicitante_id: string;
  data_inicio: string;
  data_fim: string;
  finalidade: string;
  status: string;
  observacao: string;
}

const VAZIO: FormState = {
  item_id: "",
  solicitante_id: SEM,
  data_inicio: "",
  data_fim: "",
  finalidade: "",
  status: "reservada",
  observacao: "",
};

export default function Reservas() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();
  const { servidores } = useReferencias();

  const [lista, setLista] = React.useState<Reserva[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [statusFiltro, setStatusFiltro] = React.useState<string>(TODOS);
  const [carregando, setCarregando] = React.useState(true);
  const [itens, setItens] = React.useState<Item[]>([]);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Reserva | null>(null);
  const [form, setForm] = React.useState<FormState>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [removendo, setRemovendo] = React.useState<Reserva | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);
  const [statusSalvandoId, setStatusSalvandoId] = React.useState<number | null>(
    null
  );

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const f: FiltroReservas = {
      pagina,
      tamanho: TAMANHO,
      status: statusFiltro !== TODOS ? statusFiltro : undefined,
    };
    reservasApi
      .listar(f)
      .then((r) => {
        setLista(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar reservas",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [pagina, statusFiltro, toast]);

  React.useEffect(carregar, [carregar]);

  // Só equipamentos patrimoniados (não consumíveis) e não baixados podem ser
  // reservados.
  React.useEffect(() => {
    itensApi
      .listar({ tamanho: 500, baixado: false })
      .then((r) => setItens(r.dados.filter((i) => !i.categoria?.consumivel)))
      .catch(() => setItens([]));
  }, []);

  function abrirCriacao() {
    setEditando(null);
    setForm(VAZIO);
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(r: Reserva) {
    setEditando(r);
    setForm({
      item_id: String(r.item_id),
      solicitante_id: r.solicitante_id ? String(r.solicitante_id) : SEM,
      data_inicio: isoParaDateInput(r.data_inicio),
      data_fim: isoParaDateInput(r.data_fim),
      finalidade: r.finalidade ?? "",
      status: r.status,
      observacao: r.observacao ?? "",
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!form.item_id) novos.item_id = "Selecione o equipamento.";
    if (!form.data_inicio) novos.data_inicio = "Informe o início.";
    if (!form.data_fim) novos.data_fim = "Informe o término.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});

    const payload: ReservaPayload = {
      item_id: Number(form.item_id),
      solicitante_id:
        form.solicitante_id !== SEM ? Number(form.solicitante_id) : null,
      data_inicio: dateInputParaISO(form.data_inicio),
      data_fim: dateInputParaISO(form.data_fim),
      finalidade: form.finalidade.trim(),
      status: form.status as ReservaPayload["status"],
      observacao: form.observacao.trim(),
    };

    setSalvando(true);
    try {
      if (editando) {
        await reservasApi.atualizar(editando.id, payload);
        toast({ titulo: "Reserva atualizada.", variant: "success" });
      } else {
        await reservasApi.criar(payload);
        toast({ titulo: "Reserva registrada.", variant: "success" });
      }
      setModalAberto(false);
      carregar();
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível salvar",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  async function mudarStatus(r: Reserva, novo: string) {
    setStatusSalvandoId(r.id);
    try {
      await reservasApi.definirStatus(r.id, novo);
      toast({ titulo: "Status atualizado.", variant: "success" });
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível alterar o status",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setStatusSalvandoId(null);
    }
  }

  async function confirmarRemocao() {
    if (!removendo) return;
    setProcessandoRemocao(true);
    try {
      await reservasApi.excluir(removendo.id);
      toast({ titulo: "Reserva removida.", variant: "success" });
      setRemovendo(null);
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível remover",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProcessandoRemocao(false);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Reservas"
        descricao="Agendamento de uso de equipamentos, com controle de conflito de datas."
        acao={
          <Button onClick={abrirCriacao}>
            <Plus className="h-4 w-4" /> Nova reserva
          </Button>
        }
      />

      <div className="mb-4 flex flex-wrap gap-2">
        <FiltroBotao
          ativo={statusFiltro === TODOS}
          onClick={() => {
            setStatusFiltro(TODOS);
            setPagina(1);
          }}
        >
          Todas
        </FiltroBotao>
        {STATUS_RESERVA.map((s) => (
          <FiltroBotao
            key={s.valor}
            ativo={statusFiltro === s.valor}
            onClick={() => {
              setStatusFiltro(s.valor);
              setPagina(1);
            }}
          >
            {s.rotulo}
          </FiltroBotao>
        ))}
      </div>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Nenhuma reserva para este filtro." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Equipamento</TableHead>
                  <TableHead>Solicitante</TableHead>
                  <TableHead>Período</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell className="font-medium">
                      {r.item?.descricao ?? `#${r.item_id}`}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {r.solicitante?.nome ?? "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatarData(r.data_inicio)} — {formatarData(r.data_fim)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={varianteStatusReserva(r.status)}>
                        {rotuloStatusReserva(r.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        {statusSalvandoId === r.id ? (
                          <Spinner className="h-4 w-4" />
                        ) : (
                          <>
                            {r.status === "reservada" && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => mudarStatus(r, "em_uso")}
                              >
                                Em uso
                              </Button>
                            )}
                            {(r.status === "reservada" ||
                              r.status === "em_uso") && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => mudarStatus(r, "devolvida")}
                              >
                                Devolver
                              </Button>
                            )}
                          </>
                        )}
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => abrirEdicao(r)}
                          aria-label="Editar"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        {ehAdministrador && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setRemovendo(r)}
                            aria-label="Remover"
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {total > TAMANHO && (
        <div className="mt-4 flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {total} reserva(s) · página {pagina} de {totalPaginas}
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

      <Dialog open={modalAberto} onOpenChange={setModalAberto}>
        <DialogContent className="max-h-[90vh] max-w-lg overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editando ? "Editar reserva" : "Nova reserva"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Equipamento" obrigatorio erro={erros.item_id}>
              <Select
                value={form.item_id}
                onValueChange={(v) => setForm({ ...form, item_id: v })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o equipamento" />
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

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Início" htmlFor="di" obrigatorio erro={erros.data_inicio}>
                <Input
                  id="di"
                  type="date"
                  value={form.data_inicio}
                  onChange={(e) =>
                    setForm({ ...form, data_inicio: e.target.value })
                  }
                />
              </FormField>
              <FormField label="Término" htmlFor="df" obrigatorio erro={erros.data_fim}>
                <Input
                  id="df"
                  type="date"
                  value={form.data_fim}
                  onChange={(e) =>
                    setForm({ ...form, data_fim: e.target.value })
                  }
                />
              </FormField>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Solicitante">
                <Select
                  value={form.solicitante_id}
                  onValueChange={(v) =>
                    setForm({ ...form, solicitante_id: v })
                  }
                >
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
              <FormField label="Status">
                <Select
                  value={form.status}
                  onValueChange={(v) => setForm({ ...form, status: v })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {STATUS_RESERVA.map((s) => (
                      <SelectItem key={s.valor} value={s.valor}>
                        {s.rotulo}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            </div>

            <FormField label="Finalidade" htmlFor="fin">
              <Input
                id="fin"
                value={form.finalidade}
                onChange={(e) =>
                  setForm({ ...form, finalidade: e.target.value })
                }
                placeholder="Ex.: Apresentação na Secretaria de Educação"
              />
            </FormField>

            <FormField label="Observação" htmlFor="obs">
              <Textarea
                id="obs"
                value={form.observacao}
                onChange={(e) =>
                  setForm({ ...form, observacao: e.target.value })
                }
              />
            </FormField>

            <DialogFooter className="gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setModalAberto(false)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={salvando}>
                {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        aberto={!!removendo}
        titulo="Remover reserva"
        descricao="Deseja remover esta reserva?"
        textoConfirmar="Remover"
        destrutivo
        processando={processandoRemocao}
        onConfirmar={confirmarRemocao}
        onCancelar={() => setRemovendo(null)}
      />
    </div>
  );
}

function FiltroBotao({
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
      onClick={onClick}
      className={cn(
        "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
        ativo
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border bg-card text-foreground hover:bg-accent"
      )}
    >
      {children}
    </button>
  );
}
