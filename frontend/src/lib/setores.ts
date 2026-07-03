import type { Setor } from "@/types";

// Nó da árvore de unidades com a profundidade calculada.
export interface SetorNo {
  setor: Setor;
  nivel: number;
}

// ordenarComoArvore recebe a lista plana de setores e devolve os nós em ordem
// hierárquica (pais antes dos filhos), cada um com seu nível de profundidade —
// para renderização indentada (Secretaria › Departamento › Unidade).
export function ordenarComoArvore(lista: Setor[]): SetorNo[] {
  const filhosDe = new Map<number | null, Setor[]>();
  for (const s of lista) {
    const chave = s.pai_id ?? null;
    if (!filhosDe.has(chave)) filhosDe.set(chave, []);
    filhosDe.get(chave)!.push(s);
  }
  for (const arr of filhosDe.values()) {
    arr.sort((a, b) => a.nome.localeCompare(b.nome, "pt-BR"));
  }
  const saida: SetorNo[] = [];
  const visitar = (paiId: number | null, nivel: number) => {
    for (const s of filhosDe.get(paiId) ?? []) {
      saida.push({ setor: s, nivel });
      visitar(s.id, nivel + 1);
    }
  };
  visitar(null, 0);
  return saida;
}

// idsDoRamo devolve o id informado + todos os descendentes. Útil para impedir
// que uma unidade seja movida para baixo de si mesma (evita ciclo).
export function idsDoRamo(lista: Setor[], raizId: number): Set<number> {
  const filhosDe = new Map<number, Setor[]>();
  for (const s of lista) {
    if (s.pai_id != null) {
      if (!filhosDe.has(s.pai_id)) filhosDe.set(s.pai_id, []);
      filhosDe.get(s.pai_id)!.push(s);
    }
  }
  const set = new Set<number>([raizId]);
  const pilha = [raizId];
  while (pilha.length) {
    const atual = pilha.pop()!;
    for (const f of filhosDe.get(atual) ?? []) {
      if (!set.has(f.id)) {
        set.add(f.id);
        pilha.push(f.id);
      }
    }
  }
  return set;
}

// prefixoIndentacao devolve espaços + conector para exibir o nível numa opção
// de <select> (que não aceita indentação por CSS de forma confiável).
export function prefixoIndentacao(nivel: number): string {
  if (nivel <= 0) return "";
  return "  ".repeat(nivel) + "└ ";
}

// caminhoSetor monta o caminho completo "Secretaria › Departamento › Unidade"
// de um setor, para exibição fora de listas hierárquicas (ex.: detalhe).
export function caminhoSetor(lista: Setor[], id?: number | null): string {
  if (id == null) return "";
  const por = new Map(lista.map((s) => [s.id, s]));
  const partes: string[] = [];
  let atual = por.get(id);
  let guarda = 0;
  while (atual && guarda++ < 100) {
    partes.unshift(atual.nome);
    atual = atual.pai_id != null ? por.get(atual.pai_id) : undefined;
  }
  return partes.join(" › ");
}
