import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";

export default function NaoEncontrado() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20 text-center">
      <p className="text-5xl font-bold text-muted-foreground/40">404</p>
      <div>
        <h1 className="text-lg font-semibold text-foreground">
          Página não encontrada
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          O endereço acessado não existe.
        </p>
      </div>
      <Button asChild variant="outline">
        <Link to="/">Voltar ao painel</Link>
      </Button>
    </div>
  );
}
