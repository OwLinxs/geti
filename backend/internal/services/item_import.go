package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pmfb/sige-ti/internal/models"
)

// ErroLinha descreve um problema em uma linha específica do CSV.
type ErroLinha struct {
	Linha     int    `json:"linha"`     // número da linha no arquivo (1 = cabeçalho)
	Descricao string `json:"descricao"` // conteúdo da linha, para contexto
	Mensagem  string `json:"mensagem"`  // motivo da rejeição
}

// ResultadoImportacao resume o resultado de uma importação em massa.
type ResultadoImportacao struct {
	// Validacao=true indica simulação (dry-run): nada foi gravado.
	Validacao  bool        `json:"validacao"`
	Total      int         `json:"total"`      // linhas de dados lidas
	Importados int         `json:"importados"` // linhas gravadas (0 em dry-run)
	Validas    int         `json:"validas"`    // linhas sem erro
	Erros      []ErroLinha `json:"erros"`
}

// CabecalhoCSVItens é o cabeçalho esperado do arquivo de importação.
var CabecalhoCSVItens = []string{
	"descricao", "categoria", "numero_patrimonio", "numero_serie",
	"marca", "modelo", "estado_conservacao", "quantidade",
	"estoque_minimo", "setor", "responsavel_matricula",
	"data_aquisicao", "valor",
}

// colunasObrigatorias são os cabeçalhos mínimos que o arquivo deve conter.
var colunasObrigatorias = []string{"descricao", "categoria"}

// ImportarCSV importa itens em massa a partir de um CSV. Resolve categoria e
// setor por NOME e responsável por MATRÍCULA (mais amigável que IDs). Valida
// linha a linha e devolve um relatório detalhado de erros.
//
// dryRun=true apenas simula (valida sem gravar) — usado para pré-visualização.
func (s *ItemService) ImportarCSV(r io.Reader, dryRun bool) (*ResultadoImportacao, error) {
	leitor := csv.NewReader(r)
	leitor.FieldsPerRecord = -1 // tolera linhas com nº de campos variável
	leitor.TrimLeadingSpace = true

	registros, err := leitor.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: arquivo CSV ilegível (%v)", ErrRegraNegocio, err)
	}
	if len(registros) < 2 {
		return nil, fmt.Errorf("%w: arquivo vazio ou sem linhas de dados", ErrRegraNegocio)
	}

	// Mapeia cabeçalho -> índice da coluna.
	idx := map[string]int{}
	for i, h := range registros[0] {
		idx[normalizarChave(h)] = i
	}
	for _, obrig := range colunasObrigatorias {
		if _, ok := idx[obrig]; !ok {
			return nil, fmt.Errorf("%w: coluna obrigatória ausente no cabeçalho: %q", ErrRegraNegocio, obrig)
		}
	}

	// Índices de lookup (nome/matrícula -> id), carregados uma vez.
	catPorNome, setorPorNome, servPorMatricula, err := s.indicesImportacao()
	if err != nil {
		return nil, err
	}

	res := &ResultadoImportacao{Validacao: dryRun}

	for n := 1; n < len(registros); n++ {
		linha := registros[n]
		numArquivo := n + 1 // 1-based; considerando o cabeçalho como linha 1

		if linhaVazia(linha) {
			continue
		}
		res.Total++

		campo := func(nome string) string {
			i, ok := idx[nome]
			if !ok || i >= len(linha) {
				return ""
			}
			return strings.TrimSpace(linha[i])
		}

		descricao := campo("descricao")
		entrada, msg := s.montarEntradaCSV(campo, catPorNome, setorPorNome, servPorMatricula)
		if msg != "" {
			res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: descricao, Mensagem: msg})
			continue
		}

		if dryRun {
			if _, err := s.validar(*entrada); err != nil {
				res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: descricao, Mensagem: mensagemDeErro(err)})
				continue
			}
			res.Validas++
			continue
		}

		if _, err := s.Criar(*entrada); err != nil {
			res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: descricao, Mensagem: mensagemDeErro(err)})
			continue
		}
		res.Validas++
		res.Importados++
	}

	return res, nil
}

