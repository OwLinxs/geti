import { Inbox } from "lucide-react";

export function EstadoVazio({
  titulo = "Nenhum registro encontrado",
  descricao,
}: {
  titulo?: string;
  descricao?: string;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 py-12 text-center text-muted-foreground">
      <Inbox className="h-10 w-10 opacity-50" />
      <p className="text-sm font-medium">{titulo}</p>
      {descricao && <p className="text-xs">{descricao}</p>}
    </div>
  );
}
