import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { conhecimentoApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { ArtigoConhecimento, ArtigoConhecimentoPayload } from "@/types";

interface Props {
  aberto: boolean;
  artigo: ArtigoConhecimento | null; // null => criação
  onFechar: () => void;
  onSalvo: () => void;
}

export function ArtigoForm({ aberto, artigo, onFechar, onSalvo }: Props) {
  const { toast } = useToast();
  const [titulo, setTitulo] = React.useState("");
  const [categoria, setCategoria] = React.useState("");
  const [conteudo, setConteudo] = React.useState("");
  const [publicado, setPublicado] = React.useState(true);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (!aberto) return;
    setTitulo(artigo?.titulo ?? "");
    setCategoria(artigo?.categoria ?? "");
    setConteudo(artigo?.conteudo ?? "");
    setPublicado(artigo?.publicado ?? true);
    setErros({});
  }, [aberto, artigo]);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novos: Record<string, string> = {};
    if (!titulo.trim()) novos.titulo = "Informe o título.";
    if (!conteudo.trim()) novos.conteudo = "Informe o conteúdo.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setErros({});

    const payload: ArtigoConhecimentoPayload = {
      titulo: titulo.trim(),
      categoria: categoria.trim(),
      conteudo: conteudo.trim(),
      publicado,
    };

    setSalvando(true);
    try {
      if (artigo) {
        await conhecimentoApi.atualizar(artigo.id, payload);
        toast({ titulo: "Artigo atualizado.", variant: "success" });
      } else {
        await conhecimentoApi.criar(payload);
        toast({ titulo: "Artigo criado.", variant: "success" });
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
          <DialogTitle>{artigo ? "Editar artigo" : "Novo artigo"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={salvar} className="space-y-4" noValidate>
          <FormField label="Título" htmlFor="titulo" obrigatorio erro={erros.titulo}>
            <Input
              id="titulo"
              value={titulo}
              onChange={(e) => setTitulo(e.target.value)}
              placeholder="Ex.: Como resetar a senha do e-mail institucional"
            />
          </FormField>

          <FormField label="Categoria" htmlFor="categoria">
            <Input
              id="categoria"
              value={categoria}
              onChange={(e) => setCategoria(e.target.value)}
              placeholder="Ex.: E-mail, Rede, Impressoras (opcional)"
            />
          </FormField>

          <FormField
            label="Conteúdo"
            htmlFor="conteudo"
            obrigatorio
            erro={erros.conteudo}
          >
            <Textarea
              id="conteudo"
              value={conteudo}
              onChange={(e) => setConteudo(e.target.value)}
              placeholder="Passo a passo, solução, orientação…"
              rows={12}
            />
          </FormField>

          <div className="flex items-center justify-between rounded-md border border-border px-3 py-2">
            <div>
              <p className="text-sm font-medium">Publicado</p>
              <p className="text-xs text-muted-foreground">
                Artigos não publicados ficam como rascunho (ainda visíveis à
                equipe).
              </p>
            </div>
            <Switch checked={publicado} onCheckedChange={setPublicado} />
          </div>

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
