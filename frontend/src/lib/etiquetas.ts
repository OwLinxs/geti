// Impressão em lote de etiquetas QR de patrimônio: monta uma folha A4 com 4
// etiquetas por linha, cada uma com QR Code, código interno de T.I., número de
// patrimônio e descrição. Abre em nova aba para visualizar/imprimir/salvar.
//
// O QR aponta para a página do bem (mesmo destino da etiqueta individual), de
// forma que escanear com o celular abre o item no sistema.

import QRCode from "qrcode";
import { codigoInterno } from "@/lib/patrimonio";
import type { Item } from "@/types";

function escapar(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

/**
 * Abre uma nova aba com as etiquetas dos itens informados prontas para
 * impressão (4 por linha, folha A4). Retorna false se o popup foi bloqueado.
 *
 * Deve ser chamada dentro do gesto de clique do usuário (abre a aba de forma
 * síncrona) para não ser bloqueada pelo navegador.
 */
export async function imprimirEtiquetasQR(itens: Item[]): Promise<boolean> {
  if (itens.length === 0) return true;

  // Abre a aba ainda no gesto do clique (evita bloqueio de popup).
  const aba = window.open("", "_blank");
  if (!aba) return false;

  aba.document.write(
    `<!doctype html><html lang="pt-BR"><head><meta charset="utf-8">` +
      `<title>Etiquetas de patrimônio</title></head>` +
      `<body style="font-family:system-ui,sans-serif">` +
      `<p style="padding:24px;text-align:center;color:#555">Gerando etiquetas…</p>` +
      `</body></html>`
  );
  aba.document.close();

  // Gera os QR Codes (base64) em paralelo.
  const origem = window.location.origin;
  const etiquetas = await Promise.all(
    itens.map(async (item) => {
      const url = `${origem}/patrimonio/${item.id}`;
      let dataUrl = "";
      try {
        dataUrl = await QRCode.toDataURL(url, { width: 220, margin: 1 });
      } catch {
        dataUrl = "";
      }
      return { item, dataUrl };
    })
  );

  const cards = etiquetas
    .map(({ item, dataUrl }) => {
      const codigo = codigoInterno(item.id);
      const patr = item.numero_patrimonio
        ? `<div class="pat">Patr. ${escapar(item.numero_patrimonio)}</div>`
        : "";
      const img = dataUrl
        ? `<img src="${dataUrl}" alt="QR"/>`
        : `<div class="semqr">sem QR</div>`;
      return (
        `<div class="etq">${img}` +
        `<div class="cod">${escapar(codigo)}</div>` +
        `<div class="desc">${escapar(item.descricao || "")}</div>` +
        `${patr}</div>`
      );
    })
    .join("");

  const html =
    `<!doctype html><html lang="pt-BR"><head><meta charset="utf-8">` +
    `<title>Etiquetas de patrimônio</title>` +
    `<style>
      @page { size: A4; margin: 10mm; }
      * { box-sizing: border-box; }
      body { font-family: system-ui, sans-serif; margin: 0; padding: 8mm; }
      .grid {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 6mm 4mm;
      }
      .etq {
        border: 1px dashed #bbb;
        border-radius: 6px;
        padding: 4mm 2mm;
        text-align: center;
        page-break-inside: avoid;
        break-inside: avoid;
      }
      .etq img { width: 34mm; height: 34mm; display: block; margin: 0 auto; }
      .semqr {
        width: 34mm; height: 34mm; margin: 0 auto;
        display: flex; align-items: center; justify-content: center;
        color: #999; font-size: 10px; border: 1px solid #eee;
      }
      .cod { font-family: monospace; font-size: 12px; margin-top: 3mm; font-weight: 600; }
      .desc {
        font-size: 10px; color: #333; margin-top: 1mm;
        overflow: hidden; display: -webkit-box;
        -webkit-line-clamp: 2; -webkit-box-orient: vertical;
      }
      .pat { font-size: 10px; color: #666; margin-top: 1mm; }
      .barra { margin-bottom: 6mm; text-align: center; }
      .barra button {
        font: inherit; padding: 6px 16px; border: 1px solid #333;
        border-radius: 6px; background: #111; color: #fff; cursor: pointer;
      }
      @media print { .barra { display: none; } body { padding: 0; } }
    </style></head>` +
    `<body>` +
    `<div class="barra"><button onclick="window.print()">Imprimir</button></div>` +
    `<div class="grid">${cards}</div>` +
    `</body></html>`;

  aba.document.open();
  aba.document.write(html);
  aba.document.close();
  aba.focus();
  // Dispara o diálogo de impressão automaticamente.
  setTimeout(() => {
    try {
      aba.print();
    } catch {
      /* usuário pode imprimir pelo botão */
    }
  }, 400);
  return true;
}
