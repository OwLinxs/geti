import * as React from "react";
import type { Perfil, Usuario } from "@/types";
import { authApi } from "@/services/api";
import {
  clearToken,
  getToken,
  setOnUnauthorized,
  setToken,
} from "@/services/api/client";

interface AuthContextValue {
  usuario: Usuario | null;
  carregando: boolean;
  autenticado: boolean;
  ehAdministrador: boolean;
  ehEquipe: boolean; // administrador ou operador (área interna de T.I.)
  ehSolicitante: boolean; // usuário final (portal de chamados)
  entrar: (email: string, senha: string) => Promise<void>;
  registrar: (nome: string, email: string, senha: string) => Promise<void>;
  sair: () => void;
}

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function useAuth() {
  const ctx = React.useContext(AuthContext);
  if (!ctx) throw new Error("useAuth deve ser usado dentro de AuthProvider");
  return ctx;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [usuario, setUsuario] = React.useState<Usuario | null>(null);
  const [carregando, setCarregando] = React.useState(true);

  const sair = React.useCallback(() => {
    clearToken();
    setUsuario(null);
  }, []);

  // Registra o callback de 401 do client axios para encerrar a sessão.
  React.useEffect(() => {
    setOnUnauthorized(() => {
      setUsuario(null);
    });
  }, []);

  // Ao montar, se houver token, recupera o usuário atual.
  React.useEffect(() => {
    const token = getToken();
    if (!token) {
      setCarregando(false);
      return;
    }
    authApi
      .euMesmo()
      .then((u) => setUsuario(u))
      .catch(() => clearToken())
      .finally(() => setCarregando(false));
  }, []);

  const entrar = React.useCallback(async (email: string, senha: string) => {
    const res = await authApi.login(email, senha);
    setToken(res.token);
    setUsuario(res.usuario);
  }, []);

  const registrar = React.useCallback(
    async (nome: string, email: string, senha: string) => {
      const res = await authApi.registrar(nome, email, senha);
      setToken(res.token);
      setUsuario(res.usuario);
    },
    []
  );

  const perfil = usuario?.perfil;
  const value: AuthContextValue = {
    usuario,
    carregando,
    autenticado: !!usuario,
    ehAdministrador: perfil === ("administrador" as Perfil),
    ehEquipe: perfil === "administrador" || perfil === "operador",
    ehSolicitante: perfil === "solicitante",
    entrar,
    registrar,
    sair,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
