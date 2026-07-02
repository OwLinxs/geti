import * as React from "react";
import { X } from "lucide-react";
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
import { FormField } from "@/components/FormField";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { auditoriaApi, type FiltroAuditoria } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { formatarDataHora } from "@/lib/format";
import type { RegistroAuditoria } from "@/types";

const TODOS = "todos";
const TAMANHO = 30;

const RECURSOS = [
  "item",
  "usuario",
  "servidor",
  "categoria",
  "setor",
  "movimentacao",
  "termo",
];

const ACOES = [
  "login",
  "criou",
  "atualizou",
  "alterou",
  "excluiu",
  "importou",
  "redefiniu senha",
  "alterou acesso",
];

export default function Auditoria() {
  const { toast } = useToast();
  const [regs, setRegs] = React.useState<RegistroAuditoria[]>([]);
  const [total, setTotal] = React.useState(0);
  const [pagina, setPagina] = React.useState(1);
  const [carregando, setCarregando] = React.useState(true);

  const [recurso, setRecurso] = React.useState<string>(TODOS);
  const [acao, setAcao] = React.useState<string>(TODOS);
  const [de, setDe] = React.useState("");
  const [ate, setAte] = React.useState("");

  const carregar = React.useCallback(() => {
    setCarregando(true);
    const filtro: FiltroAuditoria = {
      recurso: recurso !== TODOS ? recurso : undefined,
      acao: acao !== TODOS ? acao : undefined,
      de: de || undefined,
      ate: ate || undefined,
      pagina,
      tamanho: TAMANHO,
    };
    auditoriaApi
      .listar(filtro)
      .then((r) => {
        setRegs(r.dados);
        setTotal(r.total);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar auditoria",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [recurso, acao, de, ate, pagina, toast]);

  React.useEffect(carregar, [carregar]);

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANHO));
  const temFiltro = recurso !== TODOS || acao !== TODOS || de || ate;

  function limpar() {
    setRecurso(TODOS);
    setAcao(TODOS);
    setDe("");
    setAte("");
    setPagina(1);
  }

  return (
    <div>
      <PageHeader
        titulo="Auditoria"
        descricao="Trilha de quem fez o quê e quando — logins e operações de escrita."
      />

      <Card className="mb-4">
        <CardContent className="flex flex-col gap-3 p-4 sm:flex-row sm:flex-wrap sm:items-end">
          <FormField label="Recurso" className="sm:w-48">
            <Select
              value={recurso}
              onValueChange={(v) => {
                setRecurso(v);
                setPagina(1);
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={TODOS}>Todos</SelectItem>
                {RECURSOS.map((r) => (
                  <SelectItem key={r} value={r}>
                    {r}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <FormField label="Ação" className="sm:w-48">
            <Select
              value={acao}
              onValueChange={(v) => {
                setAcao(v);
                setPagina(1);
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={TODOS}>Todas</SelectItem>
                {ACOES.map((a) => (
                  <SelectItem key={a} value={a}>
                    {a}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <FormField label="De" className="sm:w-40">
            <Input
              type="date"
              value={de}
              onChange={(e) => {
                setDe(e.target.value);
                setPagina(1);
              }}
            />
          </FormField>

          <FormField label="Até" className="sm:w-40">
            <Input
              type="date"
              value={ate}
              onChange={(e) => {
                setAte(e.target.value);
                setPagina(1);
              }}
            />
          </FormField>

          {temFiltro && (
            <Button variant="ghost" size="sm" onClick={limpar}>
              <X className="h-4 w-4" /> Limpar
            </Button>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : regs.length === 0 ? (
            <EstadoVazio
              titulo="Nenhum registro"
              descricao="Não há eventos de auditoria para os filtros selecionados."
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-44">Data/Hora</TableHead>
                  <TableHead>Usuário</TableHead>
                  <TableHead>Ação</TableHead>
                  <TableHead>Recurso</TableHead>
                  <TableHead>Detalhe</TableHead>
                  <TableHead className="w-28">IP</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {regs.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell className="whitespace-nowrap text-sm text-muted-foreground">
                      {formatarDataHora(r.criado_em)}
                    </TableCell>
                    <TableCell className="text-sm">
                      {r.usuario_nome || r.usuario_email || "—"}
                    </TableCell>
                    <TableCell>
                      <Badge variant={corAcao(r)}>{r.acao}</Badge>
                    </TableCell>
                    <TableCell className="text-sm">
                      {r.recurso || "—"}
                      {r.recurso_id ? ` #${r.recurso_id}` : ""}
                    </TableCell>
                    <TableCell className="max-w-[280px] truncate text-sm text-muted-foreground">
                      {r.detalhe || `${r.metodo ?? ""} ${r.caminho ?? ""}`.trim()}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {r.ip || "—"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <div className="mt-3 flex items-center justify-between text-sm text-muted-foreground">
        <span>
          {total} evento(s) · página {pagina} de {totalPaginas}
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
    </div>
  );
}

// corAcao dá um leve destaque visual: falhas de login em vermelho, exclusões
// em vermelho, o resto neutro.
function corAcao(r: RegistroAuditoria): "default" | "destructive" | "secondary" {
  if (r.acao === "login" && r.status !== 200) return "destructive";
  if (r.acao === "excluiu") return "destructive";
  return "secondary";
}
