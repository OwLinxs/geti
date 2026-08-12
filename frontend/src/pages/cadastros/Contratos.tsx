import * as React from "react";
import { Plus, Pencil, Trash2, AlertTriangle } from "lucide-react";
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
  contratosApi,
  fornecedoresApi,
  type FiltroContratos,
} from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import {
  STATUS_CONTRATO,
  rotuloStatusContrato,
  varianteStatusContrato,
} from "@/lib/rotulos";
import {
  dateInputParaISO,
  formatarData,
  formatarMoeda,
  isoParaDateInput,
} from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Contrato, ContratoPayload, Fornecedor } from "@/types";

const TAMANHO = 20;
const TODOS = "todos";

interface FormState {
  numero: string;
  fornecedor_id: string;
  objeto: string;
  data_inicio: string;
  data_fim: string;
  valor: string;
  status: string;
  observacao: string;
}

const VAZIO: FormState = {
  numero: "",
  fornecedor_id: "",
  objeto: "",
  data_inicio: "",
  data_fim: "",
  valor: "",
  status: "vigente",
  observacao: "",
};

export default function Contratos() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();

  const [lista, setLista] = React.useState<Contrato[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [statusFiltro, setStatusFiltro] = React.useState<string>(TODOS);
  const [soVencendo, setSoVencendo] = React.useState(false);
  const [carregando, setCarregando] = React.useState(true);
  const [fornecedores, setFornecedores] = React.useState<Fornecedor[]>([]);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Contrato | null>(null);
  const [form, setForm] = React.useState<FormState>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [removendo, setRemovendo] = React.useState<Contrato | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const f: FiltroContratos = {
      pagina,
      tamanho: TAMANHO,
      status: statusFiltro !== TODOS ? statusFiltro : undefined,
      vencendo: soVencendo ? 30 : undefined,
    };
    contratosApi
      .listar(f)
      .then((r) => {
        setLista(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar contratos",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [pagina, statusFiltro, soVencendo, toast]);

  React.useEffect(carregar, [carregar]);

  React.useEffect(() => {
    fornecedoresApi
      .listar()
      .then(setFornecedores)
      .catch(() => setFornecedores([]));
  }, []);

  function abrirCriacao() {
    setEditando(null);
    setForm(VAZIO);
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(c: Contrato) {
    setEditando(c);
    setForm({
      numero: c.numero,
      fornecedor_id: String(c.fornecedor_id),
      objeto: c.objeto,
      data_inicio: isoParaDateInput(c.data_inicio),
      data_fim: isoParaDateInput(c.data_fim),
      valor: c.valor != null ? String(c.valor) : "",
      status: c.status,
      observacao: c.observacao ?? "",
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!form.numero.trim()) novos.numero = "Informe o número.";
    if (!form.objeto.trim()) novos.objeto = "Informe o objeto.";
    if (!form.fornecedor_id) novos.fornecedor_id = "Selecione o fornecedor.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});

    const payload: ContratoPayload = {
      numero: form.numero.trim(),
      fornecedor_id: Number(form.fornecedor_id),
      objeto: form.objeto.trim(),
      data_inicio: dateInputParaISO(form.data_inicio),
      data_fim: dateInputParaISO(form.data_fim),
      valor: form.valor ? Number(form.valor) : null,
      status: form.status as ContratoPayload["status"],
      observacao: form.observacao.trim(),
    };

    setSalvando(true);
    try {
      if (editando) {
        await contratosApi.atualizar(editando.id, payload);
        toast({ titulo: "Contrato atualizado.", variant: "success" });
      } else {
        await contratosApi.criar(payload);
        toast({ titulo: "Contrato criado.", variant: "success" });
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

  async function confirmarRemocao() {
    if (!removendo) return;
    setProcessandoRemocao(true);
    try {
      await contratosApi.excluir(removendo.id);
      toast({ titulo: "Contrato removido.", variant: "success" });
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

  // Destaca contratos vencendo em até 30 dias (vigentes).
  function vencendoEmBreve(c: Contrato): boolean {
    if (c.status !== "vigente" || !c.data_fim) return false;
    const fim = new Date(c.data_fim).getTime();
    const agora = Date.now();
    return fim >= agora && fim <= agora + 30 * 24 * 60 * 60 * 1000;
  }

  return (
    <div>
      <PageHeader
        titulo="Contratos"
        descricao="Contratos e serviços com fornecedores, com controle de vigência."
        acao={
          ehAdministrador && (
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> Novo contrato
            </Button>
          )
        }
      />

      <div className="mb-4 flex flex-wrap gap-2">
        <FiltroBotao
          ativo={statusFiltro === TODOS && !soVencendo}
          onClick={() => {
            setStatusFiltro(TODOS);
            setSoVencendo(false);
            setPagina(1);
          }}
        >
          Todos
        </FiltroBotao>
        {STATUS_CONTRATO.map((s) => (
          <FiltroBotao
            key={s.valor}
            ativo={statusFiltro === s.valor && !soVencendo}
            onClick={() => {
              setStatusFiltro(s.valor);
              setSoVencendo(false);
              setPagina(1);
            }}
          >
            {s.rotulo}
          </FiltroBotao>
        ))}
        <FiltroBotao
          ativo={soVencendo}
          onClick={() => {
            setSoVencendo(true);
            setStatusFiltro(TODOS);
            setPagina(1);
          }}
        >
          <AlertTriangle className="mr-1 inline h-3.5 w-3.5" />
          Vencendo (30 dias)
        </FiltroBotao>
      </div>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Nenhum contrato para este filtro." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Número</TableHead>
                  <TableHead>Fornecedor</TableHead>
                  <TableHead>Objeto</TableHead>
                  <TableHead>Vencimento</TableHead>
                  <TableHead>Valor</TableHead>
                  <TableHead>Status</TableHead>
                  {ehAdministrador && (
                    <TableHead className="w-24 text-right">Ações</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell className="font-medium">{c.numero}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {c.fornecedor?.nome ?? "—"}
                    </TableCell>
                    <TableCell className="max-w-xs truncate text-muted-foreground">
                      {c.objeto}
                    </TableCell>
                    <TableCell>
                      <span className="flex items-center gap-1">
                        {c.data_fim ? formatarData(c.data_fim) : "—"}
                        {vencendoEmBreve(c) && (
                          <AlertTriangle className="h-3.5 w-3.5 text-amber-500" />
                        )}
                      </span>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatarMoeda(c.valor)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={varianteStatusContrato(c.status)}>
                        {rotuloStatusContrato(c.status)}
                      </Badge>
                    </TableCell>
                    {ehAdministrador && (
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => abrirEdicao(c)}
                            aria-label="Editar"
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setRemovendo(c)}
                            aria-label="Remover"
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    )}
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
            {total} contrato(s) · página {pagina} de {totalPaginas}
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
              {editando ? "Editar contrato" : "Novo contrato"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Número" htmlFor="numero" obrigatorio erro={erros.numero}>
                <Input
                  id="numero"
                  value={form.numero}
                  onChange={(e) => setForm({ ...form, numero: e.target.value })}
                  placeholder="Ex.: 012/2026"
                />
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
                    {STATUS_CONTRATO.map((s) => (
                      <SelectItem key={s.valor} value={s.valor}>
                        {s.rotulo}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            </div>

            <FormField
              label="Fornecedor"
              obrigatorio
              erro={erros.fornecedor_id}
            >
              <Select
                value={form.fornecedor_id}
                onValueChange={(v) => setForm({ ...form, fornecedor_id: v })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  {fornecedores.map((f) => (
                    <SelectItem key={f.id} value={String(f.id)}>
                      {f.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Objeto" htmlFor="objeto" obrigatorio erro={erros.objeto}>
              <Textarea
                id="objeto"
                value={form.objeto}
                onChange={(e) => setForm({ ...form, objeto: e.target.value })}
                placeholder="Ex.: Manutenção de impressoras / locação de licenças…"
              />
            </FormField>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <FormField label="Início" htmlFor="di">
                <Input
                  id="di"
                  type="date"
                  value={form.data_inicio}
                  onChange={(e) =>
                    setForm({ ...form, data_inicio: e.target.value })
                  }
                />
              </FormField>
              <FormField label="Vencimento" htmlFor="df" erro={erros.data_fim}>
                <Input
                  id="df"
                  type="date"
                  value={form.data_fim}
                  onChange={(e) =>
                    setForm({ ...form, data_fim: e.target.value })
                  }
                />
              </FormField>
              <FormField label="Valor (R$)" htmlFor="valor" erro={erros.valor}>
                <Input
                  id="valor"
                  type="number"
                  min={0}
                  step="0.01"
                  value={form.valor}
                  onChange={(e) => setForm({ ...form, valor: e.target.value })}
                />
              </FormField>
            </div>

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
        titulo="Remover contrato"
        descricao={`Deseja remover o contrato "${removendo?.numero}"?`}
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
