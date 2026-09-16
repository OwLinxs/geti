import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Pencil, Printer, Plus, Trash2 } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { OrdemServicoForm } from "./OrdemServicoForm";
import {
  itensApi,
  ordensServicoApi,
  usuariosApi,
} from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { useReferencias } from "@/hooks/useReferencias";
import { baixarBlob, formatarData } from "@/lib/format";
import {
  STATUS_OS,
  rotuloPrioridadeOS,
  variantePrioridadeOS,
} from "@/lib/rotulos";
import type { Item, OrdemServico, Usuario } from "@/types";

// Passo em edição no cliente (id opcional: novos passos ainda não persistidos).
interface PassoEdit {
  descricao: string;
  concluido: boolean;
  observacao: string;
}

export default function OrdemServicoDetalhe() {
  const { id } = useParams<{ id: string }>();
  const osId = Number(id);
  const navigate = useNavigate();
  const { toast } = useToast();
  const { ehAdministrador, usuario } = useAuth();
  const { setores, servidores } = useReferencias();

  const [os, setOs] = React.useState<OrdemServico | null>(null);
  const [carregando, setCarregando] = React.useState(true);
  const [passos, setPassos] = React.useState<PassoEdit[]>([]);
  const [salvandoPassos, setSalvandoPassos] = React.useState(false);
  const [mudandoStatus, setMudandoStatus] = React.useState(false);

  const [itens, setItens] = React.useState<Item[]>([]);
  const [tecnicos, setTecnicos] = React.useState<Usuario[]>([]);
  const [editAberto, setEditAberto] = React.useState(false);
  const [imprimirAberto, setImprimirAberto] = React.useState(false);
  const [excluindo, setExcluindo] = React.useState(false);
  const [processandoExcluir, setProcessandoExcluir] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    ordensServicoApi
      .buscarPorId(osId)
      .then((r) => {
        setOs(r);
        setPassos(
          (r.passos ?? []).map((p) => ({
            descricao: p.descricao,
            concluido: p.concluido,
            observacao: p.observacao ?? "",
          }))
        );
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar a OS",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [osId, toast]);

  React.useEffect(carregar, [carregar]);

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

  async function alterarStatus(novo: string) {
    if (!os || novo === os.status) return;
    setMudandoStatus(true);
    try {
      const atualizada = await ordensServicoApi.definirStatus(os.id, novo);
      setOs(atualizada);
      toast({ titulo: "Status atualizado.", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível alterar o status",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setMudandoStatus(false);
    }
  }

  async function salvarChecklist() {
    if (!os) return;
    setSalvandoPassos(true);
    try {
      const atualizada = await ordensServicoApi.salvarPassos(
        os.id,
        passos
          .filter((p) => p.descricao.trim())
          .map((p) => ({
            descricao: p.descricao.trim(),
            concluido: p.concluido,
            observacao: p.observacao.trim(),
          }))
      );
      setOs(atualizada);
      setPassos(
        (atualizada.passos ?? []).map((p) => ({
          descricao: p.descricao,
          concluido: p.concluido,
          observacao: p.observacao ?? "",
        }))
      );
      toast({ titulo: "Checklist salvo.", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível salvar o checklist",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvandoPassos(false);
    }
  }

  function atualizarPasso(idx: number, patch: Partial<PassoEdit>) {
    setPassos((prev) =>
      prev.map((p, i) => (i === idx ? { ...p, ...patch } : p))
    );
  }
  function removerPasso(idx: number) {
    setPassos((prev) => prev.filter((_, i) => i !== idx));
  }
  function adicionarPasso() {
    setPassos((prev) => [
      ...prev,
      { descricao: "", concluido: false, observacao: "" },
    ]);
  }

  async function confirmarExcluir() {
    if (!os) return;
    setProcessandoExcluir(true);
    try {
      await ordensServicoApi.excluir(os.id);
      toast({ titulo: "OS excluída.", variant: "success" });
      navigate("/manutencao");
    } catch (err) {
      toast({
        titulo: "Não foi possível excluir",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
      setProcessandoExcluir(false);
    }
  }

  if (carregando) return <CarregandoTela texto="Carregando OS..." />;
  if (!os)
    return (
      <div className="py-12 text-center text-muted-foreground">
        Ordem de serviço não encontrada.
      </div>
    );

  const podeExcluir =
    ehAdministrador &&
    (os.status === "aberta" || os.status === "cancelada");

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate("/manutencao")}
      >
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Button>

      <PageHeader
        titulo={`OS ${os.numero}`}
        descricao={os.equipamento_snapshot}
        acao={
          <>
            <Button variant="outline" onClick={() => setImprimirAberto(true)}>
              <Printer className="h-4 w-4" /> Imprimir
            </Button>
            <Button variant="outline" onClick={() => setEditAberto(true)}>
              <Pencil className="h-4 w-4" /> Editar
            </Button>
            {podeExcluir && (
              <Button
                variant="outline"
                onClick={() => setExcluindo(true)}
                className="text-destructive"
              >
                <Trash2 className="h-4 w-4" /> Excluir
              </Button>
            )}
          </>
        }
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Dados. */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base">Dados da OS</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <Campo rotulo="Status">
              <div className="flex items-center gap-2">
                <Select
                  value={os.status}
                  onValueChange={alterarStatus}
                  disabled={mudandoStatus}
                >
                  <SelectTrigger className="h-8 w-44">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {STATUS_OS.map((s) => (
                      <SelectItem key={s.valor} value={s.valor}>
                        {s.rotulo}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {mudandoStatus && <Spinner className="h-4 w-4" />}
              </div>
            </Campo>
            <Campo rotulo="Prioridade">
              <Badge variant={variantePrioridadeOS(os.prioridade)}>
                {rotuloPrioridadeOS(os.prioridade)}
              </Badge>
            </Campo>
            <Campo rotulo="Equipamento">{os.equipamento_snapshot}</Campo>
            {os.patrimonio_snapshot && (
              <Campo rotulo="Patrimônio">{os.patrimonio_snapshot}</Campo>
            )}
            {os.equipamento_identificacao && (
              <Campo rotulo="Série / ident.">
                {os.equipamento_identificacao}
              </Campo>
            )}
            {os.item_id ? (
              <Campo rotulo="Origem">Item do inventário</Campo>
            ) : (
              <Campo rotulo="Origem">Máquina externa</Campo>
            )}
            <Campo rotulo="Solicitante">
              {os.solicitante?.nome ?? os.solicitante_nome_snapshot ?? "—"}
            </Campo>
            <Campo rotulo="Departamento">{os.setor?.nome ?? "—"}</Campo>
            <Campo rotulo="Técnico">{os.tecnico?.nome ?? "—"}</Campo>
            <Campo rotulo="Abertura">{formatarData(os.data_abertura)}</Campo>
            {os.data_conclusao && (
              <Campo rotulo="Conclusão">
                {formatarData(os.data_conclusao)}
              </Campo>
            )}
            <Campo rotulo="Aberta por">{os.aberto_por?.nome ?? "—"}</Campo>
          </CardContent>
        </Card>

        {/* Defeito/diagnóstico/solução + checklist. */}
        <div className="space-y-6 lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Atendimento</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <Texto rotulo="Defeito relatado" valor={os.defeito_relatado} />
              <Texto rotulo="Diagnóstico" valor={os.diagnostico} />
              <Texto rotulo="Solução aplicada" valor={os.solucao_aplicada} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-base">
                Checklist de procedimentos
              </CardTitle>
              <Button
                size="sm"
                onClick={salvarChecklist}
                disabled={salvandoPassos}
              >
                {salvandoPassos ? (
                  <Spinner className="h-4 w-4" />
                ) : (
                  "Salvar checklist"
                )}
              </Button>
            </CardHeader>
            <CardContent className="space-y-2">
              {passos.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  Nenhum passo. Adicione procedimentos abaixo.
                </p>
              )}
              {passos.map((p, idx) => (
                <div
                  key={idx}
                  className="flex items-start gap-2 rounded-md border border-border p-2"
                >
                  <input
                    type="checkbox"
                    className="mt-2.5 h-4 w-4 shrink-0"
                    checked={p.concluido}
                    onChange={(e) =>
                      atualizarPasso(idx, { concluido: e.target.checked })
                    }
                  />
                  <div className="flex-1 space-y-1">
                    <Input
                      value={p.descricao}
                      onChange={(e) =>
                        atualizarPasso(idx, { descricao: e.target.value })
                      }
                      placeholder="Descrição do procedimento"
                    />
                    <Input
                      value={p.observacao}
                      onChange={(e) =>
                        atualizarPasso(idx, { observacao: e.target.value })
                      }
                      placeholder="Observação (opcional)"
                      className="text-xs"
                    />
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => removerPasso(idx)}
                    aria-label="Remover passo"
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              ))}
              <Button variant="outline" size="sm" onClick={adicionarPasso}>
                <Plus className="h-4 w-4" /> Adicionar passo
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      <OrdemServicoForm
        aberto={editAberto}
        ordem={os}
        itens={itens}
        setores={setores}
        servidores={servidores}
        tecnicos={tecnicos}
        usuarioAtual={usuario}
        onFechar={() => setEditAberto(false)}
        onSalvo={() => {
          setEditAberto(false);
          carregar();
        }}
      />

      <ImprimirDialog
        aberto={imprimirAberto}
        os={os}
        onFechar={() => setImprimirAberto(false)}
      />

      <ConfirmDialog
        aberto={excluindo}
        titulo="Excluir ordem de serviço"
        descricao={`Deseja excluir a OS ${os.numero}? Esta ação não pode ser desfeita.`}
        textoConfirmar="Excluir"
        destrutivo
        processando={processandoExcluir}
        onConfirmar={confirmarExcluir}
        onCancelar={() => setExcluindo(false)}
      />
    </div>
  );
}

// ImprimirDialog gera o PDF, permitindo acrescentar passos extras só na
// impressão (um por linha), sem alterar o checklist salvo.
function ImprimirDialog({
  aberto,
  os,
  onFechar,
}: {
  aberto: boolean;
  os: OrdemServico;
  onFechar: () => void;
}) {
  const { toast } = useToast();
  const [extras, setExtras] = React.useState("");
  const [baixando, setBaixando] = React.useState(false);

  React.useEffect(() => {
    if (aberto) setExtras("");
  }, [aberto]);

  async function baixar() {
    // Abre a aba de forma síncrona (ainda no gesto do clique) para evitar
    // bloqueio de popup. Em HTTP (contexto inseguro) o download direto de blob
    // é bloqueado pelo navegador; por isso exibimos o PDF numa nova aba para
    // visualizar/imprimir/salvar. Se o popup for bloqueado, cai no download.
    const aba = window.open("", "_blank");
    setBaixando(true);
    try {
      const linhas = extras
        .split("\n")
        .map((l) => l.trim())
        .filter(Boolean);
      const blob = await ordensServicoApi.baixarDocumento(os.id, linhas);
      const url = window.URL.createObjectURL(blob);
      if (aba) {
        aba.location.href = url;
      } else {
        baixarBlob(blob, `ordem-servico-${os.numero}.pdf`);
      }
      window.setTimeout(() => window.URL.revokeObjectURL(url), 60000);
      onFechar();
    } catch (err) {
      if (aba) aba.close();
      toast({
        titulo: "Não foi possível gerar o documento",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setBaixando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Imprimir documento da OS</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            O documento inclui os dados da OS e o checklist salvo. Você pode
            acrescentar passos extras abaixo (um por linha) — eles saem apenas
            neste documento, sem alterar o checklist salvo.
          </p>
          <FormField label="Passos extras (um por linha)">
            <Textarea
              value={extras}
              onChange={(e) => setExtras(e.target.value)}
              placeholder={"Ex.: Trocar pasta térmica\nSubstituir bateria CMOS"}
              rows={4}
            />
          </FormField>
        </div>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onFechar}>
            Cancelar
          </Button>
          <Button onClick={baixar} disabled={baixando}>
            {baixando ? <Spinner className="h-4 w-4" /> : "Baixar PDF"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function Campo({
  rotulo,
  children,
}: {
  rotulo: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-border/60 pb-2 last:border-0">
      <span className="text-muted-foreground">{rotulo}</span>
      <span className="text-right font-medium text-foreground">{children}</span>
    </div>
  );
}

function Texto({ rotulo, valor }: { rotulo: string; valor?: string }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        {rotulo}
      </p>
      <p className="whitespace-pre-wrap text-foreground">
        {valor?.trim() ? valor : "—"}
      </p>
    </div>
  );
}
