import * as React from "react";
import { Outlet } from "react-router-dom";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";

// Casca institucional: sidebar fixa em desktop, drawer em mobile.
export function AppShell() {
  const [menuAberto, setMenuAberto] = React.useState(false);

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar aberto={menuAberto} />

      {/* Overlay para fechar o menu em mobile. */}
      {menuAberto && (
        <div
          className="fixed inset-0 z-30 bg-black/40 lg:hidden"
          onClick={() => setMenuAberto(false)}
          aria-hidden
        />
      )}

      <div className="flex flex-1 flex-col lg:pl-0">
        <Header onToggleMenu={() => setMenuAberto((v) => !v)} />
        <main className="flex-1 p-4 lg:p-6">
          <div className="mx-auto w-full max-w-7xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
