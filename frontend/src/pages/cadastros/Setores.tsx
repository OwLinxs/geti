import * as React from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { setoresApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import type { Setor, SetorPayload } from "@/types";

const VAZIO: SetorPayload = { nome: "", sigla: "", localizacao: "" };

export default function Setores() {
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();
  const [lista, setLista] = React.useState<Setor[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  const [modalAberto, setModalAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Setor | null>(null);
  const [form, setForm] = React.useState<SetorPayload>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  const [removendo, setRemovendo] = React.useState<Setor | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    setoresApi
      .listar()
      .then(setLista)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar setores",
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

  function abrirEdicao(s: Setor) {
    setEditando(s);
    setForm({ nome: s.nome, sigla: s.sigla, localizacao: s.localizacao });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setErros({});
    if (!form.nome.trim()) {
      setErros({ nome: "Informe o nome do setor." });
      return;
    }
    setSalvando(true);
    try {
      if (editando) {
        await setoresApi.atualizar(editando.id, form);
        toast({ titulo: "Setor atualizado.", variant: "success" });
      } else {
        await setoresApi.criar(form);
        toast({ titulo: "Setor criado.", variant: "success" });
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
      await setoresApi.remover(removendo.id);
      toast({ titulo: "Setor removido.", variant: "success" });
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
        titulo="Setores"
        descricao="Localizações organizacionais (departamentos, secretarias, salas)."
        acao={
          ehAdministrador && (
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> Novo setor
            </Button>
          )
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre o primeiro setor." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Sigla</TableHead>
                  <TableHead>Localização</TableHead>
                  {ehAdministrador && (
                    <TableHead className="w-24 text-right">Ações</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-medium">{s.nome}</TableCell>
                    <TableCell>{s.sigla || "—"}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {s.localizacao || "—"}
                    </TableCell>
                    {ehAdministrador && (
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
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setRemovendo(s)}
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
            <DialogTitle>{editando ? "Editar setor" : "Novo setor"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Nome" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={form.nome}
                onChange={(e) => setForm({ ...form, nome: e.target.value })}
                placeholder="Ex.: Secretaria de Educação"
              />
            </FormField>
            <FormField label="Sigla" htmlFor="sigla" erro={erros.sigla}>
              <Input
                id="sigla"
                value={form.sigla}
                onChange={(e) => setForm({ ...form, sigla: e.target.value })}
                placeholder="Ex.: SEMED"
              />
            </FormField>
            <FormField
              label="Localização"
              htmlFor="localizacao"
              erro={erros.localizacao}
            >
              <Input
                id="localizacao"
                value={form.localizacao}
                onChange={(e) =>
                  setForm({ ...form, localizacao: e.target.value })
                }
                placeholder="Prédio / sala (opcional)"
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
        titulo="Remover setor"
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
