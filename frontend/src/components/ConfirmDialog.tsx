import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

interface ConfirmDialogProps {
  aberto: boolean;
  titulo: string;
  descricao?: string;
  textoConfirmar?: string;
  destrutivo?: boolean;
  processando?: boolean;
  onConfirmar: () => void;
  onCancelar: () => void;
}

export function ConfirmDialog({
  aberto,
  titulo,
  descricao,
  textoConfirmar = "Confirmar",
  destrutivo,
  processando,
  onConfirmar,
  onCancelar,
}: ConfirmDialogProps) {
  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onCancelar()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{titulo}</DialogTitle>
          {descricao && <DialogDescription>{descricao}</DialogDescription>}
        </DialogHeader>
        <DialogFooter className="mt-2 gap-2">
          <Button variant="outline" onClick={onCancelar} disabled={processando}>
            Cancelar
          </Button>
          <Button
            variant={destrutivo ? "destructive" : "default"}
            onClick={onConfirmar}
            disabled={processando}
          >
            {textoConfirmar}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