// montarEntradaCSV converte uma linha em EntradaItem, resolvendo referências.
// Devolve mensagem não-vazia quando a linha é inválida no nível de parsing.
func (s *ItemService) montarEntradaCSV(
	campo func(string) string,
	catPorNome, setorPorNome map[string]uint,
	servPorMatricula map[string]uint,
) (*EntradaItem, string) {
	descricao := campo("descricao")
	if descricao == "" {
		return nil, "Descrição é obrigatória."
	}

	nomeCat := campo("categoria")
	if nomeCat == "" {
		return nil, "Categoria é obrigatória."
	}
	catID, ok := catPorNome[normalizarChave(nomeCat)]
	if !ok {
		return nil, fmt.Sprintf("Categoria %q não cadastrada.", nomeCat)
	}

	in := &EntradaItem{
		Descricao:        descricao,
		CategoriaID:      catID,
		NumeroPatrimonio: ptrOuNil(campo("numero_patrimonio")),
		NumeroSerie:      ptrOuNil(campo("numero_serie")),
		Marca:            campo("marca"),
		Modelo:           campo("modelo"),
	}

	if e := campo("estado_conservacao"); e != "" {
		in.EstadoConservacao = models.EstadoConservacao(normalizarChave(e))
	}

	if q := campo("quantidade"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil {
			return nil, fmt.Sprintf("Quantidade inválida: %q.", q)
		}
		in.Quantidade = v
	}
	if em := campo("estoque_minimo"); em != "" {
		v, err := strconv.Atoi(em)
		if err != nil {
			return nil, fmt.Sprintf("Estoque mínimo inválido: %q.", em)
		}
		in.EstoqueMinimo = v
	}

	if nomeSetor := campo("setor"); nomeSetor != "" {
		id, ok := setorPorNome[normalizarChave(nomeSetor)]
		if !ok {
			return nil, fmt.Sprintf("Setor %q não cadastrado.", nomeSetor)
		}
		in.SetorID = &id
	}

	if mat := campo("responsavel_matricula"); mat != "" {
		id, ok := servPorMatricula[strings.ToLower(mat)]
		if !ok {
			return nil, fmt.Sprintf("Responsável (matrícula %q) não cadastrado.", mat)
		}
		in.ResponsavelID = &id
	}

	if d := campo("data_aquisicao"); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			return nil, fmt.Sprintf("Data de aquisição inválida: %q (use AAAA-MM-DD).", d)
		}
		in.DataAquisicao = &t
	}

	if v := campo("valor"); v != "" {
		// Aceita vírgula ou ponto como separador decimal.
		f, err := strconv.ParseFloat(strings.Replace(v, ",", ".", 1), 64)
		if err != nil {
			return nil, fmt.Sprintf("Valor inválido: %q.", v)
		}
		in.Valor = &f
	}

	return in, ""
}

// indicesImportacao carrega os mapas de lookup usados na importação.
func (s *ItemService) indicesImportacao() (cat, setor, serv map[string]uint, err error) {
	categorias, err := s.categoriaRepo.Listar()
	if err != nil {
		return nil, nil, nil, err
	}
	setores, err := s.setorRepo.Listar()
	if err != nil {
		return nil, nil, nil, err
	}
	servidores, err := s.servidorRepo.Listar()
	if err != nil {
		return nil, nil, nil, err
	}

	cat = make(map[string]uint, len(categorias))
	for _, c := range categorias {
		cat[normalizarChave(c.Nome)] = c.ID
	}
	setor = make(map[string]uint, len(setores))
	for _, st := range setores {
		setor[normalizarChave(st.Nome)] = st.ID
	}
	serv = make(map[string]uint, len(servidores))
	for _, sv := range servidores {
		serv[strings.ToLower(sv.Matricula)] = sv.ID
	}
	return cat, setor, serv, nil
}

// mensagemDeErro extrai uma mensagem legível de um erro de domínio.
func mensagemDeErro(err error) string {
	var ev *ErroValidacao
	if errors.As(err, &ev) {
		partes := make([]string, 0, len(ev.Campos))
		for _, m := range ev.Campos {
			partes = append(partes, m)
		}
		if len(partes) > 0 {
			return strings.Join(partes, " ")
		}
	}
	return err.Error()
}

func linhaVazia(linha []string) bool {
	for _, c := range linha {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func ptrOuNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// normalizarChave deixa uma string minúscula, sem acentos e sem espaços nas
// bordas — para casar cabeçalhos e nomes de forma tolerante.
func normalizarChave(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return removerAcentos(s)
}

// removerAcentos troca os acentos mais comuns do português por letras simples.
func removerAcentos(s string) string {
	repl := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)
	return repl.Replace(s)
}
