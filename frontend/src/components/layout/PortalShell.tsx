import { Outlet, useNavigate } from "react-router-dom";
import { LifeBuoy, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

// Shell enxuto do portal do solicitante (usuário final). Sem a navegação
// interna da equipe de T.I.
export function PortalShell() {
  const { usuario, sair } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-muted/30">
      <header className="border-b border-border bg-card">
        <div className="mx-auto flex h-16 max-w-3xl items-center justify-between px-4">
          <button
            className="flex items-center gap-2"
            onClick={() => navigate("/")}
          >
            <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary text-primary-foreground">
              <LifeBuoy className="h-5 w-5" />
            </div>
            <div className="text-left leading-tight">
              <p className="text-sm font-semibold text-foreground">
                Chamados de T.I.
              </p>
              <p className="text-[11px] text-muted-foreground">
                Portal do solicitante
              </p>
            </div>
          </button>

          <div className="flex items-center gap-3">
            <span className="hidden text-sm text-muted-foreground sm:inline">
              {usuario?.nome}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                sair();
                navigate("/login", { replace: true });
              }}
            >
              <LogOut className="h-4 w-4" /> Sair
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
