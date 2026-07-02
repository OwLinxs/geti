import * as React from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { termosApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { baixarBlob } from "@/lib/format";
import type {
  Item,
  Servidor,
  TermoPayload,
  TermoResponsabilidade,
} from "@/types";

interface TermoFormProps {
  aberto: boolean;
  itemFixo?: Item | null;
  itens?: Item[];
  servidores: Servidor[];
  onFechar: () => void;
  onEmitido: (t: TermoResponsabilidade) => void;
}

export function TermoForm({
  aberto,
  itemFixo,
  itens,
  servidores,
  onFechar,
  onEmitido,
}: TermoFormProps) {
  const { toast } = useToast();
  const [itemId, setItemId] = React.useState<number | null>(
    itemFixo?.id ?? null
  );
  const [servidorId, setServidorId] = React.useState<number | null>(null);
  const [observacao, setObservacao] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (aberto) {
      setItemId(itemFixo?.id ?? null);
      setServidorId(itemFixo?.responsavel_id ?? null);
      setObservacao("");
      setErros({});
    }
  }, [aberto, itemFixo]);

  async function emitir(e: React.FormEvent) {
    e.preventDefault();
    const novosErros: Record<string, string> = {};
    if (!itemId) novosErros.item_id = "Selecione o item.";
    if (!servidorId) novosErros.servidor_id = "Selecione o servidor.";
    if (Object.keys(novosErros).length) {
      setErros(novosErros);
      return;
    }
    setErros({});

    const payload: TermoPayload = {
      item_id: itemId!,
      servidor_id: servidorId!,
      observacao: observacao.trim() || undefined,
    };

    setSalvando(true);
    try {
      const termo = await termosApi.emitir(payload);
      toast({
        titulo: `Termo ${termo.numero} emitido.`,
        descricao: "Iniciando download do PDF...",
        variant: "success",
      });
      // Baixa o PDF automaticamente após emitir.
      try {
        const pdf = await termosApi.baixarPdf(termo.id);
        baixarBlob(pdf, `termo-${termo.numero}.pdf`);
      } catch {
        toast({
          titulo: "Termo emitido, mas o PDF não pôde ser baixado.",
          descricao: "Tente baixá-lo novamente na lista de termos.",
          variant: "default",
        });
      }
      onEmitido(termo);
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível emitir o termo",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Emitir termo de responsabilidade</DialogTitle>
          <DialogDescription>
            Gera o recibo de entrega do equipamento ao servidor. O PDF é baixado
            automaticamente.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={emitir} className="space-y-4" noValidate>
          {itemFixo ? (
            <div className="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm">
              <span className="text-muted-foreground">Item: </span>
              <span className="font-medium text-foreground">
                {itemFixo.descricao}
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

          <FormField label="Servidor" obrigatorio erro={erros.servidor_id}>
            <Select
              value={servidorId ? String(servidorId) : ""}
              onValueChange={(v) => setServidorId(Number(v))}
            >
              <SelectTrigger>
                <SelectValue placeholder="Selecione o servidor" />
              </SelectTrigger>
              <SelectContent>
                {servidores.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.nome} ({s.matricula})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <FormField label="Observação" htmlFor="obs_termo">
            <Textarea
              id="obs_termo"
              value={observacao}
              onChange={(e) => setObservacao(e.target.value)}
              placeholder="Opcional"
            />
          </FormField>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Emitir e baixar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
