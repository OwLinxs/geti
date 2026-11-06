import * as React from "react";
import { Outlet } from "react-router-dom";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";
import { WidgetConversasWhatsApp } from "@/components/WidgetConversasWhatsApp";
import { useAuth } from "@/contexts/AuthContext";

// Casca institucional: sidebar fixa em desktop, drawer em mobile.
export function AppShell() {
  const [menuAberto, setMenuAberto] = React.useState(false);
  const [recolhido, setRecolhido] = React.useState(() => {
    try {
      return localStorage.getItem("sidebar_recolhida") === "1";
    } catch {
      return false;
    }
  });
  const { ehEquipe } = useAuth();

  // No desktop o botão recolhe/expande a sidebar (largura); no mobile abre/fecha
  // o drawer. Decide pelo breakpoint lg (1024px).
  const alternarMenu = React.useCallback(() => {
    const desktop =
      typeof window !== "undefined" &&
      window.matchMedia("(min-width: 1024px)").matches;
    if (desktop) {
      setRecolhido((v) => {
        const novo = !v;
        try {
          localStorage.setItem("sidebar_recolhida", novo ? "1" : "0");
        } catch {
          /* ignore */
        }
        return novo;
      });
    } else {
      setMenuAberto((v) => !v);
    }
  }, []);

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar aberto={menuAberto} recolhido={recolhido} />

      {/* Overlay para fechar o menu em mobile. */}
      {menuAberto && (
        <div
          className="fixed inset-0 z-30 bg-black/40 lg:hidden"
          onClick={() => setMenuAberto(false)}
          aria-hidden
        />
      )}

      <div className="flex flex-1 flex-col lg:pl-0">
        <Header onToggleMenu={alternarMenu} />
        <main className="flex-1 p-4 lg:p-6">
          <div className="mx-auto w-full max-w-7xl">
            <Outlet />
          </div>
        </main>
      </div>

      {/* Widget flutuante de conversas do WhatsApp (só para a equipe de T.I.). */}
      {ehEquipe && <WidgetConversasWhatsApp />}
    </div>
  );
}
