import * as React from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import { permissoesApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { Permissao, Usuario } from "@/types";

interface Props {
  usuario: Usuario | null;
  aberto: boolean;
  onClose: () => void;
}

// catálogo é estático; carrega uma vez e reaproveita entre aberturas.
let catalogoCache: Permissao[] | null = null;

export function DialogPermissoesUsuario({ usuario, aberto, onClose }: Props) {
  const { toast } = useToast();
  const [catalogo, setCatalogo] = React.useState<Permissao[]>(
    catalogoCache ?? []
  );
  const [selecionadas, setSelecionadas] = React.useState<Set<string>>(
    new Set()
  );
  const [perfilAlvo, setPerfilAlvo] = React.useState<string>("");
  const [editavel, setEditavel] = React.useState(false);
  const [carregando, setCarregando] = React.useState(false);
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    if (!aberto || !usuario) return;
    let cancelado = false;
    setCarregando(true);
    Promise.all([
      catalogoCache
        ? Promise.resolve(catalogoCache)
        : permissoesApi.catalogo(),
      permissoesApi.doUsuario(usuario.id),
    ])
      .then(([cat, estado]) => {
        if (cancelado) return;
        catalogoCache = cat;
        setCatalogo(cat);
        setSelecionadas(new Set(estado.concedidas));
        setPerfilAlvo(estado.perfil);
        setEditavel(estado.editavel);
      })
      .catch((err) => {
        if (cancelado) return;
        toast({
          titulo: "Erro ao carregar permissões",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
        onClose();
      })
      .finally(() => !cancelado && setCarregando(false));
    return () => {
      cancelado = true;
    };
  }, [aberto, usuario, toast, onClose]);

  function alternar(chave: string, valor: boolean) {
    setSelecionadas((atual) => {
      const nova = new Set(atual);
      if (valor) nova.add(chave);
      else nova.delete(chave);
      return nova;
    });
  }

  async function salvar() {
    if (!usuario) return;
    setSalvando(true);
    try {
      await permissoesApi.salvar(usuario.id, Array.from(selecionadas));
      toast({
        titulo: "Permissões atualizadas.",
        descricao: `As permissões de ${usuario.nome} foram salvas.`,
        variant: "success",
      });
      onClose();
    } catch (err) {
      toast({
        titulo: "Não foi possível salvar as permissões",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  // Agrupa o catálogo por grupo, preservando a ordem de primeira aparição.
  const grupos = React.useMemo(() => {
    const map = new Map<string, Permissao[]>();
    for (const p of catalogo) {
      const arr = map.get(p.grupo) ?? [];
      arr.push(p);
      map.set(p.grupo, arr);
    }
    return Array.from(map.entries());
  }, [catalogo]);

  const ehAdmin = perfilAlvo === "administrador";
  const ehSolicitante = perfilAlvo === "solicitante";

  return (
    <Dialog open={aberto} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[85vh] max-w-2xl overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>Permissões de {usuario?.nome}</DialogTitle>
        </DialogHeader>

        {carregando ? (
          <div className="flex justify-center py-10">
            <Spinner className="h-6 w-6" />
          </div>
        ) : ehAdmin ? (
          <p className="py-6 text-sm text-muted-foreground">
            <Badge variant="default" className="mr-2">
              Administrador
            </Badge>
            Administradores têm acesso total ao sistema. As permissões não são
            editáveis para evitar travar o acesso administrativo.
          </p>
        ) : ehSolicitante ? (
          <p className="py-6 text-sm text-muted-foreground">
            Solicitantes acessam apenas o portal de chamados (abrir e acompanhar
            os próprios chamados). Não participam das permissões da equipe.
          </p>
        ) : (
          <>
            <p className="text-sm text-muted-foreground">
              Marque o que este operador pode fazer. Por padrão, o operador tem
              acesso à operação do dia a dia; as ações abaixo são adicionais.
            </p>
            <div className="flex-1 overflow-y-auto pr-1 space-y-5 py-2">
              {grupos.map(([grupo, itens]) => (
                <div key={grupo}>
                  <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    {grupo}
                  </h4>
                  <div className="space-y-3">
                    {itens.map((p) => (
                      <label
                        key={p.chave}
                        htmlFor={`perm-${p.chave}`}
                        className="flex items-start justify-between gap-3 rounded-md border p-3 cursor-pointer hover:bg-accent/50"
                      >
                        <div className="min-w-0">
                          <div className="text-sm font-medium">{p.rotulo}</div>
                          <div className="text-xs text-muted-foreground">
                            {p.descricao}
                          </div>
                        </div>
                        <Switch
                          id={`perm-${p.chave}`}
                          checked={selecionadas.has(p.chave)}
                          onCheckedChange={(v) => alternar(p.chave, v)}
                          disabled={!editavel}
                        />
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </>
        )}

        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {ehAdmin || ehSolicitante ? "Fechar" : "Cancelar"}
          </Button>
          {editavel && !ehAdmin && !ehSolicitante && (
            <Button type="button" onClick={salvar} disabled={salvando || carregando}>
              {salvando ? <Spinner className="h-4 w-4" /> : "Salvar permissões"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
