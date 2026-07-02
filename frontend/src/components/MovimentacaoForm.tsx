import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
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
import { UserPlus } from "lucide-react";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { NovoServidorDialog } from "@/components/NovoServidorDialog";
import { movimentacoesApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import {
  TIPOS_ENTRADA,
  TIPOS_SAIDA,
  ehBaixa,
  ehEntrada,
} from "@/lib/rotulos";
import type {
  Item,
  MovimentacaoPayload,
  ResultadoMovimentacao,
  Servidor,
  Setor,
  TipoMovimentacao,
} from "@/types";

const SEM = "0";

interface MovimentacaoFormProps {
  aberto: boolean;
  // Item pré-selecionado (na tela de detalhe) ou lista para escolher.
  itemFixo?: Item | null;
  itens?: Item[];
  setores: Setor[];
  servidores: Servidor[];
  onFechar: () => void;
  onRegistrado: (r: ResultadoMovimentacao) => void;
  // Notifica o pai quando um servidor é criado no cadastro rápido, para que a
  // lista de referência global seja atualizada.
  onServidorCriado?: (s: Servidor) => void;
}

export function MovimentacaoForm({
  aberto,
  itemFixo,
  itens,
  setores,
  servidores,
  onFechar,
  onRegistrado,
  onServidorCriado,
}: MovimentacaoFormProps) {
  const { toast } = useToast();
  const [itemId, setItemId] = React.useState<number | null>(
    itemFixo?.id ?? null
  );
  const [tipo, setTipo] = React.useState<TipoMovimentacao | "">("");
  const [quantidade, setQuantidade] = React.useState("1");
  const [setorOrigem, setSetorOrigem] = React.useState<string>(SEM);
  const [setorDestino, setSetorDestino] = React.useState<string>(SEM);
  const [servidorId, setServidorId] = React.useState<string>(SEM);
  const [observacao, setObservacao] = React.useState("");
  const [motivoBaixa, setMotivoBaixa] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  // Lista local de servidores: espelha a prop, mas permite acrescentar um novo
  // cadastrado na hora sem esperar o recarregamento do pai.
  const [listaServidores, setListaServidores] = React.useState<Servidor[]>(servidores);
  const [novoServidorAberto, setNovoServidorAberto] = React.useState(false);
  React.useEffect(() => setListaServidores(servidores), [servidores]);

  React.useEffect(() => {
    if (aberto) {
      setItemId(itemFixo?.id ?? null);
      setTipo("");
      setQuantidade("1");
      setSetorOrigem(SEM);
      setSetorDestino(SEM);
      setServidorId(SEM);
      setObservacao("");
      setMotivoBaixa("");
      setErros({});
    }
  }, [aberto, itemFixo]);

  const itemAtual = itemFixo ?? itens?.find((i) => i.id === itemId) ?? null;
  const entrada = tipo ? ehEntrada(tipo) : false;
  const baixa = tipo ? ehBaixa(tipo) : false;

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novosErros: Record<string, string> = {};
    if (!itemId) novosErros.item_id = "Selecione o item.";
    if (!tipo) novosErros.tipo = "Selecione o tipo de movimentação.";
    const qtd = Number(quantidade);
    if (Number.isNaN(qtd) || qtd <= 0)
      novosErros.quantidade = "Informe uma quantidade maior que zero.";
    if (
      !entrada &&
      itemAtual &&
      qtd > itemAtual.quantidade
    ) {
      novosErros.quantidade = `Saída maior que o estoque disponível (${itemAtual.quantidade}).`;
    }
    if (baixa && !motivoBaixa.trim())
      novosErros.motivo_baixa = "Informe o motivo da baixa.";
    if (Object.keys(novosErros).length) {
      setErros(novosErros);
      return;
    }
    setErros({});

    const payload: MovimentacaoPayload = {
      item_id: itemId!,
      tipo: tipo as TipoMovimentacao,
      quantidade: qtd,
      setor_origem_id: setorOrigem !== SEM ? Number(setorOrigem) : null,
      setor_destino_id: setorDestino !== SEM ? Number(setorDestino) : null,
      servidor_id: servidorId !== SEM ? Number(servidorId) : null,
      observacao: observacao.trim() || undefined,
      motivo_baixa: baixa ? motivoBaixa.trim() : undefined,
    };

    setSalvando(true);
    try {
      const r = await movimentacoesApi.registrar(payload);
      toast({
        titulo: "Movimentação registrada.",
        descricao: r.alerta_estoque
          ? "Atenção: o estoque ficou abaixo do mínimo."
          : undefined,
        variant: r.alerta_estoque ? "default" : "success",
      });
      onRegistrado(r);
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível registrar",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <>
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>Registrar movimentação</DialogTitle>
        </DialogHeader>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          {itemFixo ? (
            <div className="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm">
              <span className="text-muted-foreground">Item: </span>
              <span className="font-medium text-foreground">
                {itemFixo.descricao}
              </span>{" "}
              <span className="text-muted-foreground">
                (estoque atual: {itemFixo.quantidade})
              </span>
            </div>
          ) : (
            <FormField label="Item" obrigatorio erro={erros.item_id}>
              <Select
                value={itemId ? String(itemId) : ""}
                onValueChange={(v) => setItemId(Number(v))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o item" />
                </SelectTrigger>
                <SelectContent>
                  {(itens ?? []).map((i) => (
                    <SelectItem key={i.id} value={String(i.id)}>
                      {i.descricao}
                      {i.numero_patrimonio ? ` — ${i.numero_patrimonio}` : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Tipo" obrigatorio erro={erros.tipo}>
              <Select
                value={tipo}
                onValueChange={(v) => setTipo(v as TipoMovimentacao)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione" />
                </SelectTrigger>
                <SelectContent>
                  {TIPOS_ENTRADA.map((t) => (
                    <SelectItem key={t.valor} value={t.valor}>
                      {t.rotulo}
                    </SelectItem>
                  ))}
                  {TIPOS_SAIDA.map((t) => (
                    <SelectItem key={t.valor} value={t.valor}>
                      {t.rotulo}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField
              label="Quantidade"
              htmlFor="qtd_mov"
              obrigatorio
              erro={erros.quantidade}
            >
              <Input
                id="qtd_mov"
                type="number"
                min={1}
                value={quantidade}
                onChange={(e) => setQuantidade(e.target.value)}
              />
            </FormField>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Departamento de origem">
              <Select value={setorOrigem} onValueChange={setSetorOrigem}>
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não informado</SelectItem>
                  {setores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Departamento de destino">
              <Select value={setorDestino} onValueChange={setSetorDestino}>
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não informado</SelectItem>
                  {setores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          </div>

          <FormField label="Servidor responsável envolvido">
            <div className="flex items-center gap-2">
              <Select value={servidorId} onValueChange={setServidorId}>
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não informado</SelectItem>
                  {listaServidores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome} ({s.matricula})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="shrink-0"
                onClick={() => setNovoServidorAberto(true)}
              >
                <UserPlus className="h-4 w-4" /> Novo
              </Button>
            </div>
          </FormField>

          {baixa && (
            <FormField
              label="Motivo da baixa"
              htmlFor="motivo"
              obrigatorio
              erro={erros.motivo_baixa}
            >
              <Input
                id="motivo"
                value={motivoBaixa}
                onChange={(e) => setMotivoBaixa(e.target.value)}
                placeholder="Ex.: Equipamento sem conserto"
              />
            </FormField>
          )}

          <FormField label="Observação" htmlFor="obs">
            <Textarea
              id="obs"
              value={observacao}
              onChange={(e) => setObservacao(e.target.value)}
            />
          </FormField>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Registrar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <NovoServidorDialog
      aberto={novoServidorAberto}
      setores={setores}
      onFechar={() => setNovoServidorAberto(false)}
      onCriado={(s) => {
        setListaServidores((prev) => [s, ...prev]);
        setServidorId(String(s.id));
        setNovoServidorAberto(false);
        onServidorCriado?.(s);
      }}
    />
    </>
  );
}
