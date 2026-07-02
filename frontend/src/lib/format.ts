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
export function baixarBlob(blob: Blob, nomeArquivo: string) {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = nomeArquivo;
  document.body.appendChild(a);
  a.click();
  a.remove();
  window.URL.revokeObjectURL(url);
}
