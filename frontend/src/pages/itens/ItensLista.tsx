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
import { CadastrarLoteDialog } from "@/components/CadastrarLoteDialog";
import { itensApi, type FiltroItens } from "@/services/api";
import { codigoInterno } from "@/lib/patrimonio";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import { ESTADOS_CONSERVACAO } from "@/lib/rotulos";
import type { Item } from "@/types";

const TODOS = "todos";
const TAMANHO = 20;

export default function ItensLista({ tipo }: { tipo: "consumo" | "patrimonio" }) {
  const ehPatrimonio = tipo === "patrimonio";
  const rotaBase = ehPatrimonio ? "/patrimonio" : "/itens";
  const { toast } = useToast();
  const navigate = useNavigate();
  const { ehAdministrador } = useAuth();
  const { categorias, setores, servidores } = useReferencias();
  const categoriasDaAba = React.useMemo(
    () => categorias.filter((c) => c.consumivel !== ehPatrimonio),
    [categorias, ehPatrimonio]
  );

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
  const [lote, setLote] = React.useState("");
  const [pendencia, setPendencia] = React.useState<"" | "patrimonio" | "local">("");

  const [formAberto, setFormAberto] = React.useState(false);
  const [importarAberto, setImportarAberto] = React.useState(false);
  const [loteAberto, setLoteAberto] = React.useState(false);
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
      tipo,
      q: busca || undefined,
      categoria_id: categoriaId !== TODOS ? Number(categoriaId) : undefined,
      setor_id: setorId !== TODOS ? Number(setorId) : undefined,
      estado: estado !== TODOS ? estado : undefined,
      lote: lote || undefined,
      sem_patrimonio: pendencia === "patrimonio" ? true : undefined,
      sem_setor: pendencia === "local" ? true : undefined,
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
  }, [busca, categoriaId, setorId, estado, lote, pendencia, pagina, tipo, toast]);

  React.useEffect(carregar, [carregar]);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));
  const temFiltro =
    busca || categoriaId !== TODOS || setorId !== TODOS || estado !== TODOS || lote || pendencia;

  function limparFiltros() {
    setTexto("");
    setBusca("");
    setCategoriaId(TODOS);
    setSetorId(TODOS);
    setEstado(TODOS);
    setLote("");
    setPendencia("");
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
        titulo={ehPatrimonio ? "Patrimônio" : "Itens"}
        descricao={ehPatrimonio
          ? "Bens controlados individualmente, com patrimônio e localização definidos ao longo do tempo."
          : "Materiais de consumo controlados por quantidade e estoque mínimo."}
        acao={
          <>
            {!ehPatrimonio && <Button variant="outline" onClick={() => setImportarAberto(true)}>
              <Upload className="h-4 w-4" /> Importar CSV
            </Button>}
            {ehPatrimonio && <Button variant="outline" onClick={() => setLoteAberto(true)}>
              <Plus className="h-4 w-4" /> Receber lote
            </Button>}
            <Button onClick={abrirCriacao}>
              <Plus className="h-4 w-4" /> {ehPatrimonio ? "Novo bem" : "Novo item"}
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
              placeholder={ehPatrimonio ? "Buscar por descrição, código TI, patrimônio, série..." : "Buscar material por descrição..."}
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
              {categoriasDaAba.map((c) => (
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

      {ehPatrimonio && <div className="mb-4 flex flex-wrap items-center gap-2">
        <Button size="sm" variant={pendencia === "patrimonio" ? "default" : "outline"} onClick={() => { setPendencia(pendencia === "patrimonio" ? "" : "patrimonio"); setPagina(1); }}>
          Sem patrimônio
        </Button>
        <Button size="sm" variant={pendencia === "local" ? "default" : "outline"} onClick={() => { setPendencia(pendencia === "local" ? "" : "local"); setPagina(1); }}>
          Sem local definido
        </Button>
        {lote && <Badge variant="secondary">Lote {lote} <button className="ml-2" onClick={() => { setLote(""); setPagina(1); }} aria-label="Limpar lote">×</button></Badge>}
      </div>}

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : itens.length === 0 ? (
            <EstadoVazio
              titulo={ehPatrimonio ? "Nenhum bem encontrado" : "Nenhum item encontrado"}
              descricao={
                temFiltro
                  ? "Ajuste os filtros ou limpe a busca."
                  : ehPatrimonio ? "Cadastre o primeiro bem." : "Cadastre o primeiro item."
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Descrição</TableHead>
                  <TableHead>Categoria</TableHead>
                  {ehPatrimonio && <TableHead>Patrimônio</TableHead>}
                  <TableHead>Departamento</TableHead>
                  {!ehPatrimonio && <TableHead className="text-center">Qtd.</TableHead>}
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
                      onClick={() => navigate(`${rotaBase}/${item.id}`)}
                    >
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          {item.descricao}
                          {!item.categoria?.consumivel && !item.numero_patrimonio && !item.baixado && (
                            <Badge variant="warning">Aguardando patrimônio</Badge>
                          )}
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
                        {ehPatrimonio && <p className="text-xs font-medium text-muted-foreground">{codigoInterno(item.id)}</p>}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.categoria?.nome ?? "—"}
                      </TableCell>
                      {ehPatrimonio && <TableCell>{item.numero_patrimonio ?? "Pendente"}</TableCell>}
                      <TableCell className="text-muted-foreground">
                        {item.setor?.nome ?? (ehPatrimonio ? "Local pendente" : "—")}
                      </TableCell>
                      {!ehPatrimonio && <TableCell className="text-center">
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
                      </TableCell>}
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
                              navigate(`${rotaBase}/${item.id}`);
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
            {total} {ehPatrimonio ? (total === 1 ? "bem" : "bens") : (total === 1 ? "item" : "itens")} · página {pagina} de{" "}
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
        categorias={categoriasDaAba}
        setores={setores}
        servidores={servidores}
        onFechar={() => setFormAberto(false)}
        onSalvo={() => {
          setFormAberto(false);
          carregar();
        }}
      />

      {!ehPatrimonio && <ImportarItensDialog
        aberto={importarAberto}
        onFechar={() => setImportarAberto(false)}
        onConcluido={carregar}
      />}
      {ehPatrimonio && <CadastrarLoteDialog
        aberto={loteAberto}
        categorias={categoriasDaAba}
        setores={setores}
        onFechar={() => setLoteAberto(false)}
        onCriado={(res) => {
          setTexto(""); setBusca(""); setCategoriaId(TODOS); setSetorId(TODOS);
          setEstado(TODOS); setPendencia(""); setLote(res.lote_codigo); setPagina(1);
        }}
      />}

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
