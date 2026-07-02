package services

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/pmfb/sige-ti/internal/config"
	"github.com/pmfb/sige-ti/internal/models"
)

// cabecalhoInstitucional desenha o cabeçalho parametrizável (logo + textos
// vindos de variáveis de ambiente, sem hardcode).
func cabecalhoInstitucional(pdf *gofpdf.Fpdf, cfg *config.Config) {
	if cfg.PrefeituraLogoPath != "" {
		if _, err := os.Stat(cfg.PrefeituraLogoPath); err == nil {
			// Logo no canto superior esquerdo (largura 20mm, altura automática).
			pdf.ImageOptions(cfg.PrefeituraLogoPath, 15, 10, 20, 0, false,
				gofpdf.ImageOptions{ImageType: "", ReadDpi: true}, 0, "")
		}
	}
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetXY(40, 12)
	pdf.MultiCell(0, 6, traduzir(cfg.PrefeituraNome), "", "L", false)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(40)
	pdf.MultiCell(0, 5, traduzir(cfg.PrefeituraDepto), "", "L", false)
	pdf.Ln(4)
	pdf.SetDrawColor(120, 120, 120)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)
}

// PDFTermoResponsabilidade gera o termo (recibo) em PDF.
func PDFTermoResponsabilidade(termo *models.TermoResponsabilidade, cfg *config.Config) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 12, 15)
	pdf.AddPage()
	cabecalhoInstitucional(pdf, cfg)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, traduzir(cfg.TermoCabecalho), "", 1, "C", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, traduzir(fmt.Sprintf("Número: %s", termo.Numero)), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 11)
	corpo := fmt.Sprintf(
		"Declaro, para os devidos fins, que recebi do %s o equipamento abaixo "+
			"discriminado, comprometendo-me a zelar pela sua guarda e conservação, "+
			"utilizando-o exclusivamente no exercício de minhas funções, e a "+
			"devolvê-lo quando solicitado ou ao término de minhas atividades.",
		cfg.PrefeituraDepto,
	)
	pdf.MultiCell(0, 6, traduzir(corpo), "", "J", false)
	pdf.Ln(4)

	// Tabela de dados.
	linha := func(rotulo, valor string) {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(45, 7, traduzir(rotulo), "1", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(0, 7, traduzir(valor), "1", 1, "L", false, 0, "")
	}
	linha("Servidor:", termo.ServidorNomeSnapshot)
	linha("Matrícula:", termo.ServidorMatriculaSnapshot)
	linha("Equipamento:", termo.ItemDescricaoSnapshot)
	if termo.PatrimonioSnapshot != "" {
		linha("Nº Patrimônio:", termo.PatrimonioSnapshot)
	}
	linha("Data de emissão:", termo.DataEmissao.Format("02/01/2006"))
	if termo.Observacao != "" {
		linha("Observações:", termo.Observacao)
	}
	pdf.Ln(16)

	// Local e data.
	local := cfg.TermoCidadeUF
	dataExtenso := fmt.Sprintf("%s, %s.", local, dataPorExtenso(termo.DataEmissao))
	if local == "" {
		dataExtenso = fmt.Sprintf("%s.", dataPorExtenso(termo.DataEmissao))
	}
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 6, traduzir(dataExtenso), "", 1, "R", false, 0, "")
	pdf.Ln(18)

	// Assinatura.
	larguraAssinatura := 90.0
	x := (210.0 - larguraAssinatura) / 2
	pdf.Line(x, pdf.GetY(), x+larguraAssinatura, pdf.GetY())
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 5, traduzir(termo.ServidorNomeSnapshot), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, traduzir("Assinatura do servidor responsável"), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PDFRelatorioItens gera um PDF tabular de uma lista de itens.
func PDFRelatorioItens(titulo string, itens []models.Item, cfg *config.Config) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // paisagem para caber colunas
	pdf.SetMargins(10, 12, 10)
	pdf.AddPage()
	cabecalhoInstitucional(pdf, cfg)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, traduzir(titulo), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(0, 5, traduzir(fmt.Sprintf("Emitido em %s", time.Now().Format("02/01/2006 15:04"))), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Cabeçalho da tabela.
	cols := []struct {
		titulo string
		largura float64
	}{
		{"Descrição", 70},
		{"Categoria", 35},
		{"Patrimônio", 30},
		{"Estado", 25},
		{"Qtd", 15},
		{"Setor", 45},
		{"Responsável", 45},
	}
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetFillColor(230, 230, 230)
	for _, col := range cols {
		pdf.CellFormat(col.largura, 7, traduzir(col.titulo), "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, i := range itens {
		vals := []string{
			i.Descricao,
			nomeCategoria(&i),
			derefStr(i.NumeroPatrimonio),
			string(i.EstadoConservacao),
			fmt.Sprintf("%d", i.Quantidade),
			nomeSetor(&i),
			nomeResponsavel(&i),
		}
		for idx, col := range cols {
			pdf.CellFormat(col.largura, 6, traduzir(truncar(vals[idx], 45)), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// trUTF8 converte UTF-8 para a codificação aceita pelas fontes core do gofpdf
// (CP1252/Latin-1), preservando a acentuação do português. Inicializado uma
// única vez.
var trUTF8 = gofpdf.New("P", "mm", "A4", "").UnicodeTranslatorFromDescriptor("")

func traduzir(s string) string {
	return trUTF8(s)
}

func truncar(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func dataPorExtenso(t time.Time) string {
	meses := []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho",
		"julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
	return fmt.Sprintf("%d de %s de %d", t.Day(), meses[int(t.Month())-1], t.Year())
}
