import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
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
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { categoriasApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import type { Categoria, CategoriaPayload } from "@/types";

const VAZIO: CategoriaPayload = { nome: "", descricao: "", consumivel: false };

export default function Categorias() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();
  const [lista, setLista] = React.useState<Categoria[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Categoria | null>(null);
  const [form, setForm] = React.useState<CategoriaPayload>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  const [removendo, setRemovendo] = React.useState<Categoria | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    categoriasApi
      .listar()
      .then(setLista)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar categorias",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  function abrirCriacao() {
    setEditando(null);
    setForm(VAZIO);
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(c: Categoria) {
    setEditando(c);
    setForm({
      nome: c.nome,
      descricao: c.descricao,
      consumivel: c.consumivel,
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setErros({});
    if (!form.nome.trim()) {
      setErros({ nome: "Informe o nome da categoria." });
      return;
    }
    setSalvando(true);
    try {
      if (editando) {
        await categoriasApi.atualizar(editando.id, form);
        toast({ titulo: "Categoria atualizada.", variant: "success" });
      } else {
        await categoriasApi.criar(form);
        toast({ titulo: "Categoria criada.", variant: "success" });
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
      await categoriasApi.remover(removendo.id);
      toast({ titulo: "Categoria removida.", variant: "success" });
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
        titulo="Categorias"
        descricao="Classificação dos itens (patrimoniados e materiais de consumo)."
        acao={
          ehAdministrador && (
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> Nova categoria
            </Button>
          )
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre a primeira categoria." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Descrição</TableHead>
                  <TableHead>Tipo</TableHead>
                  {ehAdministrador && (
                    <TableHead className="w-24 text-right">Ações</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell className="font-medium">{c.nome}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {c.descricao || "—"}
                    </TableCell>
                    <TableCell>
                      {c.consumivel ? (
                        <Badge variant="secondary">Consumo</Badge>
                      ) : (
                        <Badge variant="muted">Patrimoniado</Badge>
                      )}
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

      <Dialog open={modalAberto} onOpenChange={setModalAberto}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editando ? "Editar categoria" : "Nova categoria"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Nome" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={form.nome}
                onChange={(e) => setForm({ ...form, nome: e.target.value })}
                placeholder="Ex.: Computador, Monitor, Toner..."
              />
            </FormField>
            <FormField label="Descrição" htmlFor="descricao" erro={erros.descricao}>
              <Textarea
                id="descricao"
                value={form.descricao}
                onChange={(e) =>
                  setForm({ ...form, descricao: e.target.value })
                }
              />
            </FormField>
            <div className="flex items-center justify-between rounded-md border border-border p-3">
              <div>
                <p className="text-sm font-medium text-foreground">
                  Material de consumo
                </p>
                <p className="text-xs text-muted-foreground">
                  Controlado por quantidade e estoque mínimo.
                </p>
              </div>
              <Switch
                checked={form.consumivel}
                onCheckedChange={(v) => setForm({ ...form, consumivel: v })}
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
        titulo="Remover categoria"
        descricao={`Deseja remover "${removendo?.nome}"? Esta ação não poderá ser desfeita.`}
        textoConfirmar="Remover"
        destrutivo
        processando={processandoRemocao}
        onConfirmar={confirmarRemocao}
        onCancelar={() => setRemovendo(null)}
      />
    </div>
  );
}
