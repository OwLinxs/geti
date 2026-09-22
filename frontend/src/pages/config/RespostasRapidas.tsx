import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
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
import { respostasRapidasApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { RespostaRapida, RespostaRapidaPayload } from "@/types";

const VAZIO: RespostaRapidaPayload = {
  titulo: "",
  conteudo: "",
  ordem: 0,
  ativo: true,
};

export default function RespostasRapidas() {
  const { toast } = useToast();
  const [lista, setLista] = React.useState<RespostaRapida[]>([]);
  const [carregando, setCarregando] = React.useState(true);
  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<RespostaRapida | null>(null);
  const [form, setForm] = React.useState<RespostaRapidaPayload>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [removendo, setRemovendo] = React.useState<RespostaRapida | null>(null);
  const [proc, setProc] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    respostasRapidasApi
      .listar()
      .then(setLista)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar",
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
    setErros({});
    setModalAberto(true);
  }
  function abrirEdicao(r: RespostaRapida) {
    setEditando(r);
    setForm({
      titulo: r.titulo,
      conteudo: r.conteudo,
      ordem: r.ordem,
      ativo: r.ativo,
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!form.titulo.trim()) novos.titulo = "Informe o título.";
    if (!form.conteudo.trim()) novos.conteudo = "Informe o texto.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});
    setSalvando(true);
    try {
      if (editando) await respostasRapidasApi.atualizar(editando.id, form);
      else await respostasRapidasApi.criar(form);
      toast({ titulo: "Salvo.", variant: "success" });
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
    setProc(true);
    try {
      await respostasRapidasApi.excluir(removendo.id);
      toast({ titulo: "Removida.", variant: "success" });
      setRemovendo(null);
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível remover",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProc(false);
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          Textos-modelo que a equipe insere no chat com um clique.
        </p>
        <Button onClick={abrirCriacao}>
          <Plus className="h-4 w-4" /> Nova resposta
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre a primeira resposta rápida." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">Ordem</TableHead>
                  <TableHead>Título</TableHead>
                  <TableHead>Texto</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead className="w-24 text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell className="text-muted-foreground">
                      {r.ordem}
                    </TableCell>
                    <TableCell className="font-medium">{r.titulo}</TableCell>
                    <TableCell className="max-w-xs truncate text-muted-foreground">
                      {r.conteudo}
                    </TableCell>
                    <TableCell>
                      {r.ativo ? (
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
                          onClick={() => abrirEdicao(r)}
                          aria-label="Editar"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setRemovendo(r)}
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
              {editando ? "Editar resposta" : "Nova resposta"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Título" htmlFor="t" obrigatorio erro={erros.titulo}>
              <Input
                id="t"
                value={form.titulo}
                onChange={(e) => setForm({ ...form, titulo: e.target.value })}
                placeholder="Ex.: Recebido"
              />
            </FormField>
            <FormField
              label="Texto da resposta"
              htmlFor="c"
              obrigatorio
              erro={erros.conteudo}
            >
              <Textarea
                id="c"
                rows={4}
                value={form.conteudo}
                onChange={(e) => setForm({ ...form, conteudo: e.target.value })}
                placeholder="Olá! Recebemos seu chamado…"
              />
            </FormField>
            <div className="grid grid-cols-2 gap-4">
              <FormField label="Ordem" htmlFor="o">
                <Input
                  id="o"
                  type="number"
                  value={String(form.ordem ?? 0)}
                  onChange={(e) =>
                    setForm({ ...form, ordem: Number(e.target.value) })
                  }
                />
              </FormField>
              <div className="flex items-end justify-between rounded-md border border-border px-3 py-2">
                <span className="text-sm font-medium">Ativa</span>
                <Switch
                  checked={form.ativo ?? true}
                  onCheckedChange={(v) => setForm({ ...form, ativo: v })}
                />
              </div>
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
        titulo="Remover resposta"
        descricao={`Remover "${removendo?.titulo}"?`}
        textoConfirmar="Remover"
        destrutivo
        processando={proc}
        onConfirmar={confirmarRemocao}
        onCancelar={() => setRemovendo(null)}
      />
    </div>
  );
}
