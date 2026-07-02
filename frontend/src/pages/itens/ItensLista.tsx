import * as React from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, Pencil, Eye, AlertTriangle, X, Trash2, Upload } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { EstadoVazio } from "@/components/EstadoVazio";
import { EstadoBadge } from "@/components/EstadoBadge";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { ItemForm } from "./ItemForm";
import { ImportarItensDialog } from "@/components/ImportarItensDialog";
import { itensApi, type FiltroItens } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import { ESTADOS_CONSERVACAO } from "@/lib/rotulos";
import type { Item } from "@/types";

const TODOS = "todos";
const TAMANHO = 20;

export default function ItensLista() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const { ehAdministrador } = useAuth();
  const { categorias, setores, servidores } = useReferencias();

  const [itens, setItens] = React.useState<Item[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [carregando, setCarregando] = React.useState(true);

  // Filtros.
  const [texto, setTexto] = React.useState("");
  const [busca, setBusca] = React.useState("");
  const [categoriaId, setCategoriaId] = React.useState<string>(TODOS);
  const [setorId, setSetorId] = React.useState<string>(TODOS);
  const [estado, setEstado] = React.useState<string>(TODOS);

  const [formAberto, setFormAberto] = React.useState(false);
  const [importarAberto, setImportarAberto] = React.useState(false);
  const [editando, setEditando] = React.useState<Item | null>(null);

  // Exclusão (admin): correção de cadastro errado.
  const [excluindoAlvo, setExcluindoAlvo] = React.useState<Item | null>(null);
  const [excluindo, setExcluindo] = React.useState(false);

  // Debounce da busca textual.
  React.useEffect(() => {
    const t = setTimeout(() => {
      setBusca(texto);
      setPagina(1);
    }, 350);
    return () => clearTimeout(t);
  }, [texto]);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const filtro: FiltroItens = {
      q: busca || undefined,
      categoria_id: categoriaId !== TODOS ? Number(categoriaId) : undefined,
      setor_id: setorId !== TODOS ? Number(setorId) : undefined,
      estado: estado !== TODOS ? estado : undefined,
      pagina,
      tamanho: TAMANHO,
    };
    itensApi
      .listar(filtro)
      .then((r) => {
        setItens(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar itens",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [busca, categoriaId, setorId, estado, pagina, toast]);

  React.useEffect(carregar, [carregar]);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));
  const temFiltro =
    busca || categoriaId !== TODOS || setorId !== TODOS || estado !== TODOS;

  function limparFiltros() {
    setTexto("");
    setBusca("");
    setCategoriaId(TODOS);
    setSetorId(TODOS);
    setEstado(TODOS);
    setPagina(1);
  }

  function abrirCriacao() {
    setEditando(null);
    setFormAberto(true);
  }

  function abrirEdicao(item: Item, e: React.MouseEvent) {
    e.stopPropagation();
    setEditando(item);
    setFormAberto(true);
  }

  function abrirExclusao(item: Item, e: React.MouseEvent) {
    e.stopPropagation();
    setExcluindoAlvo(item);
  }

  async function confirmarExclusao() {
    if (!excluindoAlvo) return;
    setExcluindo(true);
    try {
      await itensApi.excluir(excluindoAlvo.id);
      toast({ titulo: "Item excluído.", variant: "success" });
      setExcluindoAlvo(null);
      // Se a página ficou vazia após excluir, recua uma página.
      if (itens.length === 1 && pagina > 1) {
        setPagina((p) => p - 1);
      } else {
        carregar();
      }
    } catch (err) {
      toast({
        titulo: "Não foi possível excluir o item",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setExcluindo(false);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Itens"
        descricao="Cadastro de equipamentos patrimoniados e materiais de consumo."
        acao={
          <>
            <Button variant="outline" onClick={() => setImportarAberto(true)}>
              <Upload className="h-4 w-4" /> Importar CSV
            </Button>
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> Novo item
            </Button>
          </>
        }
      />

      <Card className="mb-4">
        <CardContent className="flex flex-col gap-3 p-4 lg:flex-row lg:items-center">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="pl-9"
              placeholder="Buscar por descrição, patrimônio, série, marca..."
              value={texto}
              onChange={(e) => setTexto(e.target.value)}
            />
          </div>

          <Select
            value={categoriaId}
            onValueChange={(v) => {
              setCategoriaId(v);
              setPagina(1);
            }}
          >
            <SelectTrigger className="lg:w-44">
              <SelectValue placeholder="Categoria" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={TODOS}>Todas categorias</SelectItem>
              {categorias.map((c) => (
                <SelectItem key={c.id} value={String(c.id)}>
                  {c.nome}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select
            value={setorId}
            onValueChange={(v) => {
              setSetorId(v);
              setPagina(1);
            }}
          >
            <SelectTrigger className="lg:w-44">
              <SelectValue placeholder="Departamento" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={TODOS}>Todos setores</SelectItem>
              {setores.map((s) => (
                <SelectItem key={s.id} value={String(s.id)}>
                  {s.nome}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select
            value={estado}
            onValueChange={(v) => {
              setEstado(v);
              setPagina(1);
            }}
          >
            <SelectTrigger className="lg:w-36">
              <SelectValue placeholder="Estado" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={TODOS}>Todos estados</SelectItem>
              {ESTADOS_CONSERVACAO.map((e) => (
                <SelectItem key={e.valor} value={e.valor}>
                  {e.rotulo}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {temFiltro && (
            <Button variant="ghost" size="sm" onClick={limparFiltros}>
              <X className="h-4 w-4" /> Limpar
            </Button>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : itens.length === 0 ? (
            <EstadoVazio
              titulo="Nenhum item encontrado"
              descricao={
                temFiltro
                  ? "Ajuste os filtros ou limpe a busca."
                  : "Cadastre o primeiro item."
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Descrição</TableHead>
                  <TableHead>Categoria</TableHead>
                  <TableHead>Patrimônio</TableHead>
                  <TableHead>Departamento</TableHead>
                  <TableHead className="text-center">Qtd.</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead className="w-24 text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {itens.map((item) => {
                  const alerta =
                    item.estoque_minimo > 0 &&
                    item.quantidade < item.estoque_minimo;
                  return (
                    <TableRow
                      key={item.id}
                      className="cursor-pointer"
                      onClick={() => navigate(`/itens/${item.id}`)}
                    >
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          {item.descricao}
                          {item.baixado && (
                            <Badge variant="muted">Baixado</Badge>
                          )}
                        </div>
                        {(item.marca || item.modelo) && (
                          <p className="text-xs text-muted-foreground">
                            {[item.marca, item.modelo]
                              .filter(Boolean)
                              .join(" · ")}
                          </p>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.categoria?.nome ?? "—"}
                      </TableCell>
                      <TableCell>{item.numero_patrimonio ?? "—"}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.setor?.nome ?? "—"}
                      </TableCell>
                      <TableCell className="text-center">
                        <span
                          className={
                            alerta
                              ? "inline-flex items-center gap-1 font-medium text-amber-600"
                              : ""
                          }
                        >
                          {alerta && <AlertTriangle className="h-3.5 w-3.5" />}
                          {item.quantidade}
                        </span>
                      </TableCell>
                      <TableCell>
                        <EstadoBadge estado={item.estado_conservacao} />
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={(e) => {
                              e.stopPropagation();
                              navigate(`/itens/${item.id}`);
                            }}
                            aria-label="Ver detalhes"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={(e) => abrirEdicao(item, e)}
                            aria-label="Editar"
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          {ehAdministrador && (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive hover:text-destructive"
                              onClick={(e) => abrirExclusao(item, e)}
                              aria-label="Excluir"
                              title="Excluir item (correção de cadastro)"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {total > 0 && (
        <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
          <span>
            {total} {total === 1 ? "item" : "itens"} · página {pagina} de{" "}
            {totalPaginas}
          </span>
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

      <ItemForm
        aberto={formAberto}
        item={editando}
        categorias={categorias}
        setores={setores}
        servidores={servidores}
        onFechar={() => setFormAberto(false)}
        onSalvo={() => {
          setFormAberto(false);
          carregar();
        }}
      />

      <ImportarItensDialog
        aberto={importarAberto}
        onFechar={() => setImportarAberto(false)}
        onConcluido={carregar}
      />

      <ConfirmDialog
        aberto={!!excluindoAlvo}
        titulo="Excluir item?"
        descricao={
          `"${excluindoAlvo?.descricao}" será removido do sistema. ` +
          "Use esta opção apenas para corrigir um cadastro feito por engano. " +
          "Itens com histórico de movimentações ou termos não podem ser excluídos — nesse caso, use a baixa patrimonial."
        }
        textoConfirmar="Excluir"
        destrutivo
        processando={excluindo}
        onConfirmar={confirmarExclusao}
        onCancelar={() => setExcluindoAlvo(null)}
      />
    </div>
  );
}
