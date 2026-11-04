import * as React from "react";
import { Plus, Pencil, Trash2, GripVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
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
import { Spinner } from "@/components/ui/spinner";
import { categoriasApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type {
  Categoria,
  CampoCategoria,
  CampoCategoriaPayload,
  TipoCampo,
} from "@/types";

const TIPOS: { valor: TipoCampo; rotulo: string }[] = [
  { valor: "texto", rotulo: "Texto" },
  { valor: "numero", rotulo: "Número" },
  { valor: "lista", rotulo: "Lista (seleção)" },
  { valor: "data", rotulo: "Data" },
  { valor: "booleano", rotulo: "Sim / Não" },
];

export function rotuloTipoCampo(t: TipoCampo): string {
  return TIPOS.find((x) => x.valor === t)?.rotulo ?? t;
}

const VAZIO: CampoCategoriaPayload = {
  nome: "",
  tipo: "texto",
  opcoes: "",
  obrigatorio: false,
  ordem: 0,
};

interface Props {
  aberto: boolean;
  categoria: Categoria | null;
  onFechar: () => void;
}

export function GerenciarCamposDialog({ aberto, categoria, onFechar }: Props) {
  const { toast } = useToast();
  const [campos, setCampos] = React.useState<CampoCategoria[]>([]);
  const [carregando, setCarregando] = React.useState(false);

  const [form, setForm] = React.useState<CampoCategoriaPayload>(VAZIO);
  const [editando, setEditando] = React.useState<CampoCategoria | null>(null);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [removendo, setRemovendo] = React.useState<CampoCategoria | null>(null);
  const [processandoRemocao, setProcessandoRemocao] = React.useState(false);

  const carregar = React.useCallback(() => {
    if (!categoria) return;
    setCarregando(true);
    categoriasApi
      .listarCampos(categoria.id)
      .then(setCampos)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar os campos",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [categoria, toast]);

  React.useEffect(() => {
    if (aberto) {
      setForm(VAZIO);
      setEditando(null);
      setErros({});
      carregar();
    }
  }, [aberto, carregar]);

  function editar(c: CampoCategoria) {
    setEditando(c);
    setForm({
      nome: c.nome,
      tipo: c.tipo,
      opcoes: c.opcoes ?? "",
      obrigatorio: c.obrigatorio,
      ordem: c.ordem,
    });
    setErros({});
  }

  function cancelarEdicao() {
    setEditando(null);
    setForm(VAZIO);
    setErros({});
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    if (!categoria) return;
    setErros({});
    if (!form.nome.trim()) {
      setErros({ nome: "Informe o nome do campo." });
      return;
    }
    if (form.tipo === "lista" && !(form.opcoes ?? "").trim()) {
      setErros({ opcoes: "Informe ao menos uma opção (uma por linha)." });
      return;
    }
    setSalvando(true);
    try {
      // Novo campo entra ao final por padrão.
      const payload: CampoCategoriaPayload = {
        ...form,
        ordem: editando ? form.ordem : campos.length,
      };
      if (editando) {
        await categoriasApi.atualizarCampo(categoria.id, editando.id, payload);
        toast({ titulo: "Campo atualizado.", variant: "success" });
      } else {
        await categoriasApi.criarCampo(categoria.id, payload);
        toast({ titulo: "Campo adicionado.", variant: "success" });
      }
      cancelarEdicao();
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível salvar o campo",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  async function confirmarRemocao() {
    if (!categoria || !removendo) return;
    setProcessandoRemocao(true);
    try {
      await categoriasApi.removerCampo(categoria.id, removendo.id);
      toast({ titulo: "Campo removido.", variant: "success" });
      setRemovendo(null);
      if (editando?.id === removendo.id) cancelarEdicao();
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível remover o campo",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProcessandoRemocao(false);
    }
  }

  return (
    <>
      <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Campos de "{categoria?.nome}"</DialogTitle>
          </DialogHeader>

          <p className="text-sm text-muted-foreground">
            Campos extras preenchidos no cadastro de itens desta categoria (ex.:
            Processador, Memória). Aparecem no formulário e no detalhe do item.
          </p>

          {/* Lista de campos existentes. */}
          <div className="rounded-md border border-border">
            {carregando ? (
              <div className="flex justify-center py-6">
                <Spinner className="h-5 w-5" />
              </div>
            ) : campos.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">
                Nenhum campo cadastrado ainda.
              </p>
            ) : (
              <ul className="divide-y divide-border">
                {campos.map((c) => (
                  <li
                    key={c.id}
                    className="flex items-center gap-2 px-3 py-2 text-sm"
                  >
                    <GripVertical className="h-4 w-4 shrink-0 text-muted-foreground/50" />
                    <div className="min-w-0 flex-1">
                      <span className="font-medium">{c.nome}</span>
                      {c.obrigatorio && (
                        <span className="ml-1 text-destructive">*</span>
                      )}
                      <Badge variant="muted" className="ml-2">
                        {rotuloTipoCampo(c.tipo)}
                      </Badge>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => editar(c)}
                      aria-label="Editar campo"
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => setRemovendo(c)}
                      aria-label="Remover campo"
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </div>

          {/* Formulário de adicionar/editar campo. */}
          <form
            onSubmit={salvar}
            className="space-y-3 rounded-md border border-border bg-muted/30 p-3"
            noValidate
          >
            <p className="text-sm font-medium">
              {editando ? "Editar campo" : "Adicionar campo"}
            </p>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <FormField label="Nome" htmlFor="campo-nome" obrigatorio erro={erros.nome}>
                <Input
                  id="campo-nome"
                  value={form.nome}
                  onChange={(e) => setForm({ ...form, nome: e.target.value })}
                  placeholder="Ex.: Processador, Memória..."
                />
              </FormField>
              <FormField label="Tipo" htmlFor="campo-tipo">
                <Select
                  value={form.tipo}
                  onValueChange={(v) =>
                    setForm({ ...form, tipo: v as TipoCampo })
                  }
                >
                  <SelectTrigger id="campo-tipo">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {TIPOS.map((t) => (
                      <SelectItem key={t.valor} value={t.valor}>
                        {t.rotulo}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            </div>

            {form.tipo === "lista" && (
              <FormField
                label="Opções (uma por linha)"
                htmlFor="campo-opcoes"
                erro={erros.opcoes}
              >
                <Textarea
                  id="campo-opcoes"
                  value={form.opcoes ?? ""}
                  onChange={(e) => setForm({ ...form, opcoes: e.target.value })}
                  placeholder={"SSD\nHD\nNVMe"}
                  rows={3}
                />
              </FormField>
            )}

            <div className="flex items-center justify-between rounded-md border border-border bg-card p-3">
              <div>
                <p className="text-sm font-medium">Obrigatório</p>
                <p className="text-xs text-muted-foreground">
                  Exige preenchimento no cadastro do item.
                </p>
              </div>
              <Switch
                checked={form.obrigatorio}
                onCheckedChange={(v) => setForm({ ...form, obrigatorio: v })}
              />
            </div>

            <div className="flex justify-end gap-2">
              {editando && (
                <Button type="button" variant="outline" onClick={cancelarEdicao}>
                  Cancelar
                </Button>
              )}
              <Button type="submit" disabled={salvando}>
                {salvando ? (
                  <Spinner className="h-4 w-4" />
                ) : editando ? (
                  "Salvar campo"
                ) : (
                  <>
                    <Plus className="h-4 w-4" /> Adicionar
                  </>
                )}
              </Button>
            </div>
          </form>

          <div className="flex justify-end">
            <Button variant="outline" onClick={onFechar}>
              Fechar
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        aberto={!!removendo}
        titulo="Remover campo"
        descricao={`Remover o campo "${removendo?.nome}"? Os valores desse campo já preenchidos nos itens também serão apagados.`}
        textoConfirmar="Remover"
        destrutivo
        processando={processandoRemocao}
        onConfirmar={confirmarRemocao}
        onCancelar={() => setRemovendo(null)}
      />
    </>
  );
}
