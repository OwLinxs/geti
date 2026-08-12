package services

import (
	"strings"

	"github.com/pmfb/sige-ti/internal/models"
	"github.com/pmfb/sige-ti/internal/repositories"
)

// ConhecimentoService gerencia a base de conhecimento (KB) do setor de T.I.
type ConhecimentoService struct {
	repo repositories.ConhecimentoRepository
}

func NewConhecimentoService(repo repositories.ConhecimentoRepository) *ConhecimentoService {
	return &ConhecimentoService{repo: repo}
}

// EntradaArtigo reúne os dados de criação/edição de um artigo.
type EntradaArtigo struct {
	Titulo    string
	Categoria string
	Conteudo  string
	Publicado bool
	AutorID   uint // definido na criação (usuário autenticado)
}

func (s *ConhecimentoService) validar(in EntradaArtigo) error {
	ev := NovoErroValidacao()
	if strings.TrimSpace(in.Titulo) == "" {
		ev.Add("titulo", "Informe o título do artigo.")
	}
	if strings.TrimSpace(in.Conteudo) == "" {
		ev.Add("conteudo", "Informe o conteúdo do artigo.")
	}
	if ev.TemErros() {
		return ev
	}
	return nil
}

func (s *ConhecimentoService) Criar(in EntradaArtigo) (*models.ArtigoConhecimento, error) {
	if err := s.validar(in); err != nil {
		return nil, err
	}
	artigo := &models.ArtigoConhecimento{
		Titulo:    strings.TrimSpace(in.Titulo),
		Categoria: strings.TrimSpace(in.Categoria),
		Conteudo:  strings.TrimSpace(in.Conteudo),
		Publicado: in.Publicado,
	}
	if in.AutorID != 0 {
		id := in.AutorID
		artigo.AutorID = &id
	}
	if err := s.repo.Criar(artigo); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(artigo.ID)
}

func (s *ConhecimentoService) Atualizar(id uint, in EntradaArtigo) (*models.ArtigoConhecimento, error) {
	artigo, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	if err := s.validar(in); err != nil {
		return nil, err
	}
	artigo.Titulo = strings.TrimSpace(in.Titulo)
	artigo.Categoria = strings.TrimSpace(in.Categoria)
	artigo.Conteudo = strings.TrimSpace(in.Conteudo)
	artigo.Publicado = in.Publicado
	// Não sobrescreve o autor original; zera a associação antes do Save para
	// não tentar upsert no usuário.
	artigo.Autor = nil
	if err := s.repo.Atualizar(artigo); err != nil {
		return nil, err
	}
	return s.repo.BuscarPorID(id)
}

func (s *ConhecimentoService) Listar(f repositories.FiltroConhecimento) ([]models.ArtigoConhecimento, int64, error) {
	return s.repo.Listar(f)
}

// Visualizar retorna o artigo e conta uma visualização (uso na leitura).
func (s *ConhecimentoService) Visualizar(id uint) (*models.ArtigoConhecimento, error) {
	artigo, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, traduzErroRepo(err)
	}
	// Contagem é best-effort: falha não impede a leitura.
	_ = s.repo.IncrementarVisualizacoes(id)
	artigo.Visualizacoes++
	return artigo, nil
}

func (s *ConhecimentoService) Excluir(id uint) error {
	if _, err := s.repo.BuscarPorID(id); err != nil {
		return traduzErroRepo(err)
	}
	return s.repo.Remover(id)
}
