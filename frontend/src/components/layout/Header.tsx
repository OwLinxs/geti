import { LogOut, Menu, UserCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/contexts/AuthContext";
import { rotuloPerfil } from "@/lib/rotulos";

export function Header({ onToggleMenu }: { onToggleMenu: () => void }) {
  const { usuario, sair } = useAuth();

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-card/95 px-4 backdrop-blur lg:px-6">
      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          onClick={onToggleMenu}
          aria-label="Abrir menu"
        >
          <Menu className="h-5 w-5" />
        </Button>
        <div className="hidden sm:block">
          <h1 className="text-sm font-semibold text-foreground">
            Sistema de Gestão de Estoque de T.I.
          </h1>
          <p className="text-xs text-muted-foreground">
            Departamento de Tecnologia da Informação
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        {usuario && (
          <div className="hidden items-center gap-2 sm:flex">
            <UserCircle2 className="h-5 w-5 text-muted-foreground" />
            <div className="text-right leading-tight">
              <p className="text-sm font-medium text-foreground">
                {usuario.nome}
              </p>
              <Badge variant="muted" className="text-[10px]">
                {rotuloPerfil(usuario.perfil)}
              </Badge>
            </div>
          </div>
        )}
        <Button variant="outline" size="sm" onClick={sair}>
          <LogOut className="h-4 w-4" />
          <span className="hidden sm:inline">Sair</span>
        </Button>
      </div>
    </header>
  );
}
