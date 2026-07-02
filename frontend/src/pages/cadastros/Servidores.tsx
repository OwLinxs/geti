import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
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
import { servidoresApi, setoresApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import type { Servidor, ServidorPayload, Setor } from "@/types";

const SEM_SETOR = "0";

export default function Servidores() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();
  const [lista, setLista] = React.useState<Servidor[]>([]);
  const [setores, setSetores] = React.useState<Setor[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Servidor | null>(null);
  const [form, setForm] = React.useState<ServidorPayload>({
    nome: "",
    matricula: "",
    setor_id: null,
    ativo: true,
  });
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  const [removendo, setRemovendo] = React.useState<Servidor | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    Promise.all([servidoresApi.listar(), setoresApi.listar()])
      .then(([servs, sets]) => {
        setLista(servs);
        setSetores(sets);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar servidores",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  function abrirCriacao() {
    setEditando(null);
    setForm({ nome: "", matricula: "", setor_id: null, ativo: true });
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(s: Servidor) {
    setEditando(s);
    setForm({
      nome: s.nome,
      matricula: s.matricula,
      setor_id: s.setor_id ?? null,
      ativo: s.ativo,
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novosErros: Record<string, string> = {};
    if (!form.nome.trim()) novosErros.nome = "Informe o nome do servidor.";
    if (!form.matricula.trim())
      novosErros.matricula = "Informe a matrícula.";
    if (Object.keys(novosErros).length) {
      setErros(novosErros);
      return;
    }
    setErros({});
    setSalvando(true);
    try {
      if (editando) {
        await servidoresApi.atualizar(editando.id, form);
        toast({ titulo: "Servidor atualizado.", variant: "success" });
      } else {
        await servidoresApi.criar(form);
        toast({ titulo: "Servidor cadastrado.", variant: "success" });
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
      await servidoresApi.remover(removendo.id);
      toast({ titulo: "Servidor removido.", variant: "success" });
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
        titulo="Servidores"
        descricao="Responsáveis por equipamentos. Conforme a LGPD, coletamos apenas nome e matrícula."
        acao={
          <Button onClick={abrirCriacao}>
            <Plus className="h-4 w-4" /> Novo servidor
          </Button>
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre o primeiro servidor." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Matrícula</TableHead>
                  <TableHead>Lotação</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead className="w-24 text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-medium">{s.nome}</TableCell>
                    <TableCell>{s.matricula}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {s.setor?.nome ?? "—"}
                    </TableCell>
                    <TableCell>
                      {s.ativo ? (
                        <Badge variant="success">Ativo</Badge>
                      ) : (
                        <Badge variant="muted">Inativo</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => abrirEdicao(s)}
                          aria-label="Editar"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        {ehAdministrador && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setRemovendo(s)}
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

      <Dialog open={modalAberto} onOpenChange={setModalAberto}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editando ? "Editar servidor" : "Novo servidor"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Nome" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={form.nome}
                onChange={(e) => setForm({ ...form, nome: e.target.value })}
              />
            </FormField>
            <FormField
              label="Matrícula"
              htmlFor="matricula"
              obrigatorio
              erro={erros.matricula}
            >
              <Input
                id="matricula"
                value={form.matricula}
                onChange={(e) =>
                  setForm({ ...form, matricula: e.target.value })
                }
              />
            </FormField>
            <FormField label="Lotação (setor)" erro={erros.setor_id}>
              <Select
                value={form.setor_id ? String(form.setor_id) : SEM_SETOR}
                onValueChange={(v) =>
                  setForm({
                    ...form,
                    setor_id: v === SEM_SETOR ? null : Number(v),
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione um setor" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM_SETOR}>Sem lotação</SelectItem>
                  {setores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <div className="flex items-center justify-between rounded-md border border-border p-3">
              <p className="text-sm font-medium text-foreground">
                Servidor ativo
              </p>
              <Switch
                checked={form.ativo ?? true}
                onCheckedChange={(v) => setForm({ ...form, ativo: v })}
              />
            </div>
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
        titulo="Remover servidor"
        descricao={`Deseja remover "${removendo?.nome}"?`}
        textoConfirmar="Remover"
        destrutivo
        processando={processandoRemocao}
        onConfirmar={confirmarRemocao}
        onCancelar={() => setRemovendo(null)}
      />
    </div>
  );
}
