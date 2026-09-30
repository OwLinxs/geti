import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
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
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { categoriasChamadoApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { CategoriaChamado, CategoriaChamadoPayload } from "@/types";

const VAZIO: CategoriaChamadoPayload = {
  nome: "",
  descricao: "",
  ordem: 0,
  ativo: true,
};

export default function CategoriasChamado({
  embutido = false,
}: {
  embutido?: boolean;
}) {
  const { toast } = useToast();
  const [lista, setLista] = React.useState<CategoriaChamado[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<CategoriaChamado | null>(null);
  const [form, setForm] = React.useState<CategoriaChamadoPayload>(VAZIO);
  const [checklistTexto, setChecklistTexto] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [removendo, setRemovendo] = React.useState<CategoriaChamado | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    categoriasChamadoApi
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
    setForm({ ...VAZIO, ordem: lista.length + 1 });
    setChecklistTexto("");
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(c: CategoriaChamado) {
    setEditando(c);
    setForm({
      nome: c.nome,
      descricao: c.descricao ?? "",
      ordem: c.ordem,
      ativo: c.ativo,
    });
    setChecklistTexto((c.checklist_padrao ?? []).join("\n"));
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    if (!form.nome.trim()) {
      setErros({ nome: "Informe o nome." });
      return;
    }
    setErros({});
    setSalvando(true);
    const payload: CategoriaChamadoPayload = {
      ...form,
      checklist_padrao: checklistTexto
        .split("\n")
        .map((l) => l.trim())
        .filter(Boolean),
    };
    try {
      if (editando) {
        await categoriasChamadoApi.atualizar(editando.id, payload);
        toast({ titulo: "Categoria atualizada.", variant: "success" });
      } else {
        await categoriasChamadoApi.criar(payload);
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
      await categoriasChamadoApi.excluir(removendo.id);
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
      {embutido ? (
        <div className="mb-4 flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Tipos de chamado usados na abertura e nos relatórios.
          </p>
          <Button onClick={abrirCriacao}>
            <Plus className="h-4 w-4" /> Nova categoria
          </Button>
        </div>
      ) : (
        <PageHeader
          titulo="Categorias de Chamado"
          descricao="Tipos de chamado (rede, e-mail, impressora…) usados na abertura e nos relatórios."
          acao={
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> Nova categoria
            </Button>
          }
        />
      )}

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre a primeira categoria de chamado." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">Ordem</TableHead>
                  <TableHead>Nome</TableHead>
                  <TableHead>Descrição</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead className="w-24 text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell className="text-muted-foreground">
                      {c.ordem}
                    </TableCell>
                    <TableCell className="font-medium">{c.nome}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {c.descricao || "—"}
                    </TableCell>
                    <TableCell>
                      {c.ativo ? (
                        <Badge variant="success">Ativa</Badge>
                      ) : (
                        <Badge variant="muted">Inativa</Badge>
                      )}
                    </TableCell>
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
                placeholder="Ex.: Rede"
              />
            </FormField>
            <FormField label="Descrição" htmlFor="descricao">
              <Input
                id="descricao"
                value={form.descricao}
                onChange={(e) =>
                  setForm({ ...form, descricao: e.target.value })
                }
              />
            </FormField>
            <FormField label="Ordem de exibição" htmlFor="ordem">
              <Input
                id="ordem"
                type="number"
                value={String(form.ordem ?? 0)}
                onChange={(e) =>
                  setForm({ ...form, ordem: Number(e.target.value) })
                }
              />
            </FormField>
            <FormField
              label="Checklist padrão (uma etapa por linha)"
              htmlFor="checklist"
            >
              <Textarea
                id="checklist"
                rows={5}
                value={checklistTexto}
                onChange={(e) => setChecklistTexto(e.target.value)}
                placeholder={"Verificar cabo de rede\nTestar ping no gateway\nReiniciar o roteador"}
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Aplicado automaticamente ao classificar um chamado nesta
                categoria (se o checklist ainda estiver em branco).
              </p>
            </FormField>
            <div className="flex items-center justify-between rounded-md border border-border px-3 py-2">
              <div>
                <p className="text-sm font-medium">Ativa</p>
                <p className="text-xs text-muted-foreground">
                  Categorias inativas não aparecem na abertura de chamados.
                </p>
              </div>
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
        titulo="Remover categoria"
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
