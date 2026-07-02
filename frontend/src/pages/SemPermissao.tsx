import { Link } from "react-router-dom";
import { ShieldX } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function SemPermissao() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-4 text-center">
      <ShieldX className="h-12 w-12 text-destructive" />
      <div>
        <h1 className="text-xl font-semibold text-foreground">
          Acesso não autorizado
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Esta área é exclusiva de administradores.
        </p>
      </div>
      <Button asChild variant="outline">
        <Link to="/">Voltar ao painel</Link>
      </Button>
    </div>
  );
}
