import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Pencil, Trash2, Eye } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { CarregandoTela } from "@/components/ui/spinner";
import { ArtigoForm } from "./ArtigoForm";
import { conhecimentoApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { formatarDataHora } from "@/lib/format";
import type { ArtigoConhecimento } from "@/types";

export default function ConhecimentoDetalhe() {
  const { id } = useParams<{ id: string }>();
  const artigoId = Number(id);
  const navigate = useNavigate();
  const { toast } = useToast();
  const { ehAdministrador } = useAuth();

  const [artigo, setArtigo] = React.useState<ArtigoConhecimento | null>(null);
  const [carregando, setCarregando] = React.useState(true);
  const [editAberto, setEditAberto] = React.useState(false);
  const [excluindo, setExcluindo] = React.useState(false);
  const [processandoExcluir, setProcessandoExcluir] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    conhecimentoApi
      .buscarPorId(artigoId)
      .then(setArtigo)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar o artigo",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [artigoId, toast]);

  React.useEffect(carregar, [carregar]);

  async function confirmarExcluir() {
    if (!artigo) return;
    setProcessandoExcluir(true);
    try {
      await conhecimentoApi.excluir(artigo.id);
      toast({ titulo: "Artigo excluído.", variant: "success" });
      navigate("/conhecimento");
    } catch (err) {
      toast({
        titulo: "Não foi possível excluir",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
      setProcessandoExcluir(false);
    }
  }

  if (carregando) return <CarregandoTela texto="Carregando artigo..." />;
  if (!artigo)
    return (
      <div className="py-12 text-center text-muted-foreground">
        Artigo não encontrado.
      </div>
    );

  return (
    <div>
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => navigate("/conhecimento")}
      >
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Button>

      <PageHeader
        titulo={artigo.titulo}
        acao={
          <>
            <Button variant="outline" onClick={() => setEditAberto(true)}>
              <Pencil className="h-4 w-4" /> Editar
            </Button>
            {ehAdministrador && (
              <Button
                variant="outline"
                className="text-destructive"
                onClick={() => setExcluindo(true)}
              >
                <Trash2 className="h-4 w-4" /> Excluir
              </Button>
            )}
          </>
        }
      />

      <div className="mb-4 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
        {artigo.categoria && (
          <Badge variant="secondary">{artigo.categoria}</Badge>
        )}
        {!artigo.publicado && <Badge variant="muted">Rascunho</Badge>}
        <span className="flex items-center gap-1">
          <Eye className="h-3.5 w-3.5" /> {artigo.visualizacoes} visualização(ões)
        </span>
        {artigo.autor?.nome && <span>· Por {artigo.autor.nome}</span>}
        <span>· Atualizado em {formatarDataHora(artigo.atualizado_em)}</span>
      </div>

      <Card>
        <CardContent className="p-6">
          <div className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
            {artigo.conteudo}
          </div>
        </CardContent>
      </Card>

      <ArtigoForm
        aberto={editAberto}
        artigo={artigo}
        onFechar={() => setEditAberto(false)}
        onSalvo={() => {
          setEditAberto(false);
          carregar();
        }}
      />

      <ConfirmDialog
        aberto={excluindo}
        titulo="Excluir artigo"
        descricao={`Deseja excluir "${artigo.titulo}"? Esta ação não pode ser desfeita.`}
        textoConfirmar="Excluir"
        destrutivo
        processando={processandoExcluir}
        onConfirmar={confirmarExcluir}
        onCancelar={() => setExcluindo(false)}
      />
    </div>
  );
}
