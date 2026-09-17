import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Plus, LayoutGrid } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { OrdemServicoForm } from "./OrdemServicoForm";
import {
  itensApi,
  ordensServicoApi,
  usuariosApi,
  type FiltroOrdensServico,
} from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import { formatarData } from "@/lib/format";
import {
  STATUS_OS,
  rotuloPrioridadeOS,
  rotuloStatusOS,
  variantePrioridadeOS,
  varianteStatusOS,
} from "@/lib/rotulos";
import { cn } from "@/lib/utils";
import type { Item, OrdemServico, Usuario } from "@/types";

const TODOS = "todos";
const TAMANHO = 20;

export default function Manutencao() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const { ehAdministrador, usuario } = useAuth();
  const { setores, servidores } = useReferencias();

  const [lista, setLista] = React.useState<OrdemServico[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [statusFiltro, setStatusFiltro] = React.useState<string>(TODOS);
  const [carregando, setCarregando] = React.useState(true);

  const [itens, setItens] = React.useState<Item[]>([]);
  const [tecnicos, setTecnicos] = React.useState<Usuario[]>([]);
  const [formAberto, setFormAberto] = React.useState(false);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const f: FiltroOrdensServico = {
      pagina,
      tamanho: TAMANHO,
      status: statusFiltro !== TODOS ? statusFiltro : undefined,
    };
    ordensServicoApi
      .listar(f)
      .then((r) => {
        setLista(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar ordens de serviço",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [pagina, statusFiltro, toast]);

  React.useEffect(carregar, [carregar]);

  // Referências para o formulário: itens do inventário (ativos) e, se admin,
  // a lista de técnicos (usuários).
  React.useEffect(() => {
    itensApi
      .listar({ tamanho: 500, baixado: false })
      .then((r) => setItens(r.dados))
      .catch(() => setItens([]));
    if (ehAdministrador) {
      usuariosApi
        .listar()
        .then((us) => setTecnicos(us.filter((u) => u.ativo)))
        .catch(() => setTecnicos([]));
    }
  }, [ehAdministrador]);

  return (
    <div>
      <PageHeader
        titulo="Manutenção"
        descricao="Ordens de serviço de manutenção de equipamentos."
        acao={
          <>
            <Button variant="outline" onClick={() => navigate("/manutencao/kanban")}>
              <LayoutGrid className="h-4 w-4" /> Kanban
            </Button>
            <Button onClick={() => setFormAberto(true)}>
              <Plus className="h-4 w-4" /> Nova OS
            </Button>
          </>
        }
      />

      {/* Filtro por status. */}
      <div className="mb-4 flex flex-wrap gap-2">
        <FiltroBotao
          ativo={statusFiltro === TODOS}
          onClick={() => {
            setStatusFiltro(TODOS);
            setPagina(1);
          }}
        >
          Todas
        </FiltroBotao>
        {STATUS_OS.map((s) => (
          <FiltroBotao
            key={s.valor}
            ativo={statusFiltro === s.valor}
            onClick={() => {
              setStatusFiltro(s.valor);
              setPagina(1);
            }}
          >
            {s.rotulo}
          </FiltroBotao>
        ))}
      </div>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio
              titulo="Nenhuma ordem de serviço"
              descricao="Abra a primeira OS ao receber um equipamento para manutenção."
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Número</TableHead>
                  <TableHead>Equipamento</TableHead>
                  <TableHead>Prioridade</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Técnico</TableHead>
                  <TableHead>Abertura</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((os) => (
                  <TableRow
                    key={os.id}
                    className="cursor-pointer"
                    onClick={() => navigate(`/manutencao/${os.id}`)}
                  >
                    <TableCell className="font-medium">{os.numero}</TableCell>
                    <TableCell>{os.equipamento_snapshot}</TableCell>
                    <TableCell>
                      <Badge variant={variantePrioridadeOS(os.prioridade)}>
                        {rotuloPrioridadeOS(os.prioridade)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={varianteStatusOS(os.status)}>
                        {rotuloStatusOS(os.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {os.tecnico?.nome ?? "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatarData(os.data_abertura)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {total > TAMANHO && (
        <div className="mt-4 flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {total} ordem(ns) · página {pagina} de {totalPaginas}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={pagina <= 1 || carregando}
              onClick={() => setPagina((p) => p - 1)}
            >
              Anterior
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={pagina >= totalPaginas || carregando}
              onClick={() => setPagina((p) => p + 1)}
            >
              {carregando ? <Spinner className="h-4 w-4" /> : "Próxima"}
            </Button>
          </div>
        </div>
      )}

      <OrdemServicoForm
        aberto={formAberto}
        ordem={null}
        itens={itens}
        setores={setores}
        servidores={servidores}
        tecnicos={tecnicos}
        usuarioAtual={usuario}
        onFechar={() => setFormAberto(false)}
        onSalvo={() => {
          setFormAberto(false);
          setPagina(1);
          carregar();
        }}
      />
    </div>
  );
}

function FiltroBotao({
  ativo,
  onClick,
  children,
}: {
  ativo: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
        ativo
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border bg-card text-foreground hover:bg-accent"
      )}
    >
      {children}
    </button>
  );
}
