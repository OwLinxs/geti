import * as React from "react";
import { FileText, FileSpreadsheet } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { ordenarComoArvore, prefixoIndentacao } from "@/lib/setores";
import { EstadoBadge } from "@/components/EstadoBadge";
import { EstadoVazio } from "@/components/EstadoVazio";
import { HistoricoTabela } from "@/components/HistoricoTabela";
import { CarregandoTela } from "@/components/ui/spinner";
import {
  relatoriosApi,
  type DefinicaoRelatorio,
  type ParamsRelatorio,
  type RelatorioTipo,
  type ResultadoRelatorio,
} from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useReferencias } from "@/hooks/useReferencias";
import { baixarBlob, formatarData } from "@/lib/format";
import { codigoInterno } from "@/lib/patrimonio";
import type { Item } from "@/types";

const TODOS = "todos";

const ESTADOS: { valor: string; rotulo: string }[] = [
  { valor: "novo", rotulo: "Novo" },
  { valor: "bom", rotulo: "Bom" },
  { valor: "regular", rotulo: "Regular" },
  { valor: "inservivel", rotulo: "Inservível" },
];

export default function Relatorios() {
  const { toast } = useToast();
  const { setores, servidores, categorias } = useReferencias();

  const [catalogo, setCatalogo] = React.useState<DefinicaoRelatorio[]>([]);
  const [tipo, setTipo] = React.useState<RelatorioTipo>("inventario");

  const [resultado, setResultado] = React.useState<ResultadoRelatorio | null>(
    null
  );
  const [carregando, setCarregando] = React.useState(true);
  const [baixando, setBaixando] = React.useState(false);

  // Filtros do construtor.
  const [secretariaId, setSecretariaId] = React.useState(TODOS);
  const [departamentoId, setDepartamentoId] = React.useState(TODOS);
  const [responsavelId, setResponsavelId] = React.useState(TODOS);
  const [categoriaId, setCategoriaId] = React.useState(TODOS);
  const [estado, setEstado] = React.useState(TODOS);
  const [dataInicio, setDataInicio] = React.useState("");
  const [dataFim, setDataFim] = React.useState("");

  const def = catalogo.find((d) => d.tipo === tipo);

  // Secretarias = setores de topo (sem pai).
  const secretarias = React.useMemo(
    () => setores.filter((s) => !s.pai_id),
    [setores]
  );

  // Carrega o catálogo de tipos uma vez.
  React.useEffect(() => {
    relatoriosApi
      .catalogo()
      .then((c) => {
        setCatalogo(c);
        if (c.length && !c.some((d) => d.tipo === tipo)) setTipo(c[0].tipo);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar tipos de relatório",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Monta os parâmetros respeitando os filtros aplicáveis ao tipo atual.
  const montarParams = React.useCallback((): ParamsRelatorio => {
    const p: ParamsRelatorio = { tipo };
    if (def?.usa_setor) {
      // Departamento escolhido = filtro exato; senão a secretaria pega tudo abaixo.
      if (departamentoId !== TODOS) p.setor_id = Number(departamentoId);
      else if (secretariaId !== TODOS) p.secretaria_id = Number(secretariaId);
    }
    if (def?.usa_responsavel && responsavelId !== TODOS)
      p.responsavel_id = Number(responsavelId);
    if (def?.usa_categoria && categoriaId !== TODOS)
      p.categoria_id = Number(categoriaId);
    if (def?.usa_estado && estado !== TODOS) p.estado = estado;
    if (def?.usa_periodo) {
      if (dataInicio) p.data_inicio = dataInicio;
      if (dataFim) p.data_fim = dataFim;
    }
    return p;
  }, [
    tipo,
    def,
    secretariaId,
    departamentoId,
    responsavelId,
    categoriaId,
    estado,
    dataInicio,
    dataFim,
  ]);

  const carregar = React.useCallback(() => {
    if (!def) return;
    setCarregando(true);
    relatoriosApi
      .gerar(montarParams())
      .then(setResultado)
      .catch((err) =>
        toast({
          titulo: "Erro ao gerar relatório",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [def, montarParams, toast]);

  React.useEffect(carregar, [carregar]);

  async function exportar(formato: "csv" | "pdf") {
    setBaixando(true);
    try {
      const { tipo: _t, ...resto } = montarParams();
      const blob = await relatoriosApi.baixar(tipo, formato, resto);
      baixarBlob(blob, `${tipo}.${formato}`);
    } catch (err) {
      toast({
        titulo: "Erro ao exportar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setBaixando(false);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Relatórios"
        descricao="Monte o relatório: escolha o tipo, aplique os filtros e exporte."
        acao={
          <>
            <Button
              variant="outline"
              onClick={() => exportar("csv")}
              disabled={baixando || carregando}
            >
              <FileSpreadsheet className="h-4 w-4" /> CSV
            </Button>
            {def?.permite_pdf && (
              <Button
                variant="outline"
                onClick={() => exportar("pdf")}
                disabled={baixando || carregando}
              >
                <FileText className="h-4 w-4" /> PDF
              </Button>
            )}
          </>
        }
      />

      {/* Seletor de tipo de relatório. */}
      <div className="mb-4 flex flex-wrap gap-2">
        {catalogo.map((d) => (
          <button
            key={d.tipo}
            onClick={() => setTipo(d.tipo)}
            title={d.descricao}
            className={
              "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors " +
              (tipo === d.tipo
                ? "border-primary bg-primary text-primary-foreground"
                : "border-border bg-card text-foreground hover:bg-accent")
            }
          >
            {d.rotulo}
          </button>
        ))}
      </div>

      {/* Filtros do tipo selecionado. */}
      {def && (
        <Card className="mb-4">
          <CardContent className="flex flex-col gap-3 p-4 sm:flex-row sm:flex-wrap sm:items-end">
            {def.descricao && (
              <p className="w-full text-sm text-muted-foreground">
                {def.descricao}
              </p>
            )}

            {def.usa_setor && (
              <>
                <FormField label="Secretaria" className="sm:w-56">
                  <Select
                    value={secretariaId}
                    onValueChange={(v) => {
                      setSecretariaId(v);
                      setDepartamentoId(TODOS);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={TODOS}>Todas as secretarias</SelectItem>
                      {secretarias.map((s) => (
                        <SelectItem key={s.id} value={String(s.id)}>
                          {s.nome}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label="Departamento" className="sm:w-56">
                  <Select value={departamentoId} onValueChange={setDepartamentoId}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={TODOS}>Todos</SelectItem>
                      {ordenarComoArvore(setores).map(({ setor: s, nivel }) => (
                        <SelectItem key={s.id} value={String(s.id)}>
                          {prefixoIndentacao(nivel)}
                          {s.nome}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
              </>
            )}

            {def.usa_responsavel && (
              <FormField label="Responsável" className="sm:w-56">
                <Select value={responsavelId} onValueChange={setResponsavelId}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={TODOS}>Todos</SelectItem>
                    {servidores.map((s) => (
                      <SelectItem key={s.id} value={String(s.id)}>
                        {s.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}

            {def.usa_categoria && (
              <FormField label="Categoria" className="sm:w-48">
                <Select value={categoriaId} onValueChange={setCategoriaId}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={TODOS}>Todas</SelectItem>
                    {categorias.map((cat) => (
                      <SelectItem key={cat.id} value={String(cat.id)}>
                        {cat.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}

            {def.usa_estado && (
              <FormField label="Estado" className="sm:w-40">
                <Select value={estado} onValueChange={setEstado}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={TODOS}>Todos</SelectItem>
                    {ESTADOS.map((e) => (
                      <SelectItem key={e.valor} value={e.valor}>
                        {e.rotulo}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}

            {def.usa_periodo && (
              <>
                <FormField label="De" className="sm:w-40">
                  <input
                    type="date"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={dataInicio}
                    onChange={(e) => setDataInicio(e.target.value)}
                  />
                </FormField>
                <FormField label="Até" className="sm:w-40">
                  <input
                    type="date"
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={dataFim}
                    onChange={(e) => setDataFim(e.target.value)}
                  />
                </FormField>
              </>
            )}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            {def?.rotulo ?? "Relatório"}
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : def?.eh_evento ? (
            (resultado?.movimentacoes?.length ?? 0) === 0 ? (
              <EstadoVazio titulo="Sem resultados para este relatório" />
            ) : (
              <HistoricoTabela
                movimentacoes={resultado?.movimentacoes ?? []}
                mostrarItem
              />
            )
          ) : (resultado?.itens?.length ?? 0) === 0 ? (
            <EstadoVazio titulo="Sem resultados para este relatório" />
          ) : (
            <TabelaItens itens={resultado?.itens ?? []} mostrarBaixa={tipo === "baixados"} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function TabelaItens({
  itens,
  mostrarBaixa,
}: {
  itens: Item[];
  mostrarBaixa?: boolean;
}) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Código TI</TableHead>
          <TableHead>Descrição</TableHead>
          <TableHead>Categoria</TableHead>
          <TableHead>Patrimônio</TableHead>
          <TableHead>Departamento</TableHead>
          <TableHead>Responsável</TableHead>
          <TableHead className="text-center">Qtd.</TableHead>
          <TableHead>Estado</TableHead>
          {mostrarBaixa ? (
            <>
              <TableHead>Data baixa</TableHead>
              <TableHead>Motivo</TableHead>
            </>
          ) : (
            <TableHead>Aquisição</TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {itens.map((i) => (
          <TableRow key={i.id}>
            <TableCell className="font-mono text-xs">
              {codigoInterno(i.id)}
            </TableCell>
            <TableCell className="font-medium">{i.descricao}</TableCell>
            <TableCell className="text-muted-foreground">
              {i.categoria?.nome ?? "—"}
            </TableCell>
            <TableCell>
              {i.categoria?.consumivel
                ? "—"
                : (i.numero_patrimonio ?? "Pendente")}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {i.setor?.nome ?? "—"}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {i.responsavel?.nome ?? "—"}
            </TableCell>
            <TableCell className="text-center">{i.quantidade}</TableCell>
            <TableCell>
              <EstadoBadge estado={i.estado_conservacao} />
            </TableCell>
            {mostrarBaixa ? (
              <>
                <TableCell className="text-muted-foreground">
                  {formatarData(i.data_baixa)}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {i.motivo_baixa || "—"}
                </TableCell>
              </>
            ) : (
              <TableCell className="text-muted-foreground">
                {formatarData(i.data_aquisicao)}
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
