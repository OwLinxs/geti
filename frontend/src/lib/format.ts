// Utilitários de formatação para PT-BR.

export function formatarData(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("pt-BR");
}

export function formatarDataHora(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "—";
  return d.toLocaleString("pt-BR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatarMoeda(valor?: number | null): string {
  if (valor === undefined || valor === null) return "—";
  return valor.toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

// Converte uma data <input type="date"> (YYYY-MM-DD) para ISO ou null.
export function dateInputParaISO(valor: string): string | null {
  if (!valor) return null;
  return new Date(`${valor}T00:00:00`).toISOString();
}

// Converte um ISO para o formato aceito por <input type="date">.
export function isoParaDateInput(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
  return d.toISOString().slice(0, 10);
}

// Dispara o download de um Blob com o nome informado.
//
// A limpeza (remover o <a> e revogar o object URL) é adiada: revogar o URL de
// forma síncrona logo após o click() faz alguns navegadores (ex.: Firefox)
// cancelarem o download — o download "começa mas não completa".
export function baixarBlob(blob: Blob, nomeArquivo: string) {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = nomeArquivo;
  a.rel = "noopener";
  document.body.appendChild(a);
  a.click();
  setTimeout(() => {
    a.remove();
    window.URL.revokeObjectURL(url);
  }, 1500);
}
