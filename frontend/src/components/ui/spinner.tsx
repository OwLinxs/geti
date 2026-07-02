import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={cn("h-5 w-5 animate-spin", className)} />;
}

export function CarregandoTela({ texto = "Carregando..." }: { texto?: string }) {
  return (
    <div className="flex h-full min-h-[40vh] flex-col items-center justify-center gap-3 text-muted-foreground">
      <Spinner className="h-7 w-7" />
      <span className="text-sm">{texto}</span>
    </div>
  );
}
