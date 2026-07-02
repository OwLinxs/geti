package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/pmfb/sige-ti/internal/models"
)

// CSVItens gera um CSV (em memória) do inventário de itens.
func CSVItens(itens []models.Item) ([]byte, error) {
	var buf bytes.Buffer
	// BOM UTF-8 para acentuação correta no Excel.
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	cabecalho := []string{
		"ID", "Descricao", "Categoria", "Patrimonio", "Serie", "Marca", "Modelo",
		"Estado", "Quantidade", "Estoque Minimo", "Setor", "Responsavel",
		"Valor", "Baixado",
	}
	if err := w.Write(cabecalho); err != nil {
		return nil, err
	}

	for _, i := range itens {
		linha := []string{
			strconv.FormatUint(uint64(i.ID), 10),
			i.Descricao,
			nomeCategoria(&i),
			derefStr(i.NumeroPatrimonio),
			derefStr(i.NumeroSerie),
			i.Marca,
			i.Modelo,
			string(i.EstadoConservacao),
			strconv.Itoa(i.Quantidade),
			strconv.Itoa(i.EstoqueMinimo),
			nomeSetor(&i),
			nomeResponsavel(&i),
			valorStr(i.Valor),
			boolStr(i.Baixado),
		}
		if err := w.Write(linha); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CSVMovimentacoes gera um CSV do histórico de movimentações.
func CSVMovimentacoes(movs []models.Movimentacao) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	if err := w.Write([]string{
		"ID", "Data", "Item", "Tipo", "Quantidade", "Saldo Resultante",
		"Origem", "Destino", "Servidor", "Registrado Por", "Observacao",
	}); err != nil {
		return nil, err
	}

	for _, m := range movs {
		item := ""
		if m.Item != nil {
			item = m.Item.Descricao
		}
		serv := ""
		if m.Servidor != nil {
			serv = m.Servidor.Nome
		}
		reg := ""
		if m.RegistradoPor != nil {
			reg = m.RegistradoPor.Nome
		}
		if err := w.Write([]string{
			strconv.FormatUint(uint64(m.ID), 10),
			m.DataEvento.Format("02/01/2006 15:04"),
			item,
			string(m.Tipo),
			strconv.Itoa(m.Quantidade),
			strconv.Itoa(m.SaldoResultante),
			m.OrigemDescricao,
			m.DestinoDescricao,
			serv,
			reg,
			m.Observacao,
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func nomeCategoria(i *models.Item) string {
	if i.Categoria != nil {
		return i.Categoria.Nome
	}
	return ""
}
func nomeSetor(i *models.Item) string {
	if i.Setor != nil {
		return i.Setor.Nome
	}
	return ""
}
func nomeResponsavel(i *models.Item) string {
	if i.Responsavel != nil {
		return i.Responsavel.Nome
	}
	return ""
}
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func valorStr(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *v)
}
func boolStr(b bool) string {
	if b {
		return "Sim"
	}
	return "Nao"
}
