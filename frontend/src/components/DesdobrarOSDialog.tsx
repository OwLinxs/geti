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
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { ordensServicoApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { cn } from "@/lib/utils";
import type { Item, OrdemServico } from "@/types";

const SEM = "0";
type OrigemEquip = "inventario" | "externa";

interface Props {
  aberto: boolean;
  osId: number;
  numeroOrigem: string;
  itens: Item[];
  onFechar: () => void;
  onCriada: (nova: OrdemServico) => void;
}

/**
 * Desdobra um chamado em uma OS irmã para um 2º equipamento (ex.: o solicitante
 * trouxe 2 computadores). A nova OS entra no mesmo atendimento; o canal de
 * conversa e o vínculo externo ficam na OS de origem.
 */
export function DesdobrarOSDialog({
  aberto,
  osId,
  numeroOrigem,
  itens,
  onFechar,
  onCriada,
}: Props) {
  const { toast } = useToast();
  const [origem, setOrigem] = React.useState<OrigemEquip>("inventario");
  const [itemId, setItemId] = React.useState<string>(SEM);
  const [equipDescricao, setEquipDescricao] = React.useState("");
  const [equipIdent, setEquipIdent] = React.useState("");
  const [defeito, setDefeito] = React.useState("");
  const [erro, setErro] = React.useState("");
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (aberto) {
      setOrigem("inventario");
      setItemId(SEM);
      setEquipDescricao("");
      setEquipIdent("");
      setDefeito("");
      setErro("");
    }
  }, [aberto]);

  // Só bens patrimoniados fazem sentido como equipamento de manutenção.
  const patrimonios = React.useMemo(
    () =>
      itens.filter(
        (i) => !!i.numero_patrimonio || !(i.categoria?.consumivel ?? false)
      ),
    [itens]
  );

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setErro("");
    if (origem === "inventario" && itemId === SEM) {
      setErro("Selecione o item do inventário.");
      return;
    }
    if (origem === "externa" && !equipDescricao.trim()) {
      setErro("Descreva a máquina externa.");
      return;
    }
    setSalvando(true);
    try {
      const nova = await ordensServicoApi.desdobrar(osId, {
        item_id: origem === "inventario" ? Number(itemId) : 0,
        equipamento_descricao:
          origem === "externa" ? equipDescricao.trim() : "",
        equipamento_identificacao:
          origem === "externa" ? equipIdent.trim() : "",
        defeito_relatado: defeito.trim(),
      });
      toast({ titulo: `OS ${nova.numero} criada no atendimento.`, variant: "success" });
      onCriada(nova);
    } catch (err) {
      toast({
        titulo: "Não foi possível desdobrar o chamado",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Adicionar equipamento ao atendimento</DialogTitle>
        </DialogHeader>

        <p className="text-sm text-muted-foreground">
          Abre uma nova OS, no mesmo atendimento de <strong>{numeroOrigem}</strong>
          , para um segundo equipamento do mesmo chamado. A conversa com o
          solicitante continua na OS de origem.
        </p>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField label="Equipamento" erro={erro}>
            <div className="mb-2 flex gap-2">
              {(
                [
                  ["inventario", "Item do inventário"],
                  ["externa", "Máquina externa"],
                ] as [OrigemEquip, string][]
              ).map(([v, rot]) => (
                <button
                  key={v}
                  type="button"
                  onClick={() => setOrigem(v)}
                  className={cn(
                    "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
                    origem === v
                      ? "border-primary bg-primary text-primary-foreground"
                      : "border-border bg-card hover:bg-accent"
                  )}
                >
                  {rot}
                </button>
              ))}
            </div>

            {origem === "inventario" ? (
              <Select value={itemId} onValueChange={setItemId}>
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o item" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Selecione…</SelectItem>
                  {patrimonios.map((i) => (
                    <SelectItem key={i.id} value={String(i.id)}>
                      {i.descricao}
                      {i.numero_patrimonio ? ` — ${i.numero_patrimonio}` : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  value={equipDescricao}
                  onChange={(e) => setEquipDescricao(e.target.value)}
                  placeholder="Descrição (ex.: Notebook do usuário)"
                />
                <Input
                  value={equipIdent}
                  onChange={(e) => setEquipIdent(e.target.value)}
                  placeholder="Série / patrimônio / tag (opcional)"
                />
              </div>
            )}
          </FormField>

          <FormField label="Defeito deste equipamento" htmlFor="defeito-irma">
            <Textarea
              id="defeito-irma"
              value={defeito}
              onChange={(e) => setDefeito(e.target.value)}
              placeholder="Opcional — em branco, repete o defeito do chamado."
            />
          </FormField>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Criar OS no atendimento"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
