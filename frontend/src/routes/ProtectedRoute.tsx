import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { CarregandoTela } from "@/components/ui/spinner";

// Garante autenticação. Opcionalmente exige perfil administrador.
export function ProtectedRoute({
  somenteAdmin = false,
}: {
  somenteAdmin?: boolean;
}) {
  const { autenticado, carregando, ehAdministrador } = useAuth();
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

  return <Outlet />;
}
