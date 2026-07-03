import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { Spinner } from "@/components/ui/spinner";
import { itensApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { ESTADOS_CONSERVACAO } from "@/lib/rotulos";
import { dateInputParaISO, isoParaDateInput } from "@/lib/format";
import { ordenarComoArvore, prefixoIndentacao } from "@/lib/setores";
import type {
  Categoria,
  EstadoConservacao,
  Item,
  ItemPayload,
  Servidor,
  Setor,
} from "@/types";

const SEM = "0";

interface ItemFormProps {
  aberto: boolean;
  item: Item | null; // null => criação
  categorias: Categoria[];
  setores: Setor[];
  servidores: Servidor[];
  onFechar: () => void;
  onSalvo: () => void;
}

interface FormState {
  descricao: string;
  categoria_id: number;
  numero_patrimonio: string;
  numero_serie: string;
  marca: string;
  modelo: string;
  estado_conservacao: EstadoConservacao;
  quantidade: string;
  estoque_minimo: string;
  setor_id: number | null;
  responsavel_id: number | null;
  data_aquisicao: string;
  valor: string;
}

function estadoInicial(item: Item | null): FormState {
  return {
    descricao: item?.descricao ?? "",
    categoria_id: item?.categoria_id ?? 0,
    numero_patrimonio: item?.numero_patrimonio ?? "",
    numero_serie: item?.numero_serie ?? "",
    marca: item?.marca ?? "",
    modelo: item?.modelo ?? "",
    estado_conservacao: item?.estado_conservacao ?? "bom",
    quantidade: String(item?.quantidade ?? 1),
    estoque_minimo: String(item?.estoque_minimo ?? 0),
    setor_id: item?.setor_id ?? null,
    responsavel_id: item?.responsavel_id ?? null,
    data_aquisicao: isoParaDateInput(item?.data_aquisicao),
    valor: item?.valor != null ? String(item.valor) : "",
  };
}

export function ItemForm({
  aberto,
  item,
  categorias,
  setores,
  servidores,
  onFechar,
  onSalvo,
}: ItemFormProps) {
  const { toast } = useToast();
  const [form, setForm] = React.useState<FormState>(estadoInicial(item));
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  // Setores em ordem hierárquica (Secretaria › Departamento › Unidade) para
  // exibição indentada no seletor.
  const setoresArvore = React.useMemo(() => ordenarComoArvore(setores), [setores]);

  React.useEffect(() => {
    if (aberto) {
      setForm(estadoInicial(item));
      setErros({});
    }
  }, [aberto, item]);

  const categoriaSelecionada = categorias.find(
    (c) => c.id === form.categoria_id
  );
  const ehConsumivel = categoriaSelecionada?.consumivel ?? false;

  function set<K extends keyof FormState>(k: K, v: FormState[K]) {
    setForm((f) => ({ ...f, [k]: v }));
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novosErros: Record<string, string> = {};
    if (!form.descricao.trim())
      novosErros.descricao = "Informe a descrição do item.";
    if (!form.categoria_id) novosErros.categoria_id = "Selecione a categoria.";
    const qtd = Number(form.quantidade);
    if (Number.isNaN(qtd) || qtd < 0)
      novosErros.quantidade = "Quantidade inválida.";
    const min = Number(form.estoque_minimo);
    if (Number.isNaN(min) || min < 0)
      novosErros.estoque_minimo = "Estoque mínimo inválido.";
    if (form.valor && Number.isNaN(Number(form.valor)))
      novosErros.valor = "Valor inválido.";
    if (Object.keys(novosErros).length) {
      setErros(novosErros);
      return;
    }
    setErros({});

    const payload: ItemPayload = {
      descricao: form.descricao.trim(),
      categoria_id: form.categoria_id,
      numero_patrimonio: form.numero_patrimonio.trim() || null,
      numero_serie: form.numero_serie.trim() || null,
      marca: form.marca.trim(),
      modelo: form.modelo.trim(),
      estado_conservacao: form.estado_conservacao,
      quantidade: qtd,
      estoque_minimo: min,
      setor_id: form.setor_id,
      responsavel_id: form.responsavel_id,
      data_aquisicao: dateInputParaISO(form.data_aquisicao),
      valor: form.valor ? Number(form.valor) : null,
    };

    setSalvando(true);
    try {
      if (item) {
        await itensApi.atualizar(item.id, payload);
        toast({ titulo: "Item atualizado.", variant: "success" });
      } else {
        await itensApi.criar(payload);
        toast({ titulo: "Item cadastrado.", variant: "success" });
      }
      onSalvo();
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível salvar o item",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{item ? "Editar item" : "Novo item"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField
            label="Descrição"
            htmlFor="descricao"
            obrigatorio
            erro={erros.descricao}
          >
            <Input
              id="descricao"
              value={form.descricao}
              onChange={(e) => set("descricao", e.target.value)}
              placeholder="Ex.: Notebook Dell Latitude 5440"
            />
          </FormField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Categoria" obrigatorio erro={erros.categoria_id}>
              <Select
                value={form.categoria_id ? String(form.categoria_id) : ""}
                onValueChange={(v) => set("categoria_id", Number(v))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  {categorias.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.nome}
                      {c.consumivel ? " (consumo)" : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Estado de conservação" obrigatorio>
              <Select
                value={form.estado_conservacao}
                onValueChange={(v) =>
                  set("estado_conservacao", v as EstadoConservacao)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ESTADOS_CONSERVACAO.map((e) => (
                    <SelectItem key={e.valor} value={e.valor}>
                      {e.rotulo}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          </div>

          {!ehConsumivel && (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField
                label="Nº de patrimônio"
                htmlFor="patrimonio"
                erro={erros.numero_patrimonio}
              >
                <Input
                  id="patrimonio"
                  value={form.numero_patrimonio}
                  onChange={(e) => set("numero_patrimonio", e.target.value)}
                  placeholder="Ex.: 2024-00123"
                />
              </FormField>
              <FormField label="Nº de série" htmlFor="serie">
                <Input
                  id="serie"
                  value={form.numero_serie}
                  onChange={(e) => set("numero_serie", e.target.value)}
                />
              </FormField>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Marca" htmlFor="marca">
              <Input
                id="marca"
                value={form.marca}
                onChange={(e) => set("marca", e.target.value)}
              />
            </FormField>
            <FormField label="Modelo" htmlFor="modelo">
              <Input
                id="modelo"
                value={form.modelo}
                onChange={(e) => set("modelo", e.target.value)}
              />
            </FormField>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField
              label={ehConsumivel ? "Quantidade em estoque" : "Quantidade"}
              htmlFor="quantidade"
              obrigatorio
              erro={erros.quantidade}
            >
              <Input
                id="quantidade"
                type="number"
                min={0}
                value={form.quantidade}
                onChange={(e) => set("quantidade", e.target.value)}
              />
            </FormField>
            <FormField
              label="Estoque mínimo"
              htmlFor="estoque_minimo"
              erro={erros.estoque_minimo}
            >
              <Input
                id="estoque_minimo"
                type="number"
                min={0}
                value={form.estoque_minimo}
                onChange={(e) => set("estoque_minimo", e.target.value)}
                disabled={!ehConsumivel}
              />
            </FormField>
          </div>
          {ehConsumivel && (
            <p className="-mt-2 text-xs text-muted-foreground">
              Para materiais de consumo, o alerta é disparado quando a
              quantidade fica abaixo do estoque mínimo.
            </p>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Departamento / localização">
              <Select
                value={form.setor_id ? String(form.setor_id) : SEM}
                onValueChange={(v) =>
                  set("setor_id", v === SEM ? null : Number(v))
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não definido</SelectItem>
                  {setoresArvore.map(({ setor: s, nivel }) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {prefixoIndentacao(nivel)}
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Servidor responsável">
              <Select
                value={
                  form.responsavel_id ? String(form.responsavel_id) : SEM
                }
                onValueChange={(v) =>
                  set("responsavel_id", v === SEM ? null : Number(v))
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não definido</SelectItem>
                  {servidores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome} ({s.matricula})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Data de aquisição" htmlFor="data_aquisicao">
              <Input
                id="data_aquisicao"
                type="date"
                value={form.data_aquisicao}
                onChange={(e) => set("data_aquisicao", e.target.value)}
              />
            </FormField>
            <FormField label="Valor (R$)" htmlFor="valor" erro={erros.valor}>
              <Input
                id="valor"
                type="number"
                min={0}
                step="0.01"
                value={form.valor}
                onChange={(e) => set("valor", e.target.value)}
                placeholder="Opcional"
              />
            </FormField>
          </div>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
