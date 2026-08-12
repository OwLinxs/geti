import { Routes, Route, Navigate } from "react-router-dom";
import { AppShell } from "@/components/layout/AppShell";
import { ProtectedRoute } from "@/routes/ProtectedRoute";

import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import ItensLista from "@/pages/itens/ItensLista";
import ItemDetalhe from "@/pages/itens/ItemDetalhe";
import Movimentacoes from "@/pages/Movimentacoes";
import Manutencao from "@/pages/manutencao/Manutencao";
import OrdemServicoDetalhe from "@/pages/manutencao/OrdemServicoDetalhe";
import Alertas from "@/pages/Alertas";
import Termos from "@/pages/Termos";
import Conhecimento from "@/pages/conhecimento/Conhecimento";
import ConhecimentoDetalhe from "@/pages/conhecimento/ConhecimentoDetalhe";
import Relatorios from "@/pages/Relatorios";
import Categorias from "@/pages/cadastros/Categorias";
import Setores from "@/pages/cadastros/Setores";
import Servidores from "@/pages/cadastros/Servidores";
import Fornecedores from "@/pages/cadastros/Fornecedores";
import Contratos from "@/pages/cadastros/Contratos";
import Reservas from "@/pages/Reservas";
import Usuarios from "@/pages/cadastros/Usuarios";
import Auditoria from "@/pages/Auditoria";
import SemPermissao from "@/pages/SemPermissao";
import NaoEncontrado from "@/pages/NaoEncontrado";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/sem-permissao" element={<SemPermissao />} />

      {/* Rotas autenticadas (qualquer perfil). */}
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/itens" element={<ItensLista />} />
          <Route path="/itens/:id" element={<ItemDetalhe />} />
          <Route path="/movimentacoes" element={<Movimentacoes />} />
          <Route path="/manutencao" element={<Manutencao />} />
          <Route path="/manutencao/:id" element={<OrdemServicoDetalhe />} />
          <Route path="/reservas" element={<Reservas />} />
          <Route path="/alertas" element={<Alertas />} />
          <Route path="/termos" element={<Termos />} />
          <Route path="/conhecimento" element={<Conhecimento />} />
          <Route path="/conhecimento/:id" element={<ConhecimentoDetalhe />} />
          <Route path="/relatorios" element={<Relatorios />} />
          <Route path="/categorias" element={<Categorias />} />
          <Route path="/setores" element={<Setores />} />
          <Route path="/servidores" element={<Servidores />} />
          <Route path="/fornecedores" element={<Fornecedores />} />
          <Route path="/contratos" element={<Contratos />} />

          {/* Rotas exclusivas de administrador. */}
          <Route element={<ProtectedRoute somenteAdmin />}>
            <Route path="/usuarios" element={<Usuarios />} />
            <Route path="/auditoria" element={<Auditoria />} />
          </Route>

          <Route path="*" element={<NaoEncontrado />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
