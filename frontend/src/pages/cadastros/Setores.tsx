import * as React from "react";
import { Plus, Pencil, Trash2, ChevronRight } from "lucide-react";
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
import { setoresApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import type { Setor, SetorPayload } from "@/types";

const SEM_PAI = "0";
const VAZIO: SetorPayload = { nome: "", sigla: "", localizacao: "", pai_id: null };

// Nó da árvore com profundidade calculada, para renderização indentada.
interface NoArvore {
  setor: Setor;
  nivel: number;
}

// achatarArvore ordena a lista plana como uma árvore (pais antes dos filhos),
// atribuindo o nível de profundidade a cada nó.
function achatarArvore(lista: Setor[]): NoArvore[] {
  const filhosDe = new Map<number | null, Setor[]>();
  for (const s of lista) {
    const chave = s.pai_id ?? null;
    if (!filhosDe.has(chave)) filhosDe.set(chave, []);
    filhosDe.get(chave)!.push(s);
  }
  for (const arr of filhosDe.values()) {
    arr.sort((a, b) => a.nome.localeCompare(b.nome, "pt-BR"));
  }
  const saida: NoArvore[] = [];
  const visitar = (paiId: number | null, nivel: number) => {
    for (const s of filhosDe.get(paiId) ?? []) {
      saida.push({ setor: s, nivel });
      visitar(s.id, nivel + 1);
    }
  };
  visitar(null, 0);
  return saida;
}

// idsDescendentes devolve o id do nó + todos os descendentes (para impedir
// escolher a si mesmo ou uma unidade filha como pai — evitaria ciclo).
function idsDescendentes(lista: Setor[], raizId: number): Set<number> {
  const filhosDe = new Map<number, Setor[]>();
  for (const s of lista) {
    if (s.pai_id != null) {
      if (!filhosDe.has(s.pai_id)) filhosDe.set(s.pai_id, []);
      filhosDe.get(s.pai_id)!.push(s);
    }
  }
  const set = new Set<number>([raizId]);
  const pilha = [raizId];
  while (pilha.length) {
    const atual = pilha.pop()!;
    for (const f of filhosDe.get(atual) ?? []) {
      if (!set.has(f.id)) {
        set.add(f.id);
        pilha.push(f.id);
      }
    }
  }
  return set;
}

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
          titulo: "Erro ao carregar departamentos",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  const arvore = React.useMemo(() => achatarArvore(lista), [lista]);

  // Opções válidas para "unidade superior": todas, menos a própria unidade e
  // suas descendentes (quando editando), apresentadas com indentação.
  const opcoesPai = React.useMemo(() => {
    const bloqueados = editando
      ? idsDescendentes(lista, editando.id)
      : new Set<number>();
    return achatarArvore(lista).filter((n) => !bloqueados.has(n.setor.id));
  }, [lista, editando]);

  function abrirCriacao(paiSugerido?: Setor) {
    setEditando(null);
    setForm({ ...VAZIO, pai_id: paiSugerido?.id ?? null });
    setErros({});
    setModalAberto(true);
  }

  function abrirEdicao(s: Setor) {
    setEditando(s);
    setForm({
      nome: s.nome,
      sigla: s.sigla,
      localizacao: s.localizacao,
      pai_id: s.pai_id ?? null,
    });
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setErros({});
    if (!form.nome.trim()) {
      setErros({ nome: "Informe o nome do departamento." });
      return;
    }
    setSalvando(true);
    try {
      if (editando) {
        await setoresApi.atualizar(editando.id, form);
        toast({ titulo: "Departamento atualizado.", variant: "success" });
      } else {
        await setoresApi.criar(form);
        toast({ titulo: "Departamento criado.", variant: "success" });
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
      toast({ titulo: "Departamento removido.", variant: "success" });
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
        titulo="Departamentos"
        descricao="Estrutura organizacional em árvore: Secretarias no topo e departamentos/divisões abaixo."
        acao={
          ehAdministrador && (
            <Button onClick={() => abrirCriacao()}>
              <Plus className="h-4 w-4" /> Nova unidade
            </Button>
          )
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio descricao="Cadastre a primeira unidade (ex.: uma Secretaria)." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Unidade</TableHead>
                  <TableHead>Sigla</TableHead>
                  <TableHead>Localização</TableHead>
                  {ehAdministrador && (
                    <TableHead className="w-32 text-right">Ações</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {arvore.map(({ setor: s, nivel }) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-medium">
                      <span
                        className="flex items-center"
                        style={{ paddingLeft: `${nivel * 20}px` }}
                      >
                        {nivel > 0 && (
                          <ChevronRight className="mr-1 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                        )}
                        {s.nome}
                      </span>
                    </TableCell>
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
                            onClick={() => abrirCriacao(s)}
                            aria-label="Adicionar unidade filha"
                            title="Adicionar unidade subordinada"
                          >
                            <Plus className="h-4 w-4" />
                          </Button>
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
            <DialogTitle>
              {editando ? "Editar unidade" : "Nova unidade"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Unidade superior" erro={erros.pai_id}>
              <Select
                value={form.pai_id ? String(form.pai_id) : SEM_PAI}
                onValueChange={(v) =>
                  setForm({ ...form, pai_id: v === SEM_PAI ? null : Number(v) })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM_PAI}>
                    Nenhuma (unidade de topo / Secretaria)
                  </SelectItem>
                  {opcoesPai.map(({ setor: s, nivel }) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {" ".repeat(nivel * 3)}
                      {nivel > 0 ? "└ " : ""}
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Nome" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={form.nome}
                onChange={(e) => setForm({ ...form, nome: e.target.value })}
                placeholder="Ex.: Departamento de T.I."
              />
            </FormField>
            <FormField label="Sigla" htmlFor="sigla" erro={erros.sigla}>
              <Input
                id="sigla"
                value={form.sigla}
                onChange={(e) => setForm({ ...form, sigla: e.target.value })}
                placeholder="Ex.: DTI"
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
        titulo="Remover unidade"
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
