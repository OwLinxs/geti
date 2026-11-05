import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { CarregandoTela } from "@/components/ui/spinner";

// Garante autenticação. Opcionalmente exige perfil administrador ou uma
// permissão delegável específica (administrador sempre passa).
export function ProtectedRoute({
  somenteAdmin = false,
  permissao,
  permissoes,
}: {
  somenteAdmin?: boolean;
  permissao?: string;
  // "qualquer uma": libera se o usuário tiver ao menos uma destas chaves.
  permissoes?: string[];
}) {
  const { autenticado, carregando, ehAdministrador, pode } = useAuth();
  const location = useLocation();

  if (carregando) {
    return <CarregandoTela texto="Verificando sessão..." />;
  }

  if (!autenticado) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (somenteAdmin && !ehAdministrador) {
    return <Navigate to="/sem-permissao" replace />;
  }

  if (permissao && !pode(permissao)) {
    return <Navigate to="/sem-permissao" replace />;
  }

  if (permissoes && permissoes.length > 0 && !permissoes.some((c) => pode(c))) {
    return <Navigate to="/sem-permissao" replace />;
  }

  return <Outlet />;
}
