import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { itensApi, type ResultadoLoteItens } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { codigoInterno } from "@/lib/patrimonio";
import { baixarBlob, dateInputParaISO } from "@/lib/format";
import { useToast } from "@/components/ui/toast";
import type { Categoria, Setor } from "@/types";

interface Props {
  aberto: boolean;
  categorias: Categoria[];
  setores: Setor[];
  onFechar: () => void;
  onCriado: (resultado: ResultadoLoteItens) => void;
}

export function CadastrarLoteDialog({ aberto, categorias, setores, onFechar, onCriado }: Props) {
  const { toast } = useToast();
  const [categoriaId, setCategoriaId] = React.useState(0);
  const [descricao, setDescricao] = React.useState("");
  const [marca, setMarca] = React.useState("");
  const [modelo, setModelo] = React.useState("");
  const [unidades, setUnidades] = React.useState("1");
  const [setorId, setSetorId] = React.useState("0");
  const [dataAquisicao, setDataAquisicao] = React.useState("");
  const [valor, setValor] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);
  const [resultado, setResultado] = React.useState<ResultadoLoteItens | null>(null);

  React.useEffect(() => {
    if (!aberto) return;
    setCategoriaId(0);
    setDescricao("");
    setMarca("");
    setModelo("");
    setUnidades("1");
    setSetorId("0");
    setDataAquisicao("");
    setValor("");
    setErros({});
    setResultado(null);
  }, [aberto]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    const qtd = Number(unidades);
    if (!descricao.trim()) novos.descricao = "Informe a descrição.";
    if (!categoriaId) novos.categoria_id = "Selecione uma categoria.";
    if (!Number.isInteger(qtd) || qtd < 1 || qtd > 200) novos.unidades = "Informe de 1 a 200 unidades.";
    if (valor && (!Number.isFinite(Number(valor)) || Number(valor) < 0)) novos.valor = "Valor inválido.";
    setErros(novos);
    if (Object.keys(novos).length) return;

    setSalvando(true);
    try {
      const res = await itensApi.criarLote({
        descricao: descricao.trim(), categoria_id: categoriaId, marca: marca.trim(),
        modelo: modelo.trim(), estado_conservacao: "novo", unidades: qtd,
        setor_id: setorId === "0" ? null : Number(setorId),
        data_aquisicao: dateInputParaISO(dataAquisicao),
        valor: valor ? Number(valor) : null,
      });
      setResultado(res);
      onCriado(res);
      toast({ titulo: `${res.quantidade} bens cadastrados individualmente.`, variant: "success" });
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else toast({ titulo: "Não foi possível cadastrar o lote", descricao: mensagemErro(err), variant: "destructive" });
    } finally {
      setSalvando(false);
    }
  }

  function baixarCodigos() {
    if (!resultado) return;
    const linhas = ["codigo_ti;descricao;lote", ...resultado.ids.map((id) =>
      `${codigoInterno(id)};"${descricao.replace(/"/g, '""')}";${resultado.lote_codigo}`)];
    baixarBlob(new Blob(["\uFEFF", linhas.join("\r\n")], { type: "text/csv;charset=utf-8" }),
      `codigos-${resultado.lote_codigo}.csv`);
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>Receber bens em lote</DialogTitle>
          <DialogDescription>
            O sistema cria um registro e um código interno para cada unidade. Patrimônio,
            série e destino podem ser preenchidos depois, em cada bem.
          </DialogDescription>
        </DialogHeader>
        {resultado ? (
          <div className="space-y-3 text-sm">
            <p className="font-medium">{resultado.quantidade} unidades criadas no lote {resultado.lote_codigo}.</p>
            <p>Códigos internos das unidades:</p>
            <div className="grid max-h-36 grid-cols-3 gap-1 overflow-y-auto rounded-md border p-2 font-mono text-xs sm:grid-cols-4">
              {resultado.ids.map((id) => <span key={id}>{codigoInterno(id)}</span>)}
            </div>
            <p className="text-muted-foreground">Identifique fisicamente cada unidade com seu código interno antes de distribuí-la.</p>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={baixarCodigos}>Baixar lista de códigos</Button>
              <Button type="button" onClick={onFechar}>Ver unidades</Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Categoria" obrigatorio erro={erros.categoria_id}>
                <Select value={categoriaId ? String(categoriaId) : ""} onValueChange={(v) => setCategoriaId(Number(v))}>
                  <SelectTrigger><SelectValue placeholder="Selecione" /></SelectTrigger>
                  <SelectContent>{categorias.filter((c) => !c.consumivel).map((c) => <SelectItem key={c.id} value={String(c.id)}>{c.nome}</SelectItem>)}</SelectContent>
                </Select>
              </FormField>
              <FormField label="Número de unidades" obrigatorio erro={erros.unidades}>
                <Input type="number" min={1} max={200} value={unidades} onChange={(e) => setUnidades(e.target.value)} />
              </FormField>
            </div>
            <FormField label="Descrição" obrigatorio erro={erros.descricao}>
              <Input value={descricao} onChange={(e) => setDescricao(e.target.value)} placeholder="Ex.: Monitor LG 24 polegadas" />
            </FormField>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Marca"><Input value={marca} onChange={(e) => setMarca(e.target.value)} /></FormField>
              <FormField label="Modelo"><Input value={modelo} onChange={(e) => setModelo(e.target.value)} /></FormField>
            </div>
            <FormField label="Local atual / departamento">
              <Select value={setorId} onValueChange={setSetorId}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">Ainda não definido</SelectItem>
                  {setores.map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.nome}</SelectItem>)}
                </SelectContent>
              </Select>
            </FormField>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <FormField label="Data de aquisição"><Input type="date" value={dataAquisicao} onChange={(e) => setDataAquisicao(e.target.value)} /></FormField>
              <FormField label="Valor por unidade" erro={erros.valor}><Input type="number" min={0} step="0.01" value={valor} onChange={(e) => setValor(e.target.value)} /></FormField>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={onFechar}>Cancelar</Button>
              <Button type="submit" disabled={salvando}>{salvando && <Spinner className="h-4 w-4" />} Cadastrar unidades</Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
