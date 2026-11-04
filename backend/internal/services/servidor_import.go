package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CabecalhoCSVServidores é o cabeçalho esperado na importação de servidores.
var CabecalhoCSVServidores = []string{"nome", "matricula", "setor", "ativo"}

// colunasObrigatoriasServidor são os cabeçalhos mínimos exigidos.
var colunasObrigatoriasServidor = []string{"nome", "matricula"}

// ImportarCSV importa servidores em massa a partir de um CSV. Resolve o setor
// por NOME ou SIGLA (mais amigável que ID). Valida linha a linha e devolve um
// relatório detalhado de erros. dryRun=true apenas simula (não grava).
func (s *ServidorService) ImportarCSV(r io.Reader, dryRun bool) (*ResultadoImportacao, error) {
	leitor := csv.NewReader(r)
	leitor.FieldsPerRecord = -1 // tolera nº de campos variável
	leitor.TrimLeadingSpace = true

	registros, err := leitor.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: arquivo CSV ilegível (%v)", ErrRegraNegocio, err)
	}
	if len(registros) < 2 {
		return nil, fmt.Errorf("%w: arquivo vazio ou sem linhas de dados", ErrRegraNegocio)
	}

	idx := map[string]int{}
	for i, h := range registros[0] {
		idx[normalizarChave(h)] = i
	}
	for _, obrig := range colunasObrigatoriasServidor {
		if _, ok := idx[obrig]; !ok {
			return nil, fmt.Errorf("%w: coluna obrigatória ausente no cabeçalho: %q", ErrRegraNegocio, obrig)
		}
	}

	setorPorChave, err := s.indiceSetores()
	if err != nil {
		return nil, err
	}

	res := &ResultadoImportacao{Validacao: dryRun}

	for n := 1; n < len(registros); n++ {
		linha := registros[n]
		numArquivo := n + 1 // 1-based, cabeçalho = linha 1

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

		nome := campo("nome")
		entrada, msg := montarEntradaServidorCSV(campo, setorPorChave)
		if msg != "" {
			res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: nome, Mensagem: msg})
			continue
		}

		if dryRun {
			if err := s.validar(*entrada, nil); err != nil {
				res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: nome, Mensagem: mensagemDeErro(err)})
				continue
			}
			res.Validas++
			continue
		}

		if _, err := s.Criar(*entrada); err != nil {
			res.Erros = append(res.Erros, ErroLinha{Linha: numArquivo, Descricao: nome, Mensagem: mensagemDeErro(err)})
			continue
		}
		res.Validas++
		res.Importados++
	}

	return res, nil
}

// montarEntradaServidorCSV converte uma linha em EntradaServidor.
func montarEntradaServidorCSV(campo func(string) string, setorPorChave map[string]uint) (*EntradaServidor, string) {
	nome := campo("nome")
	if nome == "" {
		return nil, "Nome é obrigatório."
	}
	matricula := campo("matricula")
	if matricula == "" {
		return nil, "Matrícula é obrigatória."
	}

	in := &EntradaServidor{Nome: nome, Matricula: matricula, Ativo: true}

	if nomeSetor := campo("setor"); nomeSetor != "" {
		id, ok := setorPorChave[normalizarChave(nomeSetor)]
		if !ok {
			return nil, fmt.Sprintf("Setor %q não cadastrado (use o nome ou a sigla).", nomeSetor)
		}
		in.SetorID = &id
	}

	if a := campo("ativo"); a != "" {
		in.Ativo = parseBoolPT(a)
	}

	return in, ""
}

// indiceSetores carrega o mapa de lookup nome/sigla -> id dos setores.
func (s *ServidorService) indiceSetores() (map[string]uint, error) {
	setores, err := s.setorRepo.Listar()
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint, len(setores)*2)
	for _, st := range setores {
		m[normalizarChave(st.Nome)] = st.ID
		if st.Sigla != "" {
			m[normalizarChave(st.Sigla)] = st.ID
		}
	}
	return m, nil
}

// parseBoolPT interpreta valores comuns em português para booleano. Vazio e
// qualquer valor não reconhecido como negativo são considerados verdadeiros.
func parseBoolPT(s string) bool {
	switch normalizarChave(s) {
	case "nao", "false", "0", "inativo", "n":
		return false
	default:
		return true
	}
}
