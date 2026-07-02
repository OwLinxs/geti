import * as React from "react";
import { categoriasApi, setoresApi, servidoresApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import type { Categoria, Servidor, Setor } from "@/types";

// Carrega listas de referência (categorias, setores, servidores) usadas em
// formulários de itens, movimentações e termos.
export function useReferencias() {
  const { toast } = useToast();
  const [categorias, setCategorias] = React.useState<Categoria[]>([]);
  const [setores, setSetores] = React.useState<Setor[]>([]);
  const [servidores, setServidores] = React.useState<Servidor[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  React.useEffect(() => {
    Promise.all([
      categoriasApi.listar(),
      setoresApi.listar(),
      servidoresApi.listar(),
    ])
      .then(([cats, sets, servs]) => {
        setCategorias(cats);
        setSetores(sets);
        setServidores(servs);
      })
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar listas",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  // Insere um servidor recém-criado (cadastro rápido) sem refazer a busca.
  const adicionarServidor = React.useCallback((s: Servidor) => {
    setServidores((prev) =>
      prev.some((x) => x.id === s.id) ? prev : [s, ...prev]
    );
  }, []);

  return { categorias, setores, servidores, carregando, adicionarServidor };
}
