import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
import { servidoresApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { Servidor, Setor } from "@/types";

const SEM = "0";

interface NovoServidorDialogProps {
  aberto: boolean;
  setores: Setor[];
  onFechar: () => void;
  onCriado: (s: Servidor) => void;
}

// Cadastro rápido de servidor responsável, sem sair do fluxo atual (ex.: ao
// registrar uma movimentação e o responsável ainda não existir).
// LGPD: coleta apenas nome e matrícula.
export function NovoServidorDialog({
  aberto,
  setores,
  onFechar,
  onCriado,
}: NovoServidorDialogProps) {
  const { toast } = useToast();
  const [nome, setNome] = React.useState("");
  const [matricula, setMatricula] = React.useState("");
  const [setorId, setSetorId] = React.useState<string>(SEM);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (aberto) {
      setNome("");
      setMatricula("");
      setSetorId(SEM);
      setErros({});
    }
  }, [aberto]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!nome.trim()) novos.nome = "Informe o nome do servidor.";
    if (!matricula.trim()) novos.matricula = "Informe a matrícula.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});
    setSalvando(true);
    try {
      const novo = await servidoresApi.criar({
        nome: nome.trim(),
        matricula: matricula.trim(),
        setor_id: setorId !== SEM ? Number(setorId) : null,
      });
      toast({ titulo: "Servidor cadastrado.", variant: "success" });
      onCriado(novo);
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível cadastrar",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onFechar()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Cadastrar servidor</DialogTitle>
          <DialogDescription>
            Cadastro rápido do responsável. Apenas nome e matrícula são
            coletados (LGPD).
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField label="Nome" htmlFor="ns_nome" obrigatorio erro={erros.nome}>
            <Input
              id="ns_nome"
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              autoFocus
            />
          </FormField>

          <FormField
            label="Matrícula"
            htmlFor="ns_mat"
            obrigatorio
            erro={erros.matricula}
          >
            <Input
              id="ns_mat"
              value={matricula}
              onChange={(e) => setMatricula(e.target.value)}
            />
          </FormField>

          <FormField label="Departamento (lotação)">
            <Select value={setorId} onValueChange={setSetorId}>
              <SelectTrigger>
                <SelectValue placeholder="Opcional" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={SEM}>Não informado</SelectItem>
                {setores.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.nome}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>

          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onFechar}>
              Cancelar
            </Button>
            <Button type="submit" disabled={salvando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Cadastrar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
