import * as React from "react";
import { Download, Upload, FileWarning, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Spinner } from "@/components/ui/spinner";
import { itensApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { baixarBlob } from "@/lib/format";
import type { ResultadoImportacao } from "@/types";

interface ImportarItensDialogProps {
  aberto: boolean;
  onFechar: () => void;
  onConcluido: () => void; // recarrega a lista após importar
}

// Importação em massa de itens via CSV, em duas etapas:
// 1) o usuário escolhe o arquivo → simulação (dry-run) mostra o que será
//    importado e os erros por linha;
// 2) confirma a gravação apenas das linhas válidas.
export function ImportarItensDialog({
  aberto,
  onFechar,
  onConcluido,
}: ImportarItensDialogProps) {
  const { toast } = useToast();
  const [arquivo, setArquivo] = React.useState<File | null>(null);
  const [previa, setPrevia] = React.useState<ResultadoImportacao | null>(null);
  const [processando, setProcessando] = React.useState(false);
  const [importado, setImportado] = React.useState<ResultadoImportacao | null>(null);
  const inputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    if (aberto) {
      setArquivo(null);
      setPrevia(null);
      setImportado(null);
      setProcessando(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }, [aberto]);

  async function baixarModelo() {
    try {
      const blob = await itensApi.baixarModeloCsv();
      baixarBlob(blob, "modelo-itens.csv");
    } catch (err) {
      toast({ titulo: "Erro ao baixar modelo", descricao: mensagemErro(err), variant: "destructive" });
    }
  }

  async function aoEscolherArquivo(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0] ?? null;
    setArquivo(f);
    setPrevia(null);
    setImportado(null);
    if (!f) return;

    setProcessando(true);
    try {
      const res = await itensApi.importar(f, true); // dry-run
      setPrevia(res);
    } catch (err) {
      toast({ titulo: "Não foi possível ler o arquivo", descricao: mensagemErro(err), variant: "destructive" });
    } finally {
      setProcessando(false);
    }
  }

  async function confirmarImportacao() {
    if (!arquivo) return;
    setProcessando(true);
    try {
      const res = await itensApi.importar(arquivo, false);
      setImportado(res);
      toast({
        titulo: `${res.importados} item(ns) importado(s).`,
        descricao: res.erros.length ? `${res.erros.length} linha(s) ignorada(s) por erro.` : undefined,
        variant: res.importados > 0 ? "success" : "default",
      });
      onConcluido();
    } catch (err) {
      toast({ titulo: "Falha na importação", descricao: mensagemErro(err), variant: "destructive" });
    } finally {
      setProcessando(false);
    }
  }

  const resultado = importado ?? previa;

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Importar itens via CSV</DialogTitle>
          <DialogDescription>
            Cadastre vários itens de uma vez. Baixe o modelo, preencha e envie.
            Categoria e setor são casados pelo nome; o responsável, pela matrícula.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Button type="button" variant="outline" onClick={baixarModelo}>
              <Download className="h-4 w-4" /> Baixar modelo
            </Button>
            <input
              ref={inputRef}
              type="file"
              accept=".csv,text/csv"
              onChange={aoEscolherArquivo}
              disabled={processando}
              className="text-sm file:mr-3 file:rounded-md file:border-0 file:bg-primary file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary-foreground hover:file:bg-primary/90"
            />
          </div>

          {processando && !resultado && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Spinner className="h-4 w-4" /> Analisando arquivo…
            </div>
          )}

          {resultado && (
            <div className="space-y-3">
              <div className="flex flex-wrap gap-4 rounded-md border border-border bg-muted/30 p-3 text-sm">
                <span><strong>{resultado.total}</strong> linha(s) lida(s)</span>
                <span className="text-emerald-600">
                  <strong>{importado ? importado.importados : resultado.validas}</strong>{" "}
                  {importado ? "importada(s)" : "válida(s)"}
                </span>
                <span className={resultado.erros.length ? "text-destructive" : ""}>
                  <strong>{resultado.erros.length}</strong> com erro
                </span>
              </div>

              {importado && (
                <div className="flex items-center gap-2 rounded-md border border-emerald-300 bg-emerald-50 p-3 text-sm text-emerald-700">
                  <CheckCircle2 className="h-4 w-4" />
                  Importação concluída. As linhas com erro (se houver) foram ignoradas.
                </div>
              )}

              {resultado.erros.length > 0 && (
                <div className="max-h-64 overflow-auto rounded-md border border-border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-16">Linha</TableHead>
                        <TableHead>Conteúdo</TableHead>
                        <TableHead>Problema</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {resultado.erros.map((e) => (
                        <TableRow key={e.linha}>
                          <TableCell className="font-mono">{e.linha}</TableCell>
                          <TableCell className="max-w-[200px] truncate text-muted-foreground">
                            {e.descricao || "—"}
                          </TableCell>
                          <TableCell className="text-destructive">
                            <span className="flex items-start gap-1">
                              <FileWarning className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                              {e.mensagem}
                            </span>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onFechar}>
            {importado ? "Fechar" : "Cancelar"}
          </Button>
          {!importado && (
            <Button
              type="button"
              onClick={confirmarImportacao}
              disabled={processando || !previa || previa.validas === 0}
            >
              {processando ? (
                <Spinner className="h-4 w-4" />
              ) : (
                <>
                  <Upload className="h-4 w-4" /> Importar{" "}
                  {previa ? `${previa.validas} item(ns)` : ""}
                </>
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
