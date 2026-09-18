import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { categoriasChamadoApi, ordensServicoApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { PRIORIDADES_OS } from "@/lib/rotulos";
import { ordenarComoArvore, prefixoIndentacao } from "@/lib/setores";
import { cn } from "@/lib/utils";
import type {
  CategoriaChamado,
  Item,
  OrdemServico,
  OrdemServicoPayload,
  Servidor,
  Setor,
  Usuario,
} from "@/types";

const SEM = "0";

type OrigemEquip = "inventario" | "externa";

interface Props {
  aberto: boolean;
  ordem: OrdemServico | null; // null => criação
  itens: Item[];
  setores: Setor[];
  servidores: Servidor[];
  tecnicos: Usuario[]; // lista de técnicos selecionáveis (admin); pode vir vazia
  usuarioAtual: Usuario | null; // fallback de técnico para operador
  onFechar: () => void;
  onSalvo: () => void;
}

export function OrdemServicoForm({
  aberto,
  ordem,
  itens,
  setores,
  servidores,
  tecnicos,
  usuarioAtual,
  onFechar,
  onSalvo,
}: Props) {
  const { toast } = useToast();
  const [origem, setOrigem] = React.useState<OrigemEquip>("inventario");
  const [itemId, setItemId] = React.useState<string>(SEM);
  const [equipDescricao, setEquipDescricao] = React.useState("");
  const [equipIdent, setEquipIdent] = React.useState("");
  const [setorId, setSetorId] = React.useState<string>(SEM);
  const [categoriaId, setCategoriaId] = React.useState<string>(SEM);
  const [categorias, setCategorias] = React.useState<CategoriaChamado[]>([]);
  const [solicitanteId, setSolicitanteId] = React.useState<string>(SEM);
  const [defeito, setDefeito] = React.useState("");
  const [diagnostico, setDiagnostico] = React.useState("");
  const [solucao, setSolucao] = React.useState("");
  const [prioridade, setPrioridade] = React.useState("normal");
  const [tecnicoId, setTecnicoId] = React.useState<string>(SEM);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  const setoresArvore = React.useMemo(
    () => ordenarComoArvore(setores),
    [setores]
  );

  // Opções de técnico: lista de admin, ou só o próprio usuário (operador).
  const opcoesTecnico = React.useMemo<Usuario[]>(() => {
    if (tecnicos.length > 0) return tecnicos;
    return usuarioAtual ? [usuarioAtual] : [];
  }, [tecnicos, usuarioAtual]);

  React.useEffect(() => {
    if (!aberto) return;
    categoriasChamadoApi
      .listar(true)
      .then(setCategorias)
      .catch(() => setCategorias([]));
    if (ordem) {
      setOrigem(ordem.item_id ? "inventario" : "externa");
      setItemId(ordem.item_id ? String(ordem.item_id) : SEM);
      setEquipDescricao(ordem.equipamento_descricao ?? "");
      setEquipIdent(ordem.equipamento_identificacao ?? "");
      setSetorId(ordem.setor_id ? String(ordem.setor_id) : SEM);
      setCategoriaId(
        ordem.categoria_chamado_id ? String(ordem.categoria_chamado_id) : SEM
      );
      setSolicitanteId(ordem.solicitante_id ? String(ordem.solicitante_id) : SEM);
      setDefeito(ordem.defeito_relatado);
      setDiagnostico(ordem.diagnostico ?? "");
      setSolucao(ordem.solucao_aplicada ?? "");
      setPrioridade(ordem.prioridade);
      setTecnicoId(ordem.tecnico_id ? String(ordem.tecnico_id) : SEM);
    } else {
      setOrigem("inventario");
      setItemId(SEM);
      setEquipDescricao("");
      setEquipIdent("");
      setSetorId(SEM);
      setCategoriaId(SEM);
      setSolicitanteId(SEM);
      setDefeito("");
      setDiagnostico("");
      setSolucao("");
      setPrioridade("normal");
      // Operador: técnico padrão é ele mesmo.
      setTecnicoId(
        tecnicos.length === 0 && usuarioAtual ? String(usuarioAtual.id) : SEM
      );
    }
    setErros({});
  }, [aberto, ordem, tecnicos.length, usuarioAtual]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!defeito.trim())
      novos.defeito_relatado = "Descreva o defeito relatado.";
    if (origem === "inventario" && itemId === SEM)
      novos.item_id = "Selecione o item do inventário.";
    if (origem === "externa" && !equipDescricao.trim())
      novos.equipamento = "Descreva a máquina externa.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});

    const payload: OrdemServicoPayload = {
      item_id: origem === "inventario" ? Number(itemId) : 0,
      equipamento_descricao: origem === "externa" ? equipDescricao.trim() : "",
      equipamento_identificacao:
        origem === "externa" ? equipIdent.trim() : "",
      setor_id: setorId !== SEM ? Number(setorId) : null,
      categoria_chamado_id: categoriaId !== SEM ? Number(categoriaId) : null,
      solicitante_id: solicitanteId !== SEM ? Number(solicitanteId) : null,
      defeito_relatado: defeito.trim(),
      diagnostico: diagnostico.trim(),
      solucao_aplicada: solucao.trim(),
      prioridade: prioridade as OrdemServicoPayload["prioridade"],
      tecnico_id: tecnicoId !== SEM ? Number(tecnicoId) : null,
    };

    setSalvando(true);
    try {
      if (ordem) {
        await ordensServicoApi.atualizar(ordem.id, payload);
        toast({ titulo: "Ordem de serviço atualizada.", variant: "success" });
      } else {
        await ordensServicoApi.criar(payload);
        toast({ titulo: "Ordem de serviço aberta.", variant: "success" });
      }
      onSalvo();
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível salvar",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {ordem ? `Editar OS ${ordem.numero}` : "Nova ordem de serviço"}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          {/* Origem do equipamento. */}
          <FormField label="Equipamento" erro={erros.equipamento}>
            <div className="mb-2 flex gap-2">
              {(
                [
                  ["inventario", "Item do inventário"],
                  ["externa", "Máquina externa"],
                ] as [OrigemEquip, string][]
              ).map(([v, rot]) => (
                <button
                  key={v}
                  type="button"
                  onClick={() => setOrigem(v)}
                  className={cn(
                    "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
                    origem === v
                      ? "border-primary bg-primary text-primary-foreground"
                      : "border-border bg-card hover:bg-accent"
                  )}
                >
                  {rot}
                </button>
              ))}
            </div>

            {origem === "inventario" ? (
              <Select value={itemId} onValueChange={setItemId}>
                <SelectTrigger>
                  <SelectValue placeholder="Selecione o item" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Selecione…</SelectItem>
                  {itens.map((i) => (
                    <SelectItem key={i.id} value={String(i.id)}>
                      {i.descricao}
                      {i.numero_patrimonio ? ` — ${i.numero_patrimonio}` : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Input
                  value={equipDescricao}
                  onChange={(e) => setEquipDescricao(e.target.value)}
                  placeholder="Descrição (ex.: Notebook do usuário)"
                />
                <Input
                  value={equipIdent}
                  onChange={(e) => setEquipIdent(e.target.value)}
                  placeholder="Série / patrimônio / tag (opcional)"
                />
              </div>
            )}
            {erros.item_id && (
              <p className="mt-1 text-sm text-destructive">{erros.item_id}</p>
            )}
          </FormField>

          <FormField
            label="Defeito relatado"
            htmlFor="defeito"
            obrigatorio
            erro={erros.defeito_relatado}
          >
            <Textarea
              id="defeito"
              value={defeito}
              onChange={(e) => setDefeito(e.target.value)}
              placeholder="Ex.: Não liga; barulho na fonte; muito lento…"
            />
          </FormField>

          <FormField label="Categoria do chamado">
            <Select value={categoriaId} onValueChange={setCategoriaId}>
              <SelectTrigger>
                <SelectValue placeholder="Opcional" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={SEM}>Não classificado</SelectItem>
                {categorias.map((cat) => (
                  <SelectItem key={cat.id} value={String(cat.id)}>
                    {cat.nome}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Prioridade">
              <Select value={prioridade} onValueChange={setPrioridade}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PRIORIDADES_OS.map((p) => (
                    <SelectItem key={p.valor} value={p.valor}>
                      {p.rotulo}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Técnico responsável">
              <Select value={tecnicoId} onValueChange={setTecnicoId}>
                <SelectTrigger>
                  <SelectValue placeholder="Não definido" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não definido</SelectItem>
                  {opcoesTecnico.map((u) => (
                    <SelectItem key={u.id} value={String(u.id)}>
                      {u.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <FormField label="Departamento de origem">
              <Select value={setorId} onValueChange={setSetorId}>
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não informado</SelectItem>
                  {setoresArvore.map(({ setor: s, nivel }) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {prefixoIndentacao(nivel)}
                      {s.nome}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="Solicitante">
              <Select value={solicitanteId} onValueChange={setSolicitanteId}>
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SEM}>Não informado</SelectItem>
                  {servidores.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.nome} ({s.matricula})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          </div>

          <FormField label="Diagnóstico" htmlFor="diagnostico">
            <Textarea
              id="diagnostico"
              value={diagnostico}
              onChange={(e) => setDiagnostico(e.target.value)}
              placeholder="Preenchido durante o atendimento (opcional)"
            />
          </FormField>

          <FormField label="Solução aplicada" htmlFor="solucao">
            <Textarea
              id="solucao"
              value={solucao}
              onChange={(e) => setSolucao(e.target.value)}
              placeholder="Preenchido ao concluir (opcional)"
            />
          </FormField>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
