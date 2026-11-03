import { NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  Package,
  Monitor,
  ArrowLeftRight,
  Wrench,
  FileText,
  BookOpen,
  BarChart3,
  Tags,
  Building2,
  Users,
  UserCog,
  AlertTriangle,
  ShieldCheck,
  CalendarClock,
  Truck,
  ScrollText,
  Tag,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/contexts/AuthContext";

interface ItemMenu {
  rotulo: string;
  para: string;
  icone: React.ComponentType<{ className?: string }>;
  // Exige uma permissão delegável para aparecer (administrador sempre vê).
  permissao?: string;
}

const PRINCIPAL: ItemMenu[] = [
  { rotulo: "Painel", para: "/", icone: LayoutDashboard },
  { rotulo: "Itens", para: "/itens", icone: Package },
  { rotulo: "Patrimônio", para: "/patrimonio", icone: Monitor },
  { rotulo: "Movimentações", para: "/movimentacoes", icone: ArrowLeftRight },
  { rotulo: "Manutenção", para: "/manutencao", icone: Wrench },
  { rotulo: "Reservas", para: "/reservas", icone: CalendarClock },
  { rotulo: "Alertas de Estoque", para: "/alertas", icone: AlertTriangle },
  { rotulo: "Termos", para: "/termos", icone: FileText },
  { rotulo: "Base de Conhecimento", para: "/conhecimento", icone: BookOpen },
  { rotulo: "Relatórios", para: "/relatorios", icone: BarChart3 },
];

const CADASTROS: ItemMenu[] = [
  { rotulo: "Categorias", para: "/categorias", icone: Tags },
  { rotulo: "Departamentos", para: "/setores", icone: Building2 },
  { rotulo: "Servidores", para: "/servidores", icone: Users },
  { rotulo: "Fornecedores", para: "/fornecedores", icone: Truck },
  { rotulo: "Contratos", para: "/contratos", icone: ScrollText },
  {
    rotulo: "Chamados (config)",
    para: "/config/chamados",
    icone: Tag,
    permissao: "config.gerenciar",
  },
  {
    rotulo: "Usuários",
    para: "/usuarios",
    icone: UserCog,
    permissao: "usuarios.gerenciar",
  },
  {
    rotulo: "Auditoria",
    para: "/auditoria",
    icone: ShieldCheck,
    permissao: "auditoria.ver",
  },
];

export function Sidebar({ aberto }: { aberto: boolean }) {
  const { pode } = useAuth();

  return (
    <aside
      className={cn(
        "fixed inset-y-0 left-0 z-40 w-64 transform border-r border-border bg-card transition-transform duration-200 lg:static lg:translate-x-0",
        aberto ? "translate-x-0" : "-translate-x-full"
      )}
    >
      <div className="flex h-16 items-center gap-2 border-b border-border px-6">
        <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Package className="h-5 w-5" />
        </div>
        <div className="leading-tight">
          <p className="text-sm font-semibold text-foreground">SIGE-TI</p>
          <p className="text-[11px] text-muted-foreground">Estoque de T.I.</p>
        </div>
      </div>

      <nav className="flex flex-col gap-6 overflow-y-auto px-3 py-4">
        <Grupo titulo="Geral" itens={PRINCIPAL} pode={pode} />
        <Grupo titulo="Cadastros" itens={CADASTROS} pode={pode} />
      </nav>
    </aside>
  );
}

function Grupo({
  titulo,
  itens,
  pode,
}: {
  titulo: string;
  itens: ItemMenu[];
  pode: (chave: string) => boolean;
}) {
  const visiveis = itens.filter((i) => !i.permissao || pode(i.permissao));
  return (
    <div>
      <p className="mb-2 px-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
        {titulo}
      </p>
      <ul className="space-y-1">
        {visiveis.map((item) => (
          <li key={item.para}>
            <NavLink
              to={item.para}
              end={item.para === "/"}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-foreground/80 hover:bg-accent hover:text-accent-foreground"
                )
              }
            >
              <item.icone className="h-4 w-4 shrink-0" />
              {item.rotulo}
            </NavLink>
          </li>
        ))}
      </ul>
    </div>
  );
}
